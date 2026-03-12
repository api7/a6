//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginConsumerRestriction(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-consumer-restrict-route"
	const allowedUsername = "skill-allowed-user"
	const deniedUsername = "skill-denied-user"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", allowedUsername, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", deniedUsername, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupConsumer(t, allowedUsername) })
	t.Cleanup(func() { cleanupConsumer(t, deniedUsername) })

	allowedConsumerJSON := `{"username":"skill-allowed-user","plugins":{"key-auth":{"key":"allowed-key-123"}}}`
	allowedFile := writeJSON(t, "allowed-consumer", allowedConsumerJSON)
	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", allowedFile)
	require.NoError(t, err, "allowed consumer create: stdout=%s stderr=%s", stdout, stderr)

	deniedConsumerJSON := `{"username":"skill-denied-user","plugins":{"key-auth":{"key":"denied-key-123"}}}`
	deniedFile := writeJSON(t, "denied-consumer", deniedConsumerJSON)
	stdout, stderr, err = runA6WithEnv(env, "consumer", "create", "-f", deniedFile)
	require.NoError(t, err, "denied consumer create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{"id":"skill-consumer-restrict-route","uri":"/skill-consumer-restrict","plugins":{"key-auth":{},"consumer-restriction":{"whitelist":["skill-allowed-user"]},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"consumer-restriction"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-consumer-restrict", map[string]string{"apikey": "allowed-key-123"}, 200, 5*time.Second)
	assert.Equal(t, 200, status)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-consumer-restrict", map[string]string{"apikey": "denied-key-123"}, 403, 5*time.Second)
	assert.Equal(t, 403, status)
}
