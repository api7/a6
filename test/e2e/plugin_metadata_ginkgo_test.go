//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func setupPluginMetadataEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create plugin-metadata test context: %s", stderr)
	return env
}

func writePluginMetadataFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func deletePluginMetadataViaAdminByName(g Gomega, pluginName string) {
	resp, err := adminAPI("DELETE", "/apisix/admin/plugin_metadata/"+pluginName, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

var _ = Describe("plugin-metadata command", Ordered, func() {
	var env []string

	BeforeEach(func() {
		env = setupPluginMetadataEnvWithKey(NewWithT(GinkgoT()), adminKey)
	})

	Describe("create", func() {
		It("creates plugin metadata from JSON and YAML files and verifies it through the CLI", func() {
			g := NewWithT(GinkgoT())
			jsonPlugin := "syslog"
			yamlPlugin := "http-logger"

			_, _, _ = runA6WithEnv(env, "plugin-metadata", "delete", jsonPlugin, "--force")
			_, _, _ = runA6WithEnv(env, "plugin-metadata", "delete", yamlPlugin, "--force")
			DeferCleanup(deletePluginMetadataViaAdminByName, g, jsonPlugin)
			DeferCleanup(deletePluginMetadataViaAdminByName, g, yamlPlugin)

			jsonFile := writePluginMetadataFile(g, "plugin-metadata.json", `{
  "log_format": {
    "host": "$host"
  }
}`)
			stdout, stderr, err := runA6WithEnv(env, "plugin-metadata", "create", jsonPlugin, "-f", jsonFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("log_format"))

			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", jsonPlugin, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("log_format"))

			yamlFile := writePluginMetadataFile(g, "plugin-metadata.yaml", `log_format:
  host: $host
`)
			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "create", yamlPlugin, "-f", yamlFile, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("log_format"))
		})

		It("preserves required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())

			_, stderr, err := runA6WithEnv(env, "plugin-metadata", "create", "syslog")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))
		})
	})

	Describe("get, update, and delete", func() {
		It("gets plugin metadata in yaml/json, updates it, and deletes it through the CLI", func() {
			g := NewWithT(GinkgoT())
			pluginName := "syslog"

			_, _, _ = runA6WithEnv(env, "plugin-metadata", "delete", pluginName, "--force")
			DeferCleanup(deletePluginMetadataViaAdminByName, g, pluginName)

			createFile := writePluginMetadataFile(g, "plugin-metadata-create.json", `{
  "log_format": {
    "host": "$host"
  }
}`)
			stdout, stderr, err := runA6WithEnv(env, "plugin-metadata", "create", pluginName, "-f", createFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName, "--output", "yaml")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("host: $host"))

			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)
			g.Expect(stdout).To(ContainSubstring("log_format"))

			updateFile := writePluginMetadataFile(g, "plugin-metadata-update.json", `{
  "log_format": {
    "host": "$host",
    "request_id": "$request_id"
  }
}`)
			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "update", pluginName, "-f", updateFile)
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("request_id"))

			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName, "--output", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("request_id"))

			stdout, stderr, err = runA6WithEnv(env, "plugin-metadata", "delete", pluginName, "--force")
			g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
			g.Expect(stdout).To(ContainSubstring("Deleted plugin metadata"))

			_, stderr, err = runA6WithEnv(env, "plugin-metadata", "get", pluginName)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})

		It("surfaces get not-found behavior and update required-flag validation through the real CLI", func() {
			g := NewWithT(GinkgoT())
			pluginName := "nonexistent-plugin-metadata-999"

			_, stderr, err := runA6WithEnv(env, "plugin-metadata", "get", pluginName)
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))

			_, stderr, err = runA6WithEnv(env, "plugin-metadata", "update", "syslog")
			g.Expect(err).To(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("required flag"))

			_, stderr, err = runA6WithEnv(env, "plugin-metadata", "delete", pluginName, "--force")
			g.Expect(err).To(HaveOccurred())
			g.Expect(strings.ToLower(stderr)).To(ContainSubstring("not found"))
		})
	})
})
