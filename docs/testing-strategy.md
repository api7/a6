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
- Avoid writing integration tests against a real APISIX instance for the v1 scope.

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
