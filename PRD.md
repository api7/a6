# Product Requirements Document (PRD): a6 CLI

## Overview
- **Project Name**: a6 (repository: api7/a6)
- **Purpose**: A command-line tool that wraps the Apache APISIX Admin API, providing a convenient terminal interface for managing APISIX resources.
- **Target Users**: DevOps engineers, API developers, and platform teams responsible for managing and automating APISIX instances.
- **Design Philosophy**: AI-first development. All documentation and specifications are written to enable AI coding agents to understand the project scope and develop it autonomously.

## Problem Statement
- Apache APISIX lacks an imperative CLI for Admin API operations. Users are currently forced to use `curl` for manual resource management.
- The existing CLI (`bin/apisix`) is strictly for lifecycle management, such as starting, stopping, and reloading the service, not for managing API resources.
- While a Dashboard exists, it requires a web browser and is not easily scriptable for terminal-based workflows.
- The ADC tool provides declarative configuration management, but there is no quick tool for ad-hoc CRUD (Create, Read, Update, Delete) operations.

## Goals and Non-Goals

### Goals
- Provide full CRUD operations for all 14 Apache APISIX Admin API resources.
- Implement context management for switching between multiple APISIX instances (similar to `gcloud` or `kubectl`).
- Support rich terminal output, including tables for interactive TTY sessions and JSON/YAML for piped output.
- Incorporate declarative configuration synchronization by absorbing ADC functionality over the long term.
- Provide shell completions for major shells (Bash, Zsh, Fish, PowerShell).
- Maintain an extensible command architecture to support future enhancements.

### Non-Goals
- Not intended to replace the APISIX Dashboard (Web UI).
- Not replacing existing lifecycle management commands like `start` or `stop` found in `bin/apisix`.
- Not building a graphical user interface (GUI).
- Not managing the installation or deployment of APISIX itself.
- Not supporting APISIX versions prior to v3.x.

## Target APISIX Version
- **Supported Version**: APISIX v3.x (utilizing Admin API v3 response formats).
- **Response Format**:
  - **Single Resource**: `{"key":"...","value":{...},"createdIndex":N,"modifiedIndex":N}`
  - **List**: `{"total":N,"list":[...]}`

## Command Design

### Command Structure
The CLI follows a noun-verb pattern:
`a6 <resource> <action> [flags]`

### Resource Commands
Covers all 14 Admin API resources:
- `a6 route list|get|create|update|delete`: Route management.
- `a6 service list|get|create|update|delete`: Service management.
- `a6 upstream list|get|create|update|delete`: Upstream management, including a `health` subcommand.
- `a6 consumer list|get|create|update|delete`: Consumer management.
- `a6 consumer credential list|get|create|update|delete`: Nested credential management for consumers.
- `a6 ssl list|get|create|update|delete|upload`: SSL certificate management.
- `a6 plugin list|get`: Plugin listing and schema inspection.
- `a6 plugin-metadata get|create|update|delete`: Plugin metadata management.
- `a6 plugin-config list|get|create|update|delete`: Plugin configuration management.
- `a6 global-rule list|get|create|update|delete`: Global rules management.
- `a6 stream-route list|get|create|update|delete`: Stream route management.
- `a6 proto list|get|create|update|delete`: Protocol buffer management.
- `a6 consumer-group list|get|create|update|delete`: Consumer group management.
- `a6 secret list|get|create|update|delete`: Secret management.
- `a6 schema get`: Schema inspection for resources.

### Utility Commands
- `a6 context create|use|list|delete|current`: Manage contexts for multiple APISIX instances.
- `a6 config sync|dump|diff|validate`: Declarative configuration operations (ADC functionality).
- `a6 completion bash|zsh|fish|powershell`: Generate shell completion scripts.
- `a6 version`: Display version information.

### Common Flags
- `--context` / `-c`: Override the active context.
- `--server` / `-s`: Override the target server URL.
- `--api-key`: Override the Admin API key.
- `--output` / `-o`: Set output format: `table` (default in TTY), `json`, or `yaml`.
- `--format`: Use a Go template for custom output formatting.
- `--verbose` / `-v`: Enable verbose output, showing HTTP request details.
- `--force`: Skip confirmation prompts for destructive actions.
- `-f` / `--file`: Read a resource definition from a JSON or YAML file.

### Resource-Specific Flags
- **list**: `--page`, `--page-size`, `--name`, `--label`, and `--uri` (routes only).
- **create/update**: `-f/--file` (JSON/YAML input) and `--ttl` (TTL on PUT operations).
- **delete**: `--force` to skip confirmation and handle routes with active plugins.

## MVP Scope (Phase 1)
1. Context management (create, use, list, delete, current).
2. Route CRUD (list, get, create, update, delete).
3. Upstream CRUD.
4. Service CRUD.
5. Consumer CRUD.
6. SSL CRUD.
7. Plugin list and schema retrieval.
8. Shell completions.
9. JSON and YAML output modes.

## Phase 2
1. Support for all remaining resources: global-rule, stream-route, proto, plugin-metadata, plugin-config, consumer-group, secret, and consumer credentials.
2. Declarative config operations: sync, dump, diff, and validate.
3. Upstream health check status display.
4. Interactive mode featuring fuzzy selection.

## Phase 3
1. Extension and plugin system for the CLI.
2. Debugging tools, including log streaming and request tracing.
3. Bulk operations for resource management.
4. Automated update mechanism.

## UX Requirements
- **TTY Detection**: Default to tables in interactive terminals and JSON when output is piped.
- **Colors**: Use ANSI colors when supported. Respect the `NO_COLOR` environment variable.
- **Confirmation Prompts**: Destructive operations like `delete` must require confirmation unless the `--force` flag is used.
- **Error Messages**: Display the HTTP status code, the APISIX error message, and a helpful suggestion for resolution.
- **Progress Feedback**: Use spinners or progress indicators for long-running operations.
- **Consistent Sorting**: Resources should be listed by ID in ascending order by default.

## Authentication
- **Precedence**: Flag (`--api-key`) > Environment Variable (`A6_API_KEY`) > Context Configuration.
- **Security**: Prevent sensitive data from appearing in shell history by supporting key input from files or stdin.

## Configuration Storage
- **Location**: `~/.config/a6/config.yaml`. Respect `XDG_CONFIG_HOME` if set.
- **Format**:
  ```yaml
  current-context: local
  contexts:
    - name: local
      server: http://localhost:9180
      api-key: edd1c9f034335f136f87ad84b625c8f1
    - name: staging
      server: https://staging.example.com:9180
      api-key: <key>
  ```
