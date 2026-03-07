package root

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	contextCmd "github.com/api7/a6/pkg/cmd/context"
	routeCmd "github.com/api7/a6/pkg/cmd/route"
	serviceCmd "github.com/api7/a6/pkg/cmd/service"
	upstreamCmd "github.com/api7/a6/pkg/cmd/upstream"
)

// NewCmdRoot creates the root command for the a6 CLI.
func NewCmdRoot(f *cmd.Factory) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "a6",
		Short:         "Apache APISIX CLI",
		Long:          "a6 is a command-line tool for managing Apache APISIX from your terminal.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global persistent flags — inherited by all subcommands.
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: json, yaml, table")
	rootCmd.PersistentFlags().String("context", "", "Override the active context")
	rootCmd.PersistentFlags().String("server", "", "Override the APISIX server URL")
	rootCmd.PersistentFlags().String("api-key", "", "Override the API key")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("force", false, "Skip confirmation prompts")

	// Command groups.
	rootCmd.AddCommand(contextCmd.NewCmdContext(f))
	rootCmd.AddCommand(routeCmd.NewCmdRoute(f))
	rootCmd.AddCommand(serviceCmd.NewCmdService(f))
	rootCmd.AddCommand(upstreamCmd.NewCmdUpstream(f))

	return rootCmd
}
