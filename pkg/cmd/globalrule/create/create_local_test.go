package create

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/iostreams"
)

type localMockConfig struct {
	baseURL string
}

func (m *localMockConfig) BaseURL() string                                 { return m.baseURL }
func (m *localMockConfig) APIKey() string                                  { return "" }
func (m *localMockConfig) CurrentContext() string                          { return "test" }
func (m *localMockConfig) Contexts() []config.Context                      { return nil }
func (m *localMockConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (m *localMockConfig) AddContext(ctx config.Context) error             { return nil }
func (m *localMockConfig) RemoveContext(name string) error                 { return nil }
func (m *localMockConfig) SetCurrentContext(name string) error             { return nil }
func (m *localMockConfig) Save() error                                     { return nil }

func TestGlobalRuleCreate_MissingID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "global-rule.json")
	err := os.WriteFile(filePath, []byte(`{"plugins":{"prometheus":{}}}`), 0o644)
	require.NoError(t, err)

	err = createRun(&Options{
		IO:     ios,
		File:   filePath,
		Client: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &localMockConfig{baseURL: "http://localhost:9180"}, nil
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), `must include an "id" field`)
}
