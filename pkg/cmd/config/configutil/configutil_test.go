package configutil

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/pkg/api"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResourceDiffAndDiffResultHelpers(t *testing.T) {
	diff := ResourceDiff{
		Create:    []ResourceItem{{Key: "create-1"}},
		Update:    []ResourceItem{{Key: "update-1"}},
		Delete:    []ResourceItem{{Key: "delete-1"}},
		Unchanged: []string{"same-1"},
	}

	assert.Equal(t, 1, diff.CreateCount())
	assert.Equal(t, 1, diff.UpdateCount())
	assert.Equal(t, 1, diff.DeleteCount())
	assert.True(t, diff.HasDifferences())

	var nilResult *DiffResult
	assert.False(t, nilResult.HasDifferences())
	assert.Nil(t, nilResult.Sections())

	result := &DiffResult{
		Upstreams: diff,
	}
	assert.True(t, result.HasDifferences())

	sections := result.Sections()
	require.Len(t, sections, 12)
	assert.Equal(t, "upstreams", sections[0].Name)
	assert.Equal(t, "stream_routes", sections[len(sections)-1].Name)
}

func TestReadConfigFile_JSONAndYAML(t *testing.T) {
	tmpDir := t.TempDir()

	jsonPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"version":"1","routes":[{"id":"route-json","uri":"/json"}]}`), 0o644))

	cfg, err := ReadConfigFile(jsonPath)
	require.NoError(t, err)
	require.Len(t, cfg.Routes, 1)
	require.NotNil(t, cfg.Routes[0].ID)
	assert.Equal(t, "route-json", *cfg.Routes[0].ID)

	yamlPath := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("version: \"1\"\nroutes:\n  - id: route-yaml\n    uri: /yaml\n"), 0o644))

	cfg, err = ReadConfigFile(yamlPath)
	require.NoError(t, err)
	require.Len(t, cfg.Routes, 1)
	require.NotNil(t, cfg.Routes[0].ID)
	assert.Equal(t, "route-yaml", *cfg.Routes[0].ID)
}

func TestReadConfigFile_ParseError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "broken.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"version":`), 0o644))

	_, err := ReadConfigFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON file")
}

func TestComputeDiffAndFormatSummary(t *testing.T) {
	local := api.ConfigFile{
		Routes: []api.Route{
			{ID: stringPtr("route-create"), URI: stringPtr("/create")},
			{ID: stringPtr("route-update"), URI: stringPtr("/new")},
			{ID: stringPtr("route-same"), URI: stringPtr("/same"), CreateTime: intPtr(11), UpdateTime: intPtr(12)},
		},
		Consumers: []api.Consumer{
			{Username: stringPtr("consumer-create")},
		},
	}
	remote := api.ConfigFile{
		Routes: []api.Route{
			{ID: stringPtr("route-update"), URI: stringPtr("/old")},
			{ID: stringPtr("route-delete"), URI: stringPtr("/delete")},
			{ID: stringPtr("route-same"), URI: stringPtr("/same"), CreateTime: intPtr(101), UpdateTime: intPtr(102)},
		},
		Consumers: []api.Consumer{
			{Username: stringPtr("consumer-delete")},
		},
	}

	diff, err := ComputeDiff(local, remote)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"route-create"}, keysOf(diff.Routes.Create))
	assert.ElementsMatch(t, []string{"route-update"}, keysOf(diff.Routes.Update))
	assert.ElementsMatch(t, []string{"route-delete"}, keysOf(diff.Routes.Delete))
	assert.Equal(t, []string{"route-same"}, diff.Routes.Unchanged)
	assert.ElementsMatch(t, []string{"consumer-create"}, keysOf(diff.Consumers.Create))
	assert.ElementsMatch(t, []string{"consumer-delete"}, keysOf(diff.Consumers.Delete))

	summary := FormatDiffSummary(diff)
	assert.Contains(t, summary, "Differences found:")
	assert.Contains(t, summary, "routes: create=1 update=1 delete=1 unchanged=1")
	assert.Contains(t, summary, "CREATE route-create")
	assert.Contains(t, summary, "UPDATE route-update")
	assert.Contains(t, summary, "DELETE route-delete")

	assert.Equal(t, "No differences found.\n", FormatDiffSummary(&DiffResult{}))
}

func TestComputeDiff_MissingKey(t *testing.T) {
	_, err := ComputeDiff(
		api.ConfigFile{Routes: []api.Route{{URI: stringPtr("/missing-id")}}},
		api.ConfigFile{},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `routes[0]: missing "id" field`)
}

