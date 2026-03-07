package debug

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/debug/trace"
)

func NewCmdDebug(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug and diagnose APISIX",
	}

	cmd.AddCommand(trace.NewCmdTrace(f))

	return cmd
}
