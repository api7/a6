package upstream

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/upstream/create"
	deletecmd "github.com/api7/a6/pkg/cmd/upstream/delete"
	"github.com/api7/a6/pkg/cmd/upstream/get"
	"github.com/api7/a6/pkg/cmd/upstream/health"
	"github.com/api7/a6/pkg/cmd/upstream/list"
	"github.com/api7/a6/pkg/cmd/upstream/update"
)

// NewCmdUpstream creates the `upstream` parent command.
func NewCmdUpstream(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upstream <command>",
		Short: "Manage APISIX upstreams",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(health.NewCmdHealth(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
