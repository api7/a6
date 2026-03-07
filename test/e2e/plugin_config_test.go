//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deletePluginConfig(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/plugin_configs/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupPluginConfigEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestPluginConfig_CRUD(t *testing.T) {
	const pluginConfigID = "test-plugin-config-1"

	deletePluginConfig(t, pluginConfigID)
	t.Cleanup(func() { deletePluginConfig(t, pluginConfigID) })

	env := setupPluginConfigEnv(t)

	createJSON := `{
		"id": "test-plugin-config-1",
		"plugins": {
			"limit-count": {
				"count": 100,
				"time_window": 60,
				"rejected_code": 503,
				"key_type": "var",
				"key": "remote_addr"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "plugin-config.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "plugin-config", "create", "-f", createFile)
	require.NoError(t, err, "plugin-config create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, pluginConfigID)

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", pluginConfigID)
	require.NoError(t, err, "plugin-config get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "limit-count")

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "list")
	require.NoError(t, err, "plugin-config list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, pluginConfigID)

	updateJSON := `{
		"plugins": {
			"limit-count": {
				"count": 200,
				"time_window": 60,
				"rejected_code": 429,
				"key_type": "var",
				"key": "remote_addr"
			}
		}
	}`
	updateFile := filepath.Join(t.TempDir(), "plugin-config-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "update", pluginConfigID, "-f", updateFile)
	require.NoError(t, err, "plugin-config update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", pluginConfigID)
	require.NoError(t, err, "plugin-config get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "200")

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "delete", pluginConfigID, "--force")
	require.NoError(t, err, "plugin-config delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "plugin-config", "get", pluginConfigID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestPluginConfig_ListEmpty(t *testing.T) {
	env := setupPluginConfigEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "plugin-config", "list")
	require.NoError(t, err, "plugin-config list failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	noItems := strings.Contains(combined, "No plugin configs") ||
		strings.Contains(combined, "no plugin configs") ||
		strings.Contains(combined, "0")
	assert.True(t, noItems || strings.TrimSpace(stdout) == "" || strings.TrimSpace(stdout) == "[]",
		"list should indicate no plugin configs found, got: %s", combined)
}

func TestPluginConfig_GetNonExistent(t *testing.T) {
	env := setupPluginConfigEnv(t)

	_, stderr, err := runA6WithEnv(env, "plugin-config", "get", "nonexistent-plugin-config-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestPluginConfig_JSONOutput(t *testing.T) {
	const pluginConfigID = "test-plugin-config-json-1"

	deletePluginConfig(t, pluginConfigID)
	t.Cleanup(func() { deletePluginConfig(t, pluginConfigID) })

	env := setupPluginConfigEnv(t)

	createJSON := `{
		"id": "test-plugin-config-json-1",
		"plugins": {
			"limit-count": {
				"count": 100,
				"time_window": 60,
				"rejected_code": 503,
				"key_type": "var",
				"key": "remote_addr"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "plugin-config-json.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "plugin-config", "create", "-f", createFile)
	require.NoError(t, err, "plugin-config create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "list", "--output", "json")
	require.NoError(t, err, "plugin-config list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, pluginConfigID)

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", pluginConfigID, "--output", "json")
	require.NoError(t, err, "plugin-config get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, pluginConfigID)
}
