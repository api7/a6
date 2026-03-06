# a6 — Apache APISIX CLI

`a6` is a command-line tool for managing [Apache APISIX](https://apisix.apache.org/) from your terminal. It wraps the APISIX Admin API to provide convenient, scriptable access to routes, upstreams, services, consumers, SSL certificates, plugins, and more.

## Status

🚧 **Under active development.** Not yet ready for production use.

## Features (Planned)

- **Resource CRUD** — Create, list, get, update, and delete all APISIX resources (routes, upstreams, services, consumers, SSL, plugins, etc.)
- **Context management** — Switch between multiple APISIX instances (`a6 context use staging`)
- **Declarative config** — Sync, dump, diff, and validate APISIX configuration from YAML files
- **Rich output** — Human-friendly tables in TTY, machine-readable JSON/YAML in pipes
- **Shell completions** — Bash, Zsh, Fish, PowerShell

## Quick Start

```bash
# Install (once published)
go install github.com/api7/a6/cmd/a6@latest

# Configure connection to your APISIX instance
a6 context create local --server http://localhost:9180 --api-key edd1c9f034335f136f87ad84b625c8f1

# List routes
a6 route list

# Get a specific route
a6 route get 1

# Create a route from JSON
a6 route create -f route.json

# Delete a route
a6 route delete 1
```

## Building from Source

```bash
git clone https://github.com/api7/a6.git
cd a6
make build
./bin/a6 --help
```

## Requirements

- Go 1.22+
- Access to an Apache APISIX instance (v3.x) with Admin API enabled

## Documentation

- [Product Requirements](PRD.md)
- [Architecture Decisions](docs/adr/)
- [Admin API Reference](docs/admin-api-spec.md)
- [Coding Standards](docs/coding-standards.md)
- [Testing Strategy](docs/testing-strategy.md)
- [Documentation Maintenance](docs/documentation-maintenance.md)
- [AI Agent Guide](AGENTS.md)

## Contributing

See [AGENTS.md](AGENTS.md) for development workflow, coding conventions, and how to add new commands.

## License

[Apache License 2.0](LICENSE)
