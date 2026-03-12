//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPersonaDeveloper(t *testing.T) {
	env := setupEnv(t)

	const upstreamID = "skill-dev-upstream"
	const routeID = "skill-dev-route"
	const serviceID = "skill-dev-service"
	const svcRouteID = "skill-dev-svc-route"
	const username = "skill-dev-user"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "route", "delete", svcRouteID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	_, _, _ = runA6WithEnv(env, "service", "delete", serviceID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", username, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupRoute(t, svcRouteID) })
	t.Cleanup(func() { cleanupService(t, serviceID) })
	t.Cleanup(func() { cleanupUpstream(t, upstreamID) })
	t.Cleanup(func() { cleanupConsumer(t, username) })

	upstreamJSON := `{"id":"skill-dev-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`
	f := writeJSON(t, "upstream", upstreamJSON)
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "upstream create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{"id":"skill-dev-route","uri":"/skill-dev-api/*","methods":["GET","POST"],"upstream_id":"skill-dev-upstream","plugins":{"proxy-rewrite":{"regex_uri":["^/skill-dev-api/(.*)","/$1"]}}}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-dev-api/get", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)

	consumerJSON := `{"username":"skill-dev-user","plugins":{"key-auth":{"key":"skill-dev-key"}}}`
	f = writeJSON(t, "consumer", consumerJSON)
	stdout, stderr, err = runA6WithEnv(env, "consumer", "create", "-f", f)
	require.NoError(t, err, "consumer create: stdout=%s stderr=%s", stdout, stderr)

	updateJSON := `{"id":"skill-dev-route","uri":"/skill-dev-api/*","methods":["GET","POST"],"upstream_id":"skill-dev-upstream","plugins":{"key-auth":{},"proxy-rewrite":{"regex_uri":["^/skill-dev-api/(.*)","/$1"]}}}`
	f = writeJSON(t, "route-update", updateJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", f)
	require.NoError(t, err, "route update: stdout=%s stderr=%s", stdout, stderr)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-dev-api/get", nil, 401, 5*time.Second)
	assert.Equal(t, 401, status)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-dev-api/get",
		map[string]string{"apikey": "skill-dev-key"}, 200, 5*time.Second)
	assert.Equal(t, 200, status)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"key-auth"`)
	assert.Contains(t, stdout, `"skill-dev-upstream"`)

	stdout, _, err = runA6WithEnv(env, "route", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, routeID)

	serviceJSON := `{"id":"skill-dev-service","upstream_id":"skill-dev-upstream","plugins":{"key-auth":{},"proxy-rewrite":{"regex_uri":["^/skill-dev-svc/(.*)","/$1"]}}}`
	f = writeJSON(t, "service", serviceJSON)
	stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", f)
	require.NoError(t, err, "service create: stdout=%s stderr=%s", stdout, stderr)

	svcRouteJSON := `{"id":"skill-dev-svc-route","uri":"/skill-dev-svc/*","service_id":"skill-dev-service"}`
	f = writeJSON(t, "svc-route", svcRouteJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "svc-route create: stdout=%s stderr=%s", stdout, stderr)

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-dev-svc/get", nil, 401, 5*time.Second)
	assert.Equal(t, 401, status)

	stdout, _, err = runA6WithEnv(env, "plugin", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-auth")

	stdout, _, err = runA6WithEnv(env, "plugin", "get", "key-auth", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, "header")

	stdout, stderr, err = runA6WithEnv(env, "route", "delete", svcRouteID, "--force")
	require.NoError(t, err, "route delete: stdout=%s stderr=%s", stdout, stderr)
}
