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

func setupGlobalRuleEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create global-rule test context: %s", stderr)
	return env
}

func writeGlobalRuleFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deleteGlobalRuleViaAdminByID(g Gomega, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/global_rules/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteGlobalRuleViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "global-rule", "delete", id, "--force")
}

func createGlobalRuleViaCLIFile(g Gomega, env []string, file string) {
	stdout, stderr, err := runA6WithEnv(env, "global-rule", "create", "-f", file)
	g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
}

var _ = Describe("global-rule command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupGlobalRuleEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates global rules from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-global-rule-json"
			yamlID := "ginkgo-global-rule-yaml"

			deleteGlobalRuleViaCLIByID(env, jsonID)
			deleteGlobalRuleViaCLIByID(env, yamlID)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, jsonID)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, yamlID)

			jsonFile := writeGlobalRuleFile(g, "global-rule.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "prometheus": {}
  }
}`, jsonID))
			stdout, stderr, err := runA6WithEnv(env, "global-rule", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			yamlFile := writeGlobalRuleFile(g, "global-rule.yaml", fmt.Sprintf(`id: %s
plugins:
  ip-restriction:
    whitelist:
      - 0.0.0.0/0
`, yamlID))
			stdout, stderr, err = runA6WithEnv(env, "global-rule", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + yamlID))
		})

		It("preserves required-flag and missing-id validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "global-rule", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			missingIDFile := writeGlobalRuleFile(g, "global-rule-missing-id.json", `{
  "plugins": {
    "prometheus": {}
  }
}`)
			_, stderr, err = runA6WithEnv(env, "global-rule", "create", "-f", missingIDFile)
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring(`must include an "id" field`))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and surfaces authentication failures", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-global-rule-list-1"
			id2 := "ginkgo-global-rule-list-2"

			deleteGlobalRuleViaCLIByID(env, id1)
			deleteGlobalRuleViaCLIByID(env, id2)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, id1)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, id2)

			file1 := writeGlobalRuleFile(g, "global-rule-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "prometheus": {}
  }
}`, id1))
			createGlobalRuleViaCLIFile(g, env, file1)

			file2 := writeGlobalRuleFile(g, "global-rule-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "ip-restriction": {
      "whitelist": ["0.0.0.0/0"]
    }
  }
}`, id2))
			createGlobalRuleViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "global-rule", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))
			g.Expect(stdout).To(ContainSubstring("id: " + id2))

			badEnv := setupGlobalRuleEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "global-rule", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})

		It("exports global rules and strips timestamps from output", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-global-rule-export-1"
			id2 := "ginkgo-global-rule-export-2"

			deleteGlobalRuleViaCLIByID(env, id1)
			deleteGlobalRuleViaCLIByID(env, id2)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, id1)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, id2)

			file1 := writeGlobalRuleFile(g, "global-rule-export-1.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "prometheus": {}
  }
}`, id1))
			createGlobalRuleViaCLIFile(g, env, file1)

			file2 := writeGlobalRuleFile(g, "global-rule-export-2.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "ip-restriction": {
      "whitelist": ["0.0.0.0/0"]
    }
  }
}`, id2))
			createGlobalRuleViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "global-rule", "export", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "global-rule-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "global-rule", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
			g.Expect(string(exported)).NotTo(ContainSubstring("create_time"))
			g.Expect(string(exported)).NotTo(ContainSubstring("update_time"))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets global rules in yaml/json, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			ruleID := "ginkgo-global-rule-lifecycle"

			deleteGlobalRuleViaCLIByID(env, ruleID)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, ruleID)

			createFile := writeGlobalRuleFile(g, "global-rule-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "plugins": {
    "prometheus": {}
  }
}`, ruleID))
			createGlobalRuleViaCLIFile(g, env, createFile)

			stdout, stderr, err := runA6WithEnv(env, "global-rule", "get", ruleID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + ruleID))
			g.Expect(stdout).To(ContainSubstring("prometheus"))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(ruleID))

			updateFile := writeGlobalRuleFile(g, "global-rule-update.json", `{
  "plugins": {
    "prometheus": {},
    "ip-restriction": {
      "whitelist": ["0.0.0.0/0"]
    }
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "global-rule", "update", ruleID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ip-restriction"))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ip-restriction"))

			stdout, stderr, err = runA6WithEnv(env, "global-rule", "delete", ruleID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("deleted"))

			_, stderr, err = runA6WithEnv(env, "global-rule", "get", ruleID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			ruleID := "nonexistent-global-rule-999"

			deleteGlobalRuleViaCLIByID(env, ruleID)
			DeferCleanup(deleteGlobalRuleViaAdminByID, g, ruleID)

			_, stderr, err := runA6WithEnv(env, "global-rule", "get", ruleID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "global-rule", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "global-rule", "delete", ruleID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
