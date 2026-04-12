//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func setupUpstreamEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create upstream test context: %s", stderr)
	return env
}

func writeUpstreamFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deleteUpstreamViaAdminByID(g Gomega, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/upstreams/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteUpstreamViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "upstream", "delete", id, "--force")
}

func createUpstreamViaCLIFile(g Gomega, env []string, file string) {
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", file)
	g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
}

var _ = Describe("upstream command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupUpstreamEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates upstreams from JSON and YAML files against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-upstream-json"
			yamlID := "ginkgo-upstream-yaml"

			deleteUpstreamViaCLIByID(env, jsonID)
			deleteUpstreamViaCLIByID(env, yamlID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, jsonID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, yamlID)

			jsonFile := writeUpstreamFile(g, "upstream.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-json",
  "type": "roundrobin",
  "nodes": {
    "%s": 1
  }
}`, jsonID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			resp, apiErr := adminAPI("GET", "/apisix/admin/upstreams/"+jsonID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			yamlFile := writeUpstreamFile(g, "upstream.yaml", fmt.Sprintf(`id: %s
name: ginkgo-upstream-yaml
type: roundrobin
nodes:
  "%s": 1
`, yamlID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-upstream-yaml"))
		})

		It("uses explicit ids and preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			upstreamID := "ginkgo-upstream-explicit-id"

			deleteUpstreamViaCLIByID(env, upstreamID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

			createFile := writeUpstreamFile(g, "upstream-id.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-explicit-id",
  "type": "roundrobin",
  "nodes": {
    "%s": 1
  }
}`, upstreamID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			resp, apiErr := adminAPI("GET", "/apisix/admin/upstreams/"+upstreamID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			_, stderr, err = runA6WithEnv(env, "upstream", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and real filter behavior", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-upstream-list-1"
			id2 := "ginkgo-upstream-list-2"

			deleteUpstreamViaCLIByID(env, id1)
			deleteUpstreamViaCLIByID(env, id2)
			DeferCleanup(deleteUpstreamViaAdminByID, g, id1)
			DeferCleanup(deleteUpstreamViaAdminByID, g, id2)

			file1 := writeUpstreamFile(g, "upstream-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-alpha",
  "type": "roundrobin",
  "scheme": "http",
  "labels": {"suite":"ginkgo-upstream-list"},
  "nodes": {
    "%s": 1
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, file1)

			file2 := writeUpstreamFile(g, "upstream-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-beta",
  "type": "roundrobin",
  "scheme": "http",
  "labels": {"suite":"ginkgo-upstream-other"},
  "nodes": {
    "%s": 1
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "upstream", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))
			g.Expect(stdout).To(ContainSubstring("id: " + id2))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "list", "--name", "ginkgo-upstream-alpha", "--label", "suite=ginkgo-upstream-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "list", "--label", "suite=does-not-exist", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("No upstreams found."))

			badEnv := setupUpstreamEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "upstream", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})

		It("exports upstreams with real label filtering and strips timestamps from output", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-upstream-export-1"
			id2 := "ginkgo-upstream-export-2"

			deleteUpstreamViaCLIByID(env, id1)
			deleteUpstreamViaCLIByID(env, id2)
			DeferCleanup(deleteUpstreamViaAdminByID, g, id1)
			DeferCleanup(deleteUpstreamViaAdminByID, g, id2)

			file1 := writeUpstreamFile(g, "upstream-export-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-export-1",
  "type": "roundrobin",
  "labels": {"suite":"ginkgo-upstream-export"},
  "nodes": {
    "%s": 1
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, file1)

			file2 := writeUpstreamFile(g, "upstream-export-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-export-2",
  "type": "roundrobin",
  "labels": {"suite":"ginkgo-upstream-other"},
  "nodes": {
    "%s": 1
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "upstream", "export", "--label", "suite=ginkgo-upstream-export", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "upstream-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "upstream", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
			g.Expect(string(exported)).NotTo(ContainSubstring("create_time"))
			g.Expect(string(exported)).NotTo(ContainSubstring("update_time"))

			_, stderr, err = runA6WithEnv(env, "upstream", "export", "--label", "suite=does-not-exist", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("No upstreams found."))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets upstreams in yaml/json, updates them, and deletes them against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			upstreamID := "ginkgo-upstream-lifecycle"

			deleteUpstreamViaCLIByID(env, upstreamID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

			createFile := writeUpstreamFile(g, "upstream-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-lifecycle",
  "type": "roundrobin",
  "scheme": "http",
  "nodes": {
    "%s": 1
  }
}`, upstreamID, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, createFile)

			stdout, stderr, err := runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-upstream-lifecycle"))
			g.Expect(stdout).To(ContainSubstring("type: roundrobin"))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(upstreamID))

			updateFile := writeUpstreamFile(g, "upstream-update.json", fmt.Sprintf(`{
  "name": "ginkgo-upstream-updated",
  "type": "roundrobin",
  "scheme": "http",
  "nodes": {
    "%s": 1
  }
}`, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err = runA6WithEnv(env, "upstream", "update", upstreamID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-upstream-updated"))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-upstream-updated"))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("deleted"))

			resp, apiErr := adminAPI("GET", "/apisix/admin/upstreams/"+upstreamID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(404))
			g.Expect(resp.Body.Close()).To(Succeed())
		})

		It("surfaces get and delete not-found behavior and update required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			upstreamID := "nonexistent-upstream-999"

			deleteUpstreamViaCLIByID(env, upstreamID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

			_, stderr, err := runA6WithEnv(env, "upstream", "get", upstreamID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "upstream", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})

	Describe("health", func() {
		It("shows health data through the real Control API in table and json formats", func() {
			g := NewWithT(GinkgoT())
			upstreamID := "ginkgo-upstream-health"
			routeID := "ginkgo-route-health"

			deleteRouteViaCLIByID(env, routeID)
			deleteUpstreamViaCLIByID(env, upstreamID)
			DeferCleanup(deleteRouteViaAdminByID, g, routeID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

			upstreamFile := writeUpstreamFile(g, "upstream-health.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-health",
  "type": "roundrobin",
  "nodes": {
    "%s": 1
  },
  "checks": {
    "active": {
      "type": "http",
      "http_path": "/get",
      "healthy": {
        "interval": 1,
        "successes": 1
      },
      "unhealthy": {
        "interval": 1,
        "http_failures": 3
      }
    }
  }
}`, upstreamID, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, upstreamFile)

			routeFile := writeRouteFile(g, "route-health.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-route-health",
  "uri": "/ginkgo-upstream-health/*",
  "upstream_id": "%s",
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/ginkgo-upstream-health/(.*)", "/$1"]
    }
  }
}`, routeID, upstreamID))
			stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			time.Sleep(3 * time.Second)
			for range 5 {
				resp, reqErr := http.Get(gatewayURL + "/ginkgo-upstream-health/get")
				if reqErr == nil && resp != nil {
					g.Expect(resp.Body.Close()).To(Succeed())
				}
				time.Sleep(500 * time.Millisecond)
			}
			time.Sleep(3 * time.Second)

			stdout, stderr, err = runA6WithEnv(env, "upstream", "health", upstreamID, "--control-url", controlURL, "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("NODE"))
			g.Expect(stdout).To(ContainSubstring("healthy"))
			g.Expect(stdout).To(ContainSubstring("Type: http"))

			stdout, stderr, err = runA6WithEnv(env, "upstream", "health", upstreamID, "--control-url", controlURL, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			var healthResp map[string]any
			g.Expect(json.Unmarshal([]byte(stdout), &healthResp)).To(Succeed())
			g.Expect(healthResp["type"]).To(Equal("http"))
			nodes, ok := healthResp["nodes"].([]any)
			g.Expect(ok).To(BeTrue())
			g.Expect(nodes).NotTo(BeEmpty())
		})

		It("surfaces missing health-check data from the real Control API", func() {
			g := NewWithT(GinkgoT())
			upstreamID := "ginkgo-upstream-no-health"

			deleteUpstreamViaCLIByID(env, upstreamID)
			DeferCleanup(deleteUpstreamViaAdminByID, g, upstreamID)

			upstreamFile := writeUpstreamFile(g, "upstream-no-health.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-upstream-no-health",
  "type": "roundrobin",
  "nodes": {
    "%s": 1
  }
}`, upstreamID, hostPortFromURL(g, httpbinURL)))
			createUpstreamViaCLIFile(g, env, upstreamFile)

			_, stderr, err := runA6WithEnv(env, "upstream", "health", upstreamID, "--control-url", controlURL)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("health check"))
		})
	})
})
