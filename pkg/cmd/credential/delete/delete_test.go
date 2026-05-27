package delete

import (
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

func TestCredentialDelete_WithForce(t *testing.T) {
	// APISIX actually returns `deleted` as a count ("1"), not the resource id.
	// The CLI must echo the requested id in the success message.
	reg := &httpmock.Registry{}
	reg.Register(http.MethodDelete, "/apisix/admin/consumers/jack/credentials/cred-1", httpmock.JSONResponse(`{"key":"/apisix/consumers/jack/credentials/cred-1","deleted":"1"}`))

	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"cred-1", "--consumer", "jack", "--force"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Credential cred-1 deleted for consumer jack.")
	assert.NotContains(t, stdout.String(), "Credential 1 deleted")
	reg.Verify(t)
}

func TestCredentialDelete_MissingConsumer(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"cred-1", "--force"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
