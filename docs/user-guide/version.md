# Version

Print the a6 build information.

```bash
a6 version
```

```text
a6 version v1.2.3
  commit:    abc1234
  built:     2026-05-27T03:23:51Z
  go:        go1.22.x
  platform:  darwin/arm64
```

| Field | Source |
|---|---|
| `version` | Release tag injected at build time (`make build`, GoReleaser), or the Go module version for binaries installed with `go install github.com/api7/a6/cmd/a6@<tag>`. `dev` only for untagged local source builds. |
| `commit`  | Short Git commit the binary was built from (build-time injection, or VCS info embedded by `go build`). `unknown` when neither is available, e.g. `go install` module builds. |
| `built`   | UTC timestamp of the build (same sources as `commit`). |
| `go`      | Go toolchain used. |
| `platform` | `<GOOS>/<GOARCH>` of the binary. |

## When to use it

- **Filing a bug report.** Paste the full `a6 version` output so maintainers know exactly which build is affected.
- **Verifying an update.** After `a6 update`, re-run `a6 version` to confirm the new version is in place.
- **CI logs.** Run `a6 version` as the first step of any pipeline that uses a6, so the build is recorded against the run.

## Related

- [Auto-Update](./auto-update.md) — keep a6 itself up to date.
- [Getting Started](./getting-started.md) — installation and first run.
