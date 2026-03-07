package update

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

var updatedRouteBody = `{
	"key": "/apisix/routes/1",
	"value": {
		"id": "1",
		"name": "updated-route",
		"uri": "/api/v2",
		"methods": ["GET", "POST"],
		"status": 1
	}
}`

func TestRouteUpdate_Success(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/routes/1", httpmock.JSONResponse(updatedRouteBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.json")
	err := os.WriteFile(filePath, []byte(`{"name":"updated-route","uri":"/api/v2","methods":["GET","POST"]}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdUpdate(f)
	c.SetArgs([]string{"1", "-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "updated-route", result["name"])
	reg.Verify(t)
}

func TestRouteUpdate_NotFound(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/routes/999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "route.json")
	err := os.WriteFile(filePath, []byte(`{"name":"test","uri":"/test"}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdUpdate(f)
	c.SetArgs([]string{"999", "-f", filePath})
	err = c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	reg.Verify(t)
}

func TestRouteUpdate_MissingArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdUpdate(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.Error(t, err)
}

func TestRouteUpdate_MissingFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdUpdate(f)
	c.SetArgs([]string{"1"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}

func TestRouteUpdate_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := updateRun(&Options{
		IO: ios,
	})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}
