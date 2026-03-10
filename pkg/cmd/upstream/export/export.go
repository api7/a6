package export

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

	Label  string
	Output string
	File   string
}

func NewCmdExport(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export upstreams as JSON or YAML",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return exportRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Label, "label", "", "Filter by label (key=value)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "yaml", "Output format: json, yaml")
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Write output to file")

	return cmd
}

func exportRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())

	items, err := fetchAll(client, opts.Label)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	upstreams := make([]api.Upstream, 0, len(items))
	for _, item := range items {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		upstreams = append(upstreams, v)
	}

	if len(upstreams) == 0 {
		fmt.Fprintln(opts.IO.ErrOut, "No upstreams found.")
		return nil
	}

	format := opts.Output
	if format == "" {
		format = "yaml"
	}

	var out io.Writer = opts.IO.Out
	if opts.File != "" {
		f, err := os.Create(opts.File)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer f.Close()
		out = f
	}

	return cmdutil.NewExporter(format, out).Write(upstreams)
}

func fetchAll(client *api.Client, label string) ([]api.ListItem[api.Upstream], error) {
	page := 1
	pageSize := 500
	items := make([]api.ListItem[api.Upstream], 0)

	for {
		query := map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		}
		if label != "" {
			query["label"] = cmdutil.NormalizeLabel(label)
		}

		body, err := client.Get("/apisix/admin/upstreams", query)
		if err != nil {
			return nil, err
		}

		var resp api.ListResponse[api.Upstream]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		items = append(items, resp.List...)
		if len(resp.List) == 0 || len(items) >= resp.Total {
			break
		}
		page++
	}

	return items, nil
}
