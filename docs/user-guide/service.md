# Service Management

The `a6 service` command allows you to manage Apache APISIX services. You can list, create, update, get, and delete services using the CLI.

## Commands

### `a6 service list`

Lists all services in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--name` | | | Filter services by name |
| `--label` | | | Filter services by label |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all services:
```bash
a6 service list
```

Filter services by name:
```bash
a6 service list --name my-service
```

Output in JSON format:
```bash
a6 service list -o json
```

Paginated output:
```bash
a6 service list --page 2 --page-size 5
```

### `a6 service get`

Gets detailed information about a specific service by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get service by ID:
```bash
a6 service get 1
```

Get service in JSON format:
```bash
a6 service get 1 -o json
```

### `a6 service create`

Creates a new service from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the service configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Create a service from a JSON file:
```bash
a6 service create -f service.json
```

Create a service from a YAML file:
```bash
a6 service create -f service.yaml
```

**Sample `service.json`:**
```json
{
  "id": "1",
  "name": "example-service",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "httpbin.org:80": 1
    }
  }
}
```

### `a6 service update`

Updates an existing service using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the service configuration file (required) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

Update service with ID `1`:
```bash
a6 service update 1 -f updated-service.json
```

### `a6 service delete`

Deletes a service by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete service with confirmation:
```bash
a6 service delete 1
```

Delete service without confirmation:
```bash
a6 service delete 1 --force
```

## Service Configuration Reference

Key fields in the service configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the service |
| `name` | Human-readable name for the service |
| `desc` | Description of the service |
| `upstream` | Inline upstream configuration |
| `upstream_id` | Reference to an existing upstream ID |
| `status` | Service status (1 for enabled, 0 for disabled) |
| `plugins` | Plugin configurations for the service |
| `hosts` | List of hostnames the service handles |
| `labels` | Key-value labels for the service |
| `enable_websocket` | Enable WebSocket proxying |

For the full schema and detailed field descriptions, refer to the [APISIX Service Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#service).

## Examples

### Basic service with upstream

Create a simple service that forwards requests to an external service.

```json
{
  "name": "httpbin-service",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "httpbin.org:80": 1
    }
  }
}
```

### Service with upstream and timeout

Configure a service with custom timeout settings.

```json
{
  "name": "api-service",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    },
    "timeout": {
      "connect": 6,
      "send": 6,
      "read": 6
    }
  }
}
```

### Service with plugins

Attach plugins to a service that will apply to all routes using it.

```json
{
  "name": "protected-service",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    }
  },
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 429
    },
    "key-auth": {}
  }
}
```
