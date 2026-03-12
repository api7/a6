//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginKeyAuth(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-key-auth-route"
	const username = "skill-key-auth-user"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", username, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupConsumer(t, username) })

	consumerJSON := `{"username":"skill-key-auth-user","plugins":{"key-auth":{"key":"skill-key-auth-secret"}}}`
	consumerFile := writeJSON(t, "consumer", consumerJSON)
	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
	require.NoError(t, err, "consumer create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{"id":"skill-key-auth-route","uri":"/skill-key-auth","plugins":{"key-auth":{},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"key-auth"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-key-auth", nil, 401, 5*time.Second)
	assert.Equal(t, 401, status)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-key-auth", map[string]string{"apikey": "skill-key-auth-secret"}, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
