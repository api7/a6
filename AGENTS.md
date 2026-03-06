# AGENTS.md — a6 Development Guide
> Entry point for developers and AI coding agents working on the a6 CLI.
> Read this FIRST before making any changes.

### Project Overview
a6 is a Go CLI wrapping the Apache APISIX Admin API.
- Binary name: `a6`
- Module path: `github.com/api7/a6`
- Go version: 1.22+
- License: Apache 2.0

### Document Map
List of project documents and their specific purpose:

| Document | Purpose | When to Read |
|---|---|---|
| `AGENTS.md` (this file) | Entry point, development guide | Always — read first |
| `PRD.md` | Product requirements, command design, scope | Before adding new features |
| `docs/admin-api-spec.md` | Complete APISIX Admin API reference | When implementing API client code |
| `docs/adr/001-tech-stack.md` | Architecture decisions, project structure, patterns | Before writing any code |
| `docs/golden-example.md` | Complete `route list` implementation as template | When adding any new command |
| `docs/coding-standards.md` | Go style, naming, formatting conventions | Before writing code |
| `docs/testing-strategy.md` | Test patterns, mocking, fixtures | Before writing tests |
| `docs/documentation-maintenance.md` | Doc update rules | After any code change |

### Project Structure
Project directory tree with annotations:

```
a6/
├── cmd/a6/main.go                 # Entry point
├── pkg/cmd/                       # All command implementations
│   ├── root/root.go              # Root command
│   ├── factory.go                # Dependency injection
│   ├── route/                    # Route commands (list, get, create, update, delete)
│   ├── upstream/                 # Upstream commands
│   ├── service/                  # Service commands
│   ├── consumer/                 # Consumer commands
│   ├── ssl/                      # SSL commands
│   ├── plugin/                   # Plugin commands
│   └── context/                  # Context management
├── pkg/api/                       # APISIX Admin API client
│   ├── client.go                 # HTTP client with auth
│   └── types_*.go                # Go types for each resource
├── pkg/iostreams/                 # I/O abstraction (TTY detection)
├── pkg/cmdutil/                   # Shared command utilities
├── pkg/tableprinter/              # Table output rendering
├── pkg/httpmock/                  # HTTP mocking for tests
├── internal/config/               # Configuration/context management
├── internal/version/              # Build version info
├── docs/                          # All documentation
├── test/fixtures/                 # JSON fixtures for tests
└── Makefile                       # Build, test, lint commands
```

### Key Architecture Patterns
Core design principles (see `docs/adr/001-tech-stack.md` for details):
1. **Factory Pattern**: Every command receives a Factory containing IOStreams, HttpClient, and Config. No global state is allowed.
2. **Command Pattern**: Uses an Options struct, a `NewCmd` function, and a `Run` function for consistent command structure.
3. **Output Pattern**: Table output for TTY, JSON for non-TTY. The `--output` flag overrides this behavior.
4. **Testing Pattern**: Uses `httpmock` stubs and test IOStreams. Real network calls are prohibited in tests.

### How to Add a New Command (Step-by-Step)
Follow these steps when adding a new command, such as `a6 upstream list`:

1. **Read the API spec**: Open `docs/admin-api-spec.md` and find the Upstream section.
2. **Create types**: Add `pkg/api/types_upstream.go` with Go structs matching the API schema.
3. **Create parent command**: Add `pkg/cmd/upstream/upstream.go`.
4. **Create the action**: Add `pkg/cmd/upstream/list/list.go` following the template in `docs/golden-example.md`.
5. **Add tests**: Create `pkg/cmd/upstream/list/list_test.go`. Include TTY, non-TTY, filter, and error tests.
6. **Add fixture**: Place realistic mock data in `test/fixtures/upstream_list.json`.
7. **Register command**: Add the new command to `pkg/cmd/root/root.go`.
8. **Update docs**: Add or update the relevant user guide, such as `docs/user-guide/upstream.md`.
9. **Run checks**: Execute `make check` to run fmt, vet, lint, and tests.

### Common Commands
```bash
make build          # Build binary to ./bin/a6
make test           # Run all tests
make test-verbose   # Run tests with verbose output
make lint           # Run golangci-lint
make fmt            # Format code
make check          # Run all checks (fmt + vet + lint + test)
make clean          # Remove build artifacts
```

### Code Style Rules
- Follow `gofmt` and `goimports`.
- Use `golangci-lint` with the project configuration.
- Error messages must be lowercase with no trailing punctuation.
- Use camelCase for local variables and PascalCase for exported variables.
- Avoid `any` or `interface{}`. Use concrete types or generics.
- All types must be explicit.
- Name test files `*_test.go` within the same package.

### Commit Message Format
```
<type>(<scope>): <description>

Types: feat, fix, refactor, test, docs, chore
Scope: route, upstream, service, consumer, ssl, plugin, context, api, config, ci
Example: feat(route): add route list command with table output
```

### Mandatory Rules (NEVER violate)
1. Every code change must have accompanying tests.
2. Every new command must follow the pattern in `docs/golden-example.md`.
3. Every new feature must update the documentation as per `docs/documentation-maintenance.md`.
4. Never suppress errors. Always handle and propagate them.
5. Never use global state. Always inject dependencies via the Factory.
6. Run `make check` before committing. All checks must pass.

### Environment Variables
| Variable | Purpose | Default |
|---|---|---|
| `A6_API_KEY` | Admin API key | (from context config) |
| `A6_SERVER` | Admin API server URL | (from context config) |
| `A6_CONFIG_DIR` | Config directory | `~/.config/a6` |
| `NO_COLOR` | Disable color output | (unset) |
