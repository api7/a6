# a6 GA UX Consistency Sweep

Tracks the CLI UX consistency review requested in [#41](https://github.com/api7/a6/issues/41). Mirrors a7's review pattern (a7 issues #35, #36, #37, #42, #49): for every command, scan for **doc-vs-code flag mismatches**, **silently-dropped positional args**, **inconsistent `--output` defaults**, and **confusing `--id` semantics**.

Most findings here were surfaced during the [GA Run 1 manual walkthrough](./ga-test-report.md); a few were added by widening the sweep across `--help` output for every `<resource> <action>` combination. Each finding lists its current disposition: ✅ fixed in PR, 🟡 deferred follow-up, 🟢 minor/cosmetic, 📘 PRD-side fix.

---

## Class 1 — Doc-vs-code mismatches (PRD drift)

The PRD lists commands or flags that don't exist in the binary, or that have moved.

| # | What PRD says | What the binary does | Disposition |
|---|---|---|---|
| UX-1 | `a6 consumer credential ...` (nested under consumer) | `a6 credential ...` (top-level command) | 📘 PRD update (or rename the top-level command — pick one) |
| UX-2 | `a6 ssl upload` (in §"Resource Commands") | No `upload` subcommand exists — `ssl` has only `create/delete/export/get/list/update` | 📘 PRD update or implement |
| UX-3 | `a6 schema get` (in §"Resource Commands") | No `schema` top-level command exists | 📘 PRD update or implement |
| UX-4 | `--verbose / -v` (in §"Common Flags") | The `--verbose` flag exists but no `-v` shorthand is registered | 📘 PRD update or wire shorthand |

These are doc/spec follow-ups, not code regressions, but they bite first-time users who reach for the documented command. The fix is one of: update PRD, implement the command, or rename/alias the implementation.

---

## Class 2 — Silently-dropped positional args / unknown subcommands

| # | Reproducer | Expected | Actual |
|---|---|---|---|
| UX-5 | `a6 route bogus` (typo'd subcommand) | non-zero exit, error message | exit=0, prints `route --help`. Same shape for any `<resource> <typo>` |
| UX-6 | `a6 route deletex --force` (typo'd subcommand, plausible-looking flag) | error: unknown subcommand `deletex` | exit=0, prints `route --help`; `--force` silently swallowed |

🟡 **Tracked as a follow-up.** Cobra has `SilenceErrors` defaults that can be tightened (use `cobra.OnlyValidArgs` or set `SilenceUsage: false` + return an error from the parent's `RunE`). Affects every parent command — fix is centralised in `pkg/cmd/root/root.go` and the per-resource parent files.

In the manual walkthrough, no resource was found to silently drop positionals on **valid** subcommands (Cobra's `ExactArgs` / `MaximumNArgs` reject extras correctly). The silent-success failure mode is specific to **unknown** subcommands.

---

## Class 3 — Inconsistent `--output` defaults

Verified by sweeping `--help` text across every `<resource> <action>` combination:

| Action group | `--help` advertises | Actual behaviour |
|---|---|---|
| `<resource> list`, `<resource> delete` | `json, yaml, table` | All three formats work |
| `<resource> get`, `<resource> create`, `<resource> update` | `json, yaml` | `-o table` returns `unsupported output format: table` |

🟡 **BUG-6** in the [Run 1 report](./ga-test-report.md). Deferred from PR #51 — fix needs a shared per-resource table renderer lifted out of each `list.go`. Touches 12 `get.go` files plus the matching `create.go` / `update.go` if we want table for those too.

Additional symptom of the same class: **TTY default differs between `get` and `list`.** `get` defaults to `yaml` in TTY (because table is unsupported); `list` defaults to `table`. A user piping `a6 route get | grep` and `a6 route list | grep` gets different formats — not wrong per se, but unexpected.

---

## Class 4 — `--id` / positional-id semantics

The sweep across every resource confirms there is **no `--id` flag** anywhere — ids are always positional, which is consistent with Cobra conventions. The convention split that matters is the **bracket style** in `Use:` strings:

| Resource group | Pattern | Convention | Required? |
|---|---|---|---|
| `route`, `service`, `upstream`, `ssl`, `plugin-config`, `global-rule`, `stream-route`, `proto`, `consumer-group` | `<resource> get [id]` | Square brackets | Optional — falls back to interactive selector when omitted in TTY |
| `consumer` | `consumer get [username]` | Square brackets | Optional (username is the id, just renamed) |
| `secret` | `secret get [manager/id]` | Square brackets | Optional; positional must include the `manager/` prefix |
| `plugin-metadata` | `plugin-metadata get <plugin_name>` | Angle brackets | Required |
| `credential` | `credential get <id>` | Angle brackets, **but also needs `--consumer`** flag | Required |

🟢 The square-vs-angle distinction is semantically correct (Cobra: `<x>` required, `[x]` optional). No bug here, but the cognitive load is real — readers have to remember that `plugin-metadata` and `credential` require the positional while everyone else doesn't.

Real `--id`-shape bug surfaced during Run 1 and fixed: **secret create** rejected mismatched ids in body vs. positional. See [Run 1 report — BUG-4](./ga-test-report.md#bug-4--secret-create-vaultid-rejected-by-apisix-when-file-has-a-bare-id).

One related observation that survived Run 1: **global-rule `--id`** (where the resource id is forced to equal the single plugin key inside `plugins:`). a7 logged the same finding and treated it as a real bug. We didn't separately fix it on a6 — but `a6 global-rule create -f file.yaml` with a mismatched `id:` should clearly reject rather than silently letting APISIX compute the id from the plugin name. 🟡 Tracked.

---

## Class 5 — Other UX inconsistencies surfaced during Run 1

| # | Symptom | Notes |
|---|---|---|
| UX-7 | `<resource> delete <id>` (no `--force`) **silently deletes in non-TTY** | PRD says destructive ops "must require confirmation unless `--force`." In non-TTY there's no way to confirm interactively, so a6 just deletes. Either the PRD needs to allow this carve-out or the CLI should bail out asking for `--force` when stdin isn't a TTY. |
| UX-8 | `a6 config sync --dry-run` and `a6 config diff` produce **identical output** | A `Dry-run — no changes applied.` banner on the sync side would help operators in shared shells distinguish them at a glance. |
| UX-9 | `a6 config diff` emits a line per known section even when empty | `consumer_groups: create=0 update=0 delete=0 unchanged=0` etc. on an empty diff is verbose; collapse zero-change sections by default with a `--verbose` opt-in. |
| UX-10 | `credential delete` output read `Credential 1 deleted` (literal "1") before PR #51 | ✅ Fixed in PR #51 as BUG-3. |
| UX-11 | `upstream health` couldn't parse APISIX 3.x healthcheck response shape | ✅ Fixed in PR #51 as BUG-1. |
| UX-12 | `upstream health` ignored `A6_CONTROL_URL` env var | ✅ Fixed in PR #51 as BUG-2. |
| UX-13 | `config validate` accepted unknown top-level sections silently | ✅ Fixed in PR #51 as BUG-5. |

---

## Summary

| Class | Total findings | Already fixed (PR #51) | Open / deferred |
|---|---|---|---|
| 1 — PRD drift | 4 | 0 | 4 (📘 PRD-side) |
| 2 — Silent unknown subcommand | 2 | 0 | 1 (centralised fix) |
| 3 — `--output` inconsistency | 1 class spanning ~36 commands | 0 | 1 (BUG-6 in #51 report) |
| 4 — `--id` semantics | 1 cross-resource + 1 global-rule case | 1 (BUG-4) | 1 (global-rule body-id) |
| 5 — Other UX | 7 items | 4 (BUG-1/2/3/5) | 3 |

**Net: 10 open UX findings.** All non-blocking for GA from a correctness standpoint; together they're the polish layer for the imperative CLI surface.

## Next step for #41

Per the issue body, each finding becomes a discrete sub-issue. After this PR lands, file 10 GitHub issues under #33 — one per open row — and reference this document so the audit trail stays linked. Closed-by-#51 items don't need new issues.
