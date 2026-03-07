package health

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
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

var healthBody = `{
	"nodes": [
		{
			"ip": "52.86.68.46",
			"port": 80,
			"status": "healthy",
			"counter": {
				"success": 2,
				"http_failure": 0,
				"tcp_failure": 0,
				"timeout_failure": 0
			}
		}
	],
	"type": "http",
	"name": "/apisix/upstreams/1"
}`

func TestUpstreamHealth_Table(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/v1/healthcheck/upstreams/1", httpmock.JSONResponse(healthBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	err := healthRun(&Options{
		IO:            ios,
		Config:        func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		ControlClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		ID:            "1",
	})

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "NODE")
	assert.Contains(t, output, "52.86.68.46:80")
	assert.Contains(t, output, "healthy")
	assert.Contains(t, output, "Type: http")
	reg.Verify(t)
}

func TestUpstreamHealth_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/v1/healthcheck/upstreams/1", httpmock.JSONResponse(healthBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	err := healthRun(&Options{
		IO:            ios,
		Config:        func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		ControlClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		ID:            "1",
	})

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "http", result["type"])
	nodes, ok := result["nodes"].([]interface{})
	require.True(t, ok)
	require.Len(t, nodes, 1)
	node := nodes[0].(map[string]interface{})
	assert.Equal(t, "52.86.68.46", node["ip"])
	assert.Equal(t, "healthy", node["status"])
	reg.Verify(t)
}

func TestUpstreamHealth_NotFound(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/v1/healthcheck/upstreams/999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()

	err := healthRun(&Options{
		IO:            ios,
		Config:        func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		ControlClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		ID:            "999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "No health check data available for upstream 999")
	assert.Contains(t, err.Error(), "health checks configured")
	reg.Verify(t)
}

func TestUpstreamHealth_CustomControlURL(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/v1/healthcheck/upstreams/1", httpmock.JSONResponse(healthBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)
	err := healthRun(&Options{
		IO:            ios,
		Config:        func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		ControlClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		ID:            "1",
		Output:        "json",
		ControlURL:    "http://custom-control:19090",
	})

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "http", result["type"])
	assert.Equal(t, 1, reg.CallCount(http.MethodGet, "/v1/healthcheck/upstreams/1"))
	reg.Verify(t)
}

func TestUpstreamHealth_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := healthRun(&Options{
		IO: ios,
	})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}
