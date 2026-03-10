package list

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/api7/a6/pkg/tableprinter"
)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	Page     int
	PageSize int
	Name     string
	Label    string
	Output   string
}

func NewCmdList(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List upstreams",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Filter by label")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml, table")

	return cmd
}

func listRun(opts *Options) error {
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

	query := map[string]string{
		"page":      fmt.Sprintf("%d", opts.Page),
		"page_size": fmt.Sprintf("%d", opts.PageSize),
	}
	if opts.Name != "" {
		query["name"] = opts.Name
	}
	if opts.Label != "" {
		query["label"] = cmdutil.NormalizeLabel(opts.Label)
	}

	body, err := client.Get("/apisix/admin/upstreams", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Upstream]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	upstreams := make([]api.Upstream, len(resp.List))
	for i, item := range resp.List {
		upstreams[i] = item.Value
	}

	format := opts.Output
	if format == "" {
		if opts.IO.IsStdoutTTY() {
			format = "table"
		} else {
			format = "json"
		}
	}

	if format == "table" {
		if len(upstreams) == 0 {
			fmt.Fprintln(opts.IO.Out, "No upstreams found.")
			return nil
		}
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "NAME", "TYPE", "NODES", "SCHEME", "STATUS")
		for _, u := range upstreams {
			id := derefStr(u.ID)
			name := derefStr(u.Name)
			typ := derefStr(u.Type)
			nodes := ""
			if len(u.Nodes) > 0 {
				keys := make([]string, 0, len(u.Nodes))
				for k := range u.Nodes {
					keys = append(keys, k)
				}
				nodes = strings.Join(keys, ", ")
			}
			scheme := derefStr(u.Scheme)
			status := ""
			if u.Status != nil {
				status = fmt.Sprintf("%d", *u.Status)
			}
			tp.AddRow(id, name, typ, nodes, scheme, status)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(upstreams)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
