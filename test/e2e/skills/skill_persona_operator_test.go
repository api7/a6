//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPersonaOperator(t *testing.T) {
	env := setupEnv(t)

	const globalRuleID = "skill-op-global-rule"
	const upstreamID = "skill-op-upstream"
	const routeID = "skill-op-route"

	_, _, _ = runA6WithEnv(env, "global-rule", "delete", globalRuleID, "--force")
	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	t.Cleanup(func() { cleanupGlobalRule(t, globalRuleID) })
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupUpstream(t, upstreamID) })

	var stdout, stderr string

	_, _, err := runA6WithEnv(env, "route", "list")
	require.NoError(t, err)

	upstreamJSON := `{"id":"skill-op-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
	f := writeJSON(t, "upstream", upstreamJSON)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "upstream create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "upstream", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, upstreamID)

	routeJSON := `{"id":"skill-op-route","uri":"/skill-op-test","upstream_id":"skill-op-upstream","plugins":{"proxy-rewrite":{"uri":"/get"}}}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-op-test", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)

	globalRuleJSON := `{
		"id": "skill-op-global-rule",
		"plugins": {
			"limit-count": {
				"count": 50000,
				"time_window": 60,
				"key_type": "var",
				"key": "remote_addr",
				"rejected_code": 429
			}
		}
	}`
	f = writeJSON(t, "global-rule", globalRuleJSON)
	stdout, stderr, err = runA6WithEnv(env, "global-rule", "create", "-f", f)
	require.NoError(t, err, "global-rule create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "global-rule", "get", globalRuleID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"limit-count"`)

	_, _, err = runA6WithEnv(env, "ssl", "list")
	require.NoError(t, err)

	stdout, _, err = runA6WithEnv(env, "route", "list", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, routeID)

	stdout, _, err = runA6WithEnv(env, "upstream", "list", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, upstreamID)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "delete", globalRuleID, "--force")
	require.NoError(t, err, "global-rule delete: stdout=%s stderr=%s", stdout, stderr)
}
