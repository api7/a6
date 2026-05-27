# Bulk Operations

Bulk operations let you delete or export multiple resources in one command.

Supported resources:

- route
- service
- upstream
- consumer
- ssl
- global-rule (does not support `--label`; APISIX does not accept labels on global rules)
- plugin-config
- consumer-group
- stream-route
- proto

## Bulk Delete

Each supported resource delete command supports `--all`. All except `global-rule` also support `--label`.

### Delete all resources

```bash
a6 route delete --all --force
a6 service delete --all --force
a6 upstream delete --all --force
```

### Delete by label

```bash
a6 route delete --label env=staging --force
a6 service delete --label team=core --force
a6 upstream delete --label app=payments --force
```

### Notes

- `--all` and `--label` are mutually exclusive.
- You cannot combine an explicit ID/username with `--all` or `--label`.
- Without `--force`, a6 prompts for confirmation before bulk deletion.
- `global-rule` only supports `--all` and ID-based deletion (no `--label`).

## Bulk Export

Each supported resource has an `export` subcommand.

### Export as YAML (default)

```bash
a6 route export
a6 service export --label env=prod
a6 upstream export --label team=platform
```

### Export as JSON

```bash
a6 route export --output json
a6 service export --output json --label env=staging
```

### Export to file

```bash
a6 route export --label env=prod --output yaml --file routes.yaml
a6 service export --output json --file services.json
```

## Flag Reference

### Delete flags

- `--all`: Delete all resources of that type.
- `--label key=value`: Delete resources matching a label.
- `--force`: Skip confirmation prompt.

### Export flags

- `--label key=value`: Export only matching resources. Not supported by `global-rule export`.
- `--output, -o`: Output format (`yaml` or `json`).
- `--file, -f`: Write output to file instead of stdout.
