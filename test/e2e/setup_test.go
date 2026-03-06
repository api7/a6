//go:build e2e

// Package e2e provides end-to-end tests for the a6 CLI.
// These tests run against a real APISIX instance and require the following
// environment to be available:
//
//   - APISIX Admin API at APISIX_ADMIN_URL (default: http://127.0.0.1:9180)
//   - APISIX Gateway at APISIX_GATEWAY_URL (default: http://127.0.0.1:9080)
//   - httpbin at HTTPBIN_URL (default: http://127.0.0.1:8080)
//
// Run with: go test -v -tags e2e -count=1 -timeout 5m ./test/e2e/...
package e2e

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var (
	binaryPath string
	adminURL   string
	gatewayURL string
	httpbinURL string
	adminKey   = "edd1c9f034335f136f87ad84b625c8f1"
)

func TestMain(m *testing.M) {
	adminURL = envOrDefault("APISIX_ADMIN_URL", "http://127.0.0.1:9180")
	gatewayURL = envOrDefault("APISIX_GATEWAY_URL", "http://127.0.0.1:9080")
	httpbinURL = envOrDefault("HTTPBIN_URL", "http://127.0.0.1:8080")

	// Build the a6 binary into a temp directory.
	tmpDir, err := os.MkdirTemp("", "a6-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "a6")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/a6")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build a6 binary: %v\n", err)
		os.Exit(1)
	}

	// Wait for APISIX Admin API to become healthy.
	if err := waitForHealthy(adminURL+"/apisix/admin/routes", 60*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "APISIX not ready: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runA6 executes the a6 binary with the given arguments and returns
// captured stdout, stderr, and any error.
func runA6(args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// adminAPI sends an HTTP request to the APISIX Admin API.
// Used for test setup and cleanup — not for testing the CLI itself.
func adminAPI(method, path string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, adminURL+path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, adminURL+path, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", adminKey)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

// waitForHealthy polls the given URL until it returns a successful response
// or the timeout is reached.
func waitForHealthy(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		req.Header.Set("X-API-KEY", adminKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 400 {
			return nil
		}
		lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for %s: %v", url, lastErr)
}

// envOrDefault returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
