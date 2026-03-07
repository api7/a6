package proto

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/proto/create"
	deletecmd "github.com/api7/a6/pkg/cmd/proto/delete"
	"github.com/api7/a6/pkg/cmd/proto/get"
	"github.com/api7/a6/pkg/cmd/proto/list"
	"github.com/api7/a6/pkg/cmd/proto/update"
)

func NewCmdProto(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto <command>",
		Short: "Manage APISIX proto definitions",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
