package delete

import (
	"bufio"
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
	"github.com/api7/a6/pkg/selector"
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
		Use:   "delete [id]",
		Short: "Delete an upstream",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ID = args[0]
			}
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")

	return cmd
}

func deleteRun(opts *Options) error {
	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("id argument is required (or run interactively in a terminal)")
		}
		id, err := selectUpstream(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete upstream %s? (y/N): ", opts.ID)
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

	_, err = client.Delete(fmt.Sprintf("/apisix/admin/upstreams/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	fmt.Fprintf(opts.IO.Out, "✓ Upstream %s deleted.\n", opts.ID)
	return nil
}

func selectUpstream(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/upstreams", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch upstreams: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Upstream]
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]selector.Item, 0, len(resp.List))
	for _, item := range resp.List {
		id := item.Key
		if item.Value.ID != nil {
			id = *item.Value.ID
		}
		if id == "" {
			continue
		}
		label := id
		if item.Value.Name != nil && *item.Value.Name != "" {
			label = fmt.Sprintf("%s (%s)", *item.Value.Name, id)
		}
		items = append(items, selector.Item{ID: id, Label: label})
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no upstreams found")
	}

	return selector.SelectOne("Select an upstream", items)
}
