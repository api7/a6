# Testing Strategy

## Test Requirements
- Every code change must add or update tests.
- Command behavior that depends on the APISIX Admin API must be covered by real e2e tests against a running APISIX instance.
- Unit tests are reserved for self-contained logic that does not require a mocked external environment.
- Treat pure unit coverage and e2e scenario coverage as separate signals. Do not use mocked Admin API command tests to inflate package coverage.

## Test File Location
Pure unit tests should be located in the same directory as the code they test. For example, `selector.go` should have its tests in `selector_test.go`.

Real APISIX e2e tests live in `test/e2e/`.

## Test Naming Convention
Follow the pattern `func Test<Function>_<Scenario>(t *testing.T) {}`.

Examples:
- `func TestRouteList_ReturnsTable(t *testing.T) {}`
- `func TestRouteList_EmptyResponse(t *testing.T) {}`
- `func TestRouteList_APIError(t *testing.T) {}`
- `func TestRouteList_JSONOutput(t *testing.T) {}`
- `func TestRouteList_NonTTY(t *testing.T) {}`

## Test Categories

### Unit Tests
Use unit tests only for logic that is fully local and deterministic:
- pure data transformations
- output rendering helpers
- selectors, parsers, and normalization helpers
- command-independent utility code

Do not add new unit tests that mock the APISIX Admin API just to verify CLI behavior.

### TTY vs Non-TTY Tests
TTY and non-TTY behavior should be covered where it belongs:
- pure output-formatting helpers can be tested with unit tests
- command-level output mode behavior should be covered by e2e specs against the real CLI and a real APISIX instance

## What NOT to Test
- Do not test cobra flag binding, as this is handled by the cobra framework.
- Do not test JSON marshaling, which is the responsibility of the standard library.
- Do not add command tests that replace the Admin API with `httpmock`, fake `RoundTripper`s, or stubbed responses.
- Avoid writing integration tests against a real APISIX instance in unit test files — use `test/e2e` for that.

## E2E Tests

E2E tests validate the CLI binary against a real APISIX instance. They live in `test/e2e/`, use the `//go:build e2e` build tag, and should be written with Ginkgo/Gomega.

### Architecture

The e2e framework:
1. Builds the `a6` binary once per test process
2. Waits for APISIX Admin API to become healthy
3. Runs specs that invoke the binary via `exec.Command`
4. Uses direct Admin API calls only for setup and cleanup
5. Organizes scenarios with a suite structure inspired by Kubernetes-style e2e suites

### Infrastructure

Three services are required:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| etcd | `bitnamilegacy/etcd:3.6` | `2379` | APISIX config store |
| APISIX | `apache/apisix:3.15.0-debian` | `9180` (Admin), `9080` (Gateway) | Target instance |
| httpbin | `ghcr.io/mccutchen/go-httpbin` | `8080` | Upstream target |

### Running E2E Tests

**Locally** (requires Docker):
```bash
make docker-up      # Start etcd + APISIX + httpbin via docker-compose
make test-e2e       # Run e2e tests
make docker-down    # Tear down
```

**In CI**: The `.github/workflows/e2e.yml` workflow handles this automatically using GitHub Actions service containers for etcd and `docker run` for APISIX (which needs a volume-mounted config file).

### E2E Test File Structure

- `test/e2e/suite_test.go` — Ginkgo suite bootstrap
- `test/e2e/setup_test.go` — binary build and shared helper functions
- `test/e2e/<resource>_ginkgo_test.go` — Ginkgo specs for resource behavior
- `test/e2e/<resource>_test.go` — legacy `testing`-style e2e files that should be migrated over time
- `test/e2e/apisix_conf/config.yaml` — APISIX config for CI (etcd at `127.0.0.1`)
- `test/e2e/apisix_conf/config-docker.yaml` — APISIX config for docker-compose (etcd at `etcd`)

### Writing E2E Tests

```go
//go:build e2e

package e2e

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("route command", func() {
    It("creates and lists routes against the real Admin API", func() {
        env := setupRouteEnvWithKey(NewWithT(GinkgoT()), adminKey)

        stdout, stderr, err := runA6WithEnv(env, "route", "list", "--output", "json")
        Expect(err).NotTo(HaveOccurred(), "stdout=%s stderr=%s", stdout, stderr)
        Expect(stdout).To(ContainSubstring("["))
    })
})
```

Prefer Kubernetes-style structure:
- top-level `Describe` grouped by command or resource
- nested `Context` blocks for modes or preconditions
- focused `It` specs for observable behavior
- shared setup via `BeforeEach`, `JustBeforeEach`, `DeferCleanup`, and helper functions

### Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `APISIX_ADMIN_URL` | `http://127.0.0.1:9180` | Admin API base URL |
| `APISIX_GATEWAY_URL` | `http://127.0.0.1:9080` | Data plane base URL |
| `HTTPBIN_URL` | `http://127.0.0.1:8080` | httpbin upstream URL |

## Running Tests
Use the following commands to run tests:
- `make test`: Runs unit tests only.
- `make test-verbose`: Runs tests with verbose output.
- `make test-e2e`: Runs the real APISIX e2e suite.
- `make coverage`: Generates a unit-test coverage report for the current `go test` target set.
- `go test ./pkg/selector`: Runs unit tests for a pure helper package.
- `go test -tags e2e ./test/e2e/...`: Runs the e2e suite directly.

## Assertions
Use the `testify` library for assertions:
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

require.NoError(t, err)           // Fatal if an error occurs
assert.Equal(t, expected, actual) // Continue if the assertion fails
assert.Contains(t, output, "ID")  // Check for a substring
```
