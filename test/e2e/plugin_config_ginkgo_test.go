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

func setupPluginConfigEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create plugin-config test context: %s", stderr)
	return env
}

func writePluginConfigFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deletePluginConfigViaAdminByID(g Gomega, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/plugin_configs/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deletePluginConfigViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "plugin-config", "delete", id, "--force")
}

func createPluginConfigViaCLIFile(g Gomega, env []string, file string) {
	stdout, stderr, err := runA6WithEnv(env, "plugin-config", "create", "-f", file)
	g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
}

var _ = Describe("plugin-config command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupPluginConfigEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates plugin configs from JSON and YAML files against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-plugin-config-json"
			yamlID := "ginkgo-plugin-config-yaml"

			deletePluginConfigViaCLIByID(env, jsonID)
			deletePluginConfigViaCLIByID(env, yamlID)
			DeferCleanup(deletePluginConfigViaAdminByID, g, jsonID)
			DeferCleanup(deletePluginConfigViaAdminByID, g, yamlID)

			jsonFile := writePluginConfigFile(g, "plugin-config.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-json",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, jsonID))

			stdout, stderr, err := runA6WithEnv(env, "plugin-config", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			resp, apiErr := adminAPI("GET", "/apisix/admin/plugin_configs/"+jsonID, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(200))
			g.Expect(resp.Body.Close()).To(Succeed())

			yamlFile := writePluginConfigFile(g, "plugin-config.yaml", fmt.Sprintf(`id: %s
name: ginkgo-plugin-config-yaml
plugins:
  prometheus: {}
`, yamlID))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-plugin-config-yaml"))
		})

		It("preserves required-flag and missing-id validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "plugin-config", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			missingIDFile := writePluginConfigFile(g, "plugin-config-missing-id.json", `{
  "plugins": {
    "prometheus": {}
  }
}`)
			_, stderr, err = runA6WithEnv(env, "plugin-config", "create", "-f", missingIDFile)
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring(`must include an "id" field`))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and real plugin summaries", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-plugin-config-list-1"
			id2 := "ginkgo-plugin-config-list-2"

			deletePluginConfigViaCLIByID(env, id1)
			deletePluginConfigViaCLIByID(env, id2)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id1)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id2)

			file1 := writePluginConfigFile(g, "plugin-config-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-alpha",
  "labels": {"suite":"ginkgo-plugin-config-list"},
  "plugins": {
    "prometheus": {},
    "limit-count": {
      "count": 10,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, id1))
			createPluginConfigViaCLIFile(g, env, file1)

			file2 := writePluginConfigFile(g, "plugin-config-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-beta",
  "labels": {"suite":"ginkgo-plugin-config-other"},
  "plugins": {
    "prometheus": {}
  }
}`, id2))
			createPluginConfigViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "plugin-config", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring("PLUGINS"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))
			g.Expect(stdout).To(ContainSubstring("id: " + id2))

			badEnv := setupPluginConfigEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "plugin-config", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})

		It("exports plugin configs with real label filtering and strips timestamps from output", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-plugin-config-export-1"
			id2 := "ginkgo-plugin-config-export-2"

			deletePluginConfigViaCLIByID(env, id1)
			deletePluginConfigViaCLIByID(env, id2)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id1)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id2)

			file1 := writePluginConfigFile(g, "plugin-config-export-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-export-1",
  "labels": {"suite":"ginkgo-plugin-config-export"},
  "plugins": {
    "prometheus": {}
  }
}`, id1))
			createPluginConfigViaCLIFile(g, env, file1)

			file2 := writePluginConfigFile(g, "plugin-config-export-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-export-2",
  "labels": {"suite":"ginkgo-plugin-config-other"},
  "plugins": {
    "prometheus": {}
  }
}`, id2))
			createPluginConfigViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "plugin-config", "export", "--label", "suite=ginkgo-plugin-config-export", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "plugin-config-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
			g.Expect(string(exported)).NotTo(ContainSubstring("create_time"))
			g.Expect(string(exported)).NotTo(ContainSubstring("update_time"))

			_, stderr, err = runA6WithEnv(env, "plugin-config", "export", "--label", "suite=does-not-exist", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("No plugin configs found."))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets plugin configs in yaml/json, updates them, and deletes them against the real Admin API", func() {
			g := NewWithT(GinkgoT())
			id := "ginkgo-plugin-config-lifecycle"

			deletePluginConfigViaCLIByID(env, id)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id)

			createFile := writePluginConfigFile(g, "plugin-config-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-plugin-config-lifecycle",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, id))
			createPluginConfigViaCLIFile(g, env, createFile)

			stdout, stderr, err := runA6WithEnv(env, "plugin-config", "get", id, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-plugin-config-lifecycle"))
			g.Expect(stdout).To(ContainSubstring("limit-count"))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", id, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id))

			updateFile := writePluginConfigFile(g, "plugin-config-update.json", `{
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 429,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "update", id, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(`"count": 200`))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "get", id, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(`"count": 200`))

			stdout, stderr, err = runA6WithEnv(env, "plugin-config", "delete", id, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("deleted"))

			resp, apiErr := adminAPI("GET", "/apisix/admin/plugin_configs/"+id, nil)
			g.Expect(apiErr).NotTo(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(404))
			g.Expect(resp.Body.Close()).To(Succeed())
		})

		It("surfaces get and delete not-found behavior and update required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			id := "nonexistent-plugin-config-999"

			deletePluginConfigViaCLIByID(env, id)
			DeferCleanup(deletePluginConfigViaAdminByID, g, id)

			_, stderr, err := runA6WithEnv(env, "plugin-config", "get", id)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "plugin-config", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "plugin-config", "delete", id, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
