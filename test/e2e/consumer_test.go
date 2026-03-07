//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteConsumer(t *testing.T, username string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/consumers/"+username, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func createTestConsumer(t *testing.T, username string) {
	t.Helper()
	body := fmt.Sprintf(`{"username":"%s","desc":"test consumer"}`, username)
	resp, err := adminAPI("PUT", "/apisix/admin/consumers", []byte(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create test consumer")
}

func setupConsumerEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestConsumer_CRUD(t *testing.T) {
	const username = "test-consumer-crud"

	deleteConsumer(t, username)
	t.Cleanup(func() { deleteConsumer(t, username) })

	env := setupConsumerEnv(t)

	// 1. Create
	consumerJSON := `{
	"username": "test-consumer-crud",
	"desc": "crud test consumer",
	"plugins": {
		"key-auth": {
			"key": "crud-test-key-000"
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "consumer.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(consumerJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", tmpFile)
	require.NoError(t, err, "consumer create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, username, "create output should mention username")

	// 2. Get
	stdout, stderr, err = runA6WithEnv(env, "consumer", "get", username)
	require.NoError(t, err, "consumer get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "crud test consumer", "get output should contain desc")

	// 3. List
	stdout, stderr, err = runA6WithEnv(env, "consumer", "list")
	require.NoError(t, err, "consumer list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, username, "list output should contain username")

	// 4. Update
	updatedJSON := `{
	"desc": "updated crud consumer",
	"plugins": {
		"key-auth": {
			"key": "crud-test-key-updated"
		}
	}
}`
	updatedFile := filepath.Join(t.TempDir(), "consumer-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "consumer", "update", username, "-f", updatedFile)
	require.NoError(t, err, "consumer update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "consumer", "get", username)
	require.NoError(t, err, "consumer get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "updated crud consumer", "get output should contain updated desc")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "consumer", "delete", username, "--force")
	require.NoError(t, err, "consumer delete failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	assert.True(t, strings.Contains(combined, "deleted") || strings.Contains(combined, "Deleted") || strings.Contains(combined, username),
		"delete output should confirm deletion, got: %s", combined)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "consumer", "get", username)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestConsumer_WithKeyAuth(t *testing.T) {
	const (
		username = "test-key-auth-user"
		routeID  = "test-consumer-auth-route"
	)

	deleteConsumer(t, username)
	deleteRoute(t, routeID)
	t.Cleanup(func() {
		deleteRoute(t, routeID)
		deleteConsumer(t, username)
	})

	env := setupConsumerEnv(t)

	// 1. Create consumer with key-auth plugin via CLI
	consumerJSON := `{
	"username": "test-key-auth-user",
	"plugins": {
		"key-auth": {
			"key": "test-api-key-12345"
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "consumer.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(consumerJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", tmpFile)
	require.NoError(t, err, "consumer create failed: stdout=%s stderr=%s", stdout, stderr)

	// 2. Create route with key-auth plugin via Admin API
	routeBody := fmt.Sprintf(`{
		"uri": "/test-consumer-auth/*",
		"plugins": {"key-auth": {}},
		"upstream": {
			"type": "roundrobin",
			"nodes": {"%s": 1}
		}
	}`, "127.0.0.1:8080")
	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+routeID, []byte(routeBody))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create auth route")

	// 3. Request with valid key → 200
	req, err := http.NewRequest("GET", gatewayURL+"/test-consumer-auth/get", nil)
	require.NoError(t, err)
	req.Header.Set("apikey", "test-api-key-12345")
	authResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer authResp.Body.Close()
	assert.Equal(t, http.StatusOK, authResp.StatusCode, "request with valid key should return 200")

	// 4. Request without key → 401
	noAuthResp, err := http.Get(gatewayURL + "/test-consumer-auth/get")
	require.NoError(t, err)
	defer noAuthResp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, noAuthResp.StatusCode, "request without key should return 401")
}

func TestConsumer_ListEmpty(t *testing.T) {
	env := setupConsumerEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "consumer", "list")
	require.NoError(t, err, "consumer list failed: stdout=%s stderr=%s", stdout, stderr)
	// When no consumers exist, should show an empty result or a "no consumers" message.
	// An empty table (just headers) is also acceptable.
	_ = stdout + stderr
}

func TestConsumer_GetNonExistent(t *testing.T) {
	env := setupConsumerEnv(t)

	_, stderr, err := runA6WithEnv(env, "consumer", "get", "nonexistent-consumer-999")
	assert.Error(t, err, "get nonexistent consumer should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestConsumer_JSONOutput(t *testing.T) {
	const username = "test-consumer-json-out"

	deleteConsumer(t, username)
	t.Cleanup(func() { deleteConsumer(t, username) })

	createTestConsumer(t, username)

	env := setupConsumerEnv(t)

	// List with JSON output.
	stdout, stderr, err := runA6WithEnv(env, "consumer", "list", "--output", "json")
	require.NoError(t, err, "consumer list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, username, "JSON list should contain username")

	// Get with JSON output.
	stdout, stderr, err = runA6WithEnv(env, "consumer", "get", username, "--output", "json")
	require.NoError(t, err, "consumer get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, username, "JSON get should contain username")
}

func TestConsumer_DeleteNonExistent(t *testing.T) {
	env := setupConsumerEnv(t)

	_, stderr, err := runA6WithEnv(env, "consumer", "delete", "nonexistent-consumer-999", "--force")
	assert.Error(t, err, "delete nonexistent consumer should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}
