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

func TestGlobalRuleDelete_Success(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodDelete, "/apisix/admin/global_rules/1", httpmock.JSONResponse(`{"key":"/apisix/global_rules/1","deleted":"1"}`))

	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"1", "--force"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Global rule 1 deleted.")
	reg.Verify(t)
}

func TestGlobalRuleDelete_NotFound(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodDelete, "/apisix/admin/global_rules/999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"999", "--force"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	reg.Verify(t)
}

func TestGlobalruleDelete_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := deleteRun(&Options{
		IO: ios,
	})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}

func TestGlobalruleDelete_AllAndLabelMutuallyExclusive(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"--all", "--label", "env=test"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all and --label are mutually exclusive")
}

func TestGlobalruleDelete_AllWithID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"1", "--all"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all cannot be used with a specific ID")
}

func TestGlobalruleDelete_LabelWithID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"1", "--label", "env=test"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--label cannot be used with a specific ID")
}