func TestExtractKey(t *testing.T) {
	key, err := extractKey(map[string]interface{}{"id": " route-1 "}, "id")
	require.NoError(t, err)
	assert.Equal(t, "route-1", key)

	_, err = extractKey(map[string]interface{}{}, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing "id" field`)

	_, err = extractKey(map[string]interface{}{"id": "   "}, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `empty "id" field`)
}

func TestNormalizeMapAndStripTimestamps(t *testing.T) {
	normalized := normalizeMap(map[string]interface{}{
		"id":          "route-1",
		"create_time": 1,
		"nested": map[string]interface{}{
			"update_time": 2,
			"keep":        "value",
		},
		"list": []interface{}{
			map[string]interface{}{"create_time": 3, "name": "node-1"},
		},
	})

	assert.NotContains(t, normalized, "create_time")

	nested, ok := normalized["nested"].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, nested, "update_time")
	assert.Equal(t, "value", nested["keep"])

	list, ok := normalized["list"].([]interface{})
	require.True(t, ok)
	entry, ok := list[0].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, entry, "create_time")
	assert.Equal(t, "node-1", entry["name"])
}

func TestToMapSlice(t *testing.T) {
	out, err := toMapSlice([]api.Route{{ID: stringPtr("route-1"), URI: stringPtr("/one")}})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "route-1", out[0]["id"])

	out, err = toMapSlice([]api.Route(nil))
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestFetchPaginated(t *testing.T) {
	client := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/apisix/admin/routes", req.URL.Path)
			switch req.URL.Query().Get("page") {
			case "1":
				assert.Equal(t, "500", req.URL.Query().Get("page_size"))
				return jsonResponse(`{"total":2,"list":[{"key":"/apisix/routes/r1","value":{"id":"r1","uri":"/one"}}]}`), nil
			case "2":
				return jsonResponse(`{"total":2,"list":[{"key":"/apisix/routes/r2","value":{"id":"r2","uri":"/two"}}]}`), nil
			default:
				return jsonResponse(`{"total":2,"list":[]}`), nil
			}
		}),
	}, "http://example.test")

	items, err := fetchPaginated[api.Route](client, "/apisix/admin/routes")
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.NotNil(t, items[0].Value.ID)
	require.NotNil(t, items[1].Value.ID)
	assert.Equal(t, "r1", *items[0].Value.ID)
	assert.Equal(t, "r2", *items[1].Value.ID)
}

func TestFetchPaginated_OptionalResourceAndParseError(t *testing.T) {
	optionalClient := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}, "http://example.test")

	items, err := fetchPaginated[api.Route](optionalClient, "/apisix/admin/stream_routes")
	require.NoError(t, err)
	assert.Nil(t, items)

	parseErrClient := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(`{"total":1,"list":`), nil
		}),
	}, "http://example.test")

	_, err = fetchPaginated[api.Route](parseErrClient, "/apisix/admin/routes")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestFetchPluginMetadata(t *testing.T) {
	client := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/apisix/admin/plugins/list":
				return jsonResponse(`["limit-count","prometheus"]`), nil
			case "/apisix/admin/plugin_metadata/limit-count":
				return jsonResponse(`{"value":{"disable":false,"create_time":1}}`), nil
			case "/apisix/admin/plugin_metadata/prometheus":
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}, "http://example.test")

	entries, err := fetchPluginMetadata(client)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "limit-count", entries[0]["plugin_name"])
	assert.Equal(t, false, entries[0]["disable"])
}

func TestFetchPluginMetadata_OptionalAndParseError(t *testing.T) {
	optionalClient := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}, "http://example.test")

	entries, err := fetchPluginMetadata(optionalClient)
	require.NoError(t, err)
	assert.Nil(t, entries)

	parseErrClient := api.NewClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/apisix/admin/plugins/list" {
				return jsonResponse(`["broken"]`), nil
			}
			return jsonResponse(`{"value":`), nil
		}),
	}, "http://example.test")

	_, err = fetchPluginMetadata(parseErrClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestExtractSecretID(t *testing.T) {
	assert.Equal(t, "vault/my-secret", extractSecretID("/apisix/secrets/vault/my-secret"))
	assert.Equal(t, "", extractSecretID("invalid"))
}

func intPtr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func keysOf(items []ResourceItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Key)
	}
	return out
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
