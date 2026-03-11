//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverageFlags_RouteListPagination_PageAndPageSize(t *testing.T) {
	env := setupRouteEnv(t)

	routeIDs := []string{
		"pg-route-1", "pg-route-2", "pg-route-3", "pg-route-4", "pg-route-5", "pg-route-6",
		"pg-route-7", "pg-route-8", "pg-route-9", "pg-route-10", "pg-route-11",
	}

	for _, id := range routeIDs {
		deleteRouteViaCLI(t, env, id)
		idCopy := id
		t.Cleanup(func() { deleteRouteViaAdmin(t, idCopy) })
	}

	for i := range routeIDs {
		createTestRouteViaCLI(
			t,
			env,
			routeIDs[i],
			"pg-route-group",
			fmt.Sprintf("/pg-route-uri-%d", i+1),
		)
	}

	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--name", "pg-route-group", "--page", "1", "--page-size", "10", "--output", "json")
	require.NoError(t, err, "route list page 1 failed: stdout=%s stderr=%s", stdout, stderr)

	var page1 []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &page1), "page 1 output should be valid JSON array: %s", stdout)
	assert.Len(t, page1, 10, "page 1 with page-size=10 should have 10 items")

	stdout, stderr, err = runA6WithEnv(env, "route", "list", "--name", "pg-route-group", "--page", "2", "--page-size", "10", "--output", "json")
	require.NoError(t, err, "route list page 2 failed: stdout=%s stderr=%s", stdout, stderr)

	var page2 []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &page2), "page 2 output should be valid JSON array: %s", stdout)
	assert.Len(t, page2, 1, "page 2 with page-size=10 should have remaining 1 item")
}

func TestCoverageFlags_RouteListOutputYAML(t *testing.T) {
	env := setupRouteEnv(t)

	const routeID = "yaml-route-1"
	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "yaml-route-name", "/yaml-route-uri")

	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--output", "yaml")
	require.NoError(t, err, "route list --output yaml failed: stdout=%s stderr=%s", stdout, stderr)

	trimmed := strings.TrimSpace(stdout)
	assert.Contains(t, stdout, "name:", "yaml output should contain 'name:' key")
	assert.Contains(t, stdout, "uri:", "yaml output should contain 'uri:' key")
	assert.False(t, strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{"), "yaml output should not start like JSON: %s", trimmed)
}

func TestCoverageFlags_UpstreamListOutputYAML(t *testing.T) {
	env := setupUpstreamEnv(t)

	const upstreamID = "yaml-upstream-1"
	deleteUpstreamViaCLI(t, env, upstreamID)
	t.Cleanup(func() { deleteUpstreamViaAdmin(t, upstreamID) })

	createTestUpstreamViaCLI(t, env, upstreamID, "yaml-upstream-name")

	stdout, stderr, err := runA6WithEnv(env, "upstream", "list", "--output", "yaml")
	require.NoError(t, err, "upstream list --output yaml failed: stdout=%s stderr=%s", stdout, stderr)

	trimmed := strings.TrimSpace(stdout)
	assert.Contains(t, stdout, "name:", "yaml output should contain 'name:' key")
	assert.Contains(t, stdout, "nodes:", "yaml output should contain 'nodes:' key")
	assert.False(t, strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{"), "yaml output should not start like JSON: %s", trimmed)
}

func TestCoverageFlags_RouteListFilterByURI(t *testing.T) {
	env := setupRouteEnv(t)

	const (
		routeIDAlpha = "uri-route-alpha"
		routeIDBeta  = "uri-route-beta"
	)

	deleteRouteViaCLI(t, env, routeIDAlpha)
	deleteRouteViaCLI(t, env, routeIDBeta)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeIDAlpha)
		deleteRouteViaAdmin(t, routeIDBeta)
	})

	createTestRouteViaCLI(t, env, routeIDAlpha, "uri-filter-alpha", "/test-uri-filter-alpha")
	createTestRouteViaCLI(t, env, routeIDBeta, "uri-filter-beta", "/test-uri-filter-beta")

	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--uri", "/test-uri-filter-alpha", "--output", "json")
	require.NoError(t, err, "route list --uri failed: stdout=%s stderr=%s", stdout, stderr)

	var routes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &routes), "route list --uri should output valid JSON: %s", stdout)
	require.Len(t, routes, 1, "uri filter should return exactly one route")
	assert.Equal(t, "uri-filter-alpha", routes[0]["name"])
	assert.Equal(t, "/test-uri-filter-alpha", routes[0]["uri"])
}

