package service

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/service/create"
	deletecmd "github.com/api7/a6/pkg/cmd/service/delete"
	"github.com/api7/a6/pkg/cmd/service/get"
	"github.com/api7/a6/pkg/cmd/service/list"
	"github.com/api7/a6/pkg/cmd/service/update"
)

// NewCmdService creates the `service` parent command.
func NewCmdService(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service <command>",
		Short: "Manage APISIX services",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
