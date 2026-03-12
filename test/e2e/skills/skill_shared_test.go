//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillShared(t *testing.T) {
	configDir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + configDir}

	t.Run("ContextManagement", func(t *testing.T) {
		stdout, stderr, err := runA6WithEnv(env, "context", "create", "dev",
			"--server", adminURL, "--api-key", adminKey)
		require.NoError(t, err, "context create dev: stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "context", "create", "staging",
			"--server", adminURL, "--api-key", adminKey)
		require.NoError(t, err, "context create staging: stdout=%s stderr=%s", stdout, stderr)

		stdout, _, err = runA6WithEnv(env, "context", "list")
		require.NoError(t, err)
		assert.Contains(t, stdout, "dev")
		assert.Contains(t, stdout, "staging")

		stdout, stderr, err = runA6WithEnv(env, "context", "use", "staging")
		require.NoError(t, err, "context use: stdout=%s stderr=%s", stdout, stderr)

		stdout, _, err = runA6WithEnv(env, "context", "current")
		require.NoError(t, err)
		assert.Contains(t, stdout, "staging")

		stdout, stderr, err = runA6WithEnv(env, "context", "delete", "staging", "--force")
		require.NoError(t, err, "context delete: stdout=%s stderr=%s", stdout, stderr)

		_, _, _ = runA6WithEnv(env, "context", "use", "dev")
	})

	t.Run("ResourceCRUD", func(t *testing.T) {
		const routeID = "skill-shared-route"
		const upstreamID = "skill-shared-upstream"

		_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
		_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
		t.Cleanup(func() { cleanupRoute(t, routeID) })
		t.Cleanup(func() { cleanupUpstream(t, upstreamID) })

		upstreamJSON := `{"id":"skill-shared-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
		f := writeJSON(t, "upstream", upstreamJSON)
		stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", f)
		require.NoError(t, err, "upstream create: stdout=%s stderr=%s", stdout, stderr)

		routeJSON := `{"id":"skill-shared-route","uri":"/skill-shared","upstream_id":"skill-shared-upstream","plugins":{"proxy-rewrite":{"uri":"/get"}}}`
		f = writeJSON(t, "route", routeJSON)
		stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
		require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

		stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
		require.NoError(t, err)
		assert.Contains(t, stdout, `"skill-shared-route"`)
		assert.Contains(t, stdout, `"/skill-shared"`)

		stdout, _, err = runA6WithEnv(env, "route", "list")
		require.NoError(t, err)
		assert.Contains(t, stdout, routeID)

		status, _ := httpGetWithRetry(t, gatewayURL+"/skill-shared", nil, 200, 5*time.Second)
		assert.Equal(t, 200, status)

		updateJSON := `{"id":"skill-shared-route","uri":"/skill-shared-updated","upstream_id":"skill-shared-upstream","plugins":{"proxy-rewrite":{"uri":"/get"}}}`
		f = writeJSON(t, "route-update", updateJSON)
		stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", f)
		require.NoError(t, err, "route update: stdout=%s stderr=%s", stdout, stderr)

		stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
		require.NoError(t, err)
		assert.Contains(t, stdout, `"/skill-shared-updated"`)

		stdout, stderr, err = runA6WithEnv(env, "route", "delete", routeID, "--force")
		require.NoError(t, err, "route delete: stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
		require.NoError(t, err, "upstream delete: stdout=%s stderr=%s", stdout, stderr)
	})

	t.Run("PluginCommands", func(t *testing.T) {
		stdout, _, err := runA6WithEnv(env, "plugin", "list")
		require.NoError(t, err)
		assert.Contains(t, stdout, "key-auth")
		assert.Contains(t, stdout, "limit-count")
		assert.Contains(t, stdout, "proxy-rewrite")

		stdout, _, err = runA6WithEnv(env, "plugin", "get", "limit-count", "--output", "json")
		require.NoError(t, err)
		assert.Contains(t, stdout, `"count"`)
	})
}
