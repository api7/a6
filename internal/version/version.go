package version

// These variables are set at build time via ldflags.
var (
	// Version is the semantic version (e.g., "v0.1.0").
	Version = "dev"
	// Commit is the short git commit hash.
	Commit = "unknown"
	// Date is the build date in UTC.
	Date = "unknown"
)
