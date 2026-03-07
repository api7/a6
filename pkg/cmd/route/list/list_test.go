package list

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

var listBody = `{
	"total": 2,
	"list": [
		{
			"key": "/apisix/routes/1",
			"value": {
				"id": "1",
				"name": "test-route",
				"uri": "/api/v1",
				"methods": ["GET", "POST"],
				"status": 1,
				"upstream_id": "ups-1"
			}
		},
		{
			"key": "/apisix/routes/2",
			"value": {
				"id": "2",
				"name": "health-check",
				"uri": "/health",
				"methods": ["GET"],
				"status": 1,
				"upstream": {
					"type": "roundrobin",
					"nodes": {"127.0.0.1:8080": 1}
				}
			}
		}
	]
}`

func TestRouteList_Table(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "URI")
	assert.Contains(t, output, "METHODS")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "UPSTREAM")
	assert.Contains(t, output, "test-route")
	assert.Contains(t, output, "/api/v1")
	assert.Contains(t, output, "health-check")
	assert.Contains(t, output, "/health")
	reg.Verify(t)
}

func TestRouteList_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	// Verify it's valid JSON
	var result []interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	reg.Verify(t)
}

func TestRouteList_YAML(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "name: test-route")
	assert.Contains(t, output, "uri: /api/v1")
	reg.Verify(t)
}

func TestRouteList_Empty(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{"total":0,"list":[]}`))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No routes found.")
	reg.Verify(t)
}

func TestRouteList_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.StringResponse(403, `{"error_msg":"forbidden"}`))

	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	reg.Verify(t)
}

func TestRouteList_WithFilters(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{"total":0,"list":[]}`))

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{"--name", "my-route", "--label", "env:prod", "--uri", "/api"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, reg.CallCount(http.MethodGet, "/apisix/admin/routes"))
	reg.Verify(t)
}
