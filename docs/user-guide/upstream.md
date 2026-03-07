# Upstream Management

The `a6 upstream` command allows you to manage Apache APISIX upstreams. You can list, create, update, get, and delete upstreams using the CLI.

## Commands

### `a6 upstream list`

Lists all upstreams in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--name` | | | Filter upstreams by name |
| `--label` | | | Filter upstreams by label |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all upstreams:
```bash
a6 upstream list
```

Filter upstreams by name:
```bash
a6 upstream list --name my-upstream
```

Output in JSON format:
```bash
a6 upstream list -o json
```

Paginated output:
```bash
a6 upstream list --page 2 --page-size 5
```

### `a6 upstream get`

Gets detailed information about a specific upstream by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get upstream by ID:
```bash
a6 upstream get 1
```

Get upstream in JSON format:
```bash
a6 upstream get 1 -o json
```

### `a6 upstream create`

Creates a new upstream from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the upstream configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Create an upstream from a JSON file:
```bash
a6 upstream create -f upstream.json
```

Create an upstream from a YAML file:
```bash
a6 upstream create -f upstream.yaml
```

**Sample `upstream.json`:**
```json
{
  "id": "1",
  "name": "example-upstream",
  "type": "roundrobin",
  "nodes": {
    "httpbin.org:80": 1
  }
}
```

### `a6 upstream update`

Updates an existing upstream using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the upstream configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Update upstream with ID `1`:
```bash
a6 upstream update 1 -f updated-upstream.json
```

### `a6 upstream delete`

Deletes an upstream by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete upstream with confirmation:
```bash
a6 upstream delete 1
```

Delete upstream without confirmation:
```bash
a6 upstream delete 1 --force
```

### `a6 upstream health`

Shows the health check status of upstream nodes. This command queries the APISIX Control API to retrieve the current health status of all nodes in an upstream.

**Prerequisites:**
- The upstream must have health checks configured (via the `checks` field)
- The upstream must have served at least one request

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `table` | Output format (table, json, yaml) |
| `--control-url` | | | Override the APISIX Control API URL (default: derived from server URL, port 9090) |

**Examples:**

Check upstream health status:
```bash
a6 upstream health 1
```

Output in JSON format:
```bash
a6 upstream health 1 -o json
```

Use a custom Control API URL:
```bash
a6 upstream health 1 --control-url http://apisix-control:9090
```

**Sample output:**
```
NODE             STATUS     SUCCESS  HTTP_FAIL  TCP_FAIL  TIMEOUT
52.86.68.46:80   healthy    2        0          0         0
100.24.156.8:80  unhealthy  0        5          0         0

Type: http
```

## Upstream Configuration Reference

Key fields in the upstream configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the upstream |
| `name` | Human-readable name for the upstream |
| `type` | Load balancing algorithm (roundrobin, chash, ewma, least_conn) |
| `nodes` | Backend nodes with weight (e.g., `{"host:port": weight}`) |
| `service_name` | Service name for service discovery |
| `discovery_type` | Service discovery type (e.g., dns, consul, nacos) |
| `hash_on` | Hash input for consistent hashing (vars, header, cookie, consumer) |
| `key` | Hash key when using chash balancing |
| `checks` | Health check configuration |
| `retries` | Number of retries on failure |
| `retry_timeout` | Timeout for retries in seconds |
| `timeout` | Timeout settings (connect, send, read) |
| `pass_host` | Host passing strategy (pass, node, rewrite) |
| `upstream_host` | Upstream host when pass_host is rewrite |
| `scheme` | Protocol scheme (http, https, grpc, grpcs) |
| `labels` | Key-value labels for the upstream |
| `status` | Upstream status (1 for enabled, 0 for disabled) |

For the full schema and detailed field descriptions, refer to the [APISIX Upstream Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#upstream).

## Examples

### Basic upstream

Create a simple upstream with round-robin load balancing.

```json
{
  "name": "httpbin-upstream",
  "type": "roundrobin",
  "nodes": {
    "httpbin.org:80": 1
  }
}
```

### Upstream with multiple nodes

Distribute traffic across multiple backend nodes with different weights.

```json
{
  "name": "multi-node-upstream",
  "type": "roundrobin",
  "nodes": {
    "127.0.0.1:8080": 3,
    "127.0.0.1:8081": 2,
    "127.0.0.1:8082": 1
  }
}
```

### Upstream with health checks

Configure active health checks to monitor backend health.

```json
{
  "name": "healthcheck-upstream",
  "type": "roundrobin",
  "nodes": {
    "127.0.0.1:8080": 1,
    "127.0.0.1:8081": 1
  },
  "checks": {
    "active": {
      "type": "http",
      "http_path": "/health",
      "healthy": {
        "interval": 2,
        "successes": 2
      },
      "unhealthy": {
        "interval": 1,
        "http_failures": 3
      }
    }
  }
}
```