func TestCoverageFlags_ServiceListFilterByName(t *testing.T) {
	env := setupServiceEnv(t)

	const (
		serviceIDAlpha = "name-svc-alpha"
		serviceIDBeta  = "name-svc-beta"
	)

	deleteServiceViaCLI(t, env, serviceIDAlpha)
	deleteServiceViaCLI(t, env, serviceIDBeta)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, serviceIDAlpha)
		deleteServiceViaAdmin(t, serviceIDBeta)
	})

	createTestServiceViaCLI(t, env, serviceIDAlpha, "svc-name-alpha")
	createTestServiceViaCLI(t, env, serviceIDBeta, "svc-name-beta")

	stdout, stderr, err := runA6WithEnv(env, "service", "list", "--name", "svc-name-alpha", "--output", "json")
	require.NoError(t, err, "service list --name failed: stdout=%s stderr=%s", stdout, stderr)

	var services []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &services), "service list --name should output valid JSON: %s", stdout)
	require.Len(t, services, 1, "name filter should return exactly one service")
	assert.Equal(t, "svc-name-alpha", services[0]["name"])
}

func TestCoverageFlags_ConfigDumpOutputJSON(t *testing.T) {
	env := setupRouteEnv(t)

	const routeID = "dump-json-route-1"
	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "dump-json-route-name", "/dump-json-route-uri")

	stdout, stderr, err := runA6WithEnv(env, "config", "dump", "--output", "json")
	require.NoError(t, err, "config dump --output json failed: stdout=%s stderr=%s", stdout, stderr)
	require.True(t, json.Valid([]byte(stdout)), "config dump --output json should produce valid JSON: %s", stdout)

	var dumped map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &dumped))
	assert.Equal(t, "1", dumped["version"])
	assert.Contains(t, stdout, "dump-json-route-name")
	assert.Contains(t, stdout, "/dump-json-route-uri")
}

func TestCoverageFlags_ConfigDiffOutputJSON(t *testing.T) {
	env := setupRouteEnv(t)

	const routeID = "diff-json-route-1"
	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "diff-json-route-old", "/diff-json-route-uri")

	configPath := filepath.Join(t.TempDir(), "config-diff.json.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: diff-json-route-1
    uri: /diff-json-route-uri
    name: diff-json-route-new
    upstream:
      type: roundrobin
      nodes:
        "127.0.0.1:8080": 1
`), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "config", "diff", "-f", configPath, "--output", "json")
	require.Error(t, err, "config diff should exit non-zero when differences exist")
	require.True(t, json.Valid([]byte(stdout)), "config diff --output json should still output valid JSON: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "routes")
	assert.Contains(t, stdout, "diff-json-route-1")
}

func TestCoverageFlags_ConfigSyncDeleteFalse(t *testing.T) {
	env := setupRouteEnv(t)

	const (
		routeAID = "sync-del-false-route-a"
		routeBID = "sync-del-false-route-b"
	)

	deleteRouteViaCLI(t, env, routeAID)
	deleteRouteViaCLI(t, env, routeBID)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeAID)
		deleteRouteViaAdmin(t, routeBID)
	})

	createTestRouteViaCLI(t, env, routeAID, "sync-del-false-route-a-name", "/sync-del-false-a")

	configPath := filepath.Join(t.TempDir(), "config-sync-delete-false.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
version: "1"
routes:
  - id: sync-del-false-route-b
    uri: /sync-del-false-b
    name: sync-del-false-route-b-name
    upstream:
      type: roundrobin
      nodes:
        "127.0.0.1:8080": 1
`), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "config", "sync", "-f", configPath, "--delete=false")
	require.NoError(t, err, "config sync --delete=false failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeAID)
	require.NoError(t, err, "route A should still exist after --delete=false sync: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sync-del-false-route-a-name")

	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeBID)
	require.NoError(t, err, "route B should be created by sync: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sync-del-false-route-b-name")
}

