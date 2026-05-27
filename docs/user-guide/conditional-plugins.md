# Conditional Plugins

Most APISIX plugins are enabled by default and "just work" once you reference them from a route, service, consumer, or global rule. A few plugins are **conditional**: they ship with APISIX but are commented out in the stock `config.yaml`, and APISIX rejects requests that reference them until you enable them in the `plugins:` list.

If a plugin isn't enabled, you'll see one of these errors when you try to use it through a6:

```text
API error (status 400): unknown plugin [<name>]
API error (status 400): plugin <name> is not enabled
```

This page lists the conditional plugins a6 has surfaced during real-environment validation, how to enable them, and how that affects the test/skills coverage you'll see in this repo.

## Plugins that need enabling

| Plugin | APISIX docs | Why it's off by default | Notes |
|---|---|---|---|
| `skywalking` | [docs](https://apisix.apache.org/docs/apisix/plugins/skywalking/) | Requires an external SkyWalking OAP collector; off so APISIX doesn't ship a noisy default. | Enabled in this repo's e2e config (`test/e2e/apisix_conf/config.yaml`, `config-docker.yaml`). |
| `skywalking-logger` | [docs](https://apisix.apache.org/docs/apisix/plugins/skywalking-logger/) | Same external dependency as `skywalking`. | Enabled in this repo's e2e config. |
| `ai-content-moderation` | [docs](https://apisix.apache.org/docs/apisix/plugins/ai-content-moderation/) | Requires an external moderation backend (e.g. AWS Comprehend). | Not enabled in the e2e config; the matching skill case is `Skip()`-gated. |
| `chaitin-waf` | [docs](https://apisix.apache.org/docs/apisix/plugins/chaitin-waf/) | Requires the Chaitin SafeLine WAF sidecar. | Enabled in plugin list but no upstream to talk to in tests. |
| `splunk-hec-logging`, `kafka-logger`, `tcp-logger`, etc. | various | Require a real external sink. | Enabled in plugin list but harmless without a configured destination. |

This list reflects what a6's own e2e walkthrough has run into. APISIX's full conditional set is broader — when in doubt, check your APISIX deployment's `config.yaml` against the [stock APISIX config](https://github.com/apache/apisix/blob/master/conf/config.yaml.example) to see which plugins are commented out.

## Enabling a conditional plugin

Edit your APISIX deployment's `config.yaml` and add the plugin name to the `plugins:` list, then restart or hot-reload APISIX:

```yaml
plugins:
  - real-ip
  - cors
  - key-auth
  - skywalking          # <-- add the line
```

```bash
apisix reload
# or, with Docker:
docker restart <apisix-container>
```

After the reload, the previously-failing a6 command should succeed.

## How this maps to the test suite

The a6 e2e suite distinguishes three states for each plugin scenario:

| State | What you'll see |
|---|---|
| **Verified in real env** | Tests run against a live APISIX with the plugin enabled. These cases run on every CI pass — no skip. |
| **Conditional / skipped** | Tests are gated behind a `Skip(...)` call (see `test/e2e/`). They run if the plugin is enabled in the test APISIX, otherwise they skip with an explanatory message. Currently this covers `ai-content-moderation`, `chaitin-waf` flows where the dependency isn't installed, and a small number of license-gated scenarios. |
| **Not covered** | Plugins for which a6 ships no skill or test. They may still work — the gateway accepts them — but a6 makes no guarantees beyond the generic CRUD path. |

The full list of currently-skipped scenarios is tracked in [issue #36](https://github.com/api7/a6/issues/36). Each skip will be either un-skipped (by enabling the dependency) or removed (by dropping the scenario from the supported set) before GA.

## Related

- [Plugin commands](./plugin.md) — `a6 plugin list` / `a6 plugin get` for plugin introspection.
- [Skills](../skills.md) — the AI-agent skill format, including how plugin skills declare their conditional dependencies.
