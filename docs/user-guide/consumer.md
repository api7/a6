# Consumer Management

The `a6 consumer` command allows you to manage Apache APISIX consumers. You can list, create, update, get, and delete consumers using the CLI.

## Commands

### `a6 consumer list`

Lists all consumers in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all consumers:
```bash
a6 consumer list
```

Output in JSON format:
```bash
a6 consumer list -o json
```

Paginated output:
```bash
a6 consumer list --page 2 --page-size 5
```

### `a6 consumer get`

Gets detailed information about a specific consumer by its username.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get consumer by username:
```bash
a6 consumer get my-consumer
```

Get consumer in JSON format:
```bash
a6 consumer get my-consumer -o json
```

### `a6 consumer create`

Creates a new consumer from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the consumer configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create a consumer from a JSON file:
```bash
a6 consumer create -f consumer.json
```

Create a consumer from a YAML file:
```bash
a6 consumer create -f consumer.yaml
```

**Sample `consumer.json`:**
```json
{
  "username": "my-consumer",
  "desc": "My API consumer",
  "plugins": {
    "key-auth": {
      "key": "my-secret-key"
    }
  }
}
```

### `a6 consumer update`

Updates an existing consumer using a configuration file. The username is specified as an argument and will be set in the request payload.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the consumer configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update a consumer:
```bash
a6 consumer update my-consumer -f updated-consumer.json
```

### `a6 consumer delete`

Deletes a consumer by its username.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete consumer with confirmation:
```bash
a6 consumer delete my-consumer
```

Delete consumer without confirmation:
```bash
a6 consumer delete my-consumer --force
```

## Consumer Configuration Reference

Key fields in the consumer configuration:

| Field | Description |
|-------|-------------|
| `username` | Unique identifier for the consumer |
| `desc` | Human-readable description |
| `plugins` | Plugin configurations (e.g., key-auth, jwt-auth) |
| `group_id` | Reference to a consumer group |
| `labels` | Key-value labels for the consumer |

For the full schema and detailed field descriptions, refer to the [APISIX Consumer Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#consumer).

## Examples

### Consumer with key-auth

Create a consumer that authenticates via API key.

```json
{
  "username": "api-user",
  "desc": "API key authenticated user",
  "plugins": {
    "key-auth": {
      "key": "secret-api-key-123"
    }
  }
}
```

### Consumer with basic-auth

Create a consumer that authenticates via basic authentication.

```json
{
  "username": "basic-user",
  "plugins": {
    "basic-auth": {
      "username": "basic-user",
      "password": "my-password"
    }
  }
}
```

### Consumer with group

Assign a consumer to a consumer group for shared plugin configurations.

```json
{
  "username": "grouped-user",
  "group_id": "company-a",
  "plugins": {
    "key-auth": {
      "key": "grouped-key-456"
    }
  }
}
```
