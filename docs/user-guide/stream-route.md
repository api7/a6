# Stream Route Management

The `a6 stream-route` command allows you to manage Apache APISIX stream routes (L4 TCP/UDP). You can list, create, update, get, and delete stream routes using the CLI.

## Commands

### `a6 stream-route list`

Lists all stream routes in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all stream routes:
```bash
a6 stream-route list
```

Output in JSON format:
```bash
a6 stream-route list -o json
```

Paginated output:
```bash
a6 stream-route list --page 2 --page-size 5
```

### `a6 stream-route get`

Gets detailed information about a specific stream route by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get stream route by ID:
```bash
a6 stream-route get 1
```

Get stream route in JSON format:
```bash
a6 stream-route get 1 -o json
```

### `a6 stream-route create`

Creates a new stream route from a JSON or YAML file.

If the payload includes `id`, APISIX create uses `PUT /apisix/admin/stream_routes/{id}`. If `id` is not set, create uses `POST /apisix/admin/stream_routes`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the stream route configuration file (required) |
| `--output` | `-o` | `yaml/json by TTY` | Output format (json, yaml) |

**Examples:**

Create a stream route from a JSON file:
```bash
a6 stream-route create -f stream-route.json
```

Create a stream route from a YAML file:
```bash
a6 stream-route create -f stream-route.yaml
```

**Sample `stream-route.json`:**
```json
{
  "id": "stream-route-1",
  "name": "tcp-proxy",
  "server_port": 9100,
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    }
  }
}
```

### `a6 stream-route update`

Updates an existing stream route using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the stream route configuration file (required) |
| `--output` | `-o` | `yaml/json by TTY` | Output format (json, yaml) |

**Examples:**

Update stream route with ID `stream-route-1`:
```bash
a6 stream-route update stream-route-1 -f updated-stream-route.json
```

### `a6 stream-route delete`

Deletes a stream route by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete stream route with confirmation:
```bash
a6 stream-route delete stream-route-1
```

Delete stream route without confirmation:
```bash
a6 stream-route delete stream-route-1 --force
```

## Stream Route Configuration Reference

Key fields in the stream route configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the stream route |
| `name` | Human-readable name for the stream route |
| `desc` | Description of the stream route |
| `remote_addr` | Match source client address |
| `server_addr` | Match destination server address |
| `server_port` | Match destination server port |
| `sni` | Match TLS SNI |
| `upstream` | Inline upstream configuration |
| `upstream_id` | Reference to an existing upstream ID |
| `service_id` | Reference to an existing service ID |
| `plugins` | Plugin configurations |
| `protocol` | L4 protocol configuration |
| `labels` | Labels for metadata and filtering |

For the full schema and detailed field descriptions, refer to the [APISIX Stream Route Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#stream_route).

## Examples

### Basic TCP stream route

```json
{
  "id": "stream-route-1",
  "name": "basic-tcp-route",
  "server_port": 9100,
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:8080": 1
    }
  }
}
```

### Stream route with SNI matching

```json
{
  "name": "sni-route",
  "sni": "tcp.example.com",
  "server_port": 9443,
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "127.0.0.1:9000": 1
    }
  }
}
```

### Stream route with upstream_id reference

```json
{
  "name": "reuse-existing-upstream",
  "server_port": 9300,
  "upstream_id": "100"
}
```
