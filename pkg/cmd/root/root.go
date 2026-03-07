package root

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/completion"
	configCmd "github.com/api7/a6/pkg/cmd/config"
	consumerCmd "github.com/api7/a6/pkg/cmd/consumer"
	consumergroupCmd "github.com/api7/a6/pkg/cmd/consumergroup"
	contextCmd "github.com/api7/a6/pkg/cmd/context"
	credentialCmd "github.com/api7/a6/pkg/cmd/credential"
	debugCmd "github.com/api7/a6/pkg/cmd/debug"
	globalruleCmd "github.com/api7/a6/pkg/cmd/globalrule"
	pluginCmd "github.com/api7/a6/pkg/cmd/plugin"
	pluginconfigCmd "github.com/api7/a6/pkg/cmd/pluginconfig"
	pluginmetadataCmd "github.com/api7/a6/pkg/cmd/pluginmetadata"
	protoCmd "github.com/api7/a6/pkg/cmd/proto"
	routeCmd "github.com/api7/a6/pkg/cmd/route"
	secretCmd "github.com/api7/a6/pkg/cmd/secret"
	serviceCmd "github.com/api7/a6/pkg/cmd/service"
	sslCmd "github.com/api7/a6/pkg/cmd/ssl"
	streamrouteCmd "github.com/api7/a6/pkg/cmd/streamroute"
	upstreamCmd "github.com/api7/a6/pkg/cmd/upstream"
	versionCmd "github.com/api7/a6/pkg/cmd/version"
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
	rootCmd.AddCommand(completion.NewCmdCompletion())
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))
	rootCmd.AddCommand(consumerCmd.NewCmdConsumer(f))
	rootCmd.AddCommand(consumergroupCmd.NewCmdConsumerGroup(f))
	rootCmd.AddCommand(contextCmd.NewCmdContext(f))
	rootCmd.AddCommand(credentialCmd.NewCmdCredential(f))
	rootCmd.AddCommand(debugCmd.NewCmdDebug(f))
	rootCmd.AddCommand(globalruleCmd.NewCmdGlobalRule(f))
	rootCmd.AddCommand(pluginCmd.NewCmdPlugin(f))
	rootCmd.AddCommand(pluginconfigCmd.NewCmdPluginConfig(f))
	rootCmd.AddCommand(pluginmetadataCmd.NewCmdPluginMetadata(f))
	rootCmd.AddCommand(protoCmd.NewCmdProto(f))
	rootCmd.AddCommand(routeCmd.NewCmdRoute(f))
	rootCmd.AddCommand(secretCmd.NewCmdSecret(f))
	rootCmd.AddCommand(serviceCmd.NewCmdService(f))
	rootCmd.AddCommand(sslCmd.NewCmdSSL(f))
	rootCmd.AddCommand(streamrouteCmd.NewCmdStreamRoute(f))
	rootCmd.AddCommand(upstreamCmd.NewCmdUpstream(f))
	rootCmd.AddCommand(versionCmd.NewCmdVersion(f))

	return rootCmd
}
