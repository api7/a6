//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoute_ServiceAndPluginConfigCombination(t *testing.T) {
	const (
		serviceID      = "test-scenario-svc-combo"
		pluginConfigID = "test-scenario-pc-combo"
		routeID        = "test-scenario-route-combo"
	)

	env := setupServiceEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	deleteServiceViaCLI(t, env, serviceID)
	deletePluginConfigViaCLI(t, env, pluginConfigID)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID)
		deleteServiceViaAdmin(t, serviceID)
		deletePluginConfigViaAdmin(t, pluginConfigID)
	})

	serviceJSON := fmt.Sprintf(`{
		"id": "%s",
		"name": "scenario-combo-service",
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`, serviceID)
	serviceFile := filepath.Join(t.TempDir(), "service.json")
	require.NoError(t, os.WriteFile(serviceFile, []byte(serviceJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", serviceFile)
	skipIfLicenseRestricted(t, stdout, stderr, err)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)

	pluginConfigJSON := fmt.Sprintf(`{
		"id": "%s",
		"plugins": {
			"proxy-rewrite": {
				"uri": "/get",
				"headers": {
					"set": {
						"X-Scenario-Coverage": "service-route-plugin-config"
					}
				}
			}
		}
	}`, pluginConfigID)
	pluginConfigFile := filepath.Join(t.TempDir(), "plugin-config.json")
	require.NoError(t, os.WriteFile(pluginConfigFile, []byte(pluginConfigJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "create", "-f", pluginConfigFile)
	skipIfLicenseRestricted(t, stdout, stderr, err)
	require.NoError(t, err, "plugin-config create failed: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/scenario-combo",
		"name": "scenario-combo-route",
		"service_id": "%s",
		"plugin_config_id": "%s"
	}`, routeID, serviceID, pluginConfigID)
	routeFile := filepath.Join(t.TempDir(), "route.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	skipIfLicenseRestricted(t, stdout, stderr, err)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)

	var body string
	for i := 0; i < 10; i++ {
		resp, err := http.Get(gatewayURL + "/scenario-combo")
		require.NoError(t, err, "gateway request should succeed")

		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, readErr)

		if resp.StatusCode == http.StatusOK {
			body = string(payload)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.NotEmpty(t, body, "gateway should eventually proxy the route")

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &result))
	assert.Equal(t, fmt.Sprintf("%s/get", httpbinURL), result["url"])

	headers, ok := result["headers"].(map[string]interface{})
	require.True(t, ok, "httpbin response should expose request headers")
	assert.Equal(t, "service-route-plugin-config", headers["X-Scenario-Coverage"])
}

func TestUpstream_MultiNodeRealTraffic(t *testing.T) {
	const (
		upstreamID = "test-scenario-upstream-multi"
		routeID    = "test-scenario-route-multi"
	)

	env := setupUpstreamEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	deleteUpstreamViaCLI(t, env, upstreamID)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID)
		deleteUpstreamViaAdmin(t, upstreamID)
	})

	serverA := newNamedTestServer(t, "backend-a")
	serverB := newNamedTestServer(t, "backend-b")

	upstreamJSON := fmt.Sprintf(`{
		"id": "%s",
		"name": "scenario-multi-node-upstream",
		"type": "roundrobin",
		"nodes": {
			"%s": 1,
			"%s": 1
		}
	}`, upstreamID, hostPort(t, serverA.URL), hostPort(t, serverB.URL))
	upstreamFile := filepath.Join(t.TempDir(), "upstream.json")
	require.NoError(t, os.WriteFile(upstreamFile, []byte(upstreamJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", upstreamFile)
	skipIfLicenseRestricted(t, stdout, stderr, err)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/scenario-multi-node",
		"name": "scenario-multi-node-route",
		"upstream_id": "%s"
	}`, routeID, upstreamID)
	routeFile := filepath.Join(t.TempDir(), "route.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	skipIfLicenseRestricted(t, stdout, stderr, err)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)

	seen := map[string]bool{}
	for i := 0; i < 12; i++ {
		resp, err := http.Get(gatewayURL + "/scenario-multi-node")
		require.NoError(t, err, "gateway request should succeed")

		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, readErr)

		if resp.StatusCode == http.StatusOK {
			seen[strings.TrimSpace(string(payload))] = true
			if seen["backend-a"] && seen["backend-b"] {
				break
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	assert.True(t, seen["backend-a"], "traffic should reach backend-a, seen=%v", seen)
	assert.True(t, seen["backend-b"], "traffic should reach backend-b, seen=%v", seen)
}

func newNamedTestServer(t *testing.T, name string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(name))
	}))
	t.Cleanup(server.Close)
	return server
}

func hostPort(t *testing.T, rawURL string) string {
	t.Helper()

	host := strings.TrimPrefix(rawURL, "http://")
	_, _, err := net.SplitHostPort(host)
	require.NoError(t, err)
	return host
}

func skipIfLicenseRestricted(t *testing.T, stdout, stderr string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	combined := stdout + stderr
	if strings.Contains(combined, "requires a sufficient license") {
		t.Skipf("environment blocks scenario coverage with a license gate: %s", strings.TrimSpace(combined))
	}
}
