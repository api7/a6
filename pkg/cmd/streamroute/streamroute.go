package streamroute

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/streamroute/create"
	deletecmd "github.com/api7/a6/pkg/cmd/streamroute/delete"
	"github.com/api7/a6/pkg/cmd/streamroute/get"
	"github.com/api7/a6/pkg/cmd/streamroute/list"
	"github.com/api7/a6/pkg/cmd/streamroute/update"
)

func NewCmdStreamRoute(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-route <command>",
		Short: "Manage APISIX stream routes",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
