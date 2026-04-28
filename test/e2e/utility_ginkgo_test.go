//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("completion command", func() {
	It("generates shell completions for supported shells", func() {
		g := NewWithT(GinkgoT())

		stdout, stderr, err := runA6("completion", "bash")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("bash completion"))
		g.Expect(stdout).To(ContainSubstring("__start_a6"))

		stdout, stderr, err = runA6("completion", "zsh")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("compdef"))

		stdout, stderr, err = runA6("completion", "fish")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("complete -c a6"))

		stdout, stderr, err = runA6("completion", "powershell")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("Register-ArgumentCompleter"))
	})
})

var _ = Describe("version command", func() {
	It("renders text and JSON version output", func() {
		g := NewWithT(GinkgoT())

		stdout, stderr, err := runA6("version")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("a6 version"))
		g.Expect(stdout).To(ContainSubstring("commit:"))
		g.Expect(stdout).To(ContainSubstring("go:"))
		g.Expect(stdout).To(ContainSubstring("platform:"))

		stdout, stderr, err = runA6("version", "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)

		var info map[string]any
		g.Expect(json.Unmarshal([]byte(stdout), &info)).To(Succeed())
		g.Expect(info["version"]).NotTo(BeEmpty())
		g.Expect(info["goVersion"]).NotTo(BeEmpty())
		g.Expect(info["platform"]).NotTo(BeEmpty())
	})
})

var _ = Describe("extension command", func() {
	It("handles empty state, aliases, help, and validation without external services", func() {
		g := NewWithT(GinkgoT())
		env := []string{"A6_CONFIG_DIR=" + GinkgoT().TempDir()}

		stdout, stderr, err := runA6WithEnv(env, "extension", "list")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("No extensions installed."))

		stdout, stderr, err = runA6WithEnv(env, "extension", "upgrade", "--all")
		combined := stdout + stderr
		g.Expect(err == nil || strings.Contains(combined, "No extensions") || strings.Contains(combined, "no extensions")).To(BeTrue(),
			"stdout=%s stderr=%s err=%v", stdout, stderr, err)

		_, stderr, err = runA6WithEnv(env, "extension", "install", "bad-format")
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("owner/repo"))

		_, stderr, err = runA6WithEnv(env, "extension", "remove", "nonexistent", "--force")
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("not found"))

		stdout, stderr, err = runA6WithEnv(env, "extension", "--help")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("install"))
		g.Expect(stdout).To(ContainSubstring("list"))
		g.Expect(stdout).To(ContainSubstring("upgrade"))
		g.Expect(stdout).To(ContainSubstring("remove"))

		aliasStdout, aliasStderr, aliasErr := runA6WithEnv(env, "ext", "--help")
		g.Expect(aliasErr).NotTo(HaveOccurred(), "stderr=%s", aliasStderr)
		g.Expect(aliasStdout).To(Equal(stdout))
	})
})

var _ = Describe("context command", func() {
	It("creates, lists, switches, and deletes local contexts", func() {
		g := NewWithT(GinkgoT())
		env := []string{"A6_CONFIG_DIR=" + GinkgoT().TempDir()}

		stdout, stderr, err := runA6WithEnv(env, "context", "create", "local", "--server", "http://localhost:9180", "--api-key", "test123")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(stdout).To(ContainSubstring("created"))

		stdout, stderr, err = runA6WithEnv(env, "context", "current")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(strings.TrimSpace(stdout)).To(Equal("local"))

		stdout, stderr, err = runA6WithEnv(env, "context", "create", "staging", "--server", "http://staging:9180")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "context", "list")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring("local"))
		g.Expect(stdout).To(ContainSubstring("staging"))
		g.Expect(stdout).To(ContainSubstring("http://localhost:9180"))
		g.Expect(stdout).To(ContainSubstring("http://staging:9180"))

		stdout, stderr, err = runA6WithEnv(env, "context", "list", "--output", "json")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(json.Valid([]byte(stdout))).To(BeTrue(), stdout)

		stdout, stderr, err = runA6WithEnv(env, "context", "use", "staging")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "context", "current")
		g.Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		g.Expect(strings.TrimSpace(stdout)).To(Equal("staging"))

		stdout, stderr, err = runA6WithEnv(env, "context", "delete", "staging", "--force")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "context", "list")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		g.Expect(stdout).To(ContainSubstring("local"))
		g.Expect(stdout).NotTo(ContainSubstring("staging"))
	})

	It("surfaces duplicate and not-found errors", func() {
		g := NewWithT(GinkgoT())
		env := []string{"A6_CONFIG_DIR=" + GinkgoT().TempDir()}

		stdout, stderr, err := runA6WithEnv(env, "context", "create", "local", "--server", "http://localhost:9180")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		_, stderr, err = runA6WithEnv(env, "context", "create", "local", "--server", "http://localhost:9180")
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("already exists"))

		_, stderr, err = runA6WithEnv(env, "context", "use", "nonexistent")
		g.Expect(err).To(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("not found"))

		stdout, stderr, err = runA6WithEnv(env, "context", "delete", "local", "--force")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)

		stdout, stderr, err = runA6WithEnv(env, "context", "list")
		g.Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
		trimmed := strings.TrimSpace(stdout)
		g.Expect(trimmed == "" || strings.Contains(trimmed, "No contexts")).To(BeTrue(), "stdout=%s", stdout)
	})
})
