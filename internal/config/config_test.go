package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempConfigPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "config.yaml")
}

func TestFileConfig_NewFileReturnsEmpty(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	assert.Equal(t, "", cfg.CurrentContext())
	assert.Empty(t, cfg.Contexts())
	assert.Equal(t, "", cfg.BaseURL())
	assert.Equal(t, "", cfg.APIKey())
}

func TestFileConfig_AddContext(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	err := cfg.AddContext(Context{
		Name:   "local",
		Server: "http://localhost:9180",
		APIKey: "test-key",
	})
	require.NoError(t, err)

	// First context should be auto-set as current.
	assert.Equal(t, "local", cfg.CurrentContext())
	assert.Equal(t, "http://localhost:9180", cfg.BaseURL())
	assert.Equal(t, "test-key", cfg.APIKey())
	assert.Len(t, cfg.Contexts(), 1)
}

func TestFileConfig_AddContext_Duplicate(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	err := cfg.AddContext(Context{Name: "local", Server: "http://localhost:9180"})
	require.NoError(t, err)

	err = cfg.AddContext(Context{Name: "local", Server: "http://other:9180"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFileConfig_MultipleContexts(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	require.NoError(t, cfg.AddContext(Context{
		Name:   "dev",
		Server: "http://dev:9180",
		APIKey: "dev-key",
	}))
	require.NoError(t, cfg.AddContext(Context{
		Name:   "prod",
		Server: "http://prod:9180",
		APIKey: "prod-key",
	}))

	// First context is auto-current.
	assert.Equal(t, "dev", cfg.CurrentContext())
	assert.Equal(t, "http://dev:9180", cfg.BaseURL())

	// Switch to prod.
	require.NoError(t, cfg.SetCurrentContext("prod"))
	assert.Equal(t, "prod", cfg.CurrentContext())
	assert.Equal(t, "http://prod:9180", cfg.BaseURL())
	assert.Equal(t, "prod-key", cfg.APIKey())
}

func TestFileConfig_SetCurrentContext_NotFound(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	err := cfg.SetCurrentContext("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileConfig_GetContext(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	require.NoError(t, cfg.AddContext(Context{
		Name:   "local",
		Server: "http://localhost:9180",
		APIKey: "my-key",
	}))

	ctx, err := cfg.GetContext("local")
	require.NoError(t, err)
	assert.Equal(t, "local", ctx.Name)
	assert.Equal(t, "http://localhost:9180", ctx.Server)
	assert.Equal(t, "my-key", ctx.APIKey)

	_, err = cfg.GetContext("nonexistent")
	assert.Error(t, err)
}

func TestFileConfig_RemoveContext(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	require.NoError(t, cfg.AddContext(Context{Name: "a", Server: "http://a:9180"}))
	require.NoError(t, cfg.AddContext(Context{Name: "b", Server: "http://b:9180"}))
	assert.Equal(t, "a", cfg.CurrentContext())

	// Remove non-current context.
	require.NoError(t, cfg.RemoveContext("b"))
	assert.Len(t, cfg.Contexts(), 1)
	assert.Equal(t, "a", cfg.CurrentContext())

	// Remove current context — should auto-switch.
	require.NoError(t, cfg.AddContext(Context{Name: "c", Server: "http://c:9180"}))
	require.NoError(t, cfg.RemoveContext("a"))
	assert.Equal(t, "c", cfg.CurrentContext())
}

func TestFileConfig_RemoveContext_NotFound(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	err := cfg.RemoveContext("ghost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileConfig_RemoveContext_Last(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	require.NoError(t, cfg.AddContext(Context{Name: "only", Server: "http://only:9180"}))
	require.NoError(t, cfg.RemoveContext("only"))

	assert.Equal(t, "", cfg.CurrentContext())
	assert.Empty(t, cfg.Contexts())
}

func TestFileConfig_SaveAndReload(t *testing.T) {
	path := tempConfigPath(t)
	cfg := NewFileConfigWithPath(path)

	require.NoError(t, cfg.AddContext(Context{
		Name:   "saved",
		Server: "http://saved:9180",
		APIKey: "saved-key",
	}))
	require.NoError(t, cfg.Save())

	// Reload from the same path.
	cfg2 := NewFileConfigWithPath(path)
	assert.Equal(t, "saved", cfg2.CurrentContext())
	assert.Equal(t, "http://saved:9180", cfg2.BaseURL())
	assert.Equal(t, "saved-key", cfg2.APIKey())
	assert.Len(t, cfg2.Contexts(), 1)
}

func TestFileConfig_Overrides(t *testing.T) {
	cfg := NewFileConfigWithPath(tempConfigPath(t))

	require.NoError(t, cfg.AddContext(Context{
		Name:   "local",
		Server: "http://localhost:9180",
		APIKey: "file-key",
	}))

	// Override takes precedence.
	cfg.SetServerOverride("http://override:9180")
	cfg.SetAPIKeyOverride("override-key")

	assert.Equal(t, "http://override:9180", cfg.BaseURL())
	assert.Equal(t, "override-key", cfg.APIKey())
}

func TestFileConfig_SaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested")
	path := filepath.Join(nested, "config.yaml")
	cfg := NewFileConfigWithPath(path)

	require.NoError(t, cfg.AddContext(Context{Name: "test", Server: "http://test:9180"}))
	require.NoError(t, cfg.Save())

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

func TestDefaultConfigPath_A6ConfigDir(t *testing.T) {
	t.Setenv("A6_CONFIG_DIR", "/tmp/a6-test-config")
	t.Setenv("XDG_CONFIG_HOME", "")
	path := defaultConfigPath()
	assert.Equal(t, "/tmp/a6-test-config/config.yaml", path)
}

func TestDefaultConfigPath_XDGConfigHome(t *testing.T) {
	t.Setenv("A6_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	path := defaultConfigPath()
	assert.Equal(t, "/tmp/xdg-test/a6/config.yaml", path)
}
