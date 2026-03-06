package cmd

import (
	"net/http"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/iostreams"
)

// Factory provides shared dependencies for all commands.
// Functions are used for lazy initialization — the HTTP client and config
// are only created when a command actually needs them.
type Factory struct {
	IOStreams *iostreams.IOStreams

	// HttpClient returns a lazy-initialized, auth-injected HTTP client.
	HttpClient func() (*http.Client, error)

	// Config returns a lazy-initialized configuration reader.
	Config func() (config.Config, error)
}
