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
			"key": "/apisix/consumers/jack",
			"value": {
				"username": "jack",
				"desc": "test consumer",
				"group_id": "group-1",
				"plugins": {
					"key-auth": {"key": "jack-key"}
				},
				"create_time": 1704067200
			}
		},
		{
			"key": "/apisix/consumers/jill",
			"value": {
				"username": "jill",
				"desc": "another consumer",
				"create_time": 1704153600
			}
		}
	]
}`

func TestConsumerList_Table(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers", httpmock.JSONResponse(listBody))

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
	assert.Contains(t, output, "USERNAME")
	assert.Contains(t, output, "DESC")
	assert.Contains(t, output, "GROUP_ID")
	assert.Contains(t, output, "PLUGINS")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "jack")
	assert.Contains(t, output, "test consumer")
	assert.Contains(t, output, "jill")
	assert.Contains(t, output, "another consumer")
	reg.Verify(t)
}

func TestConsumerList_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers", httpmock.JSONResponse(listBody))

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
	var result []interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	reg.Verify(t)
}

func TestConsumerList_YAML(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers", httpmock.JSONResponse(listBody))

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
	assert.Contains(t, output, "username: jack")
	assert.Contains(t, output, "desc: test consumer")
	reg.Verify(t)
}

func TestConsumerList_Empty(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers", httpmock.JSONResponse(`{"total":0,"list":[]}`))

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
	assert.Contains(t, stdout.String(), "No consumers found.")
	reg.Verify(t)
}

func TestConsumerList_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers", httpmock.StringResponse(403, `{"error_msg":"forbidden"}`))

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
