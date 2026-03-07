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

func deleteConsumerGroup(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/consumer_groups/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupConsumerGroupEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestConsumerGroup_CRUD(t *testing.T) {
	const groupID = "test-consumer-group-1"

	deleteConsumerGroup(t, groupID)
	t.Cleanup(func() { deleteConsumerGroup(t, groupID) })

	env := setupConsumerGroupEnv(t)

	createJSON := `{
		"id": "test-consumer-group-1",
		"plugins": {
			"limit-count": {
				"count": 200,
				"time_window": 60,
				"rejected_code": 503,
				"key_type": "var",
				"key": "remote_addr"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "consumer-group.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "consumer-group", "create", "-f", createFile)
	require.NoError(t, err, "consumer-group create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, groupID)

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID)
	require.NoError(t, err, "consumer-group get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "limit-count")

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "list")
	require.NoError(t, err, "consumer-group list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, groupID)

	updateJSON := `{
		"plugins": {
			"limit-count": {
				"count": 300,
				"time_window": 60,
				"rejected_code": 503,
				"key_type": "var",
				"key": "remote_addr"
			},
			"prometheus": {}
		}
	}`
	updateFile := filepath.Join(t.TempDir(), "consumer-group-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "update", groupID, "-f", updateFile)
	require.NoError(t, err, "consumer-group update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID)
	require.NoError(t, err, "consumer-group get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "prometheus")

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "delete", groupID, "--force")
	require.NoError(t, err, "consumer-group delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestConsumerGroup_ListEmpty(t *testing.T) {
	const cleanupID = "test-consumer-group-list-empty-clean"
	deleteConsumerGroup(t, cleanupID)

	env := setupConsumerGroupEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "consumer-group", "list")
	require.NoError(t, err, "consumer-group list failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	noGroups := strings.Contains(combined, "No consumer groups") ||
		strings.Contains(combined, "no consumer groups") ||
		strings.Contains(combined, "0")
	assert.True(t, noGroups || strings.TrimSpace(stdout) == "" || strings.TrimSpace(stdout) == "[]",
		"list should indicate no consumer groups found, got: %s", combined)
}

func TestConsumerGroup_GetNonExistent(t *testing.T) {
	env := setupConsumerGroupEnv(t)

	_, stderr, err := runA6WithEnv(env, "consumer-group", "get", "nonexistent-consumer-group-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestConsumerGroup_JSONOutput(t *testing.T) {
	const groupID = "test-consumer-group-json-1"

	deleteConsumerGroup(t, groupID)
	t.Cleanup(func() { deleteConsumerGroup(t, groupID) })

	env := setupConsumerGroupEnv(t)

	createJSON := `{
		"id": "test-consumer-group-json-1",
		"plugins": {
			"limit-count": {
				"count": 200,
				"time_window": 60,
				"rejected_code": 503,
				"key_type": "var",
				"key": "remote_addr"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "consumer-group-json.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "consumer-group", "create", "-f", createFile)
	require.NoError(t, err, "consumer-group create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "list", "--output", "json")
	require.NoError(t, err, "consumer-group list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, groupID)

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID, "--output", "json")
	require.NoError(t, err, "consumer-group get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, groupID)
}
