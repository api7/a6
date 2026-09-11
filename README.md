# a6 — Apache APISIX CLI

`a6` is a command-line tool for managing [Apache APISIX](https://apisix.apache.org/) from your terminal. It wraps the APISIX Admin API to provide convenient, scriptable access to routes, upstreams, services, consumers, SSL certificates, plugins, and more.

Built with an **AI-first development approach** — the codebase includes structured documentation for coding agents, and a companion [AI agent skill](#ai-agent-skills) teaches those agents how to operate APISIX with `a6`.

## Features

- **Resource CRUD** — Create, list, get, update, and delete all 14 APISIX Admin API resources:
  - Route, Upstream, Service, Consumer, SSL Certificate
  - Plugin Config, Global Rule, Consumer Group
  - Stream Route, Secret, Plugin Metadata, Proto
- **Context management** — Switch between multiple APISIX instances (`a6 context create`, `a6 context use`, `a6 context list`)
- **Declarative configuration** — Sync, dump, diff, and validate APISIX configuration from YAML files (`a6 config sync`, `a6 config dump`, `a6 config diff`, `a6 config validate`)
- **Debug commands** — Stream real-time APISIX error logs (`a6 debug logs`) and trace request flow (`a6 debug trace`)
- **Rich output** — Human-friendly tables in TTY, machine-readable JSON/YAML in pipes (`--output json|yaml|table`)
- **Shell completions** — Bash, Zsh, Fish, PowerShell (`a6 completion`)
- **Self-update** — Update the CLI binary to the latest version (`a6 update`)
- **Export** — Export resource configurations to standalone YAML or JSON (`a6 route export`, `a6 upstream export --label env=prod`)
- **AI agent skill** — an [`a6` skill](https://skills.sh/api7/agent-skills/a6) that teaches AI coding agents to configure APISIX through this CLI (`npx skills add api7/agent-skills --skill a6`)

## Installation

### Using Go

If you have Go 1.22+ installed:

```bash
go install github.com/api7/a6/cmd/a6@latest
```

This installs the `a6` binary to your `$GOPATH/bin` (or `$HOME/go/bin` by default). Make sure it's in your `PATH`.

### Build from Source

```bash
git clone https://github.com/api7/a6.git
cd a6
make build
# Binary is at ./bin/a6
```

## Quick Start

This walkthrough takes about 5 minutes and exercises most a6 features against a local APISIX instance.

### Prerequisites

- a6 installed (see [Installation](#installation) above)
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose

### 1. Start APISIX

```bash
# If you built from source, start APISIX from the repo directory:
docker compose -f test/e2e/docker-compose.yml up -d

# Wait ~10 seconds for APISIX to be ready, then verify
curl -s http://localhost:9180/apisix/admin/routes -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' | head -c 50
```

### 2. Configure a context

```bash
# Create a context pointing to the local APISIX instance
a6 context create local --server http://localhost:9180 --api-key edd1c9f034335f136f87ad84b625c8f1

# Verify the active context
a6 context current

# List all contexts
a6 context list
```

### 3. Create resources

```bash
# Create an upstream pointing to httpbin
cat <<'EOF' > /tmp/upstream.json
{
  "id": "httpbin",
  "name": "httpbin",
  "type": "roundrobin",
  "nodes": {"httpbin:8080": 1}
}
EOF
a6 upstream create -f /tmp/upstream.json

# Create a route that references the upstream
cat <<'EOF' > /tmp/route.json
{
  "id": "test-route",
  "name": "test-route",
  "uri": "/get",
  "methods": ["GET"],
  "upstream_id": "httpbin",
  "labels": {"env": "test", "team": "backend"}
}
EOF
a6 route create -f /tmp/route.json
```

### 4. Read and explore

```bash
# List all routes (table output in terminal)
a6 route list

# Get a specific route in JSON
a6 route get test-route --output json

# Get it in YAML
a6 route get test-route --output yaml

# List available plugins
a6 plugin list
```

### 5. Update a resource

```bash
# Update the route to add a new method
cat <<'EOF' > /tmp/route-update.json
{
  "id": "test-route",
  "name": "test-route",
  "uri": "/get",
  "methods": ["GET", "POST"],
  "upstream_id": "httpbin",
  "labels": {"env": "test", "team": "backend"}
}
EOF
a6 route update test-route -f /tmp/route-update.json
```

### 6. Export and declarative config

```bash
# Export all routes to YAML
a6 route export

# Export routes with a label filter
a6 route export --label env=test

# Dump the entire APISIX configuration to a YAML file
a6 config dump -f /tmp/apisix-dump.yaml

# Validate a config file
a6 config validate -f /tmp/apisix-dump.yaml

# Preview what a sync would change (dry run)
a6 config sync -f /tmp/apisix-dump.yaml --dry-run
```

### 7. Clean up

```bash
# Delete the route and upstream
a6 route delete test-route
a6 upstream delete httpbin

# Verify they're gone
a6 route list
a6 upstream list

# Stop the local APISIX stack
docker compose -f test/e2e/docker-compose.yml down
```

## Requirements

- Go 1.22+
- Apache APISIX 3.x with Admin API enabled

## AI Agent Skills

The `a6` agent skill teaches AI coding agents (Claude Code, Cursor, Codex, GitHub Copilot, Windsurf, OpenCode and 70+ others) how to configure APISIX through the a6 CLI. The skill content lives in the [api7/agent-skills](https://github.com/api7/agent-skills) repository and is published at [skills.sh/api7/agent-skills/a6](https://skills.sh/api7/agent-skills/a6).

```bash
# install into the current project (add -g for a global install, -a <agent> to pick an agent)
npx skills add api7/agent-skills --skill a6
```

Without Node.js, `install.sh` in this repository copies the skill into `~/.claude/skills/a6` (or `--dir <path>`):

```bash
curl -fsSL https://raw.githubusercontent.com/api7/a6/main/install.sh | sh
```

One skill covers everything; the agent reads the detailed reference for a topic only when a task needs it:

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
| **Advanced Recipes** | 3 | multi-tenant, api-versioning, graphql-proxy |
| **Personas** | 2 | operator, developer |

Skill content changes go to [api7/agent-skills](https://github.com/api7/agent-skills); this repository's CI (`make test-skills`) validates the shell examples against the current CLI. See [docs/skills.md](docs/skills.md) for details.

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
