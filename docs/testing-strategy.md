# Testing Strategy

## Test Requirements
- Every exported function must have at least one corresponding test.
- Every command must be tested for:
  - Success cases
  - Error cases
  - TTY output
  - Non-TTY output
- Aim for a code coverage target of 80% or higher for packages within the `pkg/` directory.

## Test File Location
Tests should be located in the same directory as the code they test. For example, `list.go` should have its tests in `list_test.go`.

Store test fixtures in `test/fixtures/<resource>_<action>.json`.

## Test Naming Convention
Follow the pattern `func Test<Function>_<Scenario>(t *testing.T) {}`.

Examples:
- `func TestRouteList_ReturnsTable(t *testing.T) {}`
- `func TestRouteList_EmptyResponse(t *testing.T) {}`
- `func TestRouteList_APIError(t *testing.T) {}`
- `func TestRouteList_JSONOutput(t *testing.T) {}`
- `func TestRouteList_NonTTY(t *testing.T) {}`

## HTTP Mocking Pattern
Use the project's internal `pkg/httpmock` package instead of external mock libraries.

```go
func TestRouteList_Success(t *testing.T) {
    // 1. Create mock registry
    reg := httpmock.NewRegistry()
    
    // 2. Register expected request and response
    reg.Register(
        httpmock.GET("/apisix/admin/routes"),
        httpmock.JSONResponse(200, loadFixture("route_list.json")),
    )
    
    // 3. Create test factory with mock client
    ios := iostreams.Test()
    f := &cmd.Factory{
        IOStreams: ios,
        HttpClient: func() (*http.Client, error) {
            return reg.GetClient(), nil
        },
    }
    
    // 4. Create and execute command
    cmd := list.NewCmdList(f)
    cmd.SetArgs([]string{})
    err := cmd.Execute()
    
    // 5. Verify results
    require.NoError(t, err)
    assert.Contains(t, ios.Out.String(), "users-api")
    reg.Verify(t)
}
```

## Test Categories

### Unit Tests
Required for every command to verify:
- Command flag parsing
- HTTP request construction (URL, query parameters, headers)
- Response parsing
- Output formatting for both table and JSON
- Error handling for API errors, network issues, and authentication failures

### TTY vs Non-TTY Tests
Every command must have tests for both TTY and non-TTY environments:

```go
func TestRouteList_TTY(t *testing.T) {
    ios := iostreams.Test()
    ios.SetStdoutTTY(true)
    // Verify table output
}

func TestRouteList_NonTTY(t *testing.T) {
    ios := iostreams.Test()
    ios.SetStdoutTTY(false)
    // Verify JSON output
}
```

## Test Fixtures
- **Location**: `test/fixtures/`
- **Naming**: `<resource>_<action>.json` (e.g., `route_list.json`)
- **Content**: Use realistic APISIX responses. Copy them from the actual API and redact any sensitive data.

Load fixtures in your tests using a helper:
```go
func loadFixture(name string) []byte {
    data, err := os.ReadFile(filepath.Join("../../../test/fixtures", name))
    if err != nil {
        panic(fmt.Sprintf("failed to load fixture %s: %v", name, err))
    }
    return data
}
```

## What NOT to Test
- Do not test cobra flag binding, as this is handled by the cobra framework.
- Do not test JSON marshaling, which is the responsibility of the standard library.
- Avoid writing integration tests against a real APISIX instance in unit test files — use e2e tests for that.

## E2E Tests

E2E tests validate the CLI binary against a real APISIX instance. They live in `test/e2e/` and use the `//go:build e2e` build tag to separate them from unit tests.

### Architecture

The e2e framework:
1. Builds the `a6` binary once in `TestMain` (in `setup_test.go`)
2. Waits for APISIX Admin API to become healthy
3. Runs tests that invoke the binary via `exec.Command`
4. Uses direct Admin API calls for setup/cleanup (not via the CLI)

### Infrastructure

Three services are required:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| etcd | `bitnamilegacy/etcd:3.6` | `2379` | APISIX config store |
| APISIX | `apache/apisix:3.11.0-debian` | `9180` (Admin), `9080` (Gateway) | Target instance |
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

- `test/e2e/setup_test.go` — `TestMain`, helper functions (`runA6`, `adminAPI`, `waitForHealthy`)
- `test/e2e/smoke_test.go` — Smoke tests verifying the binary runs and APISIX is reachable
- `test/e2e/<resource>_test.go` — Per-resource CRUD lifecycle tests (added per PR)
- `test/e2e/apisix_conf/config.yaml` — APISIX config for CI (etcd at `127.0.0.1`)
- `test/e2e/apisix_conf/config-docker.yaml` — APISIX config for docker-compose (etcd at `etcd`)

### Writing E2E Tests

```go
//go:build e2e

package e2e

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRoute_CRUD(t *testing.T) {
    // Setup: create a route via Admin API
    body := []byte(`{"uri": "/test", "upstream": {"nodes": {"httpbin:8080": 1}, "type": "roundrobin"}}`)
    resp, err := adminAPI("PUT", "/apisix/admin/routes/test-route-1", body)
    require.NoError(t, err)
    defer resp.Body.Close()
    require.Equal(t, 201, resp.StatusCode)

    // Test: list routes via CLI
    stdout, _, err := runA6("route", "list")
    require.NoError(t, err)
    assert.Contains(t, stdout, "test-route-1")

    // Cleanup: delete the route
    resp, err = adminAPI("DELETE", "/apisix/admin/routes/test-route-1", nil)
    require.NoError(t, err)
    resp.Body.Close()
}
```

### E2E Test Naming Convention

Follow: `Test<Resource>_<Scenario>` (e.g., `TestRoute_CRUD`, `TestSmoke_BinaryRuns`)

### Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `APISIX_ADMIN_URL` | `http://127.0.0.1:9180` | Admin API base URL |
| `APISIX_GATEWAY_URL` | `http://127.0.0.1:9080` | Data plane base URL |
| `HTTPBIN_URL` | `http://127.0.0.1:8080` | httpbin upstream URL |

## Running Tests
Use the following commands to run tests:
- `make test`: Runs all tests with race detection.
- `make test-verbose`: Runs tests with verbose output.
- `make coverage`: Generates and opens a coverage report.
- `go test ./pkg/cmd/route/list/...`: Runs tests for a specific package.

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
