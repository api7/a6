package get

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	ID     string
	Output string
}

func NewCmdGet(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a stream route by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.ID = args[0]
			return getRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml")

	return cmd
}

func getRun(opts *Options) error {
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

	body, err := client.Get(fmt.Sprintf("/apisix/admin/stream_routes/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.SingleResponse[api.StreamRoute]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	format := opts.Output
	if format == "" {
		if opts.IO.IsStdoutTTY() {
			format = "yaml"
		} else {
			format = "json"
		}
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(resp.Value)
}
