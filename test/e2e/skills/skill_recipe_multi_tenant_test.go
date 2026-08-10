//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeMultiTenant(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-multi-tenant-route"
	const tenantA = "skill-tenant-a"
	const tenantB = "skill-tenant-b"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", tenantA, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", tenantB, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupConsumer(t, tenantA) })
	t.Cleanup(func() { cleanupConsumer(t, tenantB) })

	f := writeJSON(t, "tenant-a", `{"username":"skill-tenant-a","plugins":{"key-auth":{"key":"tenant-a-key"}}}`)
	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", f)
	require.NoError(t, err, "tenant-a create: stdout=%s stderr=%s", stdout, stderr)

	f = writeJSON(t, "tenant-b", `{"username":"skill-tenant-b","plugins":{"key-auth":{"key":"tenant-b-key"}}}`)
	stdout, stderr, err = runA6WithEnv(env, "consumer", "create", "-f", f)
	require.NoError(t, err, "tenant-b create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{
		"id": "skill-multi-tenant-route",
		"uri": "/skill-multi-tenant",
		"plugins": {
			"key-auth": {},
			"limit-count": {
				"count": 100,
				"time_window": 86400,
				"key_type": "var",
				"key": "consumer_name"
			},
			"traffic-split": {
				"rules": [
					{
						"match": [{"vars": [["consumer_name", "==", "skill-tenant-a"]]}],
						"weighted_upstreams": [{
							"upstream": {
								"type": "roundrobin",
								"pass_host": "rewrite",
								"upstream_host": "skill-tenant-a.internal",
								"nodes": {"127.0.0.1:8080": 1}
							},
							"weight": 1
						}]
					},
					{
						"match": [{"vars": [["consumer_name", "==", "skill-tenant-b"]]}],
						"weighted_upstreams": [{
							"upstream": {
								"type": "roundrobin",
								"pass_host": "rewrite",
								"upstream_host": "skill-tenant-b.internal",
								"nodes": {"127.0.0.1:8080": 1}
							},
							"weight": 1
						}]
					}
				]
			},
			"proxy-rewrite": {
				"uri": "/get"
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"traffic-split"`)
	assert.Contains(t, stdout, `"limit-count"`)

	status, body := httpGetWithRetry(t, gatewayURL+"/skill-multi-tenant",
		map[string]string{"apikey": "tenant-a-key", "X-Tenant-ID": "skill-tenant-b"}, 200, 5*time.Second)
	assert.Equal(t, 200, status)
	assert.Contains(t, body, "skill-tenant-a.internal")
	assert.NotContains(t, body, "skill-tenant-b.internal")

	status, body = httpGetWithRetry(t, gatewayURL+"/skill-multi-tenant",
		map[string]string{"apikey": "tenant-b-key", "X-Tenant-ID": "skill-tenant-a"}, 200, 5*time.Second)
	assert.Equal(t, 200, status)
	assert.Contains(t, body, "skill-tenant-b.internal")
	assert.NotContains(t, body, "skill-tenant-a.internal")

	status, _ = httpGetWithRetry(t, gatewayURL+"/skill-multi-tenant",
		map[string]string{"apikey": "unknown-key"}, 401, 5*time.Second)
	assert.Equal(t, 401, status)
}
