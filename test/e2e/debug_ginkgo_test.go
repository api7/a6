//go:build e2e

package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("debug logs command", func() {
	It("reads APISIX container logs when Docker access is available", func() {
		g := NewWithT(GinkgoT())
		env := setupCLIEnvWithKey(g, adminKey)

		stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "10")
		skipIfDockerLogsUnavailable(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(strings.TrimSpace(stdout)).NotTo(BeEmpty())

		stdout, stderr, err = runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "5", "--since", "1h")
		skipIfDockerLogsUnavailable(stdout, stderr, err)
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
	})

	It("auto-detects APISIX containers when available", func() {
		g := NewWithT(GinkgoT())
		env := setupCLIEnvWithKey(g, adminKey)

		stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--tail", "10")
		skipIfDockerLogsUnavailable(stdout, stderr, err)
		if err != nil && strings.Contains(stderr, "no APISIX container found") {
			Skip("no auto-detectable APISIX container is running")
		}
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(strings.TrimSpace(stdout)).NotTo(BeEmpty())
	})

	It("surfaces missing container errors", func() {
		g := NewWithT(GinkgoT())
		env := setupCLIEnvWithKey(g, adminKey)

		_, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "non-existent-container-xyz", "--tail", "5")
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(SatisfyAny(
			ContainSubstring("No such container"),
			ContainSubstring("not found"),
		))
	})
})

var _ = Describe("debug trace command", func() {
	It("surfaces not-found errors without proxy traffic", func() {
		g := NewWithT(GinkgoT())
		env := setupCLIEnvWithKey(g, adminKey)

		_, stderr, err := runA6WithEnv(env, "debug", "trace", "non-existent-debug-route")
		g.Expect(err).To(HaveOccurred())
		g.Expect(strings.ToLower(stderr)).To(SatisfyAny(
			ContainSubstring("resource not found"),
			ContainSubstring("not found"),
			ContainSubstring("404"),
		))
	})
})

func skipIfDockerLogsUnavailable(stdout, stderr string, err error) {
	if err == nil {
		return
	}

	combined := strings.ToLower(stdout + stderr)
	if strings.Contains(combined, "cannot connect to the docker daemon") ||
		strings.Contains(combined, "permission denied") ||
		strings.Contains(combined, "no such container") ||
		strings.Contains(combined, "no apisix container found") {
		Skip("Docker logs are unavailable in this environment: " + strings.TrimSpace(stdout+stderr))
	}
}
