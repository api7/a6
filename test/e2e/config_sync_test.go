//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteConfigSyncUpstreamViaAdmin(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/upstreams/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func deleteConfigSyncUpstreamViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	runA6WithEnv(env, "upstream", "delete", id, "--force")
}

func deleteConfigSyncRouteViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	runA6WithEnv(env, "route", "delete", id, "--force")
}

func TestConfig_SyncAndDiff(t *testing.T) {
	const (
		routeID    = "test-sync-route-1"
		upstreamID = "test-sync-upstream-1"
	)

	env := setupRouteEnv(t)

	deleteConfigSyncRouteViaCLI(t, env, routeID)
	deleteConfigSyncUpstreamViaCLI(t, env, upstreamID)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID)
		deleteConfigSyncUpstreamViaAdmin(t, upstreamID)
	})
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: test-sync-route-1
    uri: /sync-test-1
    name: sync-route-1
    upstream_id: test-sync-upstream-1
upstreams:
  - id: test-sync-upstream-1
    name: sync-upstream-1
    type: roundrobin
    nodes:
      "127.0.0.1:8080": 1
`), 0o644))

	var stdout, stderr string
	var err error
	for i := 0; i < 3; i++ {
		stdout, stderr, err = runA6WithEnv(env, "config", "sync", "-f", configPath)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	require.NoError(t, err, "config sync failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sync-route-1")

	stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID)
	require.NoError(t, err, "upstream get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sync-upstream-1")

	stdout, stderr, err = runA6WithEnv(env, "config", "diff", "-f", configPath)
	require.NoError(t, err, "config diff failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, strings.ToLower(stdout), "no differences")

	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: test-sync-route-1
    uri: /sync-test-1
    name: sync-route-1-updated
    upstream_id: test-sync-upstream-1
upstreams:
  - id: test-sync-upstream-1
    name: sync-upstream-1
    type: roundrobin
    nodes:
      "127.0.0.1:8080": 1
`), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "config", "diff", "-f", configPath)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(fmt.Sprintf("%s\n%s", stdout, stderr)), "differences")

	stdout, stderr, err = runA6WithEnv(env, "config", "sync", "-f", configPath)
	require.NoError(t, err, "config sync update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sync-route-1-updated")

	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: test-sync-route-1
    uri: /sync-test-1
    name: sync-route-1-updated
    upstream:
      type: roundrobin
      nodes:
        "127.0.0.1:8080": 1
`), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "config", "sync", "-f", configPath)
	require.NoError(t, err, "config sync delete upstream failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID)
	require.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"expected upstream not found, got: %s", stderr)
}

func TestConfig_SyncDryRun(t *testing.T) {
	const routeID = "test-sync-dry-run-route-1"

	env := setupRouteEnv(t)

	deleteConfigSyncRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: test-sync-dry-run-route-1
    uri: /sync-dry-run
    name: sync-dry-run-route
    upstream:
      type: roundrobin
      nodes:
        "127.0.0.1:8080": 1
`), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "config", "sync", "-f", configPath, "--dry-run")
	require.NoError(t, err, "config sync dry-run failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "CREATE test-sync-dry-run-route-1")

	_, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"expected route not found after dry-run, got: %s", stderr)
}
