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
		Short:   "List services",
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
		query["label"] = opts.Label
	}

	body, err := client.Get("/apisix/admin/services", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Service]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	services := make([]api.Service, len(resp.List))
	for i, item := range resp.List {
		services[i] = item.Value
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
		if len(services) == 0 {
			fmt.Fprintln(opts.IO.Out, "No services found.")
			return nil
		}
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "NAME", "UPSTREAM", "PLUGINS", "STATUS")
		for _, s := range services {
			id := derefStr(s.ID)
			name := derefStr(s.Name)
			upstream := derefStr(s.UpstreamID)
			if upstream == "" && s.Upstream != nil && len(s.Upstream.Nodes) > 0 {
				nodes := make([]string, 0, len(s.Upstream.Nodes))
				for k := range s.Upstream.Nodes {
					nodes = append(nodes, k)
				}
				upstream = strings.Join(nodes, ", ")
			}
			plugins := ""
			if n := len(s.Plugins); n > 0 {
				plugins = fmt.Sprintf("%d plugins", n)
			}
			status := ""
			if s.Status != nil {
				status = fmt.Sprintf("%d", *s.Status)
			}
			tp.AddRow(id, name, upstream, plugins, status)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(services)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
