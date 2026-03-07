package pluginmetadata

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	pluginmetadataCreate "github.com/api7/a6/pkg/cmd/pluginmetadata/create"
	pluginmetadataDelete "github.com/api7/a6/pkg/cmd/pluginmetadata/delete"
	pluginmetadataGet "github.com/api7/a6/pkg/cmd/pluginmetadata/get"
	pluginmetadataUpdate "github.com/api7/a6/pkg/cmd/pluginmetadata/update"
)

func NewCmdPluginMetadata(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin-metadata <command>",
		Short: "Manage APISIX plugin metadata",
	}

	cmd.AddCommand(pluginmetadataGet.NewCmdGet(f))
	cmd.AddCommand(pluginmetadataCreate.NewCmdCreate(f))
	cmd.AddCommand(pluginmetadataUpdate.NewCmdUpdate(f))
	cmd.AddCommand(pluginmetadataDelete.NewCmdDelete(f))

	return cmd
}
