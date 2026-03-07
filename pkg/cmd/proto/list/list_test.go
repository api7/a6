package list

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
			"key": "/apisix/protos/1",
			"value": {
				"id": "1",
				"name": "helloworld",
				"desc": "Hello world proto",
				"content": "syntax = \"proto3\";\npackage helloworld;\nservice Greeter {\n    rpc SayHello (HelloRequest) returns (HelloReply) {}\n}\nmessage HelloRequest { string name = 1; }\nmessage HelloReply { string message = 1; }",
				"create_time": 1704067200
			}
		},
		{
			"key": "/apisix/protos/2",
			"value": {
				"id": "2",
				"name": "echo",
				"desc": "Echo service proto",
				"content": "syntax = \"proto3\";\npackage echo;\nservice Echo {\n    rpc Echo (EchoRequest) returns (EchoResponse) {}\n}\nmessage EchoRequest { string msg = 1; }\nmessage EchoResponse { string msg = 1; }",
				"create_time": 1704153600
			}
		}
	]
}`

func TestProtoList_Table(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/protos", httpmock.JSONResponse(listBody))

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
	expectedCreated := time.Unix(1704067200, 0).Format("2006-01-02 15:04:05")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DESC")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "helloworld")
	assert.Contains(t, output, "Hello world proto")
	assert.Contains(t, output, expectedCreated)
	reg.Verify(t)
}

func TestProtoList_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/protos", httpmock.JSONResponse(listBody))

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

func TestProtoList_Empty(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/protos", httpmock.JSONResponse(`{"total":0,"list":[]}`))

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
	assert.Contains(t, stdout.String(), "No proto definitions found.")
	reg.Verify(t)
}

func TestProtoList_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/protos", httpmock.StringResponse(403, `{"error_msg":"forbidden"}`))

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
