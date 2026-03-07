# Plugin Metadata Management

The `a6 plugin-metadata` command allows you to manage Apache APISIX plugin metadata.

Plugin metadata is configured per plugin name and stored at `/apisix/admin/plugin_metadata/:plugin_name`.
Unlike other resources, plugin metadata does not have a list endpoint.

## Commands

### `a6 plugin-metadata get`

Gets metadata for a specific plugin by plugin name.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get metadata for `syslog`:
```bash
a6 plugin-metadata get syslog
```

Get metadata in JSON format:
```bash
a6 plugin-metadata get syslog -o json
```

### `a6 plugin-metadata create`

Creates or sets plugin metadata for a plugin from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the plugin metadata file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create metadata for `syslog` from JSON file:
```bash
a6 plugin-metadata create syslog -f plugin-metadata.json
```

Create metadata for `syslog` from YAML file:
```bash
a6 plugin-metadata create syslog -f plugin-metadata.yaml
```

### `a6 plugin-metadata update`

Updates plugin metadata for a plugin using a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the plugin metadata file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update metadata for `syslog`:
```bash
a6 plugin-metadata update syslog -f plugin-metadata-updated.json
```

### `a6 plugin-metadata delete`

Deletes plugin metadata for a plugin name.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete metadata for `syslog` with confirmation:
```bash
a6 plugin-metadata delete syslog
```

Delete metadata for `syslog` without confirmation:
```bash
a6 plugin-metadata delete syslog --force
```

## Example Plugin Metadata Files

JSON example:

```json
{
  "log_format": {
    "host": "$host",
    "request_id": "$request_id"
  }
}
```

YAML example:

```yaml
log_format:
  host: $host
  request_id: $request_id
```
