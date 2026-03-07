# Route Management

The `a6 route` command allows you to manage Apache APISIX routes. You can list, create, update, get, and delete routes using the CLI.

## Commands

### `a6 route list`

Lists all routes in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--name` | | | Filter routes by name |
| `--label` | | | Filter routes by label |
| `--uri` | | | Filter routes by URI |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all routes:
```bash
a6 route list
```

Filter routes by name:
```bash
a6 route list --name my-route
```

Output in JSON format:
```bash
a6 route list -o json
```

Paginated output:
```bash
a6 route list --page 2 --page-size 5
```

### `a6 route get`

Gets detailed information about a specific route by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get route by ID:
```bash
a6 route get 1
```

Get route in JSON format:
```bash
a6 route get 1 -o json
```

### `a6 route create`

Creates a new route from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the route configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Create a route from a JSON file:
```bash
a6 route create -f route.json
```

Create a route from a YAML file:
```bash
a6 route create -f route.yaml
```

**Sample `route.json`:**
```json
{
  "id": "1",
  "name": "example-route",
  "uri": "/get",
  "methods": ["GET"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "httpbin.org:80": 1
    }
  }
}
```

### `a6 route update`

Updates an existing route using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the route configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Update route with ID `1`:
```bash
a6 route update 1 -f updated-route.json
```

### `a6 route delete`

Deletes a route by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete route with confirmation:
```bash
a6 route delete 1
```

Delete route without confirmation:
```bash
a6 route delete 1 --force
```

## Route Configuration Reference

Key fields in the route configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the route |
| `name` | Human-readable name for the route |
| `uri` | The URI pattern to match |
| `methods` | HTTP methods allowed (e.g., ["GET", "POST"]) |
| `upstream` | Inline upstream configuration |
| `upstream_id` | Reference to an existing upstream ID |
| `status` | Route status (1 for enabled, 0 for disabled) |
| `plugins` | Plugin configurations for the route |

For the full schema and detailed field descriptions, refer to the [APISIX Route Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#route).

## Examples

### Basic route with upstream

Create a simple route that forwards requests to an external service.

```json
{
  "uri": "/httpbin/*",
  "name": "httpbin-route",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "httpbin.org:80": 1
    }
  }
}
```

### Route with multiple methods

Restrict the route to specific HTTP methods.

```json
{
  "uri": "/api/data",
  "methods": ["GET", "POST", "PUT"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    }
  }
}
```

### Route with upstream_id reference

Reference an existing upstream instead of defining it inline.

```json
{
  "uri": "/service/*",
  "upstream_id": "100"
}
```
