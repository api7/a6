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
	All   bool
	Label string
}

func NewCmdDelete(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a global rule",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ID = args[0]
			}
			if opts.All && opts.Label != "" {
				return fmt.Errorf("--all and --label are mutually exclusive")
			}
			if opts.All && opts.ID != "" {
				return fmt.Errorf("--all cannot be used with a specific ID")
			}
			if opts.Label != "" && opts.ID != "" {
				return fmt.Errorf("--label cannot be used with a specific ID")
			}
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Delete all global rules")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Delete global rules matching label (key=value)")

	return cmd
}

func deleteRun(opts *Options) error {
	if opts.All || opts.Label != "" {
		return bulkDeleteGlobalRules(opts)
	}

	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("id argument is required (or run interactively in a terminal)")
		}
		id, err := selectGlobalrule(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete global rule %s? (y/N): ", opts.ID)
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

	_, err = client.Delete(fmt.Sprintf("/apisix/admin/global_rules/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	fmt.Fprintf(opts.IO.Out, "✓ Global rule %s deleted.\n", opts.ID)
	return nil
}

func bulkDeleteGlobalRules(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	ids, err := listAllGlobalRuleIDs(client, opts.Label)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	if len(ids) == 0 {
		fmt.Fprintln(opts.IO.ErrOut, "No global rules found.")
		return nil
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete %d global rules? (y/N): ", len(ids))
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

	for _, id := range ids {
		_, err := client.Delete(fmt.Sprintf("/apisix/admin/global_rules/%s", id), nil)
		if err != nil {
			return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
		}
		fmt.Fprintf(opts.IO.Out, "✓ Global rule %s deleted.\n", id)
	}

	fmt.Fprintf(opts.IO.Out, "✓ %d global rules deleted.\n", len(ids))
	return nil
}

func listAllGlobalRuleIDs(client *api.Client, label string) ([]string, error) {
	page := 1
	pageSize := 500
	ids := make([]string, 0)

	for {
		query := map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		}
		if label != "" {
			query["label"] = cmdutil.NormalizeLabel(label)
		}

		body, err := client.Get("/apisix/admin/global_rules", query)
		if err != nil {
			return nil, err
		}

		var resp api.ListResponse[api.GlobalRule]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, item := range resp.List {
			if item.Value.ID != nil && *item.Value.ID != "" {
				ids = append(ids, *item.Value.ID)
			}
		}

		if len(resp.List) == 0 || len(ids) >= resp.Total {
			break
		}
		page++
	}

	return ids, nil
}

func selectGlobalrule(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/global_rules", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch global rules: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.GlobalRule]
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
		items = append(items, selector.Item{ID: id, Label: label})
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no global rules found")
	}

	return selector.SelectOne("Select a global rule", items)
}
