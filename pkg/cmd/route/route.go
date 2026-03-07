package route

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/route/create"
	deletecmd "github.com/api7/a6/pkg/cmd/route/delete"
	"github.com/api7/a6/pkg/cmd/route/get"
	"github.com/api7/a6/pkg/cmd/route/list"
	"github.com/api7/a6/pkg/cmd/route/update"
)

// NewCmdRoute creates the `route` parent command.
func NewCmdRoute(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route <command>",
		Short: "Manage APISIX routes",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
