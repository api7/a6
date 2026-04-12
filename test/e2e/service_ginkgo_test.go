//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func setupServiceEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create service test context: %s", stderr)
	return env
}

func writeServiceFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deleteServiceViaAdminByID(g Gomega, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/services/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteServiceViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "service", "delete", id, "--force")
}

var _ = Describe("service command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupServiceEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates services from JSON and YAML files against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-service-json"
			yamlID := "ginkgo-service-yaml"

			deleteServiceViaCLIByID(env, jsonID)
			deleteServiceViaCLIByID(env, yamlID)
			DeferCleanup(deleteServiceViaAdminByID, g, jsonID)
			DeferCleanup(deleteServiceViaAdminByID, g, yamlID)

			jsonFile := writeServiceFile(g, "service.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-json",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, jsonID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			resp, apiErr := adminAPI("GET", "/apisix/admin/services/"+jsonID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			yamlFile := writeServiceFile(g, "service.yaml", fmt.Sprintf(`id: %s
name: ginkgo-service-yaml
upstream:
  type: roundrobin
  nodes:
    "%s": 1
`, yamlID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-service-yaml"))
		})

		It("uses explicit ids and preserves CLI-level required-flag validation", func() {
			g := NewWithT(GinkgoT())
			serviceID := "ginkgo-service-explicit-id"

			deleteServiceViaCLIByID(env, serviceID)
			DeferCleanup(deleteServiceViaAdminByID, g, serviceID)

			createFile := writeServiceFile(g, "service-id.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-explicit-id",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, serviceID, hostPortFromURL(g, httpbinURL)))

			stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			resp, apiErr := adminAPI("GET", "/apisix/admin/services/"+serviceID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			_, stderr, err = runA6WithEnv(env, "service", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

		})
	})

	Describe("list and export", func() {
		It("renders list output modes and real filter behavior", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-service-list-1"
			id2 := "ginkgo-service-list-2"

			deleteServiceViaCLIByID(env, id1)
			deleteServiceViaCLIByID(env, id2)
			DeferCleanup(deleteServiceViaAdminByID, g, id1)
			DeferCleanup(deleteServiceViaAdminByID, g, id2)

			file1 := writeServiceFile(g, "service-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-alpha",
  "labels": {"suite":"ginkgo-service-list"},
  "plugins": {"limit-count":{"count":10,"time_window":60}},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", file1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeServiceFile(g, "service-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-beta",
  "labels": {"suite":"ginkgo-service-other"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", file2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "service", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "service", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))

			stdout, stderr, err = runA6WithEnv(env, "service", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))

			stdout, stderr, err = runA6WithEnv(env, "service", "list", "--name", "ginkgo-service-alpha", "--label", "suite=ginkgo-service-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "service", "list", "--label", "suite=does-not-exist", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("No services found."))

			badEnv := setupServiceEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "service", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})

		It("exports services with real label filtering and strips timestamps from output", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-service-export-1"
			id2 := "ginkgo-service-export-2"

			deleteServiceViaCLIByID(env, id1)
			deleteServiceViaCLIByID(env, id2)
			DeferCleanup(deleteServiceViaAdminByID, g, id1)
			DeferCleanup(deleteServiceViaAdminByID, g, id2)

			file1 := writeServiceFile(g, "service-export-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-export-1",
  "labels": {"suite":"ginkgo-service-export"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id1, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", file1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeServiceFile(g, "service-export-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-service-export-2",
  "labels": {"suite":"ginkgo-service-other"},
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "%s": 1
    }
  }
}`, id2, hostPortFromURL(g, httpbinURL)))
			stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", file2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "service", "export", "--label", "suite=ginkgo-service-export", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "service-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "service", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
			g.Expect(string(exported)).NotTo(ContainSubstring("create_time"))
			g.Expect(string(exported)).NotTo(ContainSubstring("update_time"))
		})
	})
})
