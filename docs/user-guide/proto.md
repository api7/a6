# Proto Management

The `a6 proto` command allows you to manage Apache APISIX proto definitions used by gRPC-related integrations. You can list, create, update, get, and delete proto resources.

## Commands

### `a6 proto list`

Lists proto definitions in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List proto definitions:
```bash
a6 proto list
```

Paginated list:
```bash
a6 proto list --page 2 --page-size 10
```

Output in JSON:
```bash
a6 proto list -o json
```

### `a6 proto get`

Gets detailed information about a proto definition by ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get a proto definition:
```bash
a6 proto get my-proto-id
```

Get proto definition in JSON:
```bash
a6 proto get my-proto-id -o json
```

### `a6 proto create`

Creates a proto definition from a JSON or YAML file.

If the payload contains `id`, APISIX is called with `PUT /apisix/admin/protos/{id}`.
If `id` is omitted, APISIX is called with `POST /apisix/admin/protos`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the proto configuration file (required) |
| `--output` | `-o` | `yaml/json (TTY/non-TTY)` | Output format (json, yaml) |

**Examples:**

Create with explicit ID (PUT path):
```bash
a6 proto create -f proto-with-id.json
```

Create without ID (POST path):
```bash
a6 proto create -f proto-no-id.yaml
```

### `a6 proto update`

Updates an existing proto definition using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the proto configuration file (required) |
| `--output` | `-o` | `yaml/json (TTY/non-TTY)` | Output format (json, yaml) |

**Examples:**

Update proto `my-proto-id`:
```bash
a6 proto update my-proto-id -f updated-proto.json
```

### `a6 proto delete`

Deletes a proto definition by ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete with confirmation:
```bash
a6 proto delete my-proto-id
```

Delete without confirmation:
```bash
a6 proto delete my-proto-id --force
```

## Proto Configuration Reference

Key fields in proto configuration:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | No | Unique identifier for the proto definition |
| `content` | Yes | Full Protocol Buffer definition text |
| `name` | No | Human-readable name |
| `desc` | No | Description |
| `labels` | No | Metadata labels as key/value pairs |

For complete schema details, refer to the APISIX Admin API documentation.

## Sample Proto Configuration

```json
{
  "id": "helloworld-proto",
  "name": "helloworld",
  "desc": "Hello world proto definition",
  "content": "syntax = \"proto3\";\npackage helloworld;\nservice Greeter {\n  rpc SayHello (HelloRequest) returns (HelloReply) {}\n}\nmessage HelloRequest { string name = 1; }\nmessage HelloReply { string message = 1; }",
  "labels": {
    "env": "test",
    "team": "gateway"
  }
}
```

## Note: Proto and gRPC Transcoding

Proto resources are commonly referenced by the APISIX gRPC transcoding plugin. Define and manage proto schemas first, then reference them in route/plugin configuration to enable HTTP-to-gRPC transcoding.
