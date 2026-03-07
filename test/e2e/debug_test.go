//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugTrace_BasicRoute(t *testing.T) {
	const routeID = "test-debug-trace-basic"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	routeBody := `{
		"uri": "/debug-trace-basic/*",
		"name": "debug-trace-basic-route",
		"methods": ["GET"],
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/debug-trace-basic/(.*)", "/$1"]
			}
		},
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`

	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+routeID, []byte(routeBody))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)

	env := setupRouteEnv(t)
	env = append(env, "APISIX_GATEWAY_URL="+gatewayURL)

	stdout, stderr, err := runA6WithEnv(env, "debug", "trace", routeID, "--path", "/debug-trace-basic/get", "--output", "json")
	require.NoError(t, err, "debug trace failed: stdout=%s stderr=%s", stdout, stderr)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	routeObj, ok := got["route"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, routeID, routeObj["id"])

	respObj, ok := got["response"].(map[string]interface{})
	require.True(t, ok)
	status, ok := respObj["status"].(float64)
	require.True(t, ok)
	assert.Equal(t, 200.0, status)

	plugins, ok := got["configured_plugins"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, plugins)
}

func TestDebugTrace_NonExistentRoute(t *testing.T) {
	env := setupRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "debug", "trace", "non-existent-debug-route")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "resource not found") || strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"), "expected not found error, got: %s", stderr)
}

func TestDebugTrace_WithMethodAndPath(t *testing.T) {
	const routeID = "test-debug-trace-post"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	routeBody := `{
		"uri": "/debug-trace-post/*",
		"name": "debug-trace-post-route",
		"methods": ["POST"],
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/debug-trace-post/(.*)", "/$1"]
			}
		},
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`

	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+routeID, []byte(routeBody))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)

	env := setupRouteEnv(t)
	env = append(env, "APISIX_GATEWAY_URL="+gatewayURL)

	stdout, stderr, err := runA6WithEnv(env,
		"debug", "trace", routeID,
		"--method", "POST",
		"--path", "/debug-trace-post/post",
		"--body", `{"hello":"world"}`,
		"--header", "Content-Type: application/json",
		"--output", "json",
	)
	require.NoError(t, err, "debug trace POST failed: stdout=%s stderr=%s", stdout, stderr)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	reqObj, ok := got["request"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "POST", reqObj["method"])
	assert.Equal(t, fmt.Sprintf("%s/debug-trace-post/post", gatewayURL), reqObj["url"])

	respObj, ok := got["response"].(map[string]interface{})
	require.True(t, ok)
	status, ok := respObj["status"].(float64)
	require.True(t, ok)
	assert.Equal(t, 200.0, status)
}

func TestDebugTrace_NoArgsNonTTY(t *testing.T) {
	_, stderr, err := runA6("debug", "trace")
	require.Error(t, err)
	assert.Contains(t, stderr, "route-id argument is required")
}
