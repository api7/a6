package root

import (
	"github.com/spf13/cobra"
)

// NewCmdRoot creates the root command for the a6 CLI.
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "a6",
		Short:         "Apache APISIX CLI",
		Long:          "a6 is a command-line tool for managing Apache APISIX from your terminal.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Resource commands will be registered here:
	// cmd.AddCommand(route.NewCmdRoute(f))
	// cmd.AddCommand(upstream.NewCmdUpstream(f))
	// cmd.AddCommand(service.NewCmdService(f))
	// cmd.AddCommand(consumer.NewCmdConsumer(f))
	// cmd.AddCommand(ssl.NewCmdSSL(f))
	// cmd.AddCommand(plugin.NewCmdPlugin(f))
	// cmd.AddCommand(context.NewCmdContext(f))

	return cmd
}
