//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginTrafficSplit(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-traffic-split-route"
	const upstreamID = "skill-traffic-upstream"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupUpstream(t, upstreamID) })

	upstreamJSON := `{"id":"skill-traffic-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
	upstreamFile := writeJSON(t, "upstream", upstreamJSON)
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", upstreamFile)
	require.NoError(t, err, "upstream create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"skill-traffic-upstream"`)

	routeJSON := `{"id":"skill-traffic-split-route","uri":"/skill-traffic-split","plugins":{"traffic-split":{"rules":[{"weighted_upstreams":[{"upstream_id":"skill-traffic-upstream","weight":100}]}]},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"traffic-split"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-traffic-split", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
