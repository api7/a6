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

const ginkgoProtoContent = "syntax = \"proto3\";\npackage ginkgo;\nmessage Ping { string message = 1; }"

func deleteProtoViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/protos", id)
}

func deleteSecretViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/secrets", id)
}

func deleteProtoViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "proto", "delete", id, "--force")
}

func deleteSecretViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "secret", "delete", id, "--force")
}

func writeProtoFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writeSecretFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

var _ = Describe("proto command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates proto definitions from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-proto-json"
			yamlID := "ginkgo-proto-yaml"

			deleteProtoViaCLIByID(env, jsonID)
			deleteProtoViaCLIByID(env, yamlID)
			DeferCleanup(deleteProtoViaAdminByID, g, jsonID)
			DeferCleanup(deleteProtoViaAdminByID, g, yamlID)

			jsonFile := writeProtoFile(g, "proto.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-proto-json",
  "desc": "ginkgo proto json",
  "content": %q,
  "labels": {
    "suite": "ginkgo-proto-create"
  }
}`, jsonID, ginkgoProtoContent))
			stdout, stderr, err := runA6WithEnv(env, "proto", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			yamlFile := writeProtoFile(g, "proto.yaml", fmt.Sprintf(`id: %s
name: ginkgo-proto-yaml
desc: ginkgo proto yaml
content: |
  syntax = "proto3";
  package ginkgo;
  message Pong { string message = 1; }
labels:
  suite: ginkgo-proto-create
`, yamlID))
			stdout, stderr, err = runA6WithEnv(env, "proto", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-proto-yaml"))

			stdout, stderr, err = runA6WithEnv(env, "proto", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("ginkgo-proto-json"))
			g.Expect(stdout).To(ContainSubstring("message Ping"))
		})

		It("preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "proto", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and exports protos with label filtering", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-proto-list-1"
			id2 := "ginkgo-proto-list-2"

			deleteProtoViaCLIByID(env, id1)
			deleteProtoViaCLIByID(env, id2)
			DeferCleanup(deleteProtoViaAdminByID, g, id1)
			DeferCleanup(deleteProtoViaAdminByID, g, id2)

			file1 := writeProtoFile(g, "proto-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-proto-list-1",
  "desc": "ginkgo proto list 1",
  "content": %q,
  "labels": {"suite":"ginkgo-proto-list"}
}`, id1, ginkgoProtoContent))
			stdout, stderr, err := runA6WithEnv(env, "proto", "create", "-f", file1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeProtoFile(g, "proto-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-proto-list-2",
  "desc": "ginkgo proto list 2",
  "content": %q,
  "labels": {"suite":"ginkgo-proto-other"}
}`, id2, ginkgoProtoContent))
			stdout, stderr, err = runA6WithEnv(env, "proto", "create", "-f", file2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "proto", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "proto", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "proto", "export", "--label", "suite=ginkgo-proto-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "proto-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "proto", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets protos, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			protoID := "ginkgo-proto-lifecycle"

			deleteProtoViaCLIByID(env, protoID)
			DeferCleanup(deleteProtoViaAdminByID, g, protoID)

			createFile := writeProtoFile(g, "proto-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-proto-lifecycle",
  "desc": "ginkgo proto lifecycle",
  "content": %q
}`, protoID, ginkgoProtoContent))
			stdout, stderr, err := runA6WithEnv(env, "proto", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "proto", "get", protoID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-proto-lifecycle"))

			updateFile := writeProtoFile(g, "proto-update.json", fmt.Sprintf(`{
  "name": "ginkgo-proto-updated",
  "desc": "ginkgo proto updated",
  "content": %q
}`, "syntax = \"proto3\";\npackage ginkgo;\nmessage Updated { string message = 1; }"))
			stdout, stderr, err = runA6WithEnv(env, "proto", "update", protoID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "proto", "get", protoID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-proto-updated"))
			g.Expect(stdout).To(ContainSubstring("message Updated"))

			stdout, stderr, err = runA6WithEnv(env, "proto", "delete", protoID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation", func() {
			g := NewWithT(GinkgoT())
			protoID := "missing-ginkgo-proto"

			deleteProtoViaCLIByID(env, protoID)
			DeferCleanup(deleteProtoViaAdminByID, g, protoID)

			_, stderr, err := runA6WithEnv(env, "proto", "get", protoID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "proto", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "proto", "delete", protoID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})

var _ = Describe("secret command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates secrets from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "vault/ginkgo-secret-json"
			yamlID := "vault/ginkgo-secret-yaml"

			deleteSecretViaCLIByID(env, jsonID)
			deleteSecretViaCLIByID(env, yamlID)
			DeferCleanup(deleteSecretViaAdminByID, g, jsonID)
			DeferCleanup(deleteSecretViaAdminByID, g, yamlID)

			jsonFile := writeSecretFile(g, "secret.json", `{
  "uri": "http://127.0.0.1:8200",
  "prefix": "/apisix/kv/ginkgo-json",
  "token": "ginkgo-secret-json-token"
}`)
			stdout, stderr, err := runA6WithEnv(env, "secret", "create", jsonID, "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-secret-json"))

			yamlFile := writeSecretFile(g, "secret.yaml", `uri: http://127.0.0.1:8200
prefix: /apisix/kv/ginkgo-yaml
token: ginkgo-secret-yaml-token
`)
			stdout, stderr, err = runA6WithEnv(env, "secret", "create", yamlID, "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-secret-yaml"))

			stdout, stderr, err = runA6WithEnv(env, "secret", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("/apisix/kv/ginkgo-json"))
		})

		It("preserves argument and required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "secret", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(SatisfyAny(ContainSubstring("required flag"), ContainSubstring("accepts")))
		})
	})

	Describe("list, get, update, and delete", func() {
		It("lists secrets, updates them, and deletes them through CLI reads", func() {
			g := NewWithT(GinkgoT())
			secretID := "vault/ginkgo-secret-lifecycle"

			deleteSecretViaCLIByID(env, secretID)
			DeferCleanup(deleteSecretViaAdminByID, g, secretID)

			createFile := writeSecretFile(g, "secret-lifecycle.json", `{
  "uri": "http://127.0.0.1:8200",
  "prefix": "/apisix/kv/ginkgo-lifecycle",
  "token": "ginkgo-secret-lifecycle-token"
}`)
			stdout, stderr, err := runA6WithEnv(env, "secret", "create", secretID, "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "secret", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring("vault/ginkgo-secret-lifecycle"))

			stdout, stderr, err = runA6WithEnv(env, "secret", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("ginkgo-secret-lifecycle"))

			updateFile := writeSecretFile(g, "secret-update.json", `{
  "prefix": "/apisix/kv/ginkgo-updated",
  "token": "ginkgo-secret-updated-token"
}`)
			stdout, stderr, err = runA6WithEnv(env, "secret", "update", secretID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "secret", "get", secretID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("/apisix/kv/ginkgo-updated"))

			stdout, stderr, err = runA6WithEnv(env, "secret", "delete", secretID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation", func() {
			g := NewWithT(GinkgoT())
			secretID := "vault/missing-ginkgo-secret"

			deleteSecretViaCLIByID(env, secretID)
			DeferCleanup(deleteSecretViaAdminByID, g, secretID)

			_, stderr, err := runA6WithEnv(env, "secret", "get", secretID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "secret", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "secret", "delete", secretID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
