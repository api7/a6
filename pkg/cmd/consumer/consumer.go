package consumer

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/consumer/create"
	deletecmd "github.com/api7/a6/pkg/cmd/consumer/delete"
	"github.com/api7/a6/pkg/cmd/consumer/get"
	"github.com/api7/a6/pkg/cmd/consumer/list"
	"github.com/api7/a6/pkg/cmd/consumer/update"
)

// NewCmdConsumer creates the `consumer` parent command.
func NewCmdConsumer(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer <command>",
		Short: "Manage APISIX consumers",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
