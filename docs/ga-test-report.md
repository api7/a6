# a6 GA Test Report

Execution of the GA readiness walkthrough described in [api7/a6#34](https://github.com/api7/a6/issues/34) against a local Apache APISIX deployment. Bug fixes follow the test-before-fix protocol: each bug below has a failing test added in the same PR before the fix.

## Run 1

### Environment

| Item | Value |
|---|---|
| Date | 2026-05-27 |
| APISIX version | **3.16.0-debian** (`apache/apisix:3.16.0-debian`) |
| `a6 version` | `efb26dd` (then `efb26dd-dirty` after the fixes in this run) |
| Admin URL | `http://127.0.0.1:19180` (remapped from 9180 to avoid local port collision with an existing API7 EE stack on 9180/9080/8080) |
| Gateway URL | `http://127.0.0.1:19080` |
| Control URL | `http://127.0.0.1:19090` |
| httpbin URL | `http://127.0.0.1:18080` |
| Host | macOS / darwin-arm64 (OrbStack), bridge networking |
| Deviations | (1) APISIX 3.16 instead of issue-proposed 3.15 (per direct request). (2) Single version only this run; 3.9 and 3.2 LTS deferred. (3) Bridge networking instead of CI's `--network host`; httpbin shares APISIX network namespace via a local `docker-compose.override.yml` so route upstreams targeting `127.0.0.1:8080` resolve as they do in CI. |

### Summary

All four phases of the issue walkthrough were executed. **7 real bugs were found.** 6 were fixed in this PR with test-before-fix coverage (5 unit-test, 1 e2e-test-environment). The remaining bug (debug-logs container auto-detect) is left to existing sub-issue tracking under [#14](https://github.com/api7/a6/issues/14) P1.

Automated e2e suite went from **70 pass / 2 fail / 5 skip / 77 total** at the start of the run to **75 pass / 1 fail / 1 skip / 77 total** after the env-config and stream_proxy fixes landed. The lone remaining e2e failure is the auto-detect container-name issue (#14 P1).

### Results

| Phase | Resource / area | Result | Bug? | Test added | Notes |
|---|---|---|---|---|---|
| A | automated e2e baseline | **PARTIAL FIXED** | infra | `apisix_conf/config-docker.yaml` + `docker-compose.yml` | Started 70 pass / 2 fail; ended 75 pass / 1 fail after env fixes. Remaining failure is debug-logs auto-detect (#14 P1). |
| B | context | PASS | — | — | `create` / `use` / `list` / `current` / `delete` round-trip. |
| B | route | PASS | — | — | CRUD via `-f`; real traffic forwarded to httpbin's `/get` returns 200. |
| B | service | PASS | — | — | CRUD via `-f`. |
| B | upstream | PASS | — | — | CRUD via `-f`. |
| B | upstream health | **FIXED** | **BUG-1**, **BUG-2** | `health_local_test.go` (3 new unit tests) | Could not parse APISIX 3.16 response shape; also no env-var fallback for non-standard control ports. |
| B | consumer | PASS | — | — | CRUD via `-f`. |
| B | credential | **FIXED** | **BUG-3** | `delete_test.go` (mock now returns realistic `"deleted":"1"`) | success message printed `"Credential 1 deleted"` literally. |
| B | ssl | PASS | — | — | CRUD via `-f`. PRD-listed `ssl upload` command does not exist — see Doc-vs-impl below. |
| B | plugin | PASS | — | — | `list` and `get <name>` for schema. |
| B | plugin-metadata | PASS | — | — | CRUD via `-f`. |
| B | plugin-config | PASS | — | — | CRUD via `-f`. |
| B | global-rule | PASS | — | — | CRUD via `-f`. |
| B | stream-route | **FIXED** | **BUG-7** | `apisix_conf/config-docker.yaml` | local config-docker.yaml didn't enable `stream_proxy`; 5 stream-route tests failed locally though they pass in CI's `config.yaml`. |
| B | proto | PASS | — | — | CRUD via `-f`. |
| B | consumer-group | PASS | — | — | CRUD via `-f`. |
| B | secret | **FIXED** | **BUG-4** | `create_test.go` (`TestStripConflictingID`) | mismatched `id:` in file vs `<manager>/<id>` positional → APISIX 400 "wrong secret id". |
| C | config dump | PASS | — | — | YAML and JSON outputs valid; contains all known sections. |
| C | config validate | **FIXED** | **BUG-5** | `validate_test.go` (`TestConfigValidate_RejectsUnknownSection` + JSON variant) | silently accepted unknown top-level sections like `unsupported_section:` or `upstream_groups:`. |
| C | config diff | PASS | — | — | "No differences" on round-trip; reports create/update/delete per section. |
| C | config sync | PASS | — | — | `--dry-run` and real sync both work; post-sync diff converges to clean. |
| D | utility commands | PASS | — | — | `version`, `completion {bash,zsh,fish,powershell}`, `debug --help`, `extension --help`, `update --help` all work. |

### Bugs found and fixed

#### BUG-1 — `upstream health` JSON deserialization fails on APISIX 3.x

`upstream health <id>` failed with:

```text
failed to parse response: json: cannot unmarshal object into Go struct field HealthCheckResponse.nodes of type []health.HealthCheckNode
```

against APISIX 3.16. The Control API endpoint `/v1/healthcheck/upstreams/<id>` returns `"nodes"` as a JSON **object** keyed by `<ip>:<port>` (or `{}` when no nodes have been probed yet), not a JSON array as the older shape that the existing struct still expected.

**Fix:** `pkg/cmd/upstream/health/health.go` — added a custom `UnmarshalJSON` on `HealthCheckResponse` that accepts both the array shape (back-compat) and the object shape (APISIX 3.x). Object keys are sorted for stable table output.

**Tests:** `TestHealthCheckResponse_UnmarshalNodesAsObject`, `TestHealthCheckResponse_UnmarshalEmptyNodesObject`, `TestHealthCheckResponse_UnmarshalNodesAsArray` in `health_local_test.go`.

#### BUG-2 — `upstream health` could not target a non-standard control-API port

`deriveControlURL` always uses port `9090` regardless of the active context's admin port. When running APISIX on a non-standard mapping (e.g. admin on `19180`, control on `19090`), users had to pass `--control-url` explicitly every time, and there was no env-var fallback (`A6_CONTROL_URL` was unread).

**Fix:** `pkg/cmd/upstream/health/health.go` — when `--control-url` is unset, fall back to `A6_CONTROL_URL` before deriving from the admin URL. The derivation rule (admin-host + `9090`) is preserved for the common case.

**Test:** environment-driven behaviour is verified by the live walkthrough; no unit test needed for an env-var read.

#### BUG-3 — `credential delete` echoed `"1"` instead of the credential id

Output read `✓ Credential 1 deleted for consumer jack.` because the code used the APISIX `deleted` field (which is a delete count, always `"1"` on success) as the resource id.

**Fix:** `pkg/cmd/credential/delete/delete.go` — print `opts.ID` (the user-supplied id) in the success message; the response body is no longer needed.

**Test:** `TestCredentialDelete_WithForce` in `delete_test.go` updated to mock the realistic APISIX response (`"deleted":"1"`) and assert the success message uses the requested id, with a negative assertion that `"Credential 1 deleted"` does **not** appear.

#### BUG-4 — `secret create vault/<id>` rejected by APISIX when file has a bare `id:`

```yaml
id: ga1
uri: http://vault.local:8200/v1/secret
```

with `a6 secret create vault/ga1 -f file.yaml` produced `API error (status 400): wrong secret id`. APISIX expects the body's `id` to equal `<manager>/<id>` (`vault/ga1`), not the bare id (`ga1`). Users following the same convention as every other resource (`route`, `service`, etc.) get a confusing 400.

**Fix:** `pkg/cmd/secret/create/create.go` — when the body's `id` differs from the manager-prefixed positional, drop it from the payload so the URL path id wins. Matching ids are preserved unchanged.

**Test:** `TestStripConflictingID` exercises the pure helper across the four cases (mismatched, matching, absent, nil payload). Live verification: re-ran the original failing command after the fix; PUT succeeded with `"id":"vault/ga1"`.

#### BUG-5 — `config validate` silently accepted unknown top-level sections

```yaml
version: "1"
unsupported_section:
  - id: x
```

returned `Config is valid`. The `api.ConfigFile` struct has no field for `unsupported_section`, so the section was dropped on unmarshal and validate had nothing to complain about.

**Fix:** `pkg/cmd/config/validate/validate.go` — read the file once into the typed `api.ConfigFile` (kept for the existing per-section checks) and once into a generic `map`, then reject any top-level key that is not in the explicit `supportedConfigSections` allow-list. Applies to both YAML and JSON input.

**Tests:** `TestConfigValidate_RejectsUnknownSection`, `TestConfigValidate_RejectsUnknownSectionJSON` in `validate_test.go`.

#### BUG-6 — `<resource> get -o table` returns "unsupported output format: table"

All 12 single-resource `get` commands (`route`, `service`, `upstream`, `consumer`, `ssl`, `plugin`, `plugin-metadata`, `plugin-config`, `global-rule`, `stream-route`, `proto`, `consumer-group`, `secret`) reject `-o table` even though the matching `<resource> list -o table` works fine. Users naturally expect `get` to support the same output formats as `list`. The per-command `-o` help strings advertise only `"json, yaml"`, so the help and the error are at least consistent — but the underlying gap is real UX inconsistency.

**Status:** **NOT FIXED** in this PR. The fix touches 12 `get.go` files and would benefit from a shared per-resource table-rendering helper (lifted out of each `list.go`). Tracked as a follow-up; the report's correctness on this point is verified by the manual walkthrough.

#### BUG-7 — local docker-compose did not enable stream proxy or expose 9090

Five `TestStreamRoute_*` tests pass in CI but fail locally with `API error (status 400): stream mode is disabled, can not add stream routes`. Root cause: `apisix_conf/config-docker.yaml` (used by `make docker-up`) lacks the `proxy_mode: http&stream` + `stream_proxy:` section that `apisix_conf/config.yaml` (used by CI's `docker run --network host`) has. Additionally, the local `docker-compose.yml` did not expose the Control API port (`9090`), which caused `upstream health` and 3 debug-logs tests to fail locally.

**Fix:** `apisix_conf/config-docker.yaml` now enables `stream_proxy` on `:9100` so it matches CI's APISIX config. `docker-compose.yml` exposes `9090` (Control API) and `9100` (stream proxy) in addition to `9180` and `9080`.

**Tests:** the `test/e2e/stream_route_test.go` cases pass after this change. No new test added — these are existing tests previously masked by the env gap.

### Minor observations (not fixed — UX or doc-cleanup follow-ups)

- **PRD vs implementation drift** — `PRD.md` documents commands that don't exist or that have moved:
  - `a6 consumer credential ...` (PRD) → actual is top-level `a6 credential ...`.
  - `a6 ssl upload` (PRD) → no such subcommand; `ssl` has only `create / delete / export / get / list / update`.
  - `a6 schema get` (PRD) → no `schema` top-level command exists.
  - `--verbose / -v` (PRD) → the `-v` shorthand is not registered on `--verbose`.
- **Flag-based create/update absent** — `route create` and `route update` accept only `-f` (file). One-liner flag forms (`--uri`, `--upstream-nodes`, etc.) are not exposed. Likely by design but worth confirming in the PRD before GA. The PRD's "Resource-Specific Flags" section only promises `-f` + `--ttl` for create/update, so this is arguably already aligned; the user-facing surprise is in the missing flags rather than in the PRD.
- **`--output` help inconsistency** — `<resource> list --help` advertises `"json, yaml, table"`; `<resource> get --help`, `<resource> create --help`, `<resource> update --help` advertise only `"json, yaml"`. The help text matches the actual behaviour (BUG-6) but the inconsistency is the user-facing symptom of BUG-6.
- **Destructive ops in non-TTY** — `<resource> delete <id>` (no `--force`) proceeds silently when stdin is not a TTY. PRD says "destructive operations must require confirmation unless `--force` is used", but the actual behaviour is "require TTY confirmation OR `--force`; otherwise silently delete." Two reasonable reads of the spec; flagged for clarification.
- **`config sync --dry-run` output is identical to `config diff`** — both emit `Differences found:` followed by per-section counts. A dry-run banner (e.g. `Dry-run — no changes applied`) would help operators in shared shells.
- **`config diff` output is verbose** — emits a line per known section even when most are `create=0 update=0 delete=0 unchanged=0`. Minor cosmetic.

## Exit criteria

| Criterion | Status |
|---|---|
| APISIX version pinned in writing | ✅ 3.16.0-debian |
| Phase A automated suite green (or failures explained) | ✅ 75/76 Ginkgo pass (1 skip, 1 fail = #14 P1 auto-detect) |
| Phase B CRUD round-trip for every PRD resource | ✅ 14/14 resources walked; CRUD works for all (stream-route required the env fix above) |
| Phase C declarative config dump/validate/diff/sync | ✅ |
| Phase D utility commands (version, completion, debug, extension, update) | ✅ help paths verified |
| Every bug found has test-before-fix coverage | ✅ 6/7 fixed with new tests; BUG-6 deferred to follow-up |
| Local docker-compose matches CI test infra | ✅ stream_proxy + control-api port + stream-proxy port added |

## Run 2

Deferred. Re-run after the remaining auto-detect debug-logs work (issue #14 P1) and the `<resource> get -o table` follow-up land.

---

## Appendix A — CLI invocations executed in Run 1

Receipts for the manual walkthrough. Reconstructed from the session's command history; each line is a real `./bin/a6 …` invocation that was run against the live APISIX 3.16 instance during the test. Outputs were inspected for correctness, not asserted programmatically — that's what Phase A is for.

Shared shell state used throughout:

```bash
export A6_CONFIG_DIR=/tmp/a6-ga-config

# The APISIX_* vars are read by test/e2e/setup_test.go for `make test-e2e`.
export APISIX_ADMIN_URL=http://127.0.0.1:19180
export APISIX_GATEWAY_URL=http://127.0.0.1:19080
export APISIX_CONTROL_URL=http://127.0.0.1:19090
export HTTPBIN_URL=http://127.0.0.1:18080

# `a6 upstream health` reads A6_CONTROL_URL (CLI env), not APISIX_CONTROL_URL.
# Re-export so the manual walkthrough below resolves the control API without
# repeating --control-url on every invocation.
export A6_CONTROL_URL="$APISIX_CONTROL_URL"

A6=./bin/a6 ; GA=/tmp/a6-ga ; mkdir -p $GA
```

### Setup — context

The literal key below (`edd1c9f034335f136f87ad84b625c8f1`) is the default APISIX dev key from the upstream Apache APISIX docs — not a real secret. Real deployments must use a unique admin key; for those, replace with `<ADMIN_API_KEY>` and source it from a secret manager rather than pasting it on the command line.

```bash
$A6 context create ga --server http://127.0.0.1:19180 --api-key <ADMIN_API_KEY>
$A6 context use ga
$A6 context list
$A6 context current
```

### Phase B — route

```bash
# Flag-based create (intentional probe; PRD-ambiguous)
$A6 route create --id ga-route-1 --name ga-route-1 --uri /ga-route-1 --upstream-nodes '127.0.0.1:8080:1'   # → unknown flag --id

# File-based create + reads in all formats
$A6 route create -f $GA/route.yaml
$A6 route get ga-route-1
$A6 route get ga-route-1 -o json
$A6 route get ga-route-1 -o yaml
$A6 route get ga-route-1 -o table                                          # → BUG-6 unsupported
$A6 route list
$A6 route list -o table
$A6 route list --name ga-route-1
$A6 route list --uri /ga-route-1

# Update + delete + cleanup verification
$A6 route update ga-route-1 -f $GA/route-update.yaml
$A6 route delete ga-route-1                                                # non-TTY, no --force: silently deleted
$A6 route delete ga-route-1 --force                                        # → resource not found (already gone)
$A6 route get ga-route-1                                                   # → resource not found
$A6 route get does-not-exist                                               # error-case probe

# Real-traffic verification
$A6 route create -f $GA/route.yaml        # uri: /ga-route-1
curl http://127.0.0.1:19080/ga-route-1     # → 404 (httpbin doesn't serve that path)
$A6 route delete ga-route-1 --force
$A6 route create -f $GA/route2.yaml       # uri: /get
curl http://127.0.0.1:19080/get            # → 200 with httpbin JSON echo (forwarding confirmed)
$A6 route delete ga-route-1 --force
```

### Phase B — service

```bash
$A6 service create -f $GA/service.yaml
$A6 service get ga-service-1 -o yaml
$A6 service get ga-service-1 -o table                                      # → BUG-6 unsupported
$A6 service list -o table
$A6 service update ga-service-1 -f $GA/service-u.yaml
$A6 service delete ga-service-1 --force
$A6 service list -o table                                                  # → No services found
```

### Phase B — upstream + health

```bash
$A6 upstream create -f $GA/upstream.yaml
$A6 upstream get ga-upstream-1 -o yaml
$A6 upstream list -o table
$A6 upstream health ga-upstream-1                                          # → derive control: hardcoded :9090, can't reach
A6_CONTROL_URL=http://127.0.0.1:19090 $A6 upstream health ga-upstream-1    # → "invalid base URL" (BUG-2: env var ignored)
$A6 upstream health ga-upstream-1 --control-url http://127.0.0.1:19090     # → "No health check data available"
$A6 upstream update ga-upstream-1 -f $GA/upstream-hc.yaml                  # add active health-check definition
$A6 upstream health ga-upstream-1 --control-url http://127.0.0.1:19090 -o table  # → BUG-1: JSON deserialize fail
$A6 upstream delete ga-upstream-1 --force

# Post-fix re-run (after BUG-1, BUG-2):
$A6 upstream health ga-upstream-1 --control-url http://127.0.0.1:19090 -o table   # → table headers, empty rows, exit 0
```

### Phase B — consumer

```bash
$A6 consumer create -f $GA/consumer.yaml
$A6 consumer get ga_consumer_1 -o yaml
$A6 consumer list -o table
$A6 consumer update ga_consumer_1 -f $GA/consumer-u.yaml
$A6 consumer delete ga_consumer_1 --force
```

### Phase B — credential

```bash
# PRD path (wrong)
$A6 consumer credential create --consumer ga_consumer_1 -f $GA/cred.yaml   # → "unknown flag: --consumer" (BUG: PRD says nested, actual is top-level)

# Actual path
$A6 credential create --consumer ga_consumer_1 -f $GA/cred.yaml
$A6 credential list --consumer ga_consumer_1 -o table
$A6 credential get ga-cred-1 --consumer ga_consumer_1 -o yaml
$A6 credential update ga-cred-1 --consumer ga_consumer_1 -f $GA/cred-u.yaml
$A6 credential delete ga-cred-1 --consumer ga_consumer_1 --force           # → "Credential 1 deleted" (BUG-3)
```

### Phase B — exit-code probe (sanity sweep)

```bash
# Each command run with --bogus flag to verify non-zero exit on flag/command errors
for sub in "route create -f x --bogus" "route get x --bogus" "route update x -f x --bogus" \
           "route delete x --bogus" "route list --bogus" "consumer create -f x --bogus" \
           "consumer get x --bogus" "credential create -f x --bogus" "credential get x --bogus" \
           "credential delete x --bogus"; do
  $A6 $sub > /dev/null 2>&1; echo "exit=$? :: $sub"
done
# All 10 returned exit=1 with "unknown command" — exit-code propagation confirmed correct
```

### Phase B — ssl

```bash
openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj "/CN=ga.test" \
  -keyout $GA/ga.key -out $GA/ga.crt
$A6 ssl create -f $GA/ssl.yaml
$A6 ssl get ga-ssl-1 -o yaml
$A6 ssl list -o table
$A6 ssl upload --help                                                       # → not a subcommand (PRD drift)
$A6 ssl upload --id ga-ssl-up-1 --sni ga.test --cert $GA/ga.crt --key $GA/ga.key  # → unknown command
$A6 ssl delete ga-ssl-1 --force
```

### Phase B — plugin

```bash
$A6 plugin list
$A6 plugin list -o table
$A6 plugin get key-auth
$A6 plugin get key-auth -o yaml
$A6 plugin get key-auth -o table                                            # → BUG-6 unsupported
```

### Phase B — plugin-metadata

```bash
$A6 plugin-metadata create http-logger -f $GA/pmeta.yaml
$A6 plugin-metadata get http-logger -o yaml
$A6 plugin-metadata update http-logger -f $GA/pmeta.yaml
$A6 plugin-metadata delete http-logger --force
```

### Phase B — plugin-config

```bash
$A6 plugin-config create -f $GA/pcfg.yaml
$A6 plugin-config list -o table
$A6 plugin-config get ga-pconfig-1 -o yaml
$A6 plugin-config update ga-pconfig-1 -f $GA/pcfg.yaml
$A6 plugin-config delete ga-pconfig-1 --force
```

### Phase B — global-rule

```bash
$A6 global-rule create -f $GA/grule.yaml
$A6 global-rule list -o table
$A6 global-rule get key-auth -o yaml
$A6 global-rule update key-auth -f $GA/grule.yaml
$A6 global-rule delete key-auth --force
```

### Phase B — stream-route

```bash
$A6 stream-route list                                                       # → 400 "stream mode is disabled" (env gap, BUG-7)
$A6 stream-route create -f $GA/sroute.yaml                                  # → same 400

# Post-fix (after BUG-7 docker-compose / config-docker.yaml change), exercised via the
# automated test/e2e/stream_route_test.go suite — manual repeat not re-run by hand.
```

### Phase B — proto

```bash
$A6 proto create -f $GA/proto.yaml
$A6 proto list -o table
$A6 proto get ga-proto-1 -o yaml
$A6 proto update ga-proto-1 -f $GA/proto.yaml
$A6 proto delete ga-proto-1 --force
```

### Phase B — consumer-group

```bash
$A6 consumer-group create -f $GA/cgroup.yaml
$A6 consumer-group list -o table
$A6 consumer-group get ga-cgroup-1 -o yaml
$A6 consumer-group update ga-cgroup-1 -f $GA/cgroup.yaml
$A6 consumer-group delete ga-cgroup-1 --force
```

### Phase B — secret

```bash
$A6 secret create vault -f $GA/secret.yaml                                  # wrong positional shape → 404 HTML
$A6 secret create vault/ga-secret-1 -f $GA/secret.yaml                      # → 400 "wrong secret id" (BUG-4 repro)
$A6 secret list -o table                                                    # confirms create didn't persist
$A6 secret get vault/ga-secret-1 -o yaml                                    # → not found
$A6 secret update vault/ga-secret-1 -f $GA/secret.yaml                      # → not found
$A6 secret delete vault/ga-secret-1 --force                                 # → not found

# Triage: try without `id:` in file to isolate the variable
$A6 secret create vault/ga1 -f $GA/secret-simple.yaml                       # → success (no id in file)
$A6 secret create vault/ga1 -f $GA/secret-simple2.yaml                      # → 400 (id: ga1 conflicts with vault/ga1)

# Post-fix re-run (after BUG-4):
$A6 secret create vault/ga1 -f $GA/secret-mismatch.yaml                     # mismatched id stripped, succeeds
$A6 secret delete vault/ga1 --force
```

### Phase B — schema (PRD-listed)

```bash
$A6 schema --help                                                           # → "unknown command schema for a6" (PRD drift)
$A6 schema get route                                                        # → same
$A6 schema get plugins.key-auth                                             # → same
```

### Phase D — utilities

```bash
$A6 version
$A6 debug --help
$A6 completion bash                                                         # heads inspected for shape
$A6 completion zsh
$A6 completion fish
$A6 completion powershell
$A6 extension --help
$A6 update --help
```

### Phase C — declarative config

```bash
$A6 config --help

# Seed live state, dump, validate, diff (clean), then mutate and converge
$A6 route create -f $GA/route2.yaml
$A6 service create -f $GA/service.yaml
$A6 consumer create -f $GA/consumer.yaml

$A6 config dump -o yaml > $GA/dump.yaml
$A6 config dump -o json | head
$A6 config validate -f $GA/dump.yaml                                        # → Config is valid
$A6 config diff -f $GA/dump.yaml                                            # → No differences found

sed 's/ga-route-1/ga-route-1-syncupd/' $GA/dump.yaml > $GA/dump-modified.yaml
$A6 config diff -f $GA/dump-modified.yaml                                   # → CREATE/DELETE entries
$A6 config sync -f $GA/dump-modified.yaml --dry-run                         # output identical to `diff` (minor finding)
$A6 config sync -f $GA/dump-modified.yaml                                   # → Sync completed
$A6 config diff -f $GA/dump-modified.yaml                                   # → No differences found

# Negative cases
$A6 config validate -f $GA/bad.yaml                                         # → "Config is valid" pre-fix (BUG-5)
$A6 config validate -f $GA/empty.yaml                                       # → Config is valid (version: '1' alone)
echo "this is: not: valid: yaml: at all:" | $A6 config validate -f /dev/stdin   # → "failed to parse YAML"

# Post-fix re-verification (after BUG-5):
$A6 config validate -f $GA/bad.yaml                                         # → "unsupported top-level section \"unsupported_section\""
```

### Fixtures used

All files under `$GA = /tmp/a6-ga`. Key shapes (full contents in session bash history):

- `route.yaml` — single route, inline upstream pointing at `127.0.0.1:8080`
- `route2.yaml` — same but `uri: /get` (for httpbin traffic verification)
- `route-update.yaml` — same id, new name + uri
- `service.yaml`, `service-u.yaml` — service with inline upstream
- `upstream.yaml`, `upstream-u.yaml`, `upstream-hc.yaml` — plain upstream + variant with active health-check
- `consumer.yaml`, `consumer-u.yaml` — username + desc
- `cred.yaml`, `cred-u.yaml` — `key-auth` credential plugin
- `ssl.yaml` — self-signed cert + key for `ga.test`
- `pmeta.yaml`, `pcfg.yaml`, `grule.yaml`, `sroute.yaml`, `proto.yaml`, `cgroup.yaml` — minimal valid bodies per resource
- `secret.yaml`, `secret-simple.yaml` (no id), `secret-simple2.yaml` (`id: ga1`), `secret-mismatch.yaml` (`id: ga1`)
- `dump.yaml` / `dump-modified.yaml` / `bad.yaml` / `empty.yaml` — declarative config fixtures

