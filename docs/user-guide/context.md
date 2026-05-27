# Context

A **context** is a named connection profile that tells a6 which APISIX instance to talk to. You can keep multiple contexts (e.g. `local`, `staging`, `prod`) in the same config file and switch between them at any time — similar to `kubectl config` or `gcloud config configurations`.

This page covers the `a6 context` command surface. For the underlying config file format, precedence rules, and environment-variable overrides, see [configuration.md](./configuration.md).

## Commands

```text
a6 context create <name>   Create a new context
a6 context use <name>      Switch the active context
a6 context list            List all contexts
a6 context current         Print the active context name
a6 context delete <name>   Delete a context
```

### Create

```bash
a6 context create local \
  --server http://localhost:9180 \
  --api-key edd1c9f034335f136f87ad84b625c8f1
```

The context is saved and immediately set as the current context.

| Flag | Required | Description |
|---|---|---|
| `--server` | yes | Admin API URL (`http://<host>:<admin-port>`). |
| `--api-key` | no | Admin API key. Omit if the server doesn't require auth. |

### Use

```bash
a6 context use staging
```

All subsequent commands hit the `staging` context's server until you switch again. You can override at the command level with `--context staging` or with the `A6_SERVER` / `A6_API_KEY` env vars on a single invocation.

### List

```bash
a6 context list
```

```text
NAME       SERVER                   CURRENT
local      http://localhost:9180    *
staging    https://stg.example:9180
```

The `*` marker shows the currently active context.

### Current

```bash
a6 context current
```

Prints just the name (one line). Handy for shell prompts:

```bash
PS1='[$(a6 context current)] $ '
```

### Delete

```bash
a6 context delete staging
```

You cannot delete the currently active context. Switch to a different one with `a6 context use <other>` first.

## Tips

- Contexts are stored in `~/.config/a6/config.yaml` (or `$A6_CONFIG_DIR/config.yaml` / `$XDG_CONFIG_HOME/a6/config.yaml` if set). The file is created lazily on the first `a6 context create`.
- API keys are stored in plain text. If that's a concern, omit `--api-key` and use the `A6_API_KEY` env var on each invocation instead.
- Use `--context <name>` on any command to target a different context for that one invocation without switching.

## Related

- [Configuration](./configuration.md) — config file format, env vars, override precedence.
- [Getting Started](./getting-started.md) — the first-context walkthrough.
