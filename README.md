# a6 — Apache APISIX CLI

`a6` is a command-line tool for managing [Apache APISIX](https://apisix.apache.org/) from your terminal. It wraps the APISIX Admin API to provide convenient, scriptable access to routes, upstreams, services, consumers, SSL certificates, plugins, and more.

Built with an **AI-first development approach** — the codebase includes structured documentation and AI agent skills that enable autonomous development by coding agents.

## Features

- **Resource CRUD** — Create, list, get, update, and delete all 14 APISIX Admin API resources:
  - Route, Upstream, Service, Consumer, SSL Certificate
  - Plugin Config, Global Rule, Consumer Group
  - Stream Route, Secret, Plugin Metadata, Proto
- **Context management** — Switch between multiple APISIX instances (`a6 context create`, `a6 context use`, `a6 context list`)
- **Declarative configuration** — Sync, dump, diff, and validate APISIX configuration from YAML files (`a6 config sync`, `a6 config dump`, `a6 config diff`, `a6 config validate`)
- **Debug commands** — Stream real-time APISIX error logs (`a6 debug logs`) and trace request flow (`a6 debug trace`)
- **Health check** — Verify APISIX connectivity and version (`a6 health`)
- **Rich output** — Human-friendly tables in TTY, machine-readable JSON/YAML in pipes (`--output json|yaml|table`)
- **Shell completions** — Bash, Zsh, Fish, PowerShell (`a6 completion`)
- **Self-update** — Update the CLI binary to the latest version (`a6 self-update`)
- **Export** — Export configurations to Kubernetes, Helm, Terraform, or Standalone YAML (`a6 export`)
- **AI agent skills** — 35+ built-in [SKILL.md files](skills/) for AI coding agents to work effectively with APISIX

## Quick Start

```bash
# Install
go install github.com/api7/a6/cmd/a6@latest

# Configure connection to your APISIX instance
a6 context create local --server http://localhost:9180 --api-key <your-admin-key>

# List routes
a6 route list

# Get a specific route in JSON
a6 route get 1 --output json

# Create a route from a file
a6 route create -f route.json

# Sync declarative configuration
a6 config sync -f apisix.yaml

# Dump current configuration to YAML
a6 config dump > apisix.yaml

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
- Apache APISIX 3.x with Admin API enabled

## AI Agent Skills

The `skills/` directory contains structured knowledge files (`SKILL.md`) that enable AI coding agents to configure APISIX through the a6 CLI. Skills are compatible with 39+ AI coding tools including Claude Code, OpenCode, Cursor, GitHub Copilot, and Windsurf.

| Category | Count | Examples |
|----------|-------|---------|
| **Shared** | 1 | Core a6 conventions and patterns |
| **Authentication** | 5 | key-auth, jwt-auth, basic-auth, hmac-auth, openid-connect |
| **Security & Rate Limiting** | 4 | ip-restriction, cors, limit-count, limit-req |
| **Traffic & Transformation** | 5 | proxy-rewrite, response-rewrite, traffic-split, redirect, grpc-transcode |
| **AI Gateway** | 4 | ai-proxy, ai-prompt-template, ai-prompt-decorator, ai-content-moderation |
| **Observability** | 6 | prometheus, skywalking, zipkin, http-logger, kafka-logger, datadog |
| **Advanced Plugins** | 5 | serverless, ext-plugin, fault-injection, consumer-restriction, wolf-rbac |
| **Operational Recipes** | 5 | blue-green, canary, circuit-breaker, health-check, mTLS |

See [docs/skills.md](docs/skills.md) for the full skill format specification, taxonomy, and authoring guide.

## Documentation

- [Product Requirements](PRD.md)
- [Development Roadmap](docs/roadmap.md)
- [AI Agent Skills Guide](docs/skills.md)
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
