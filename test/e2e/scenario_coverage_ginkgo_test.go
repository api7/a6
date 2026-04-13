//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("scenario coverage", func() {
	It("covers the route + service + plugin-config combination with real APISIX traffic", func() {
		g := NewWithT(GinkgoT())
		const (
			serviceID      = "test-scenario-svc-combo"
			pluginConfigID = "test-scenario-pc-combo"
			routeID        = "test-scenario-route-combo"
		)

		env := setupCLIEnvWithKey(g, adminKey)

		deleteRouteViaCLIByID(env, routeID)
		deleteServiceViaCLIByID(env, serviceID)
		deletePluginConfigViaCLIByID(env, pluginConfigID)
		DeferCleanup(deleteRouteViaAdminByID, g, routeID)
		DeferCleanup(deleteServiceViaAdminByID, g, serviceID)
		DeferCleanup(deletePluginConfigViaAdminByID, g, pluginConfigID)

		serviceFile := writeServiceFile(g, "scenario-service.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "scenario-combo-service",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, serviceID, hostPortFromURL(g, httpbinURL)))
		stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", serviceFile)
		skipIfLicenseRestrictedGomega(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "service create failed: stdout=%s stderr=%s", stdout, stderr)

		pluginConfigFile := writePluginConfigFile(g, "scenario-plugin-config.json", fmt.Sprintf(`{
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
}`, pluginConfigID))
		stdout, stderr, err = runA6WithEnv(env, "plugin-config", "create", "-f", pluginConfigFile)
		skipIfLicenseRestrictedGomega(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "plugin-config create failed: stdout=%s stderr=%s", stdout, stderr)

		routeFile := writeRouteFile(g, "scenario-route.json", fmt.Sprintf(`{
  "id": "%s",
  "uri": "/scenario-combo",
  "name": "scenario-combo-route",
  "service_id": "%s",
  "plugin_config_id": "%s"
}`, routeID, serviceID, pluginConfigID))
		stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
		skipIfLicenseRestrictedGomega(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "route create failed: stdout=%s stderr=%s", stdout, stderr)

		var body string
		var lastErr error
		var lastStatus int
		for i := 0; i < 10; i++ {
			resp, reqErr := http.Get(gatewayURL + "/scenario-combo")
			if reqErr != nil {
				lastErr = reqErr
				time.Sleep(500 * time.Millisecond)
				continue
			}

			payload, readErr := io.ReadAll(resp.Body)
			g.Expect(resp.Body.Close()).To(Succeed())
			g.Expect(readErr).NotTo(HaveOccurred())
			lastStatus = resp.StatusCode

			if resp.StatusCode == http.StatusOK {
				body = string(payload)
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		g.Expect(body).NotTo(BeEmpty(), "gateway should eventually proxy the route, last_err=%v last_status=%d", lastErr, lastStatus)

		var result map[string]interface{}
		g.Expect(json.Unmarshal([]byte(body), &result)).To(Succeed())
		proxiedURL, ok := result["url"].(string)
		g.Expect(ok).To(BeTrue(), "httpbin response should expose url as string")
		u, parseErr := url.Parse(proxiedURL)
		g.Expect(parseErr).NotTo(HaveOccurred())
		g.Expect(u.Path).To(Equal("/get"))

		headers, ok := result["headers"].(map[string]interface{})
		g.Expect(ok).To(BeTrue(), "httpbin response should expose request headers")
		expectHeaderContains(g, headers["X-Scenario-Coverage"], "service-route-plugin-config")
	})

	It("covers multi-node upstream traffic with real APISIX load balancing", func() {
		g := NewWithT(GinkgoT())
		const (
			upstreamID = "test-scenario-upstream-multi"
			routeID    = "test-scenario-route-multi"
		)

		env := setupCLIEnvWithKey(g, adminKey)

		deleteRouteViaCLIByID(env, routeID)
		deleteUpstreamViaCLIByID(env, upstreamID)
		DeferCleanup(deleteRouteViaAdminByID, g, routeID)
		DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

		serverA := newNamedTestServer("backend-a")
		serverB := newNamedTestServer("backend-b")
		DeferCleanup(serverA.Close)
		DeferCleanup(serverB.Close)

		upstreamFile := writeUpstreamFile(g, "scenario-upstream.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "scenario-multi-node-upstream",
  "type": "roundrobin",
  "nodes": {
    "%s": 1,
    "%s": 1
  }
}`, upstreamID, hostPortFromURL(g, serverA.URL), hostPortFromURL(g, serverB.URL)))
		stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", upstreamFile)
		skipIfLicenseRestrictedGomega(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

		routeFile := writeRouteFile(g, "scenario-multi-route.json", fmt.Sprintf(`{
  "id": "%s",
  "uri": "/scenario-multi-node",
  "name": "scenario-multi-node-route",
  "upstream_id": "%s"
}`, routeID, upstreamID))
		stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
		skipIfLicenseRestrictedGomega(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "route create failed: stdout=%s stderr=%s", stdout, stderr)

		seen := map[string]bool{}
		var lastErr error
		var lastStatus int
		for i := 0; i < 10; i++ {
			resp, reqErr := http.Get(gatewayURL + "/scenario-multi-node")
			if reqErr != nil {
				lastErr = reqErr
				time.Sleep(500 * time.Millisecond)
				continue
			}

			payload, readErr := io.ReadAll(resp.Body)
			g.Expect(resp.Body.Close()).To(Succeed())
			g.Expect(readErr).NotTo(HaveOccurred())
			lastStatus = resp.StatusCode

			if resp.StatusCode == http.StatusOK {
				seen[strings.TrimSpace(string(payload))] = true
				if seen["backend-a"] && seen["backend-b"] {
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}

		g.Expect(seen["backend-a"]).To(BeTrue(), "traffic should reach backend-a, seen=%v last_err=%v last_status=%d", seen, lastErr, lastStatus)
		g.Expect(seen["backend-b"]).To(BeTrue(), "traffic should reach backend-b, seen=%v last_err=%v last_status=%d", seen, lastErr, lastStatus)
	})
})

func newNamedTestServer(name string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(name))
	}))
}

func expectHeaderContains(g Gomega, raw interface{}, want string) {
	switch v := raw.(type) {
	case string:
		g.Expect(v).To(ContainSubstring(want))
	case []interface{}:
		values := make([]string, 0, len(v))
		for _, item := range v {
			values = append(values, fmt.Sprintf("%v", item))
		}
		g.Expect(values).To(ContainElement(want))
	default:
		Fail(fmt.Sprintf("unexpected header value type: %T", raw))
	}
}
