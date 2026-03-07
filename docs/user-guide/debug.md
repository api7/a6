# Debug and Tracing

The `a6 debug` command group helps diagnose APISIX request behavior.

## `a6 debug trace`

Trace a request through a specific route by fetching route configuration from the Admin API, optionally reading plugin priorities from the Control API, then sending a real request through the APISIX gateway.

```bash
a6 debug trace <route-id>
```

When `<route-id>` is omitted in a terminal session, `a6` opens an interactive route picker. In non-interactive mode, route ID is required.

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--method` | | route first method or `GET` | HTTP method for the probe request |
| `--path` | | route `uri` | Request path for the probe request |
| `--header` | | | Request header in `Key: Value` format (repeatable) |
| `--body` | | | Request body for the probe request |
| `--host` | | route first host | Host header for the probe request |
| `--gateway-url` | | derived from Admin URL (`:9080`) | APISIX data plane gateway URL |
| `--control-url` | | derived from Admin URL (`:9090`) | APISIX Control API URL |
| `--output` | `-o` | `table` for TTY, `json` for non-TTY | Output format (`table`, `json`, `yaml`) |

Gateway URL precedence:

1. `--gateway-url`
2. `APISIX_GATEWAY_URL`
3. Derived from current Admin API host using port `9080`

Control URL precedence:

1. `--control-url`
2. Derived from current Admin API host using port `9090`

### Examples

Basic trace:

```bash
a6 debug trace my-route
```

POST with custom path and headers:

```bash
a6 debug trace my-route \
  --method POST \
  --path /orders \
  --header "Content-Type: application/json" \
  --header "X-Debug: 1" \
  --body '{"order_id":"123"}'
```

JSON output for scripts:

```bash
a6 debug trace my-route -o json
```

Custom gateway and control URLs:

```bash
a6 debug trace my-route \
  --gateway-url http://127.0.0.1:9080 \
  --control-url http://127.0.0.1:9090
```

### Table Output

Typical table output includes:

- Route summary (ID, URI, methods, hosts, upstream)
- Probe request summary (method and URL)
- Gateway response status and latency
- Configured plugins ordered by execution priority
- Upstream status and executed plugins (if APISIX debug headers are enabled)

If APISIX does not return `Apisix-Plugins`, the output explains that APISIX debug mode should be enabled to expose executed plugins.
