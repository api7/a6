package delete

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	ID    string
	Force bool
}

func NewCmdDelete(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a stream route",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.ID = args[0]
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")

	return cmd
}

func deleteRun(opts *Options) error {
	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete stream route %s? (y/N): ", opts.ID)
		reader := bufio.NewReader(opts.IO.In)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(opts.IO.ErrOut, "Aborted.")
			return nil
		}
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	baseURL := cfg.BaseURL()

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, baseURL)

	_, err = client.Delete(fmt.Sprintf("/apisix/admin/stream_routes/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	fmt.Fprintf(opts.IO.Out, "✓ Stream route %s deleted.\n", opts.ID)
	return nil
}
