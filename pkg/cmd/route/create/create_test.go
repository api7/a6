package create

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

var createdRouteBody = `{
	"key": "/apisix/routes/1",
	"value": {
		"id": "1",
		"name": "test-route",
		"uri": "/api/v1",
		"methods": ["GET"],
		"status": 1
	}
}`

func TestRouteCreate_FromJSONFile(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPost, "/apisix/admin/routes", httpmock.JSONResponse(createdRouteBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.json")
	err := os.WriteFile(filePath, []byte(`{"name":"test-route","uri":"/api/v1","methods":["GET"]}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test-route", result["name"])
	reg.Verify(t)
}

func TestRouteCreate_FromYAMLFile(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPost, "/apisix/admin/routes", httpmock.JSONResponse(createdRouteBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.yaml")
	yamlContent := "name: test-route\nuri: /api/v1\nmethods:\n  - GET\n"
	err := os.WriteFile(filePath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test-route", result["name"])
	reg.Verify(t)
}

func TestRouteCreate_WithID(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/routes/my-route", httpmock.JSONResponse(createdRouteBody))

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.json")
	err := os.WriteFile(filePath, []byte(`{"id":"my-route","name":"test-route","uri":"/api/v1"}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	reg.Verify(t)
}

func TestRouteCreate_MissingFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}

func TestRouteCreate_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPost, "/apisix/admin/routes", httpmock.StringResponse(400, `{"error_msg":"invalid route"}`))

	ios, _, _, _ := iostreams.Test()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.json")
	err := os.WriteFile(filePath, []byte(`{"uri":"/test"}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", filePath})
	err = c.Execute()

	require.Error(t, err)
	reg.Verify(t)
}
