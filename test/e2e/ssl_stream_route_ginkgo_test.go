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

func deleteSSLViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/ssls", id)
}

func deleteStreamRouteViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/stream_routes", id)
}

func deleteSSLViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "ssl", "delete", id, "--force")
}

func deleteStreamRouteViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "stream-route", "delete", id, "--force")
}

func writeSSLFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writeStreamRouteFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func readGinkgoTestCert(g Gomega) (string, string) {
	modRoot, err := resolveModuleRoot()
	g.Expect(err).NotTo(HaveOccurred())

	certBytes, err := os.ReadFile(filepath.Join(modRoot, "test/e2e/testdata/test.crt"))
	g.Expect(err).NotTo(HaveOccurred())
	keyBytes, err := os.ReadFile(filepath.Join(modRoot, "test/e2e/testdata/test.key"))
	g.Expect(err).NotTo(HaveOccurred())
	return string(certBytes), string(keyBytes)
}

var _ = Describe("ssl command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates SSL certificates from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-ssl-json"
			yamlID := "ginkgo-ssl-yaml"
			cert, key := readGinkgoTestCert(g)

			deleteSSLViaCLIByID(env, jsonID)
			deleteSSLViaCLIByID(env, yamlID)
			DeferCleanup(deleteSSLViaAdminByID, g, jsonID)
			DeferCleanup(deleteSSLViaAdminByID, g, yamlID)

			jsonFile := writeSSLFile(g, "ssl.json", fmt.Sprintf(`{
  "id": "%s",
  "cert": %q,
  "key": %q,
  "snis": ["ginkgo-json.example.com"],
  "labels": {
    "suite": "ginkgo-ssl-create"
  }
}`, jsonID, cert, key))
			stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			yamlFile := writeSSLFile(g, "ssl.yaml", fmt.Sprintf(`id: %s
cert: |
%s
key: |
%s
snis:
  - ginkgo-yaml.example.com
labels:
  suite: ginkgo-ssl-create
`, yamlID, indentPEM(cert), indentPEM(key)))
			stdout, stderr, err = runA6WithEnv(env, "ssl", "create", "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-yaml.example.com"))

			stdout, stderr, err = runA6WithEnv(env, "ssl", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("ginkgo-json.example.com"))
		})

		It("preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "ssl", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and exports SSL certificates with label filtering", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-ssl-list-1"
			id2 := "ginkgo-ssl-list-2"
			cert, key := readGinkgoTestCert(g)

			deleteSSLViaCLIByID(env, id1)
			deleteSSLViaCLIByID(env, id2)
			DeferCleanup(deleteSSLViaAdminByID, g, id1)
			DeferCleanup(deleteSSLViaAdminByID, g, id2)

			file1 := writeSSLFile(g, "ssl-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "cert": %q,
  "key": %q,
  "snis": ["ginkgo-ssl-list-1.example.com"],
  "labels": {"suite":"ginkgo-ssl-list"}
}`, id1, cert, key))
			stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", file1)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeSSLFile(g, "ssl-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "cert": %q,
  "key": %q,
  "snis": ["ginkgo-ssl-list-2.example.com"],
  "labels": {"suite":"ginkgo-ssl-other"}
}`, id2, cert, key))
			stdout, stderr, err = runA6WithEnv(env, "ssl", "create", "-f", file2)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "ssl", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring("ginkgo-ssl-list-1.example.com"))
			g.Expect(stdout).To(ContainSubstring("ginkgo-ssl-list-2.example.com"))

			stdout, stderr, err = runA6WithEnv(env, "ssl", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "ssl", "export", "--label", "suite=ginkgo-ssl-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "ssl-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "ssl", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets SSL certificates, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			sslID := "ginkgo-ssl-lifecycle"
			cert, key := readGinkgoTestCert(g)

			deleteSSLViaCLIByID(env, sslID)
			DeferCleanup(deleteSSLViaAdminByID, g, sslID)

			createFile := writeSSLFile(g, "ssl-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "cert": %q,
  "key": %q,
  "snis": ["ginkgo-ssl-lifecycle.example.com"]
}`, sslID, cert, key))
			stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-ssl-lifecycle.example.com"))

			updateFile := writeSSLFile(g, "ssl-update.json", fmt.Sprintf(`{
  "cert": %q,
  "key": %q,
  "snis": ["ginkgo-ssl-lifecycle.example.com", "ginkgo-ssl-updated.example.com"]
}`, cert, key))
			stdout, stderr, err = runA6WithEnv(env, "ssl", "update", sslID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-ssl-updated.example.com"))

			stdout, stderr, err = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation", func() {
			g := NewWithT(GinkgoT())
			sslID := "missing-ginkgo-ssl"

			deleteSSLViaCLIByID(env, sslID)
			DeferCleanup(deleteSSLViaAdminByID, g, sslID)

			_, stderr, err := runA6WithEnv(env, "ssl", "get", sslID)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "ssl", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})

var _ = Describe("stream-route command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupCLIEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates stream routes from JSON and YAML files and verifies them through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonID := "ginkgo-stream-route-json"
			yamlID := "ginkgo-stream-route-yaml"

			deleteStreamRouteViaCLIByID(env, jsonID)
			deleteStreamRouteViaCLIByID(env, yamlID)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, jsonID)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, yamlID)

			jsonFile := writeStreamRouteFile(g, "stream-route.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-stream-route-json",
  "server_port": 19110,
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:19810": 1
    }
  },
  "labels": {
    "suite": "ginkgo-stream-route-create"
  }
			}`, jsonID))
			stdout, stderr, err := runA6WithEnv(env, "stream-route", "create", "-f", jsonFile)
			skipIfStreamModeDisabled(stdout, stderr, err)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(jsonID))

			yamlFile := writeStreamRouteFile(g, "stream-route.yaml", fmt.Sprintf(`id: %s
name: ginkgo-stream-route-yaml
server_port: 19111
upstream:
  type: roundrobin
  nodes:
    "127.0.0.1:19811": 1
labels:
  suite: ginkgo-stream-route-create
`, yamlID))
			stdout, stderr, err = runA6WithEnv(env, "stream-route", "create", "-f", yamlFile, "--output", "yaml")
			skipIfStreamModeDisabled(stdout, stderr, err)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-stream-route-yaml"))

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", jsonID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("ginkgo-stream-route-json"))
			g.Expect(stdout).To(ContainSubstring("19110"))
		})

		It("preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "stream-route", "create")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("list and export", func() {
		It("renders list output modes and exports stream routes with label filtering", func() {
			g := NewWithT(GinkgoT())
			id1 := "ginkgo-stream-route-list-1"
			id2 := "ginkgo-stream-route-list-2"

			deleteStreamRouteViaCLIByID(env, id1)
			deleteStreamRouteViaCLIByID(env, id2)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, id1)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, id2)

			file1 := writeStreamRouteFile(g, "stream-route-list-1.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-stream-route-list-1",
  "server_port": 19112,
  "upstream": {"type":"roundrobin","nodes":{"127.0.0.1:19812":1}},
  "labels": {"suite":"ginkgo-stream-route-list"}
}`, id1))
			stdout, stderr, err := runA6WithEnv(env, "stream-route", "create", "-f", file1)
			skipIfStreamModeDisabled(stdout, stderr, err)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			file2 := writeStreamRouteFile(g, "stream-route-list-2.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-stream-route-list-2",
  "server_port": 19113,
  "upstream": {"type":"roundrobin","nodes":{"127.0.0.1:19813":1}},
  "labels": {"suite":"ginkgo-stream-route-other"}
}`, id2))
			stdout, stderr, err = runA6WithEnv(env, "stream-route", "create", "-f", file2)
			skipIfStreamModeDisabled(stdout, stderr, err)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "list", "--output", "table")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ID"))
			g.Expect(stdout).To(ContainSubstring("ginkgo-stream-route-list-1"))
			g.Expect(stdout).To(ContainSubstring("ginkgo-stream-route-list-2"))

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).To(ContainSubstring(id2))

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "export", "--label", "suite=ginkgo-stream-route-list", "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring(id1))
			g.Expect(stdout).NotTo(ContainSubstring(id2))
			g.Expect(stdout).NotTo(ContainSubstring("create_time"))
			g.Expect(stdout).NotTo(ContainSubstring("update_time"))

			outFile := filepath.Join(GinkgoT().TempDir(), "stream-route-export.yaml")
			stdout, stderr, err = runA6WithEnv(env, "stream-route", "export", "-f", outFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			exported, readErr := os.ReadFile(outFile)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(exported)).To(ContainSubstring(id1))
			g.Expect(string(exported)).To(ContainSubstring(id2))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets stream routes, updates them, and deletes them through the CLI", func() {
			g := NewWithT(GinkgoT())
			streamRouteID := "ginkgo-stream-route-lifecycle"

			deleteStreamRouteViaCLIByID(env, streamRouteID)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, streamRouteID)

			createFile := writeStreamRouteFile(g, "stream-route-lifecycle.json", fmt.Sprintf(`{
  "id": "%s",
  "name": "ginkgo-stream-route-lifecycle",
  "server_port": 19114,
  "upstream": {"type":"roundrobin","nodes":{"127.0.0.1:19814":1}}
}`, streamRouteID))
			stdout, stderr, err := runA6WithEnv(env, "stream-route", "create", "-f", createFile)
			skipIfStreamModeDisabled(stdout, stderr, err)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("name: ginkgo-stream-route-lifecycle"))

			updateFile := writeStreamRouteFile(g, "stream-route-update.json", `{
  "name": "ginkgo-stream-route-updated",
  "server_port": 19115,
  "upstream": {"type":"roundrobin","nodes":{"127.0.0.1:19815":1}}
}`)
			stdout, stderr, err = runA6WithEnv(env, "stream-route", "update", streamRouteID, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "get", streamRouteID, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("ginkgo-stream-route-updated"))
			g.Expect(stdout).To(ContainSubstring("19115"))

			stdout, stderr, err = runA6WithEnv(env, "stream-route", "delete", streamRouteID, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(strings.ToLower(stdout + stderr)).To(ContainSubstring("deleted"))
		})

		It("surfaces get and delete not-found behavior and update required-flag validation", func() {
			g := NewWithT(GinkgoT())
			streamRouteID := "missing-ginkgo-stream-route"

			deleteStreamRouteViaCLIByID(env, streamRouteID)
			DeferCleanup(deleteStreamRouteViaAdminByID, g, streamRouteID)

			_, stderr, err := runA6WithEnv(env, "stream-route", "get", streamRouteID)
			skipIfStreamModeDisabled("", stderr, err)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "stream-route", "update")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "stream-route", "delete", streamRouteID, "--force")
			skipIfStreamModeDisabled("", stderr, err)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})

func indentPEM(value string) string {
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return strings.Join(lines, "\n")
}

func skipIfStreamModeDisabled(stdout, stderr string, err error) {
	if err == nil {
		return
	}
	if strings.Contains(stdout+stderr, "stream mode is disabled") {
		Skip("APISIX stream mode is disabled in this environment")
	}
}
