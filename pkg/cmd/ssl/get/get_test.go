package get

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

var getBody = `{
	"key": "/apisix/ssls/1",
	"value": {
		"id": "1",
		"snis": ["test.example.com"],
		"status": 1,
		"type": "server",
		"cert": "-----BEGIN CERTIFICATE-----\nfake-cert-data\n-----END CERTIFICATE-----",
		"key": "-----BEGIN RSA PRIVATE KEY-----\nfake-key-data\n-----END RSA PRIVATE KEY-----"
	}
}`

func TestSSLGet_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls/1", httpmock.JSONResponse(getBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"1"})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "1", result["id"])
	assert.Contains(t, result["cert"], "CERTIFICATE")
	snis, ok := result["snis"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, snis, "test.example.com")
	reg.Verify(t)
}

func TestSSLGet_YAML(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls/1", httpmock.JSONResponse(getBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"1"})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "test.example.com")
	assert.Contains(t, output, "type: server")
	reg.Verify(t)
}

func TestSSLGet_NotFound(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls/999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"999"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	reg.Verify(t)
}

func TestSslGet_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := getRun(&Options{
		IO: ios,
	})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}
