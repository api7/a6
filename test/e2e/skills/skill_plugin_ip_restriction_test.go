//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginIPRestriction(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ip-restrict-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-ip-restrict-route","uri":"/skill-ip-restrict","plugins":{"ip-restriction":{"whitelist":["192.168.99.99"]},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"ip-restriction"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-ip-restrict", nil, 403, 5*time.Second)
	assert.Equal(t, 403, status)

	routeUpdateJSON := `{"id":"skill-ip-restrict-route","uri":"/skill-ip-restrict","plugins":{"ip-restriction":{"whitelist":["0.0.0.0/0"]},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeUpdateFile := writeJSON(t, "route-update", routeUpdateJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", routeUpdateFile)
	require.NoError(t, err, "route update: stdout=%s stderr=%s", stdout, stderr)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-ip-restrict", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
