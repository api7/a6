# Development Roadmap

This document defines the per-PR development plan for the a6 CLI. Each PR is self-contained and ships implementation code, e2e tests against a real APISIX instance, and user-facing documentation updates.

**Audience**: AI coding agents and human developers. Each PR section contains enough detail to be implemented autonomously.

> **Status: ✅ Phase 1-3 COMPLETE (27 PRs)** — Phase 4 (AI Agent Skills, PR-28 through PR-36) is in progress.

---

## Table of Contents

- [CI & Testing Infrastructure](#pr-1-ci--e2e-testing-infrastructure)
- [Foundation Packages](#pr-2-foundation-packages)
- [Context Management](#pr-3-context-management)
- [Route CRUD](#pr-4-route-crud)
- [Upstream CRUD](#pr-5-upstream-crud)
- [Service CRUD](#pr-6-service-crud)
- [Consumer CRUD](#pr-7-consumer-crud)
- [SSL CRUD](#pr-8-ssl-crud)
- [Plugin Commands](#pr-9-plugin-list--get)
- [Shell Completions & Version](#pr-10-shell-completions--version)
- [Phase 2 PRs](#phase-2-prs)
- [Phase 3 PRs](#phase-3-prs)
- [Phase 4 — AI Agent Skills](#phase-4--ai-agent-skills)

---

## PR Dependency Graph

```
PR-1 (CI + E2E infra)
  └── PR-2 (Foundation packages)
        ├── PR-3 (Context management)
        │     └── PR-4 (Route CRUD) ← first resource command
        │           ├── PR-5 (Upstream CRUD)
        │           ├── PR-6 (Service CRUD)
        │           ├── PR-7 (Consumer CRUD)
        │           ├── PR-8 (SSL CRUD)
        │           └── PR-9 (Plugin list/get)
        └── PR-10 (Shell completions + version)
```

All Phase 2 PRs depend on Phase 1 completion. Within Phase 1, PR-5 through PR-9 can be developed in parallel after PR-4 is merged (PR-4 establishes the resource command pattern).

---

## E2E Testing Architecture

All PRs from PR-4 onward include e2e tests that run against a real APISIX instance.

### GitHub Actions Service Containers

The e2e CI workflow uses three service containers running alongside the GitHub Actions runner:

| Service | Image | Port (host) | Purpose |
|---------|-------|-------------|---------|
| etcd | `bitnamilegacy/etcd:3.6` | `2379` | APISIX configuration store |
| apisix | `apache/apisix:3.15.0-debian` | `9180` (Admin), `9080` (Data) | Target APISIX instance |
| httpbin | `ghcr.io/mccutchen/go-httpbin` | `8080` | Upstream target for route testing |

### APISIX Configuration for Tests

```yaml
# test/e2e/apisix_conf/config.yaml
apisix:
  node_listen: 9080
  enable_ipv6: false

deployment:
  admin:
    allow_admin:
      - 0.0.0.0/0
    admin_key:
      - name: "admin"
        key: edd1c9f034335f136f87ad84b625c8f1
        role: admin
  etcd:
    host:
      - "http://etcd:2379"
    prefix: "/apisix"
    timeout: 30
```

### E2E Test Pattern

- **Build tag**: `//go:build e2e` — separates e2e tests from unit tests
- **Location**: `test/e2e/` directory
- **Binary**: Built once in `TestMain`, invoked via `exec.Command`
- **Environment variables**: `APISIX_ADMIN_URL` (default: `http://localhost:9180`), `APISIX_GATEWAY_URL` (default: `http://localhost:9080`), `HTTPBIN_URL` (default: `http://localhost:8080`)

```go
// test/e2e/setup_test.go — shared across all e2e tests
func TestMain(m *testing.M) {
    // Build the a6 binary once
    // Set up environment variables
    // Wait for APISIX health check
    // Run tests
    // Exit
}

func runA6(args ...string) (stdout string, stderr string, err error) {
    // Execute the a6 binary with given args
    // Return captured output
}

func cleanupResource(resourceType, id string) {
    // DELETE resource via Admin API (cleanup helper)
}
```

### Local Development

Developers can run e2e tests locally using docker-compose:

```bash
# Start APISIX stack
docker-compose -f test/e2e/docker-compose.yml up -d

# Run e2e tests
make test-e2e

# Tear down
docker-compose -f test/e2e/docker-compose.yml down -v
```

---

## Phase 1 — MVP ✅

### PR-1: CI & E2E Testing Infrastructure

**Goal**: Establish CI pipelines and the e2e test framework. No business logic.

#### Files to Create

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | Unit test + lint CI on every push/PR |
| `.github/workflows/e2e.yml` | E2E test CI with real APISIX services |
| `test/e2e/docker-compose.yml` | Local dev docker-compose for e2e |
| `test/e2e/apisix_conf/config.yaml` | APISIX config for test environment |
| `test/e2e/setup_test.go` | TestMain, binary build, helper functions |
| `test/e2e/smoke_test.go` | Smoke test — verify APISIX is reachable |

#### Files to Modify

| File | Change |
|------|--------|
| `Makefile` | Add `test-e2e` and `docker-up` / `docker-down` targets |
| `AGENTS.md` | Add e2e testing section to document map and commands |

#### `.github/workflows/ci.yml` — Unit Test & Lint

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make fmt
      - run: make vet
      - run: make test

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

#### `.github/workflows/e2e.yml` — E2E with Real APISIX

```yaml
name: E2E Tests
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  APISIX_DOCKER_TAG: "3.15.0-debian"

jobs:
  e2e:
    runs-on: ubuntu-latest
    services:
      etcd:
        image: bitnamilegacy/etcd:3.6
        ports:
          - 2379:2379
        env:
          ETCD_DATA_DIR: /etcd_data
          ETCD_ENABLE_V2: "true"
          ALLOW_NONE_AUTHENTICATION: "yes"
          ETCD_ADVERTISE_CLIENT_URLS: "http://etcd:2379"
          ETCD_LISTEN_CLIENT_URLS: "http://0.0.0.0:2379"
        options: >-
          --health-cmd "etcdctl endpoint health"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 10

      apisix:
        image: apache/apisix:3.15.0-debian
        ports:
          - 9180:9180
          - 9080:9080
        volumes:
          - ${{ github.workspace }}/test/e2e/apisix_conf/config.yaml:/usr/local/apisix/conf/config.yaml:ro
        options: >-
          --health-cmd "curl -sf http://localhost:9180/apisix/admin/routes || exit 1"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 10

      httpbin:
        image: ghcr.io/mccutchen/go-httpbin
        ports:
          - 8080:8080

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Wait for APISIX
        run: |
          for i in $(seq 1 30); do
            if curl -sf http://localhost:9180/apisix/admin/routes -H "X-API-KEY: edd1c9f034335f136f87ad84b625c8f1"; then
              echo "APISIX is ready"
              exit 0
            fi
            echo "Waiting for APISIX... ($i/30)"
            sleep 2
          done
          echo "APISIX failed to start"
          exit 1

      - name: Run E2E tests
        run: make test-e2e
        env:
          APISIX_ADMIN_URL: http://localhost:9180
          APISIX_GATEWAY_URL: http://localhost:9080
          HTTPBIN_URL: http://localhost:8080

      - name: Upload APISIX logs on failure
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: apisix-logs
          path: /tmp/apisix-logs/
```

> **Note on service container volumes**: GitHub Actions service containers may not support volume mounts from `${{ github.workspace }}` in all runners. If this is the case, the APISIX config must be injected via a startup script or the workflow must use `docker run` in a step instead of the `services:` block. The implementer should verify this and adjust accordingly — either use a step-based container approach or copy the config into the container after checkout.

#### `test/e2e/docker-compose.yml` — Local Dev

```yaml
version: "3"

services:
  apisix:
    image: apache/apisix:${APISIX_DOCKER_TAG:-3.15.0-debian}
    restart: always
    volumes:
      - ./apisix_conf/config.yaml:/usr/local/apisix/conf/config.yaml:ro
    depends_on:
      - etcd
    ports:
      - "9180:9180/tcp"
      - "9080:9080/tcp"
    networks:
      - a6-test

  etcd:
    image: bitnamilegacy/etcd:3.6
    restart: always
    environment:
      ETCD_DATA_DIR: /etcd_data
      ETCD_ENABLE_V2: "true"
      ALLOW_NONE_AUTHENTICATION: "yes"
      ETCD_ADVERTISE_CLIENT_URLS: "http://etcd:2379"
      ETCD_LISTEN_CLIENT_URLS: "http://0.0.0.0:2379"
    ports:
      - "2379:2379/tcp"
    networks:
      - a6-test

  httpbin:
    image: ghcr.io/mccutchen/go-httpbin
    restart: always
    ports:
      - "8080:8080/tcp"
    networks:
      - a6-test

networks:
  a6-test:
    driver: bridge
```

#### `test/e2e/setup_test.go` — Framework

```go
//go:build e2e

package e2e

import (
    "bytes"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
)

var (
    binaryPath string
    adminURL   string
    gatewayURL string
    httpbinURL string
    adminKey   = "edd1c9f034335f136f87ad84b625c8f1"
)

func TestMain(m *testing.M) {
    // Resolve environment
    adminURL = envOrDefault("APISIX_ADMIN_URL", "http://localhost:9180")
    gatewayURL = envOrDefault("APISIX_GATEWAY_URL", "http://localhost:9080")
    httpbinURL = envOrDefault("HTTPBIN_URL", "http://localhost:8080")

    // Build the a6 binary
    tmpDir, err := os.MkdirTemp("", "a6-e2e-*")
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
        os.Exit(1)
    }
    defer os.RemoveAll(tmpDir)

    binaryPath = filepath.Join(tmpDir, "a6")
    buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/a6")
    buildCmd.Stdout = os.Stdout
    buildCmd.Stderr = os.Stderr
    if err := buildCmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "failed to build a6: %v\n", err)
        os.Exit(1)
    }

    // Wait for APISIX to be healthy
    if err := waitForHealthy(adminURL+"/apisix/admin/routes", 30*time.Second); err != nil {
        fmt.Fprintf(os.Stderr, "APISIX not ready: %v\n", err)
        os.Exit(1)
    }

    os.Exit(m.Run())
}

// runA6 executes the a6 binary with the given arguments and returns stdout, stderr, and error.
func runA6(args ...string) (string, string, error) {
    cmd := exec.Command(binaryPath, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    return stdout.String(), stderr.String(), err
}

// adminAPI sends a request to the APISIX Admin API. Used for setup/cleanup.
func adminAPI(method, path string, body []byte) (*http.Response, error) {
    var bodyReader *bytes.Reader
    if body != nil {
        bodyReader = bytes.NewReader(body)
    }
    var req *http.Request
    var err error
    if bodyReader != nil {
        req, err = http.NewRequest(method, adminURL+path, bodyReader)
    } else {
        req, err = http.NewRequest(method, adminURL+path, nil)
    }
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-KEY", adminKey)
    req.Header.Set("Content-Type", "application/json")
    return http.DefaultClient.Do(req)
}

func waitForHealthy(url string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        req, _ := http.NewRequest(http.MethodGet, url, nil)
        req.Header.Set("X-API-KEY", adminKey)
        resp, err := http.DefaultClient.Do(req)
        if err == nil && resp.StatusCode < 400 {
            resp.Body.Close()
            return nil
        }
        time.Sleep(1 * time.Second)
    }
    return fmt.Errorf("timeout waiting for %s", url)
}

func envOrDefault(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

#### `test/e2e/smoke_test.go` — Smoke Test

```go
//go:build e2e

package e2e

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSmoke_BinaryRuns(t *testing.T) {
    stdout, _, err := runA6("--help")
    require.NoError(t, err)
    assert.Contains(t, stdout, "a6")
}

func TestSmoke_APISIXReachable(t *testing.T) {
    resp, err := adminAPI("GET", "/apisix/admin/routes", nil)
    require.NoError(t, err)
    defer resp.Body.Close()
    assert.Equal(t, 200, resp.StatusCode)
}
```

#### Makefile Additions

```makefile
## test-e2e: Run end-to-end tests (requires running APISIX)
test-e2e:
	go test -v -tags e2e -count=1 -timeout 5m ./test/e2e/...

## docker-up: Start local APISIX stack for e2e development
docker-up:
	docker-compose -f test/e2e/docker-compose.yml up -d
	@echo "Waiting for APISIX..."
	@for i in $$(seq 1 30); do \
		curl -sf http://localhost:9180/apisix/admin/routes -H "X-API-KEY: edd1c9f034335f136f87ad84b625c8f1" > /dev/null 2>&1 && break; \
		sleep 2; \
	done
	@echo "APISIX is ready"

## docker-down: Stop local APISIX stack
docker-down:
	docker-compose -f test/e2e/docker-compose.yml down -v
```

#### E2E Test Scenarios

| Test | Description |
|------|-------------|
| `TestSmoke_BinaryRuns` | `a6 --help` exits 0 and outputs help text |
| `TestSmoke_APISIXReachable` | Admin API GET /routes returns 200 |

#### Documentation Updates

- Update `AGENTS.md`: Add e2e test section to document map and commands table
- Update `docs/testing-strategy.md`: Add E2E test section describing the pattern, build tags, and how to run locally

#### Commit Message

```
feat(ci): add CI workflows and e2e testing infrastructure

Add GitHub Actions workflows for unit tests, linting, and e2e tests
running against real APISIX service containers. Include docker-compose
for local e2e development and smoke tests to verify the setup.
```

---

### PR-2: Foundation Packages

**Goal**: Implement the shared packages that all commands depend on: Factory, IOStreams, API Client, Config, and output utilities.

**Depends on**: PR-1

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/cmd/factory.go` | Factory struct (IOStreams, HttpClient, Config) |
| `pkg/iostreams/iostreams.go` | I/O abstraction with TTY detection |
| `pkg/api/client.go` | HTTP client with auth transport |
| `pkg/api/types.go` | Shared API types (ListResponse, APIError) |
| `internal/config/config.go` | Config file read/write, context management |
| `internal/version/version.go` | Build version variables |
| `pkg/cmdutil/exporter.go` | JSON/YAML/table output helper |
| `pkg/cmdutil/errors.go` | Error formatting utilities |
| `pkg/tableprinter/table.go` | Table rendering with color support |
| `pkg/httpmock/httpmock.go` | HTTP mock registry for unit tests |

#### Files to Create (Tests)

| File | Purpose |
|------|---------|
| `pkg/iostreams/iostreams_test.go` | IOStreams unit tests |
| `pkg/api/client_test.go` | API client unit tests |
| `internal/config/config_test.go` | Config read/write tests |
| `pkg/cmdutil/exporter_test.go` | Output format tests |
| `pkg/httpmock/httpmock_test.go` | Mock registry tests |

#### Files to Modify

| File | Change |
|------|--------|
| `cmd/a6/main.go` | Wire Factory with real IOStreams, HttpClient, Config |
| `pkg/cmd/root/root.go` | Accept Factory, add global flags (--output, --context, --server, --api-key, --verbose, --force) |
| `go.mod` | Add dependencies (testify, yaml.v3, viper) |

#### Key Design Decisions

1. **Config file**: `~/.config/a6/config.yaml` (respects `XDG_CONFIG_HOME` and `A6_CONFIG_DIR`)
2. **Auth precedence**: `--api-key` flag > `A6_API_KEY` env > context config
3. **Server precedence**: `--server` flag > `A6_SERVER` env > context config
4. **API Client**: Thin `net/http` wrapper with `apiKeyTransport` RoundTripper
5. **Output**: `--output` flag on root command, inherited by all subcommands

#### Implementation Notes

- Follow `docs/golden-example.md` exactly for Factory, IOStreams, and API Client patterns
- The `internal/config/config.go` must implement a `Config` interface for testability
- `pkg/httpmock/httpmock.go` must support `Register(method, path, response)`, `GetClient()`, and `Verify(t)`
- All packages must have unit tests; no e2e tests in this PR (no commands to test yet)

#### Documentation Updates

- No user-facing doc changes (internal infrastructure only)
- Update `AGENTS.md` if any package paths differ from the planned structure

#### Commit Message

```
feat(core): add foundation packages — Factory, IOStreams, API client, Config

Implement shared infrastructure used by all commands: dependency
injection via Factory, terminal I/O abstraction, HTTP client with
admin API auth, YAML config management, and output formatting utilities.
```

---

### PR-3: Context Management

**Goal**: Implement `a6 context create|use|list|delete|current` for managing connections to multiple APISIX instances.

**Depends on**: PR-2

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/cmd/context/context.go` | Parent `context` command |
| `pkg/cmd/context/create/create.go` | `a6 context create <name> --server --api-key` |
| `pkg/cmd/context/create/create_test.go` | Unit tests |
| `pkg/cmd/context/use/use.go` | `a6 context use <name>` |
| `pkg/cmd/context/use/use_test.go` | Unit tests |
| `pkg/cmd/context/list/list.go` | `a6 context list` |
| `pkg/cmd/context/list/list_test.go` | Unit tests |
| `pkg/cmd/context/delete/delete.go` | `a6 context delete <name>` |
| `pkg/cmd/context/delete/delete_test.go` | Unit tests |
| `pkg/cmd/context/current/current.go` | `a6 context current` |
| `pkg/cmd/context/current/current_test.go` | Unit tests |
| `docs/user-guide/getting-started.md` | Getting started guide with context setup |
| `docs/user-guide/configuration.md` | Config file format and context management docs |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `context` command |

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/context_test.go` | `TestContext_CreateAndUse` | Create a context, set it active, verify with `current` |
| | `TestContext_List` | Create multiple contexts, verify list output |
| | `TestContext_Delete` | Create → delete → verify gone from list |
| | `TestContext_CreateDuplicate` | Create same name twice → error |
| | `TestContext_UseNonExistent` | Use non-existent context → error |
| | `TestContext_DeleteActive` | Delete the active context → error or warning |

#### Implementation Notes

- Context commands are local-only (no APISIX API calls) — they read/write `~/.config/a6/config.yaml`
- E2E tests should use a temporary config directory via `A6_CONFIG_DIR` env var
- `create` must accept `--server` and `--api-key` flags
- `list` output: table with columns NAME, SERVER, ACTIVE (star marker for current)
- `delete` must prompt for confirmation if deleting the active context (unless `--force`)

#### Documentation Updates

- Create `docs/user-guide/getting-started.md`: Installation, first context creation, first API call
- Create `docs/user-guide/configuration.md`: Config file location, format, env vars, context commands

#### Commit Message

```
feat(context): add context management commands

Implement create, use, list, delete, and current subcommands for
managing connections to multiple APISIX instances. Includes e2e tests
and user documentation.
```

---

### PR-4: Route CRUD

**Goal**: Implement `a6 route list|get|create|update|delete` — the first resource command and the pattern template for all subsequent resources.

**Depends on**: PR-3

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_route.go` | Route Go types matching Admin API schema |
| `pkg/cmd/route/route.go` | Parent `route` command |
| `pkg/cmd/route/list/list.go` | `a6 route list` (THE golden example) |
| `pkg/cmd/route/list/list_test.go` | Unit tests (TTY, non-TTY, filter, error) |
| `pkg/cmd/route/get/get.go` | `a6 route get <id>` |
| `pkg/cmd/route/get/get_test.go` | Unit tests |
| `pkg/cmd/route/create/create.go` | `a6 route create -f file.json` |
| `pkg/cmd/route/create/create_test.go` | Unit tests |
| `pkg/cmd/route/update/update.go` | `a6 route update <id> -f file.json` |
| `pkg/cmd/route/update/update_test.go` | Unit tests |
| `pkg/cmd/route/delete/delete.go` | `a6 route delete <id>` |
| `pkg/cmd/route/delete/delete_test.go` | Unit tests |
| `test/fixtures/route_list.json` | List response fixture |
| `test/fixtures/route_get.json` | Single route fixture |
| `docs/user-guide/route.md` | Route command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `route` command |
| `pkg/api/client.go` | Add `Post`, `Put`, `Patch`, `Delete` methods (if not already) |

#### Flags

| Command | Flags |
|---------|-------|
| `route list` | `--page`, `--page-size`, `--name`, `--label`, `--uri`, `--output` |
| `route get` | `--output` |
| `route create` | `-f/--file` (required), `--output` |
| `route update` | `-f/--file` (required), `--output` |
| `route delete` | `--force` (skip confirmation) |

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/route_test.go` | `TestRoute_CRUD` | Full lifecycle: create → get → list → update → delete |
| | `TestRoute_CreateFromFile` | Create route from JSON file, verify via get |
| | `TestRoute_ListEmpty` | List routes when none exist → empty result |
| | `TestRoute_ListWithFilters` | Create named routes, filter by `--name` |
| | `TestRoute_GetNonExistent` | Get non-existent ID → error with helpful message |
| | `TestRoute_DeleteNonExistent` | Delete non-existent ID → error |
| | `TestRoute_DeleteWithForce` | Delete with `--force` skips confirmation |
| | `TestRoute_JSONOutput` | Verify `--output json` produces valid JSON |
| | `TestRoute_YAMLOutput` | Verify `--output yaml` produces valid YAML |
| | `TestRoute_TrafficForwarding` | Create route with httpbin upstream, curl gateway → 200 |

#### Key Test: Traffic Forwarding

This test validates the full data plane path:
1. Create a route: `uri: /test-httpbin`, upstream nodes: `httpbin:8080`
2. Curl `http://localhost:9080/test-httpbin/get`
3. Verify 200 response from httpbin
4. Cleanup: delete the route

> **Note**: In GitHub Actions, the httpbin service name resolves within the Docker network. Since the APISIX container runs as a service, its upstream must use the service name `httpbin` (not `localhost`). The e2e test creates the route via Admin API with `"nodes": {"httpbin:8080": 1}`.

#### Documentation Updates

- Create `docs/user-guide/route.md`: Full route command reference with synopsis, flags, and examples
- Update `docs/user-guide/getting-started.md`: Add "Your first route" example after context setup

#### Commit Message

```
feat(route): add route CRUD commands

Implement list, get, create, update, and delete for APISIX routes.
This is the canonical resource command pattern — all subsequent
resource commands follow this structure. Includes e2e tests with
traffic forwarding validation and user documentation.
```

---

### PR-5: Upstream CRUD

**Goal**: Implement `a6 upstream list|get|create|update|delete`.

**Depends on**: PR-4

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_upstream.go` | Upstream Go types |
| `pkg/cmd/upstream/upstream.go` | Parent `upstream` command |
| `pkg/cmd/upstream/list/list.go` | `a6 upstream list` |
| `pkg/cmd/upstream/list/list_test.go` | Unit tests |
| `pkg/cmd/upstream/get/get.go` | `a6 upstream get <id>` |
| `pkg/cmd/upstream/get/get_test.go` | Unit tests |
| `pkg/cmd/upstream/create/create.go` | `a6 upstream create -f` |
| `pkg/cmd/upstream/create/create_test.go` | Unit tests |
| `pkg/cmd/upstream/update/update.go` | `a6 upstream update <id> -f` |
| `pkg/cmd/upstream/update/update_test.go` | Unit tests |
| `pkg/cmd/upstream/delete/delete.go` | `a6 upstream delete <id>` |
| `pkg/cmd/upstream/delete/delete_test.go` | Unit tests |
| `test/fixtures/upstream_list.json` | List response fixture |
| `test/fixtures/upstream_get.json` | Single upstream fixture |
| `docs/user-guide/upstream.md` | Upstream command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `upstream` command |

#### Table Columns (list)

`ID | NAME | TYPE | NODES | SCHEME | STATUS`

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/upstream_test.go` | `TestUpstream_CRUD` | Full lifecycle: create → get → list → update → delete |
| | `TestUpstream_CreateWithNodes` | Create upstream with multiple nodes, verify via get |
| | `TestUpstream_DeleteInUse` | Delete upstream referenced by a route → error (unless `--force`) |
| | `TestUpstream_ListWithFilters` | Create named upstreams, filter by `--name` |
| | `TestUpstream_RouteWithUpstreamID` | Create upstream → create route with `upstream_id` → verify traffic forwarding |

#### Documentation Updates

- Create `docs/user-guide/upstream.md`

#### Commit Message

```
feat(upstream): add upstream CRUD commands

Implement list, get, create, update, and delete for APISIX upstreams.
Includes e2e tests with upstream-route integration and user documentation.
```

---

### PR-6: Service CRUD

**Goal**: Implement `a6 service list|get|create|update|delete`.

**Depends on**: PR-4

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_service.go` | Service Go types |
| `pkg/cmd/service/service.go` | Parent `service` command |
| `pkg/cmd/service/list/list.go` | `a6 service list` |
| `pkg/cmd/service/list/list_test.go` | Unit tests |
| `pkg/cmd/service/get/get.go` | `a6 service get <id>` |
| `pkg/cmd/service/get/get_test.go` | Unit tests |
| `pkg/cmd/service/create/create.go` | `a6 service create -f` |
| `pkg/cmd/service/create/create_test.go` | Unit tests |
| `pkg/cmd/service/update/update.go` | `a6 service update <id> -f` |
| `pkg/cmd/service/update/update_test.go` | Unit tests |
| `pkg/cmd/service/delete/delete.go` | `a6 service delete <id>` |
| `pkg/cmd/service/delete/delete_test.go` | Unit tests |
| `test/fixtures/service_list.json` | List response fixture |
| `test/fixtures/service_get.json` | Single service fixture |
| `docs/user-guide/service.md` | Service command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `service` command |

#### Table Columns (list)

`ID | NAME | UPSTREAM | PLUGINS | STATUS`

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/service_test.go` | `TestService_CRUD` | Full lifecycle |
| | `TestService_WithUpstream` | Create service with embedded upstream, verify via get |
| | `TestService_RouteWithServiceID` | Create service → create route with `service_id` → verify traffic |
| | `TestService_DeleteInUse` | Delete service referenced by route → error |

#### Documentation Updates

- Create `docs/user-guide/service.md`

#### Commit Message

```
feat(service): add service CRUD commands

Implement list, get, create, update, and delete for APISIX services.
Includes e2e tests with service-route integration and user documentation.
```

---

### PR-7: Consumer CRUD

**Goal**: Implement `a6 consumer list|get|create|update|delete`.

**Depends on**: PR-4

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_consumer.go` | Consumer Go types |
| `pkg/cmd/consumer/consumer.go` | Parent `consumer` command |
| `pkg/cmd/consumer/list/list.go` | `a6 consumer list` |
| `pkg/cmd/consumer/list/list_test.go` | Unit tests |
| `pkg/cmd/consumer/get/get.go` | `a6 consumer get <username>` |
| `pkg/cmd/consumer/get/get_test.go` | Unit tests |
| `pkg/cmd/consumer/create/create.go` | `a6 consumer create -f` |
| `pkg/cmd/consumer/create/create_test.go` | Unit tests |
| `pkg/cmd/consumer/update/update.go` | `a6 consumer update -f` |
| `pkg/cmd/consumer/update/update_test.go` | Unit tests |
| `pkg/cmd/consumer/delete/delete.go` | `a6 consumer delete <username>` |
| `pkg/cmd/consumer/delete/delete_test.go` | Unit tests |
| `test/fixtures/consumer_list.json` | List response fixture |
| `test/fixtures/consumer_get.json` | Single consumer fixture |
| `docs/user-guide/consumer.md` | Consumer command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `consumer` command |

#### Table Columns (list)

`USERNAME | DESC | GROUP_ID | PLUGINS | CREATED`

#### Implementation Notes

- Consumers are identified by `username`, not `id` — all commands use username as the positional argument
- Consumer API uses `PUT` for both create and update (idempotent)
- The `list` endpoint path is `/apisix/admin/consumers` (no username)
- The `get` endpoint path is `/apisix/admin/consumers/:username`

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/consumer_test.go` | `TestConsumer_CRUD` | Full lifecycle using username |
| | `TestConsumer_WithKeyAuth` | Create consumer with key-auth plugin → verify auth works via route |
| | `TestConsumer_ListEmpty` | List with no consumers |
| | `TestConsumer_GetNonExistent` | Get non-existent username → error |

#### Documentation Updates

- Create `docs/user-guide/consumer.md`

#### Commit Message

```
feat(consumer): add consumer CRUD commands

Implement list, get, create, update, and delete for APISIX consumers.
Consumers use username as identifier. Includes e2e tests with
key-auth plugin validation and user documentation.
```

---

### PR-8: SSL CRUD

**Goal**: Implement `a6 ssl list|get|create|update|delete`.

**Depends on**: PR-4

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_ssl.go` | SSL Go types |
| `pkg/cmd/ssl/ssl.go` | Parent `ssl` command |
| `pkg/cmd/ssl/list/list.go` | `a6 ssl list` |
| `pkg/cmd/ssl/list/list_test.go` | Unit tests |
| `pkg/cmd/ssl/get/get.go` | `a6 ssl get <id>` |
| `pkg/cmd/ssl/get/get_test.go` | Unit tests |
| `pkg/cmd/ssl/create/create.go` | `a6 ssl create -f` |
| `pkg/cmd/ssl/create/create_test.go` | Unit tests |
| `pkg/cmd/ssl/update/update.go` | `a6 ssl update <id> -f` |
| `pkg/cmd/ssl/update/update_test.go` | Unit tests |
| `pkg/cmd/ssl/delete/delete.go` | `a6 ssl delete <id>` |
| `pkg/cmd/ssl/delete/delete_test.go` | Unit tests |
| `test/fixtures/ssl_list.json` | List response fixture |
| `test/fixtures/ssl_get.json` | Single SSL fixture |
| `test/e2e/testdata/test.crt` | Self-signed test certificate |
| `test/e2e/testdata/test.key` | Test private key |
| `docs/user-guide/ssl.md` | SSL command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `ssl` command |

#### Table Columns (list)

`ID | SNI | STATUS | TYPE | VALIDITY | CREATED`

#### Implementation Notes

- SSL creation requires PEM certificate and key data — the `create` command should support both `-f` (JSON/YAML with embedded cert/key) and potentially `--cert`/`--key` flags for file paths
- E2E tests use self-signed certificates generated once and stored in `test/e2e/testdata/`
- Table output should show the SNI(s), not the full cert content

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/ssl_test.go` | `TestSSL_CRUD` | Full lifecycle with self-signed cert |
| | `TestSSL_CreateFromFile` | Create SSL from JSON file with cert/key |
| | `TestSSL_ListWithStatus` | Create enabled/disabled SSLs, verify status in output |
| | `TestSSL_GetShowsSNI` | Get SSL shows SNI, type, status (not raw cert) |

#### Documentation Updates

- Create `docs/user-guide/ssl.md`

#### Commit Message

```
feat(ssl): add SSL certificate CRUD commands

Implement list, get, create, update, and delete for APISIX SSL
certificates. Includes self-signed test certificates for e2e testing
and user documentation.
```

---

### PR-9: Plugin List & Get

**Goal**: Implement `a6 plugin list` and `a6 plugin get <name>`.

**Depends on**: PR-4

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/api/types_plugin.go` | Plugin Go types |
| `pkg/cmd/plugin/plugin.go` | Parent `plugin` command |
| `pkg/cmd/plugin/list/list.go` | `a6 plugin list` (lists all plugin names) |
| `pkg/cmd/plugin/list/list_test.go` | Unit tests |
| `pkg/cmd/plugin/get/get.go` | `a6 plugin get <name>` (shows plugin schema) |
| `pkg/cmd/plugin/get/get_test.go` | Unit tests |
| `test/fixtures/plugin_list.json` | Plugin list fixture |
| `test/fixtures/plugin_get.json` | Single plugin schema fixture |
| `docs/user-guide/plugin.md` | Plugin command reference |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `plugin` command |

#### Implementation Notes

- Plugin `list` uses `GET /apisix/admin/plugins/list` — returns an array of plugin name strings
- Plugin `get` uses `GET /apisix/admin/plugins/:plugin_name` — returns the plugin's JSON schema
- Plugin commands are read-only (no create/update/delete)
- `list` supports `--subsystem` flag (`http` or `stream`)
- `get` output shows the plugin schema — useful for understanding available configuration

#### Table Columns (list)

`NAME | PRIORITY | PHASE`

(If the full plugin info endpoint is used with `?all=true`, more columns are available. Otherwise, a simple name list.)

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/plugin_test.go` | `TestPlugin_List` | List plugins → non-empty list containing known plugins (limit-count, key-auth, etc.) |
| | `TestPlugin_Get` | Get `key-auth` plugin → valid JSON schema with `type`, `properties` |
| | `TestPlugin_GetNonExistent` | Get non-existent plugin → error |
| | `TestPlugin_ListStream` | List with `--subsystem stream` → returns stream plugins |

#### Documentation Updates

- Create `docs/user-guide/plugin.md`

#### Commit Message

```
feat(plugin): add plugin list and get commands

Implement read-only plugin commands: list all available plugins and
get a specific plugin's JSON schema. Includes e2e tests validating
against real APISIX plugin data and user documentation.
```

---

### PR-10: Shell Completions & Version

**Goal**: Implement `a6 completion bash|zsh|fish|powershell` and `a6 version`.

**Depends on**: PR-2

#### Files to Create

| File | Purpose |
|------|---------|
| `pkg/cmd/completion/completion.go` | Shell completion generation command |
| `pkg/cmd/completion/completion_test.go` | Unit tests (verify output contains expected shell syntax) |
| `pkg/cmd/version/version.go` | `a6 version` command |
| `pkg/cmd/version/version_test.go` | Unit tests |

#### Files to Modify

| File | Change |
|------|--------|
| `pkg/cmd/root/root.go` | Register `completion` and `version` commands |

#### Implementation Notes

- Cobra provides built-in completion generation — wrap it with consistent UX
- `a6 completion bash` outputs a script that can be `source`d or added to `.bashrc`
- `a6 version` outputs: version, commit hash, build date, Go version, OS/arch
- Version info comes from `internal/version/` set via ldflags at build time

#### E2E Test Scenarios

| File | Test | Description |
|------|------|-------------|
| `test/e2e/completion_test.go` | `TestCompletion_Bash` | `a6 completion bash` outputs valid bash completion script |
| | `TestCompletion_Zsh` | `a6 completion zsh` outputs valid zsh completion script |
| `test/e2e/version_test.go` | `TestVersion_Output` | `a6 version` outputs version info |

#### Documentation Updates

- Update `docs/user-guide/getting-started.md`: Add "Shell Completion" section with setup instructions for each shell

#### Commit Message

```
feat(completion): add shell completion and version commands

Add completion generation for bash, zsh, fish, and powershell.
Add version command displaying build info. Includes e2e tests and
setup instructions in the getting started guide.
```

---

## Phase 2 PRs ✅

Phase 2 extends a6 with remaining resources and advanced features. Each PR follows the same pattern established in Phase 1.

### PR-11: Global Rule CRUD

- **Commands**: `a6 global-rule list|get|create|update|delete`
- **API**: `/apisix/admin/global_rules`
- **Table columns**: `ID | PLUGINS | CREATED | UPDATED`
- **E2E tests**: CRUD lifecycle, verify global plugins apply to all routes
- **Docs**: `docs/user-guide/global-rule.md`

### PR-12: Stream Route CRUD

- **Commands**: `a6 stream-route list|get|create|update|delete`
- **API**: `/apisix/admin/stream_routes`
- **Table columns**: `ID | NAME | REMOTE_ADDR | SERVER_PORT | UPSTREAM | SNI`
- **E2E tests**: CRUD lifecycle (L4 traffic testing may be limited in CI)
- **Docs**: `docs/user-guide/stream-route.md`

### PR-13: Proto CRUD

- **Commands**: `a6 proto list|get|create|update|delete`
- **API**: `/apisix/admin/protos`
- **Table columns**: `ID | NAME | DESC | CREATED`
- **E2E tests**: CRUD lifecycle with protobuf content
- **Docs**: `docs/user-guide/proto.md`

### PR-14: Plugin Metadata

- **Commands**: `a6 plugin-metadata get|create|update|delete`
- **API**: `/apisix/admin/plugin_metadata/:plugin_name`
- **Note**: No `list` — metadata is keyed by plugin name
- **E2E tests**: Set and retrieve metadata for a known plugin
- **Docs**: `docs/user-guide/plugin-metadata.md`

### PR-15: Plugin Config CRUD

- **Commands**: `a6 plugin-config list|get|create|update|delete`
- **API**: `/apisix/admin/plugin_configs`
- **Table columns**: `ID | NAME | PLUGINS | CREATED`
- **E2E tests**: CRUD lifecycle, create route referencing plugin_config_id
- **Docs**: `docs/user-guide/plugin-config.md`

### PR-16: Consumer Group CRUD

- **Commands**: `a6 consumer-group list|get|create|update|delete`
- **API**: `/apisix/admin/consumer_groups`
- **Table columns**: `ID | NAME | PLUGINS | CREATED`
- **E2E tests**: CRUD lifecycle, create consumer with group_id
- **Docs**: `docs/user-guide/consumer-group.md`

### PR-17: Secret Manager CRUD

- **Commands**: `a6 secret list|get|create|update|delete`
- **API**: `/apisix/admin/secrets/:manager/:id`
- **Table columns**: `ID | MANAGER | URI/REGION | CREATED`
- **Note**: Path includes manager type (vault, aws, gcp)
- **E2E tests**: CRUD lifecycle (may need mock secret manager or skip integration)
- **Docs**: `docs/user-guide/secret.md`

### PR-18: Consumer Credential CRUD

- **Commands**: `a6 consumer credential list|get|create|update|delete`
- **API**: `/apisix/admin/consumers/:username/credentials`
- **Note**: Nested resource — requires `--consumer` flag or positional arg for username
- **E2E tests**: Create consumer → create credential → verify auth flow
- **Docs**: `docs/user-guide/consumer-credential.md`

### PR-19: Declarative Config — Dump & Validate

- **Commands**: `a6 config dump`, `a6 config validate`
- **`dump`**: Export all resources from APISIX to a YAML file
- **`validate`**: Validate a YAML file against APISIX schema endpoints
- **E2E tests**: Create resources → dump → verify YAML contains them; validate good/bad files
- **Docs**: `docs/user-guide/declarative-config.md`

### PR-20: Declarative Config — Sync & Diff

- **Commands**: `a6 config sync`, `a6 config diff`
- **`diff`**: Compare local YAML with remote APISIX state
- **`sync`**: Apply local YAML to APISIX (create/update/delete resources)
- **Depends on**: PR-19
- **E2E tests**: Create YAML → sync → verify resources exist; modify → diff → verify output; sync again → verify changes applied
- **Docs**: Update `docs/user-guide/declarative-config.md`

### PR-21: Upstream Health Check Status

- **Commands**: `a6 upstream health <id>`
- **Shows**: Health check status of upstream nodes
- **E2E tests**: Create upstream with health check config, query health status
- **Docs**: Update `docs/user-guide/upstream.md` with health subcommand

### PR-22: Interactive Mode (Fuzzy Selection)

- **Feature**: Interactive fuzzy selection for resource IDs/names
- **Example**: `a6 route delete` (no ID) → presents fuzzy picker with route list
- **Dependency**: Add `github.com/charmbracelet/bubbletea` or similar TUI library
- **E2E tests**: Limited (interactive mode is hard to e2e test) — focus on unit tests with mock I/O
- **Docs**: Update getting started guide with interactive mode section

---

## Phase 3 PRs ✅

Phase 3 adds advanced features including debug tooling, bulk operations, auto-update, and CLI extensions.

### PR-23: Debug — Request Tracing

- **Commands**: `a6 debug trace <route-id>`
- Send a test request and show plugin execution trace, response status, latency
- Fetches route from Admin API, sends probe to gateway, shows debug headers
- Supports `--method`, `--path`, `--header`, `--gateway-url`, `--control-url` flags
- **E2E tests**: Trace a route, verify output contains status and latency
- **Docs**: Updated `docs/user-guide/getting-started.md` with debug trace section

### PR-24: Debug — Log Streaming

- **Commands**: `a6 debug logs`
- Stream APISIX error/access logs from Docker containers via `docker logs`
- Supports `--container`, `--tail`, `--follow`, `--since`, `--timestamps` flags
- **E2E tests**: Verify log streaming output
- **Docs**: Updated getting started guide with log streaming examples

### PR-25: Bulk Operations

- **Commands**: `a6 <resource> delete --all|--label`, `a6 <resource> export --all|--label`
- Bulk delete and export operations for all resources that support labels
- `--all` and `--label key=value` filtering with `--force` for confirmation skip
- **E2E tests**: Bulk delete by label, bulk export with JSON/YAML output
- **Docs**: `docs/user-guide/bulk-operations.md`, updated getting started guide

### PR-26: Auto-Update

- **Commands**: `a6 update`
- Self-update from GitHub releases with platform-specific binary replacement
- Background update check with 24h cache, non-blocking notice on startup
- Archive detection (tar.gz/zip), atomic binary replacement, rollback on failure
- Zero new dependencies — stdlib only
- **E2E tests**: Version output, update check behavior
- **Docs**: `docs/user-guide/auto-update.md`, updated getting started guide

### PR-27: CLI Extension System

- **Commands**: `a6 extension install|list|upgrade|remove` (alias: `a6 ext`)
- GitHub-centric extension management — download platform-specific binaries from GitHub Releases
- Extensions become top-level `a6` commands via subprocess execution with `DisableFlagParsing`
- Manifest-based tracking in `~/.config/a6/extensions/`
- DI-friendly Manager struct with injectable `fetchRelease` and `downloadAsset`
- Zero new dependencies
- **E2E tests**: List empty, install invalid format, remove not found, help output, alias
- **Docs**: `docs/user-guide/extensions.md`, updated getting started guide

---

## Phase 4 — AI Agent Skills

Phase 4 adds AI agent skill files (`SKILL.md`) that enable AI coding agents to work effectively with APISIX through the a6 CLI. See `docs/skills.md` for the full skill format specification.

### PR-28: Skills Infrastructure + Shared Skill

**Goal**: Establish the skills directory structure, CI validation, and the first shared skill.

#### Files to Create

| File | Purpose |
|------|---------|
| `skills/a6-shared/SKILL.md` | Core shared skill — project conventions, patterns, development workflow |
| `scripts/validate-skills.sh` | CI validation script for SKILL.md frontmatter |
| `docs/skills.md` | Skills format, taxonomy, and authoring guide |

#### Files to Modify

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | Add `validate-skills` job |
| `Makefile` | Add `validate-skills` target |
| `AGENTS.md` | Add skills directory to project structure and document map |
| `docs/roadmap.md` | Add PR-28 through PR-36 |

#### Validation Rules

- Required frontmatter fields: `name`, `description`
- `name` must match directory name
- `name` must follow kebab-case: `^[a-z0-9]+(-[a-z0-9]+)*$`
- `description` must be non-empty

### PR-29: Authentication Plugin Skills

- **Skills**: `a6-plugin-key-auth`, `a6-plugin-jwt-auth`, `a6-plugin-basic-auth`, `a6-plugin-hmac-auth`, `a6-plugin-openid-connect`
- **Type**: Plugin skills covering APISIX authentication plugins
- **Content**: Plugin overview, config reference, step-by-step enable/configure, consumer binding, common patterns

### PR-30: Security + Rate Limiting Skills

- **Skills**: `a6-plugin-ip-restriction`, `a6-plugin-cors`, `a6-plugin-limit-count`, `a6-plugin-limit-req`
- **Type**: Plugin skills covering security and rate limiting
- **Content**: Plugin overview, config reference, enable/configure steps, common patterns

### PR-31: Traffic + Transformation Skills

- **Skills**: `a6-plugin-proxy-rewrite`, `a6-plugin-response-rewrite`, `a6-plugin-traffic-split`, `a6-plugin-redirect`, `a6-plugin-grpc-transcode`
- **Type**: Plugin skills covering traffic management and request/response transformation
- **Content**: Plugin overview, rewrite rules, weighted traffic splitting, redirect patterns

### PR-32: Operational Recipe Skills

- **Skills**: `a6-recipe-blue-green`, `a6-recipe-canary`, `a6-recipe-circuit-breaker`, `a6-recipe-health-check`, `a6-recipe-mtls`
- **Type**: Recipe skills for common operational tasks
- **Content**: Goal, prerequisites, step-by-step with a6 commands, verification, rollback

### PR-33: AI Gateway Skills

- **Skills**: `a6-plugin-ai-proxy`, `a6-plugin-ai-prompt-template`, `a6-plugin-ai-prompt-decorator`, `a6-plugin-ai-content-moderation`
- **Type**: Plugin skills for APISIX AI Gateway features
- **Content**: AI proxy config (OpenAI/Azure/Claude), prompt management, content filtering

### PR-34: Observability Skills

- **Skills**: `a6-plugin-prometheus`, `a6-plugin-skywalking`, `a6-plugin-zipkin`, `a6-plugin-http-logger`, `a6-plugin-kafka-logger`, `a6-plugin-datadog`
- **Type**: Plugin skills covering monitoring, tracing, and logging
- **Content**: Plugin config, metrics/traces/logs setup, dashboard integration

### PR-35: Advanced Plugin Skills

- **Skills**: `a6-plugin-serverless`, `a6-plugin-ext-plugin`, `a6-plugin-fault-injection`, `a6-plugin-consumer-restriction`, `a6-plugin-wolf-rbac`
- **Type**: Plugin skills for advanced/specialized plugins
- **Content**: Serverless function config, external plugin runner, fault injection patterns

### PR-36: Advanced Recipes + Personas

- **Skills**: `a6-recipe-multi-tenant`, `a6-recipe-api-versioning`, `a6-recipe-graphql-proxy`, `a6-persona-operator`, `a6-persona-developer`
- **Type**: Advanced recipes and role-based personas
- **Content**: Multi-tenant gateway setup, API version routing, GraphQL proxying, operator/developer workflow guidance

---

## Summary Table

| PR | Scope | Phase | Status | Validation |
|----|-------|-------|--------|------------|
| PR-1 | CI + E2E Infrastructure | 1 | ✅ | Smoke tests |
| PR-2 | Foundation Packages | 1 | ✅ | Unit tests only |
| PR-3 | Context Management | 1 | ✅ | Context CRUD |
| PR-4 | Route CRUD | 1 | ✅ | Full CRUD + traffic forwarding |
| PR-5 | Upstream CRUD | 1 | ✅ | Full CRUD + route integration |
| PR-6 | Service CRUD | 1 | ✅ | Full CRUD + route integration |
| PR-7 | Consumer CRUD | 1 | ✅ | Full CRUD + auth validation |
| PR-8 | SSL CRUD | 1 | ✅ | Full CRUD with self-signed certs |
| PR-9 | Plugin List/Get | 1 | ✅ | Read-only plugin queries |
| PR-10 | Completions + Version | 1 | ✅ | Output validation |
| PR-11 | Global Rule CRUD | 2 | ✅ | Full CRUD |
| PR-12 | Stream Route CRUD | 2 | ✅ | Full CRUD |
| PR-13 | Proto CRUD | 2 | ✅ | Full CRUD |
| PR-14 | Plugin Metadata | 2 | ✅ | Get/Set/Delete |
| PR-15 | Plugin Config CRUD | 2 | ✅ | Full CRUD |
| PR-16 | Consumer Group CRUD | 2 | ✅ | Full CRUD |
| PR-17 | Secret Manager CRUD | 2 | ✅ | Full CRUD |
| PR-18 | Consumer Credential | 2 | ✅ | Nested CRUD + auth |
| PR-19 | Config Dump/Validate | 2 | ✅ | Dump + validate |
| PR-20 | Config Sync/Diff | 2 | ✅ | Full declarative sync |
| PR-21 | Upstream Health | 2 | ✅ | Health status query |
| PR-22 | Interactive Mode | 2 | ✅ | Unit tests (limited e2e) |
| PR-23 | Debug Tracing | 3 | ✅ | Route trace + output validation |
| PR-24 | Debug Logs | 3 | ✅ | Log streaming output |
| PR-25 | Bulk Operations | 3 | ✅ | Bulk delete/export by label |
| PR-26 | Auto-Update | 3 | ✅ | Version + update check |
| PR-27 | CLI Extensions | 3 | ✅ | Extension management lifecycle |
| PR-28 | Skills Infra + Shared Skill | 4 | 🚧 | CI validation |
| PR-29 | Auth Plugin Skills | 4 | ⬚ | CI validation |
| PR-30 | Security + Rate Limiting Skills | 4 | ⬚ | CI validation |
| PR-31 | Traffic + Transformation Skills | 4 | ⬚ | CI validation |
| PR-32 | Operational Recipe Skills | 4 | ⬚ | CI validation |
| PR-33 | AI Gateway Skills | 4 | ⬚ | CI validation |
| PR-34 | Observability Skills | 4 | ⬚ | CI validation |
| PR-35 | Advanced Plugin Skills | 4 | ⬚ | CI validation |
| PR-36 | Advanced Recipes + Personas | 4 | ⬚ | CI validation |
