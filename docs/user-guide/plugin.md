# Plugin Management

The `a6 plugin` command allows you to inspect Apache APISIX plugins. Plugin commands are read-only and do not create, update, or delete plugin resources.

## Commands

### `a6 plugin list`

Lists available plugins in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--subsystem` | | | Filter by subsystem (`http` or `stream`) |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all HTTP plugins:
```bash
a6 plugin list
```

List stream plugins:
```bash
a6 plugin list --subsystem stream
```

Output plugin list in JSON format:
```bash
a6 plugin list -o json
```

### `a6 plugin get <name>`

Gets the schema for a specific plugin by name.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `json` | Output format (json, yaml) |

**Examples:**

Get schema for `key-auth` plugin:
```bash
a6 plugin get key-auth
```

Get schema in YAML format:
```bash
a6 plugin get key-auth -o yaml
```

## Notes

- `a6 plugin list` and `a6 plugin get` are read-only commands.
- Plugin list returns a plain array of plugin names.
- Plugin get returns the raw plugin schema.
