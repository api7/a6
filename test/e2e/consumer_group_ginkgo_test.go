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

func setupConsumerGroupEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create consumer-group test context: %s", stderr)
	return env
}

func writeConsumerGroupFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deleteConsumerGroupViaAdminByID(g Gomega, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/consumer_groups/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteConsumerGroupViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "consumer-group", "delete", id, "--force")
}

func createConsumerGroupViaCLIFile(g Gomega, env []string, file string) {
	stdout, stderr, err := runA6WithEnv(env, "consumer-group", "create", "-f", file)
	g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
}

var _ = Describe("consumer-group command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupConsumerGroupEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates consumer groups from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-consumer-group-json"
			yamlID := "ginkgo-consumer-group-yaml"

			deleteConsumerGroupViaCLIByID(env, jsonID)
			deleteConsumerGroupViaCLIByID(env, yamlID)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, jsonID)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, yamlID)

			jsonFile := writeConsumerGroupFile(g, "consumer-group.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-json",
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, jsonID))
			stdout, stderr, err := runA6WithEnv(env, "consumer-group", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-consumer-group-json"))

			yamlFile := writeConsumerGroupFile(g, "consumer-group.yaml", fmt.Sprintf(`id: %s
name: ginkgo-consumer-group-yaml
plugins:
  limit-count:
    count: 100
    time_window: 60
    rejected_code: 503
    key_type: var
    key: remote_addr
`, yamlID))
			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-consumer-group-yaml"))
		})

		It("preserves required-flag and missing-id validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "consumer-group", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			missingIDFile := writeConsumerGroupFile(g, "consumer-group-missing-id.json", `{
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`)
			_, stderr, err = runA6WithEnv(env, "consumer-group", "create", "-f", missingIDFile)
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring(`must include an "id" field`))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and surfaces authentication failures", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-consumer-group-list-1"
			id2 := "ginkgo-consumer-group-list-2"

			deleteConsumerGroupViaCLIByID(env, id1)
			deleteConsumerGroupViaCLIByID(env, id2)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, id1)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, id2)

			file1 := writeConsumerGroupFile(g, "consumer-group-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-alpha",
  "labels": {"tier":"premium"},
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, id1))
			createConsumerGroupViaCLIFile(g, env, file1)

			file2 := writeConsumerGroupFile(g, "consumer-group-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-beta",
  "labels": {"tier":"standard"},
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    },
    "prometheus": {}
  }
}`, id2))
			createConsumerGroupViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "consumer-group", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "list", "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("id: " + id1))
			g.Expect(stdout).To(ContainSubstring("id: " + id2))

			badEnv := setupConsumerGroupEnvWithKey(g, "invalid-api-key")
			_, stderr, err = runA6WithEnv(badEnv, "consumer-group", "list")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
				ContainSubstring("authentication failed"),
				ContainSubstring("permission denied"),
			))
		})

		It("exports consumer groups with real label filtering and strips timestamps from output", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-consumer-group-export-1"
			id2 := "ginkgo-consumer-group-export-2"

			deleteConsumerGroupViaCLIByID(env, id1)
			deleteConsumerGroupViaCLIByID(env, id2)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, id1)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, id2)

			file1 := writeConsumerGroupFile(g, "consumer-group-export-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-export-1",
  "labels": {"tier":"premium"},
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, id1))
			createConsumerGroupViaCLIFile(g, env, file1)

			file2 := writeConsumerGroupFile(g, "consumer-group-export-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-export-2",
  "labels": {"tier":"standard"},
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, id2))
			createConsumerGroupViaCLIFile(g, env, file2)

			stdout, stderr, err := runA6WithEnv(env, "consumer-group", "export", "--label", "tier=premium", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "consumer-group-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "export", "-f", outFile, "--output", "yaml")
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
		It("gets consumer groups in yaml/json, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			groupID := "ginkgo-consumer-group-lifecycle"

			deleteConsumerGroupViaCLIByID(env, groupID)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, groupID)

			createFile := writeConsumerGroupFile(g, "consumer-group-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-consumer-group-lifecycle",
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  }
}`, groupID))
			createConsumerGroupViaCLIFile(g, env, createFile)

			stdout, stderr, err := runA6WithEnv(env, "consumer-group", "get", groupID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-consumer-group-lifecycle"))
			g.Expect(stdout).To(ContainSubstring("limit-count"))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(groupID))

			updateFile := writeConsumerGroupFile(g, "consumer-group-update.json", `{
  "plugins": {
    "limit-count": {
      "count": 300,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    },
    "prometheus": {}
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "update", groupID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("prometheus"))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("prometheus"))

			stdout, stderr, err = runA6WithEnv(env, "consumer-group", "delete", groupID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("deleted"))

			_, stderr, err = runA6WithEnv(env, "consumer-group", "get", groupID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			groupID := "nonexistent-consumer-group-999"

			deleteConsumerGroupViaCLIByID(env, groupID)
			DeferCleanup(deleteConsumerGroupViaAdminByID, g, groupID)

			_, stderr, err := runA6WithEnv(env, "consumer-group", "get", groupID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "consumer-group", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "consumer-group", "delete", groupID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
