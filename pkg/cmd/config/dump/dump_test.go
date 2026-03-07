package dump

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/httpmock"
	"github.com/api7/a6/pkg/iostreams"
)

type mockConfig struct {
	baseURL string
}

func (m *mockConfig) BaseURL() string                                 { return m.baseURL }
func (m *mockConfig) APIKey() string                                  { return "" }
func (m *mockConfig) CurrentContext() string                          { return "test" }
func (m *mockConfig) Contexts() []config.Context                      { return nil }
func (m *mockConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (m *mockConfig) AddContext(ctx config.Context) error             { return nil }
func (m *mockConfig) RemoveContext(name string) error                 { return nil }
func (m *mockConfig) SetCurrentContext(name string) error             { return nil }
func (m *mockConfig) Save() error                                     { return nil }

func registerEmptyResources(reg *httpmock.Registry, skip map[string]bool) {
	resources := []string{
		"/apisix/admin/routes",
		"/apisix/admin/services",
		"/apisix/admin/upstreams",
		"/apisix/admin/consumers",
		"/apisix/admin/ssl",
		"/apisix/admin/global_rules",
		"/apisix/admin/plugin_configs",
		"/apisix/admin/consumer_groups",
		"/apisix/admin/stream_routes",
		"/apisix/admin/protos",
		"/apisix/admin/secrets",
	}
	for _, path := range resources {
		if skip[path] {
			continue
		}
		reg.Register(http.MethodGet, path, httpmock.JSONResponse(`{"total":0,"list":[]}`))
	}
	if !skip["/apisix/admin/plugins/list"] {
		reg.Register(http.MethodGet, "/apisix/admin/plugins/list", httpmock.JSONResponse(`[]`))
	}
}

func newFactory(reg *httpmock.Registry, ios *iostreams.IOStreams) *cmd.Factory {
	return &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}
}

func TestConfigDump_RoutesOnly(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 1,
		"list": [
			{
				"key": "/apisix/routes/1",
				"value": {
					"id": "1",
					"name": "hello-route",
					"uri": "/hello",
					"create_time": 1714100000,
					"update_time": 1714200000
				}
			}
		]
	}`))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	c := NewCmdDump(newFactory(reg, ios))
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	out := stdout.String()
	assert.Contains(t, out, "version: \"1\"")
	assert.Contains(t, out, "routes:")
	assert.Contains(t, out, "name: hello-route")
	assert.NotContains(t, out, "create_time")
	assert.NotContains(t, out, "update_time")
	reg.Verify(t)
}

func TestConfigDump_MultipleResources(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{
		"/apisix/admin/routes":       true,
		"/apisix/admin/services":     true,
		"/apisix/admin/secrets":      true,
		"/apisix/admin/plugins/list": true,
	})

	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/routes/1","value":{"id":"1","uri":"/hello"}}]
	}`))
	reg.Register(http.MethodGet, "/apisix/admin/services", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/services/1","value":{"id":"1","name":"svc-1","upstream_id":"1"}}]
	}`))
	reg.Register(http.MethodGet, "/apisix/admin/secrets", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/secrets/vault/my-vault","value":{"uri":"https://vault.example.com"}}]
	}`))
	reg.Register(http.MethodGet, "/apisix/admin/plugins/list", httpmock.JSONResponse(`["limit-count"]`))
	reg.Register(http.MethodGet, "/apisix/admin/plugin_metadata/limit-count", httpmock.JSONResponse(`{
		"key":"/apisix/plugin_metadata/limit-count",
		"value":{"policy":"local"}
	}`))

	ios, _, stdout, _ := iostreams.Test()

	c := NewCmdDump(newFactory(reg, ios))
	c.SetArgs([]string{"--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "1", result["version"])
	routes := result["routes"].([]interface{})
	assert.Len(t, routes, 1)
	services := result["services"].([]interface{})
	assert.Len(t, services, 1)

	secrets := result["secrets"].([]interface{})
	secret0 := secrets[0].(map[string]interface{})
	assert.Equal(t, "vault/my-vault", secret0["id"])

	metadata := result["plugin_metadata"].([]interface{})
	meta0 := metadata[0].(map[string]interface{})
	assert.Equal(t, "limit-count", meta0["plugin_name"])
	assert.Equal(t, "local", meta0["policy"])

	reg.Verify(t)
}

func TestConfigDump_EmptyAPISIX(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)

	ios, _, stdout, _ := iostreams.Test()

	c := NewCmdDump(newFactory(reg, ios))
	c.SetArgs([]string{"--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "1", result["version"])
	assert.NotContains(t, result, "routes")
	assert.NotContains(t, result, "services")
	assert.NotContains(t, result, "plugin_metadata")
	reg.Verify(t)
}

func TestConfigDump_YAMLOutput(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)

	ios, _, stdout, _ := iostreams.Test()

	c := NewCmdDump(newFactory(reg, ios))
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = yaml.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "1", result["version"])
	reg.Verify(t)
}

func TestConfigDump_FileFlag(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/routes/1","value":{"id":"1","uri":"/hello"}}]
	}`))

	ios, _, stdout, _ := iostreams.Test()
	outFile := filepath.Join(t.TempDir(), "dump.yaml")

	c := NewCmdDump(newFactory(reg, ios))
	c.SetArgs([]string{"-f", outFile})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, "", stdout.String())

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "version: \"1\"")
	assert.Contains(t, string(content), "uri: /hello")
	reg.Verify(t)
}
