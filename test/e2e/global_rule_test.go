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

func deleteGlobalRule(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/global_rules/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupGlobalRuleEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestGlobalRule_CRUD(t *testing.T) {
	const ruleID = "test-global-rule-crud-1"

	deleteGlobalRule(t, ruleID)
	t.Cleanup(func() { deleteGlobalRule(t, ruleID) })

	env := setupGlobalRuleEnv(t)

	createJSON := `{
		"id": "test-global-rule-crud-1",
		"plugins": {
			"prometheus": {}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "global-rule.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "global-rule", "create", "-f", createFile)
	require.NoError(t, err, "global-rule create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, ruleID)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID)
	require.NoError(t, err, "global-rule get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "prometheus")

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "list")
	require.NoError(t, err, "global-rule list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, ruleID)

	updateJSON := `{
		"plugins": {
			"prometheus": {},
			"ip-restriction": {
				"whitelist": ["0.0.0.0/0"]
			}
		}
	}`
	updateFile := filepath.Join(t.TempDir(), "global-rule-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "update", ruleID, "-f", updateFile)
	require.NoError(t, err, "global-rule update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID)
	require.NoError(t, err, "global-rule get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "ip-restriction")

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "delete", ruleID, "--force")
	require.NoError(t, err, "global-rule delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestGlobalRule_ListEmpty(t *testing.T) {
	const cleanupID = "test-global-rule-list-empty-clean"
	deleteGlobalRule(t, cleanupID)

	env := setupGlobalRuleEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "global-rule", "list")
	require.NoError(t, err, "global-rule list failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	noRules := strings.Contains(combined, "No global rules") ||
		strings.Contains(combined, "no global rules") ||
		strings.Contains(combined, "0")
	assert.True(t, noRules || strings.TrimSpace(stdout) == "" || strings.TrimSpace(stdout) == "[]",
		"list should indicate no global rules found, got: %s", combined)
}

func TestGlobalRule_GetNonExistent(t *testing.T) {
	env := setupGlobalRuleEnv(t)

	_, stderr, err := runA6WithEnv(env, "global-rule", "get", "nonexistent-global-rule-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestGlobalRule_DeleteNonExistent(t *testing.T) {
	env := setupGlobalRuleEnv(t)

	_, stderr, err := runA6WithEnv(env, "global-rule", "delete", "nonexistent-global-rule-999", "--force")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestGlobalRule_JSONOutput(t *testing.T) {
	const ruleID = "test-global-rule-json-1"

	deleteGlobalRule(t, ruleID)
	t.Cleanup(func() { deleteGlobalRule(t, ruleID) })

	env := setupGlobalRuleEnv(t)

	createJSON := `{
		"id": "test-global-rule-json-1",
		"plugins": {
			"prometheus": {}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "global-rule-json.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "global-rule", "create", "-f", createFile)
	require.NoError(t, err, "global-rule create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "list", "--output", "json")
	require.NoError(t, err, "global-rule list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, ruleID)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID, "--output", "json")
	require.NoError(t, err, "global-rule get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, ruleID)
}
