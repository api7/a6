//go:build e2e

package e2e

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func setupCLIEnvWithKey(g Gomega, apiKey string) []string {
	env := []string{
		"A6_CONFIG_DIR=" + GinkgoT().TempDir(),
	}
	_, stderr, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", apiKey)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create test context: %s", stderr)
	return env
}

func writeTempFile(g Gomega, name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	g.Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

func hostPortFromURL(g Gomega, rawURL string) string {
	parsed, err := url.Parse(rawURL)
	g.Expect(err).NotTo(HaveOccurred())
	return parsed.Host
}

func deleteResourceViaAdminByID(g Gomega, resourcePath, id string) {
	resp, err := adminAPI("DELETE", resourcePath+"/"+id, nil)
	if err == nil && resp != nil {
		g.Expect(resp.Body.Close()).To(Succeed())
	}
}

func deleteRouteViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/routes", id)
}

func deleteServiceViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/services", id)
}

func deleteUpstreamViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/upstreams", id)
}

func deletePluginConfigViaAdminByID(g Gomega, id string) {
	deleteResourceViaAdminByID(g, "/apisix/admin/plugin_configs", id)
}

func deleteRouteViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "route", "delete", id, "--force")
}

func deleteServiceViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "service", "delete", id, "--force")
}

func deleteUpstreamViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "upstream", "delete", id, "--force")
}

func deletePluginConfigViaCLIByID(env []string, id string) {
	_, _, _ = runA6WithEnv(env, "plugin-config", "delete", id, "--force")
}

func setupRouteEnvWithKey(g Gomega, apiKey string) []string {
	return setupCLIEnvWithKey(g, apiKey)
}

func setupServiceEnvWithKey(g Gomega, apiKey string) []string {
	return setupCLIEnvWithKey(g, apiKey)
}

func setupUpstreamEnvWithKey(g Gomega, apiKey string) []string {
	return setupCLIEnvWithKey(g, apiKey)
}

func writeRouteFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writeServiceFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writeUpstreamFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func writePluginConfigFile(g Gomega, name, body string) string {
	return writeTempFile(g, name, body)
}

func skipIfLicenseRestrictedGomega(stdout, stderr string, err error) {
	if err == nil {
		return
	}
	combined := stdout + stderr
	if strings.Contains(combined, "requires a sufficient license") {
		Skip("environment blocks scenario coverage with a license gate: " + strings.TrimSpace(combined))
	}
}
