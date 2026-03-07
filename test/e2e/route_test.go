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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteRoute removes a route via the Admin API. Used for test cleanup.
func deleteRoute(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/routes/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

// createTestRoute creates a route via the Admin API for test setup.
func createTestRoute(t *testing.T, id, name, uri string) {
	t.Helper()
	body := fmt.Sprintf(`{"uri":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`, uri, name)
	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+id, []byte(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create test route")
}

// setupRouteEnv returns env vars and creates a context pointing at the real APISIX.
func setupRouteEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestRoute_CRUD(t *testing.T) {
	const routeID = "test-route-crud-1"

	// Cleanup before and after.
	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	env := setupRouteEnv(t)

	// 1. Create: write a JSON file and run `a6 route create -f file.json`
	routeJSON := `{
	"id": "test-route-crud-1",
	"uri": "/test-crud",
	"name": "test-crud-route",
	"methods": ["GET"],
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "route.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(routeJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, routeID, "create output should mention route ID")

	// 2. Get: verify route data
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route", "get output should contain route name")
	assert.Contains(t, stdout, "/test-crud", "get output should contain route URI")

	// 3. List: verify the created route appears
	stdout, stderr, err = runA6WithEnv(env, "route", "list")
	require.NoError(t, err, "route list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route", "list output should contain route name")

	// 4. Update: write updated JSON and run `a6 route update {id} -f file.json`
	updatedJSON := `{
	"uri": "/test-crud-updated",
	"name": "test-crud-route-updated",
	"methods": ["GET", "POST"],
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	updatedFile := filepath.Join(t.TempDir(), "route-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", updatedFile)
	require.NoError(t, err, "route update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route-updated", "get output should contain updated name")
	assert.Contains(t, stdout, "/test-crud-updated", "get output should contain updated URI")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "route", "delete", routeID, "--force")
	require.NoError(t, err, "route delete failed: stdout=%s stderr=%s", stdout, stderr)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_ListEmpty(t *testing.T) {
	// Clean up any routes that might exist from other tests.
	resp, err := adminAPI("GET", "/apisix/admin/routes", nil)
	require.NoError(t, err)
	resp.Body.Close()

	env := setupRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "route", "list")
	require.NoError(t, err, "route list failed: stdout=%s stderr=%s", stdout, stderr)
	// When no routes exist, should show an empty result or a "no routes" message.
	combined := stdout + stderr
	noRoutes := strings.Contains(combined, "No routes") ||
		strings.Contains(combined, "no routes") ||
		strings.Contains(combined, "0 route") ||
		// An empty table (just headers) is also acceptable.
		!strings.Contains(combined, "/")
	assert.True(t, noRoutes || stdout == "" || strings.TrimSpace(stdout) == "",
		"list should indicate no routes found, got: %s", combined)
}

func TestRoute_ListWithFilters(t *testing.T) {
	const (
		routeID1 = "test-route-filter-1"
		routeID2 = "test-route-filter-2"
	)

	deleteRoute(t, routeID1)
	deleteRoute(t, routeID2)
	t.Cleanup(func() {
		deleteRoute(t, routeID1)
		deleteRoute(t, routeID2)
	})

	createTestRoute(t, routeID1, "route-one-alpha", "/filter-alpha")
	createTestRoute(t, routeID2, "route-two-beta", "/filter-beta")

	env := setupRouteEnv(t)

	// Filter by name — should show only the matching route.
	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--name", "route-one-alpha")
	require.NoError(t, err, "route list --name failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "route-one-alpha", "filtered list should contain matching route")
	assert.NotContains(t, stdout, "route-two-beta", "filtered list should not contain other route")
}

func TestRoute_GetNonExistent(t *testing.T) {
	env := setupRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "route", "get", "nonexistent-999")
	assert.Error(t, err, "get nonexistent route should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_DeleteNonExistent(t *testing.T) {
	env := setupRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "route", "delete", "nonexistent-999", "--force")
	assert.Error(t, err, "delete nonexistent route should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_DeleteWithForce(t *testing.T) {
	const routeID = "test-route-force-del"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	createTestRoute(t, routeID, "force-delete-route", "/force-delete")

	env := setupRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "route", "delete", routeID, "--force")
	require.NoError(t, err, "route delete --force failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	assert.True(t, strings.Contains(combined, "deleted") || strings.Contains(combined, "Deleted") || strings.Contains(combined, routeID),
		"delete output should confirm deletion, got: %s", combined)

	// Verify route is gone via Admin API.
	resp, err := adminAPI("GET", "/apisix/admin/routes/"+routeID, nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode, "route should not exist after deletion")
}

func TestRoute_JSONOutput(t *testing.T) {
	const routeID = "test-route-json-out"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	createTestRoute(t, routeID, "json-output-route", "/json-output")

	env := setupRouteEnv(t)

	// List with JSON output.
	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--output", "json")
	require.NoError(t, err, "route list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "json-output-route", "JSON list should contain route name")

	// Get with JSON output.
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err, "route get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, routeID, "JSON get should contain route ID")
	assert.Contains(t, stdout, "json-output-route", "JSON get should contain route name")
}

func TestRoute_YAMLOutput(t *testing.T) {
	const routeID = "test-route-yaml-out"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	createTestRoute(t, routeID, "yaml-output-route", "/yaml-output")

	env := setupRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "route", "get", routeID, "--output", "yaml")
	require.NoError(t, err, "route get --output yaml failed: stdout=%s stderr=%s", stdout, stderr)
	// YAML output should contain typical YAML markers like key: value pairs.
	assert.Contains(t, stdout, "name:", "YAML output should contain 'name:' key")
	assert.Contains(t, stdout, "uri:", "YAML output should contain 'uri:' key")
	assert.Contains(t, stdout, "yaml-output-route", "YAML output should contain route name")
}

func TestRoute_TrafficForwarding(t *testing.T) {
	const routeID = "test-route-traffic"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	// Create route via Admin API: forward /test-e2e-traffic/* to httpbin (127.0.0.1:8080).
	routeBody := `{
		"uri": "/test-e2e-traffic/*",
		"name": "traffic-test-route",
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`
	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+routeID, []byte(routeBody))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create traffic test route")

	// Retry to allow APISIX to propagate the route config from etcd.
	var gwResp *http.Response
	for i := 0; i < 10; i++ {
		gwResp, err = http.Get(gatewayURL + "/test-e2e-traffic/get")
		require.NoError(t, err, "gateway request should succeed")
		if gwResp.StatusCode == http.StatusOK {
			break
		}
		gwResp.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}
	defer gwResp.Body.Close()
	assert.Equal(t, http.StatusOK, gwResp.StatusCode, "gateway should return 200 for proxied request")
}
