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
		Short:   "List stream routes",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Page size")
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

	body, err := client.Get("/apisix/admin/stream_routes", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.StreamRoute]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	streamRoutes := make([]api.StreamRoute, len(resp.List))
	for i, item := range resp.List {
		streamRoutes[i] = item.Value
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
		if len(streamRoutes) == 0 {
			fmt.Fprintln(opts.IO.Out, "No stream routes found.")
			return nil
		}
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "NAME", "REMOTE_ADDR", "SERVER_PORT", "UPSTREAM", "SNI")
		for _, r := range streamRoutes {
			id := derefStr(r.ID)
			name := derefStr(r.Name)
			remoteAddr := derefStr(r.RemoteAddr)
			serverPort := derefInt(r.ServerPort)
			sni := derefStr(r.SNI)

			upstream := derefStr(r.UpstreamID)
			if upstream == "" && r.Upstream != nil && len(r.Upstream.Nodes) > 0 {
				nodes := make([]string, 0, len(r.Upstream.Nodes))
				for k := range r.Upstream.Nodes {
					nodes = append(nodes, k)
				}
				upstream = strings.Join(nodes, ", ")
			}

			tp.AddRow(id, name, remoteAddr, serverPort, upstream, sni)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(streamRoutes)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}
