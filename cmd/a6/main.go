package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/root"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

func main() {
	ios := iostreams.System()

	cfg := config.NewFileConfig()

	// Apply environment variable overrides.
	if v := os.Getenv("A6_SERVER"); v != "" {
		cfg.SetServerOverride(v)
	}
	if v := os.Getenv("A6_API_KEY"); v != "" {
		cfg.SetAPIKeyOverride(v)
	}

	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			apiKey := cfg.APIKey()
			if apiKey == "" {
				return nil, fmt.Errorf("no API key configured; use 'a6 context create' or set A6_API_KEY")
			}
			return api.NewAuthenticatedClient(apiKey), nil
		},
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	rootCmd := root.NewCmdRoot(f)

	// Wire flag overrides into config after flags are parsed.
	rootCmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		if v, _ := c.Flags().GetString("server"); v != "" {
			cfg.SetServerOverride(v)
		}
		if v, _ := c.Flags().GetString("api-key"); v != "" {
			cfg.SetAPIKeyOverride(v)
		}
		return nil
	}

	if err := rootCmd.Execute(); err != nil {
		if !cmdutil.IsSilent(err) {
			fmt.Fprintln(ios.ErrOut, err)
		}
		os.Exit(1)
	}
}
