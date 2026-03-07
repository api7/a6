# Declarative Configuration

The `a6 config` command group provides tools to export and validate APISIX declarative configuration files.

## Config File Format

The declarative config file supports YAML and JSON and uses this top-level structure:

```yaml
version: "1"
routes: []
services: []
upstreams: []
consumers: []
ssl: []
global_rules: []
plugin_configs: []
consumer_groups: []
stream_routes: []
protos: []
secrets: []
plugin_metadata: []
```

Supported resource sections:

- `routes`
- `services`
- `upstreams`
- `consumers`
- `ssl`
- `global_rules`
- `plugin_configs`
- `consumer_groups`
- `stream_routes`
- `protos`
- `secrets`
- `plugin_metadata`

Notes:

- `version` must be `"1"`.
- `create_time` and `update_time` are excluded from dumped output.
- `plugin_metadata` entries include `plugin_name` and plugin-specific metadata fields.
- `secrets` IDs use compound format such as `vault/my-vault`.

## `a6 config dump`

Dump resources from APISIX Admin API into a declarative config file.

```bash
a6 config dump [--output yaml|json] [--file output.yaml]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (`yaml`, `json`) |
| `--file` | `-f` | | Write output to file instead of stdout |

Examples:

```bash
# Dump as YAML to stdout
a6 config dump

# Dump as JSON to stdout
a6 config dump -o json

# Dump to file
a6 config dump -f apisix-config.yaml
```

## `a6 config validate`

Validate a declarative config file structure.

```bash
a6 config validate -f config.yaml
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--file` | `-f` | Yes | Path to a YAML/JSON declarative config file |

Validation checks include:

- `version` exists and equals `"1"`
- Required fields (for example: route requires `uri` or `uris`, consumer requires `username`)
- ID format validation (alphanumeric, `.`, `_`, `-`, max 64 chars)
- Duplicate ID detection within each resource type

Examples:

```bash
# Validate a YAML file
a6 config validate -f apisix-config.yaml

# Validate a JSON file
a6 config validate -f apisix-config.json
```

On success, the command prints:

```text
Config is valid
```
