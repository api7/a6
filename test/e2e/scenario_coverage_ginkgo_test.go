//go:build e2e

package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("scenario coverage", func() {
	It("covers the route + service + plugin-config resource combination through CLI reads", func() {
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

		stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring(`"service_id": "` + serviceID + `"`))
		g.Expect(stdout).To(ContainSubstring(`"plugin_config_id": "` + pluginConfigID + `"`))

		stdout, stderr, err = runA6WithEnv(env, "service", "get", serviceID, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring(`"name": "scenario-combo-service"`))
		g.Expect(stdout).To(ContainSubstring(hostPortFromURL(g, httpbinURL)))

		stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", pluginConfigID, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring("proxy-rewrite"))
		g.Expect(stdout).To(ContainSubstring("service-route-plugin-config"))
	})

	It("covers multi-node upstream and route binding through CLI reads", func() {
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

		upstreamFile := writeUpstreamFile(g, "scenario-upstream.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "scenario-multi-node-upstream",
  "type": "roundrobin",
  "nodes": {
    "127.0.0.1:19801": 1,
    "127.0.0.1:19802": 1
  }
}`, upstreamID))
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

		stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring(`"name": "scenario-multi-node-upstream"`))
		g.Expect(stdout).To(ContainSubstring("127.0.0.1:19801"))
		g.Expect(stdout).To(ContainSubstring("127.0.0.1:19802"))

		stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring(`"upstream_id": "` + upstreamID + `"`))
	})
})
