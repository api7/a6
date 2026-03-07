package config

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/config/dump"
	"github.com/api7/a6/pkg/cmd/config/validate"
)

func NewCmdConfig(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage declarative APISIX configuration",
	}

	cmd.AddCommand(dump.NewCmdDump(f))
	cmd.AddCommand(validate.NewCmdValidate(f))

	return cmd
}
