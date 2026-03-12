//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeCanary(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-canary-route"
	const stableUpstreamID = "skill-canary-stable"
	const canaryUpstreamID = "skill-canary-new"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", stableUpstreamID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", canaryUpstreamID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupUpstream(t, stableUpstreamID) })
	t.Cleanup(func() { cleanupUpstream(t, canaryUpstreamID) })

	f := writeJSON(t, "stable-upstream", `{"id":"skill-canary-stable","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`)
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "stable upstream create: stdout=%s stderr=%s", stdout, stderr)

	f = writeJSON(t, "canary-upstream", `{"id":"skill-canary-new","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "canary upstream create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{
		"id": "skill-canary-route",
		"uri": "/skill-canary",
		"upstream_id": "skill-canary-stable",
		"plugins": {
			"traffic-split": {
				"rules": [{
					"weighted_upstreams": [
						{"upstream_id": "skill-canary-new", "weight": 10},
						{"weight": 90}
					]
				}]
			},
			"proxy-rewrite": {
				"uri": "/get"
			}
		}
	}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"traffic-split"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-canary", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
