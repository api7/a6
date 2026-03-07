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

func deleteStreamRoute(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/stream_routes/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupStreamRouteEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestStreamRoute_CRUD(t *testing.T) {
	const streamRouteID = "test-stream-route-crud-1"

	deleteStreamRoute(t, streamRouteID)
	t.Cleanup(func() { deleteStreamRoute(t, streamRouteID) })

	env := setupStreamRouteEnv(t)

	createJSON := `{
	"id": "test-stream-route-crud-1",
	"name": "test-tcp-proxy",
	"server_port": 9100,
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	createFile := filepath.Join(t.TempDir(), "stream-route-create.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "stream-route", "create", "-f", createFile)
	require.NoError(t, err, "stream-route create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, streamRouteID)

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID)
	require.NoError(t, err, "stream-route get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-tcp-proxy")
	assert.Contains(t, stdout, "9100")

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "list")
	require.NoError(t, err, "stream-route list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-tcp-proxy")

	updateJSON := `{
	"name": "test-tcp-proxy-updated",
	"server_port": 9101,
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	updateFile := filepath.Join(t.TempDir(), "stream-route-update.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "update", streamRouteID, "-f", updateFile)
	require.NoError(t, err, "stream-route update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID)
	require.NoError(t, err, "stream-route get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-tcp-proxy-updated")
	assert.Contains(t, stdout, "9101")

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "delete", streamRouteID, "--force")
	require.NoError(t, err, "stream-route delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestStreamRoute_ListEmpty(t *testing.T) {
	env := setupStreamRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "stream-route", "list")
	require.NoError(t, err, "stream-route list failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	noResources := strings.Contains(combined, "No stream routes") ||
		strings.Contains(combined, "no stream routes") ||
		strings.Contains(combined, "0 stream") ||
		!strings.Contains(combined, "test-stream-route")
	assert.True(t, noResources || stdout == "" || strings.TrimSpace(stdout) == "",
		"list should indicate no stream routes found, got: %s", combined)
}

func TestStreamRoute_GetNonExistent(t *testing.T) {
	env := setupStreamRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "stream-route", "get", "nonexistent-stream-route-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestStreamRoute_DeleteNonExistent(t *testing.T) {
	env := setupStreamRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "stream-route", "delete", "nonexistent-stream-route-999", "--force")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestStreamRoute_JSONOutput(t *testing.T) {
	const streamRouteID = "test-stream-route-json-out"

	deleteStreamRoute(t, streamRouteID)
	t.Cleanup(func() { deleteStreamRoute(t, streamRouteID) })

	body := `{
	"name":"json-stream-route",
	"server_port":9300,
	"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}
}`
	resp, err := adminAPI("PUT", "/apisix/admin/stream_routes/"+streamRouteID, []byte(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)

	env := setupStreamRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "stream-route", "list", "--output", "json")
	require.NoError(t, err, "stream-route list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "json-stream-route")

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID, "--output", "json")
	require.NoError(t, err, "stream-route get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, streamRouteID)
	assert.Contains(t, stdout, "json-stream-route")
}
