package plugin

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/plugin/get"
	"github.com/api7/a6/pkg/cmd/plugin/list"
)

func NewCmdPlugin(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin <command>",
		Short: "Manage APISIX plugins",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))

	return cmd
}
