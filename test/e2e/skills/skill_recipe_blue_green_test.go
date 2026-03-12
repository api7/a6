//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeBlueGreen(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-blue-green-route"
	const blueUpstreamID = "skill-blue-upstream"
	const greenUpstreamID = "skill-green-upstream"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", blueUpstreamID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", greenUpstreamID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupUpstream(t, blueUpstreamID) })
	t.Cleanup(func() { cleanupUpstream(t, greenUpstreamID) })

	blueJSON := `{"id":"skill-blue-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
	f := writeJSON(t, "blue-upstream", blueJSON)
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "blue upstream create: stdout=%s stderr=%s", stdout, stderr)

	greenJSON := `{"id":"skill-green-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
	f = writeJSON(t, "green-upstream", greenJSON)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "green upstream create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{
		"id": "skill-blue-green-route",
		"uri": "/skill-blue-green",
		"upstream_id": "skill-blue-upstream",
		"plugins": {
			"traffic-split": {
				"rules": [{
					"weighted_upstreams": [
						{"upstream_id": "skill-green-upstream", "weight": 100}
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
	assert.Contains(t, stdout, `"skill-green-upstream"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-blue-green", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
