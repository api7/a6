# ADR-001: Technology Stack and Architecture

## Status
Accepted

## Context
We are building `a6`, a CLI tool for managing Apache APISIX through its Admin API. This project follows an "AI-first development" approach where AI coding agents perform the primary development. We need a tech stack that is well-documented, widely understood by AI models, and produces reliable, performant CLI binaries.

## Decisions

### 1. Language: Go 1.22+
Go is the primary language for `a6`.
- **Single binary distribution**: Simplifies installation for users across different platforms.
- **Excellent CLI ecosystem**: Industry-standard libraries like cobra and viper provide a solid foundation.
- **Strong typing**: Catches common errors at compile time, which is particularly helpful for AI-generated code.
- **AI-Friendly**: Go's syntax and standard patterns are well-understood by modern AI coding models.
- **Fast compilation**: Enables quick iteration cycles during development.

**Rejected alternatives**:
- **Rust**: Steeper learning curve for AI agents and a smaller CLI ecosystem compared to Go.
- **Python**: Complexity in distribution and slower runtime performance.
- **Node.js**: Requires a runtime environment, making distribution less straightforward.

### 2. Project Structure
The project follows a structure inspired by successful Go CLI tools, ensuring clarity and maintainability.

```
a6/
├── cmd/a6/main.go                    # Entry point — minimal, just calls root command
├── pkg/
│   ├── cmd/                           # Command implementations
│   │   ├── root/root.go              # Root command, registers all subcommands
│   │   ├── factory.go                # Factory struct for dependency injection
│   │   ├── route/                    # Route resource commands
│   │   │   ├── route.go             # Route parent command
│   │   │   ├── list/list.go         # route list subcommand
│   │   │   ├── list/list_test.go    # Tests
│   │   │   ├── list/http.go         # HTTP request logic (separated for testing)
│   │   │   ├── get/get.go
│   │   │   ├── create/create.go
│   │   │   ├── update/update.go
│   │   │   ├── delete/delete.go
│   │   │   └── shared/display.go    # Shared display logic for route resources
│   │   ├── upstream/                 # Same pattern as route/
│   │   ├── service/
│   │   ├── consumer/
│   │   ├── ssl/
│   │   ├── plugin/
│   │   └── context/                  # Context management commands
│   ├── api/                          # Admin API client
│   │   ├── client.go                # HTTP client wrapper
│   │   ├── types.go                 # Shared API types
│   │   ├── types_route.go           # Route-specific types
│   │   ├── types_upstream.go
│   │   └── ... (one types file per resource)
│   ├── iostreams/                    # I/O abstraction
│   │   └── iostreams.go             # Stdin/Stdout/Stderr + TTY detection
│   ├── cmdutil/                      # Command utilities
│   │   ├── exporter.go              # JSON/YAML/table export
│   │   └── errors.go               # Error formatting
│   ├── tableprinter/                 # Table output
│   │   └── table.go                 # Table rendering with color support
│   └── httpmock/                     # HTTP test utilities
│       └── httpmock.go              # Request recording and response stubbing
├── internal/
│   ├── config/                       # Configuration management
│   │   └── config.go                # Context/config read/write
│   └── version/                      # Build version info
│       └── version.go
├── docs/                             # Documentation
├── test/fixtures/                    # Test fixture JSON/YAML files
├── AGENTS.md                         # AI agent development guide
├── PRD.md                           # Product requirements
├── Makefile
├── go.mod
└── go.sum
```

**Rationale**:
- `pkg/` contains importable packages.
- `internal/` is reserved for non-importable internal logic.
- A dedicated directory for each subcommand keeps logic self-contained and modular.

### 3. Key Dependencies
- **github.com/spf13/cobra**: The foundation for the command-line interface.
- **github.com/spf13/viper**: Manages configuration files and environment variables.
- **github.com/olekukonez/tablewriter** or **github.com/charmbracelet/lipgloss**: For structured table output (final choice to be made during implementation).
- **github.com/stretchr/testify**: Provides robust assertions for testing.
- **gopkg.in/yaml.v3**: Handles YAML parsing.
- **encoding/json** (stdlib): Handles JSON parsing.
- **net/http** (stdlib): Used for all HTTP communication with a custom `RoundTripper` for authentication.

### 4. Architecture Patterns

#### Factory Pattern (Dependency Injection)
Every command receives a `Factory` struct to ensure testability and clear dependency management. It includes:
- `IOStreams`: Abstractions for stdin, stdout, and stderr, including TTY detection.
- `HttpClient`: Pre-configured client with authentication based on the active context.
- `Config`: Access to context and configuration management.

This pattern allows commands to be tested in isolation without performing actual network or disk I/O.

#### Command Pattern
Each command is structured into four distinct parts:
1. **Options struct**: Stores flag values and resolved dependencies.
2. **NewCmd* function**: Initializes the `cobra.Command`, defines flags, and wires the `Options`.
3. **Run function**: Contains the core business logic.
4. **Logic Separation**: HTTP request construction resides in `http.go`, while output formatting is handled in `display.go`.

#### API Client Pattern
- Operates as a thin wrapper around the Go standard library's `net/http`.
- Authentication is injected via `http.RoundTripper` middleware, making it transparent to callers.
- Methods return parsed structs instead of raw responses.
- API errors are parsed into structured types that include APISIX-specific error codes.

#### Output Pattern
- **TTY Detection**: If stdout is a terminal, the CLI provides colorful table output.
- **Piping/Non-TTY**: Default output is JSON to support machine readability.
- **Overrides**: Users can explicitly set the output format using the `--output` flag (json, yaml, table) or a custom Go template via `--format`.

#### Testing Pattern
- **Unit Tests**: Use an `httpmock` registry to stub responses, which are then injected via the `Factory`.
- **Test Fixtures**: JSON files in `test/fixtures/` provide consistent data for mock responses.
- **Environment Variants**: Tests cover both interactive (TTY) and piped (non-TTY) scenarios.
- **Scope**: Version 1 focuses on unit tests with realistic mock data instead of full integration tests.

### 5. Configuration Design
- **Location**: `~/.config/a6/config.yaml` (follows `XDG_CONFIG_HOME` and `A6_CONFIG_DIR` if set).
- **Format**: YAML structure containing an array of contexts and a pointer to the current context.
- **Authentication Precedence**: `--api-key` flag > `A6_API_KEY` env var > context config.
- **Server Precedence**: `--server` flag > `A6_SERVER` env var > context config.

### 6. Error Handling
- **No Panics**: All errors are handled through the standard `error` interface.
- **HTTP Errors**: Body content is parsed to show the status, error message, and a helpful suggestion.
- **Connectivity Issues**: Network errors are wrapped with user-friendly descriptions.
- **Auth Failures**: Specific messages guide the user to check their API key or use `a6 context create`.

## Consequences
- **Explicit Handling**: Go requires manual error checking, which adds verbosity but increases safety.
- **Opinionated Frameworks**: Cobra and Viper impose specific patterns on flag and config handling.
- **Initial Setup**: The Factory pattern requires more boilerplate initially, but it significantly simplifies testing.
- **Modularity**: The directory-per-command approach keeps the codebase organized as it grows.
