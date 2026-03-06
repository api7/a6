# Coding Standards

## Go Version
Go 1.22+ is required. Use generics where they improve code clarity or reduce duplication.

## Formatting
The use of `gofmt` and `goimports` is mandatory. Run these tools before every commit to ensure consistency.

While there is no hard line length limit, prefer wrapping around 100-120 characters to maintain readability. Use tabs for indentation as per standard Go conventions.

## Naming Conventions
- **Packages**: Use lowercase, single-word names (e.g., `api`, `config`, `iostreams`).
- **Exported symbols**: Use PascalCase for types and functions.
- **Unexported symbols**: Use camelCase for types and functions.
- **Constants**: Use PascalCase for exported constants and camelCase for unexported ones.
- **Acronyms**: Capitalize acronyms fully (e.g., `HTTPClient`, `APIURL`, `SSLCert`).
- **Files**: Use snake_case for filenames (e.g., `types_route.go`, `list_test.go`).
- **Test functions**: Follow the pattern `Test<Function>_<Scenario>` (e.g., `TestRouteList_EmptyResponse`).

## Error Handling
Always return errors and avoid ignoring them with `_`. Never use `panic()` for expected error conditions.

Error messages should be lowercase and without trailing punctuation. Wrap errors with context using the `%w` verb:
```go
return fmt.Errorf("failed to list routes: %w", err)
```

Use custom error types for API errors:
```go
type APIError struct {
    StatusCode int
    Message    string
    ErrorMsg   string // APISIX error_msg field
}
```

## Struct Design
- **Required fields**: Use value types like `string` or `int`.
- **Optional fields**: Use pointer types such as `*string` or `*int`. This ensures that JSON `omitempty` tags work correctly by distinguishing between a zero value and a missing value.
- **JSON tags**: Match the exact APISIX API field names, which typically use snake_case.

Example:
```go
type Route struct {
    ID       string   `json:"id"`
    URI      string   `json:"uri,omitempty"`
    Name     *string  `json:"name,omitempty"`
    Methods  []string `json:"methods,omitempty"`
    Status   int      `json:"status"`
}
```

## Import Organization
Group imports into three blocks separated by blank lines:
1. Standard library packages
2. Third-party packages
3. Internal project packages (`github.com/api7/a6/...`)

## Package Design
The `pkg/` directory contains code that can be imported by external consumers. Use the `internal/` directory for code that should not be accessible outside this module. Each package should have a single, clear responsibility. Avoid circular dependencies.

## Forbidden Patterns
- Avoid global variables, except for version information.
- Do not use `init()` functions, unless they are necessary for cobra command registration.
- Avoid `interface{}` or `any` when a concrete type can be used.
- Do not use unchecked type assertions.
- Never call `os.Exit()` directly outside of `cmd/a6/main.go`.
- Use `IOStreams` for all output instead of `fmt.Println`.

## Dependency Rules
Prefer the standard library over external packages whenever possible. The following external dependencies are approved:
- `github.com/spf13/cobra` (command framework)
- `github.com/spf13/viper` (configuration)
- `github.com/stretchr/testify` (test assertions)
- `gopkg.in/yaml.v3` (YAML processing)
- Table printer (TBD: `tablewriter` or `lipgloss`)

Adding a new dependency requires explicit justification in the pull request description.