func TestCoverageFlags_DebugTraceHostFlag(t *testing.T) {
	env := setupRouteEnv(t)
	env = append(env, "APISIX_GATEWAY_URL="+gatewayURL)

	const routeID = "trace-host-route-1"
	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	routeBody := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/trace-host-test/*",
		"name": "trace-host-route-name",
		"hosts": ["example.com"],
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/trace-host-test/(.*)", "/$1"]
			}
		},
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`, routeID)
	routeFile := filepath.Join(t.TempDir(), "trace-host-route.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeBody), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create for trace host test failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env,
		"debug", "trace", routeID,
		"--path", "/trace-host-test/get",
		"--host", "example.com",
		"--output", "json",
	)
	require.NoError(t, err, "debug trace --host failed: stdout=%s stderr=%s", stdout, stderr)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got), "trace output should be valid JSON: %s", stdout)

	requestObj, ok := got["request"].(map[string]interface{})
	require.True(t, ok, "trace output should contain request object")
	headersObj, ok := requestObj["headers"].(map[string]interface{})
	require.True(t, ok, "trace request should contain headers")
	assert.Equal(t, "example.com", headersObj["Host"])
}

func TestCoverageFlags_DebugTraceGatewayURLFlag(t *testing.T) {
	env := setupRouteEnv(t)

	const routeID = "trace-gateway-route-1"
	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	routeBody := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/trace-gateway-test/*",
		"name": "trace-gateway-route-name",
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/trace-gateway-test/(.*)", "/$1"]
			}
		},
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`, routeID)
	routeFile := filepath.Join(t.TempDir(), "trace-gateway-route.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeBody), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create for trace gateway test failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env,
		"debug", "trace", routeID,
		"--path", "/trace-gateway-test/get",
		"--gateway-url", gatewayURL,
		"--output", "json",
	)
	require.NoError(t, err, "debug trace --gateway-url failed: stdout=%s stderr=%s", stdout, stderr)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got), "trace output should be valid JSON: %s", stdout)

	requestObj, ok := got["request"].(map[string]interface{})
	require.True(t, ok, "trace output should contain request object")
	requestURL, ok := requestObj["url"].(string)
	require.True(t, ok, "trace request.url should be string")
	assert.True(t, strings.HasPrefix(requestURL, gatewayURL), "request URL should use provided gateway URL, got: %s", requestURL)
}

func TestCoverageFlags_DebugLogsTypeFlag(t *testing.T) {
	env := setupRouteEnv(t)
	if _, _, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "1"); err != nil {
		t.Skipf("skipping debug logs type tests because docker logs is unavailable: %v", err)
	}

	stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "5", "--type", "error")
	require.NoError(t, err, "debug logs --type error failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "5", "--type", "access")
	require.NoError(t, err, "debug logs --type access failed: stdout=%s stderr=%s", stdout, stderr)
}

func TestCoverageFlags_DebugLogsOutputJSONFlag(t *testing.T) {
	env := setupRouteEnv(t)
	if _, _, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "1"); err != nil {
		t.Skipf("skipping debug logs json test because docker logs is unavailable: %v", err)
	}

	stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "5", "--output", "json")
	require.NoError(t, err, "debug logs --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.NotEmpty(t, strings.TrimSpace(stdout), "debug logs --output json should produce output")
}

func TestCoverageFlags_CompletionFish(t *testing.T) {
	stdout, stderr, err := runA6("completion", "fish")
	require.NoError(t, err, "completion fish failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "complete -c a6", "fish completion should contain fish markers")
}

func TestCoverageFlags_CompletionPowerShell(t *testing.T) {
	stdout, stderr, err := runA6("completion", "powershell")
	require.NoError(t, err, "completion powershell failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "Register-ArgumentCompleter", "powershell completion should contain completer registration")
}
