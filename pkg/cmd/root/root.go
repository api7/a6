package root

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
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

	// Resource commands will be registered here:
	// rootCmd.AddCommand(route.NewCmdRoute(f))
	// rootCmd.AddCommand(upstream.NewCmdUpstream(f))
	// rootCmd.AddCommand(service.NewCmdService(f))
	// rootCmd.AddCommand(consumer.NewCmdConsumer(f))
	// rootCmd.AddCommand(ssl.NewCmdSSL(f))
	// rootCmd.AddCommand(plugin.NewCmdPlugin(f))
	// rootCmd.AddCommand(context.NewCmdContext(f))

	return rootCmd
}
