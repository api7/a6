package create

import (
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

var createdSecretBody = `{
	"key": "/apisix/secrets/vault/my-vault-1",
	"value": {
		"id": "my-vault-1",
		"uri": "http://127.0.0.1:8200",
		"prefix": "/apisix/kv",
		"token": "test-token-12345"
	}
}`

func TestSecretCreate_Success(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/secrets/vault/my-vault-1", httpmock.JSONResponse(createdSecretBody))

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "secret.json")
	err := os.WriteFile(filePath, []byte(`{"uri":"http://127.0.0.1:8200","prefix":"/apisix/kv","token":"test-token-12345"}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"vault/my-vault-1", "-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	reg.Verify(t)
}

func TestStripConflictingID(t *testing.T) {
	// File declares `id: my-vault` but positional is `vault/my-vault`; mismatched
	// id must be stripped so APISIX doesn't reject with "wrong secret id".
	p := stripConflictingID(map[string]interface{}{
		"id":    "my-vault",
		"token": "t",
	}, "vault/my-vault")
	_, has := p["id"]
	assert.False(t, has, "mismatched body id should be stripped")
	assert.Equal(t, "t", p["token"])

	// Matching id is preserved.
	p = stripConflictingID(map[string]interface{}{
		"id":    "vault/my-vault",
		"token": "t",
	}, "vault/my-vault")
	assert.Equal(t, "vault/my-vault", p["id"])

	// No id key, unchanged.
	p = stripConflictingID(map[string]interface{}{"token": "t"}, "vault/my-vault")
	_, has = p["id"]
	assert.False(t, has)

	// Nil payload safe.
	assert.Nil(t, stripConflictingID(nil, "vault/x"))

	// Non-string id (e.g. YAML `id: 1` parsed as int) must also be stripped —
	// APISIX rejects it with "wrong secret id" otherwise.
	p = stripConflictingID(map[string]interface{}{
		"id":    1,
		"token": "t",
	}, "vault/my-vault")
	_, has = p["id"]
	assert.False(t, has, "non-string body id should be stripped")
	assert.Equal(t, "t", p["token"])
}

func TestSecretCreate_MissingFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"vault/my-vault-1"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}
