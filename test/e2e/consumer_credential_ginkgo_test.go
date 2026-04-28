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

func deleteConsumerViaAdminByUsername(g Gomega, username string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/consumers", username)
}

func deleteCredentialViaAdminByID(g Gomega, consumer, id string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/consumers/"+consumer+"/credentials/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteConsumerViaCLIByUsername(env []string, username string) {
	_, _, _ = runA6WithEnv(env, "consumer", "delete", username, "--force")
}

func deleteCredentialViaCLIByID(env []string, consumer, id string) {
	_, _, _ = runA6WithEnv(env, "credential", "delete", id, "--consumer", consumer, "--force")
}

func writeConsumerFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writeCredentialFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

var _ = Describe("consumer command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates consumers from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonUsername := "ginkgo-consumer-json"
			yamlUsername := "ginkgo-consumer-yaml"

			deleteConsumerViaCLIByUsername(env, jsonUsername)
			deleteConsumerViaCLIByUsername(env, yamlUsername)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, jsonUsername)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, yamlUsername)

			jsonFile := writeConsumerFile(g, "consumer.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo consumer json",
  "plugins": {
    "key-auth": {
      "key": "ginkgo-consumer-json-key"
    }
  },
  "labels": {
    "suite": "ginkgo-consumer-create"
  }
}`, jsonUsername))
			stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonUsername))

			yamlFile := writeConsumerFile(g, "consumer.yaml", fmt.Sprintf(`username: %s
desc: ginkgo consumer yaml
plugins:
  key-auth:
    key: ginkgo-consumer-yaml-key
labels:
  suite: ginkgo-consumer-create
`, yamlUsername))
			stdout, stderr, err = runA6WithEnv(env, "consumer", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("username: " + yamlUsername))

			stdout, stderr, err = runA6WithEnv(env, "consumer", "get", jsonUsername, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("ginkgo consumer json"))
			g.Expect(stdout).To(ContainSubstring("key-auth"))
		})

		It("preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "consumer", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and exports consumers with label filtering", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-consumer-list-1"
			id2 := "ginkgo-consumer-list-2"

			deleteConsumerViaCLIByUsername(env, id1)
			deleteConsumerViaCLIByUsername(env, id2)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, id1)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, id2)

			file1 := writeConsumerFile(g, "consumer-list-1.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo consumer list 1",
  "labels": {"suite":"ginkgo-consumer-list"}
}`, id1))
			stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", file1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeConsumerFile(g, "consumer-list-2.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo consumer list 2",
  "labels": {"suite":"ginkgo-consumer-other"}
}`, id2))
			stdout, stderr, err = runA6WithEnv(env, "consumer", "create", "-f", file2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "consumer", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("USERNAME"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "consumer", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "consumer", "export", "--label", "suite=ginkgo-consumer-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "consumer-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "consumer", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets consumers, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			username := "ginkgo-consumer-lifecycle"

			deleteConsumerViaCLIByUsername(env, username)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, username)

			createFile := writeConsumerFile(g, "consumer-lifecycle.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo consumer lifecycle",
  "plugins": {
    "key-auth": {
      "key": "ginkgo-consumer-lifecycle-key"
    }
  }
}`, username))
			stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "consumer", "get", username, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("desc: ginkgo consumer lifecycle"))

			updateFile := writeConsumerFile(g, "consumer-update.json", `{
  "desc": "ginkgo consumer updated",
  "plugins": {
    "key-auth": {
      "key": "ginkgo-consumer-updated-key"
    }
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "consumer", "update", username, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "consumer", "get", username, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo consumer updated"))
			g.Expect(stdout).To(ContainSubstring("ginkgo-consumer-updated-key"))

			stdout, stderr, err = runA6WithEnv(env, "consumer", "delete", username, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation", func() {
			g := NewWithT(GinkgoT())
			username := "nonexistent-ginkgo-consumer"

			deleteConsumerViaCLIByUsername(env, username)
			DeferCleanup(deleteConsumerViaAdminByUsername, g, username)

			_, stderr, err := runA6WithEnv(env, "consumer", "get", username)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "consumer", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "consumer", "delete", username, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})

var _ = Describe("credential command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	It("creates, lists, updates, and deletes credentials through CLI reads", func() {
		g := NewWithT(GinkgoT())
		const (
			consumer = "ginkgo-credential-consumer"
			credID   = "ginkgo-credential-key-auth"
		)

		deleteCredentialViaCLIByID(env, consumer, credID)
		deleteConsumerViaCLIByUsername(env, consumer)
		DeferCleanup(deleteCredentialViaAdminByID, g, consumer, credID)
		DeferCleanup(deleteConsumerViaAdminByUsername, g, consumer)

		consumerFile := writeConsumerFile(g, "credential-consumer.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo credential consumer"
}`, consumer))
		stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		createFile := writeCredentialFile(g, "credential-create.json", fmt.Sprintf(`{
  "id": "%s",
  "desc": "ginkgo credential",
  "plugins": {
    "key-auth": {
      "key": "ginkgo-credential-key"
    }
  }
}`, credID))
		stdout, stderr, err = runA6WithEnv(env, "credential", "create", "--consumer", consumer, "-f", createFile)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring(credID))

		stdout, stderr, err = runA6WithEnv(env, "credential", "get", credID, "--consumer", consumer, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
		g.Expect(stdout).To(ContainSubstring("ginkgo credential"))
		g.Expect(stdout).To(ContainSubstring("key-auth"))

		stdout, stderr, err = runA6WithEnv(env, "credential", "list", "--consumer", consumer, "--output", "table")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring("ID"))
		g.Expect(stdout).To(ContainSubstring(credID))
		g.Expect(stdout).To(ContainSubstring("key-auth"))

		updateFile := writeCredentialFile(g, "credential-update.json", `{
  "desc": "ginkgo credential updated",
  "plugins": {
    "key-auth": {
      "key": "ginkgo-credential-updated-key"
    }
  }
}`)
		stdout, stderr, err = runA6WithEnv(env, "credential", "update", credID, "--consumer", consumer, "-f", updateFile)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "credential", "get", credID, "--consumer", consumer, "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring("ginkgo credential updated"))
		g.Expect(stdout).To(ContainSubstring("ginkgo-credential-updated-key"))

		stdout, stderr, err = runA6WithEnv(env, "credential", "delete", credID, "--consumer", consumer, "--force")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
	})

	It("surfaces missing credential data and required flags through the real CLI", func() {
		g := NewWithT(GinkgoT())
		const consumer = "ginkgo-credential-validation-consumer"

		deleteConsumerViaCLIByUsername(env, consumer)
		DeferCleanup(deleteConsumerViaAdminByUsername, g, consumer)

		consumerFile := writeConsumerFile(g, "credential-validation-consumer.json", fmt.Sprintf(`{
  "username": "%s",
  "desc": "ginkgo credential validation consumer"
}`, consumer))
		stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		_, stderr, err = runA6WithEnv(env, "credential", "create", "--consumer", consumer)
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("required flag"))

		missingIDFile := writeCredentialFile(g, "credential-missing-id.json", `{
  "plugins": {
    "key-auth": {
      "key": "ginkgo-credential-missing-id"
    }
  }
}`)
		_, stderr, err = runA6WithEnv(env, "credential", "create", "--consumer", consumer, "-f", missingIDFile)
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("id"))

		_, stderr, err = runA6WithEnv(env, "credential", "get", "missing-ginkgo-credential", "--consumer", consumer)
		g.Expect(err).To(HaveOccurred())
		g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
	})
})
