package version

import "runtime/debug"

// These variables are set at build time via ldflags.
var (
	// Version is the semantic version (e.g., "v0.1.0").
	Version = "dev"
	// Commit is the short git commit hash.
	Commit = "unknown"
	// Date is the build date in UTC.
	Date = "unknown"
)

// init falls back to the Go module build info for any value that was not
// injected via ldflags, so that binaries installed with
// `go install github.com/api7/a6/cmd/a6@<tag>` report the module version
// instead of "dev".
func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		Version, Commit, Date = resolve(Version, Commit, Date, info)
	}
}

// resolve returns version, commit, and date with unset (default) values
// filled in from build info. Values already set via ldflags take precedence.
func resolve(version, commit, date string, info *debug.BuildInfo) (string, string, string) {
	if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}

	var revision, vcsTime string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.time":
			vcsTime = s.Value
		}
	}
	if commit == "unknown" && revision != "" {
		if len(revision) > 7 {
			revision = revision[:7]
		}
		commit = revision
	}
	if date == "unknown" && vcsTime != "" {
		date = vcsTime
	}

	return version, commit, date
}
