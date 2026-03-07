//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPluginEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test-plugin",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err)
	t.Cleanup(func() {
		runA6WithEnv(env, "context", "delete", "test-plugin", "--force")
	})
	return env
}

func TestPlugin_List(t *testing.T) {
	env := setupPluginEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "plugin", "list", "--output", "json")
	require.NoError(t, err, "plugin list failed: stdout=%s stderr=%s", stdout, stderr)
	require.True(t, json.Valid([]byte(stdout)), "plugin list output should be valid JSON")

	var plugins []string
	require.NoError(t, json.Unmarshal([]byte(stdout), &plugins))
	assert.NotEmpty(t, plugins)
	assert.Contains(t, plugins, "limit-count")
	assert.Contains(t, plugins, "key-auth")
}

func TestPlugin_Get(t *testing.T) {
	env := setupPluginEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "plugin", "get", "key-auth", "--output", "json")
	require.NoError(t, err, "plugin get failed: stdout=%s stderr=%s", stdout, stderr)
	require.True(t, json.Valid([]byte(stdout)), "plugin get output should be valid JSON")

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &schema))
	_, ok := schema["type"]
	assert.True(t, ok)
	_, ok = schema["properties"]
	assert.True(t, ok)
}

func TestPlugin_GetNonExistent(t *testing.T) {
	env := setupPluginEnv(t)

	_, stderr, err := runA6WithEnv(env, "plugin", "get", "nonexistent-plugin-xyz")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"))
}

func TestPlugin_ListStream(t *testing.T) {
	env := setupPluginEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "plugin", "list", "--subsystem", "stream", "--output", "json")
	require.NoError(t, err, "plugin list stream failed: stdout=%s stderr=%s", stdout, stderr)
	require.True(t, json.Valid([]byte(stdout)), "plugin list stream output should be valid JSON")

	var plugins []string
	require.NoError(t, json.Unmarshal([]byte(stdout), &plugins))
}
