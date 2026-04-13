//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("route command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupRouteEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates routes from JSON and YAML files against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-route-json"
			yamlID := "ginkgo-route-yaml"

			deleteRouteViaCLIByID(env, jsonID)
			deleteRouteViaCLIByID(env, yamlID)
			DeferCleanup(deleteRouteViaAdminByID, g, jsonID)
			DeferCleanup(deleteRouteViaAdminByID, g, yamlID)

			jsonFile := writeRouteFile(g, "route.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-json",
  "uri": "/ginkgo-json",
  "methods": ["GET"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, jsonID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			getJSON, getErr := adminAPI("GET", "/apisix/admin/routes/"+jsonID, nil)
			g.Expect(getErr).NotTo(HaveOccurred())
			g.Expect(getJSON.StatusCode).To(Equal(200))
			g.Expect(getJSON.Body.Close()).To(Succeed())

			yamlFile := writeRouteFile(g, "route.yaml", fmt.Sprintf(`id: %s
name: ginkgo-route-yaml
uri: /ginkgo-yaml
methods:
  - GET
upstream:
  type: roundrobin
  nodes:
    "%s": 1
`, yamlID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-route-yaml"))

			getYAML, getErr := adminAPI("GET", "/apisix/admin/routes/"+yamlID, nil)
			g.Expect(getErr).NotTo(HaveOccurred())
			g.Expect(getYAML.StatusCode).To(Equal(200))
			g.Expect(getYAML.Body.Close()).To(Succeed())
		})

		It("uses the route id from the file and surfaces real validation errors", func() {
			g := NewWithT(GinkgoT())
			routeID := "ginkgo-route-explicit-id"

			deleteRouteViaCLIByID(env, routeID)
			DeferCleanup(deleteRouteViaAdminByID, g, routeID)

			createFile := writeRouteFile(g, "route-id.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-explicit-id",
  "uri": "/ginkgo-id",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, routeID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			resp, apiErr := adminAPI("GET", "/apisix/admin/routes/"+routeID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			_, stderr, err = runA6WithEnv(env, "route", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			invalidFile := writeRouteFile(g, "route-invalid.json", `{"name":"missing-upstream","uri":"/broken"}`)
			_, stderr, err = runA6WithEnv(env, "route", "create", "-f", invalidFile)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("error"))
		})
	})

	Describe("list", func() {
		It("renders table, json, and yaml output from real routes", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-route-list-1"
			id2 := "ginkgo-route-list-2"

			deleteRouteViaCLIByID(env, id1)
			deleteRouteViaCLIByID(env, id2)
			DeferCleanup(deleteRouteViaAdminByID, g, id1)
			DeferCleanup(deleteRouteViaAdminByID, g, id2)

			createFile1 := writeRouteFile(g, "route-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-list-1",
  "uri": "/ginkgo-list-1",
  "labels": {"suite":"ginkgo-route-list"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", createFile1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			createFile2 := writeRouteFile(g, "route-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-list-2",
  "uri": "/ginkgo-list-2",
  "labels": {"suite":"ginkgo-route-list"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", createFile2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "route", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "route", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "route", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))
			g.Expect(stdout).To(ContainSubstring("id: " + id2))
		})

		It("supports real filters, empty results, and authentication failures", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-route-filter-1"
			id2 := "ginkgo-route-filter-2"

			deleteRouteViaCLIByID(env, id1)
			deleteRouteViaCLIByID(env, id2)
			DeferCleanup(deleteRouteViaAdminByID, g, id1)
			DeferCleanup(deleteRouteViaAdminByID, g, id2)

			createFile1 := writeRouteFile(g, "route-filter-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-alpha",
  "uri": "/ginkgo-filter-alpha",
  "labels": {"suite":"ginkgo-filter"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", createFile1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			createFile2 := writeRouteFile(g, "route-filter-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-beta",
  "uri": "/ginkgo-filter-beta",
  "labels": {"suite":"ginkgo-other"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", createFile2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "route", "list", "--name", "ginkgo-route-alpha", "--uri", "/ginkgo-filter-alpha", "--label", "suite=ginkgo-filter", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "route", "list", "--label", "suite=does-not-exist", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("No routes found."))

			badEnv := setupRouteEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "route", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})
	})

	Describe("get/update/delete", func() {
		It("gets routes in yaml/json, updates them, and deletes them against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			routeID := "ginkgo-route-lifecycle"

			deleteRouteViaCLIByID(env, routeID)
			DeferCleanup(deleteRouteViaAdminByID, g, routeID)

			createFile := writeRouteFile(g, "route-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-lifecycle",
  "uri": "/ginkgo-lifecycle",
  "methods": ["GET"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, routeID, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-route-lifecycle"))
			g.Expect(stdout).To(ContainSubstring("uri: /ginkgo-lifecycle"))

			stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(routeID))

			updateFile := writeRouteFile(g, "route-update.json", `{
  "name": "ginkgo-route-updated",
  "uri": "/ginkgo-updated",
  "methods": ["GET", "POST"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    }
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-route-updated"))

			stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-route-updated"))
			g.Expect(stdout).To(ContainSubstring("/ginkgo-updated"))

			stdout, stderr, err = runA6WithEnv(env, "route", "delete", routeID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("deleted"))

			resp, apiErr := adminAPI("GET", "/apisix/admin/routes/"+routeID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(404))
			g.Expect(resp.Body.Close()).To(Succeed())
		})

		It("surfaces get/delete not-found behavior and update argument validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			deleteRouteViaCLIByID(env, "nonexistent-route-999")
			DeferCleanup(deleteRouteViaAdminByID, g, "nonexistent-route-999")

			_, stderr, err := runA6WithEnv(env, "route", "get", "nonexistent-route-999")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "route", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "route", "delete", "nonexistent-route-999", "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
