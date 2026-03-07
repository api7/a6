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

func deletePluginMetadata(t *testing.T, pluginName string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/plugin_metadata/"+pluginName, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupPluginMetadataEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestPluginMetadata_CRUD(t *testing.T) {
	const pluginName = "syslog"

	deletePluginMetadata(t, pluginName)
	t.Cleanup(func() { deletePluginMetadata(t, pluginName) })

	env := setupPluginMetadataEnv(t)

	createJSON := `{
		"log_format": {
			"host": "$host"
		}
	}`
	createFile := filepath.Join(t.TempDir(), "plugin-metadata.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "plugin-metadata", "create", pluginName, "-f", createFile)
	require.NoError(t, err, "plugin-metadata create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "log_format")

	stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName)
	require.NoError(t, err, "plugin-metadata get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "host: $host")

	updateJSON := `{
		"log_format": {
			"host": "$host",
			"request_id": "$request_id"
		}
	}`
	updateFile := filepath.Join(t.TempDir(), "plugin-metadata-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "update", pluginName, "-f", updateFile)
	require.NoError(t, err, "plugin-metadata update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName)
	require.NoError(t, err, "plugin-metadata get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "request_id: $request_id")

	stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "delete", pluginName, "--force")
	require.NoError(t, err, "plugin-metadata delete failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "Deleted plugin metadata for syslog")

	_, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestPluginMetadata_GetNonExistent(t *testing.T) {
	env := setupPluginMetadataEnv(t)

	_, stderr, err := runA6WithEnv(env, "plugin-metadata", "get", "nonexistent-plugin-metadata-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestPluginMetadata_JSONOutput(t *testing.T) {
	const pluginName = "syslog"

	deletePluginMetadata(t, pluginName)
	t.Cleanup(func() { deletePluginMetadata(t, pluginName) })

	env := setupPluginMetadataEnv(t)

	createJSON := `{
		"log_format": {
			"host": "$host"
		}
	}`
	createFile := filepath.Join(t.TempDir(), "plugin-metadata-json.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "plugin-metadata", "create", pluginName, "-f", createFile)
	require.NoError(t, err, "plugin-metadata create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName, "--output", "json")
	require.NoError(t, err, "plugin-metadata get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "log_format")
}
