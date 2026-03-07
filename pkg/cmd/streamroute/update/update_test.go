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

var updatedStreamRouteBody = `{
	"key": "/apisix/stream_routes/1",
	"value": {
		"id": "1",
		"name": "updated-stream-route",
		"server_port": 9101,
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}
}`

func TestStreamRouteUpdate_Success(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/stream_routes/1", httpmock.JSONResponse(updatedStreamRouteBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "stream-route.json")
	err := os.WriteFile(filePath, []byte(`{"name":"updated-stream-route","server_port":9101}`), 0o644)
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
	assert.Equal(t, "updated-stream-route", result["name"])
	reg.Verify(t)
}

func TestStreamRouteUpdate_MissingFile(t *testing.T) {
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
