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
		Short: "Delete a route",
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
	cmd.Flags().BoolVar(&opts.All, "all", false, "Delete all routes")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Delete routes matching label (key=value)")

	return cmd
}

func deleteRun(opts *Options) error {
	if opts.All || opts.Label != "" {
		return bulkDeleteRoutes(opts)
	}

	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("id argument is required (or run interactively in a terminal)")
		}
		id, err := selectRoute(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete route %s? (y/N): ", opts.ID)
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

	_, err = client.Delete(fmt.Sprintf("/apisix/admin/routes/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	fmt.Fprintf(opts.IO.Out, "✓ Route %s deleted.\n", opts.ID)
	return nil
}

func bulkDeleteRoutes(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())

	ids, err := listAllRouteIDs(client, opts.Label)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	if len(ids) == 0 {
		fmt.Fprintln(opts.IO.ErrOut, "No routes found.")
		return nil
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete %d routes? (y/N): ", len(ids))
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
		_, err := client.Delete(fmt.Sprintf("/apisix/admin/routes/%s", id), nil)
		if err != nil {
			return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
		}
		fmt.Fprintf(opts.IO.Out, "✓ Route %s deleted.\n", id)
	}

	fmt.Fprintf(opts.IO.Out, "✓ %d routes deleted.\n", len(ids))
	return nil
}

// parseLabel splits a "key=value" label into key and value parts.
// APISIX Admin API only supports filtering by label key, so we send the key
// as the query parameter and perform client-side filtering by value.
func parseLabel(label string) (key, value string) {
	parts := strings.SplitN(label, "=", 2)
	key = parts[0]
	if len(parts) == 2 {
		value = parts[1]
	}
	return
}

func listAllRouteIDs(client *api.Client, label string) ([]string, error) {
	page := 1
	pageSize := 500
	ids := make([]string, 0)

	var labelKey, labelValue string
	if label != "" {
		labelKey, labelValue = parseLabel(label)
	}

	for {
		query := map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		}
		if labelKey != "" {
			query["label"] = labelKey
		}

		body, err := client.Get("/apisix/admin/routes", query)
		if err != nil {
			return nil, err
		}

		var resp api.ListResponse[api.Route]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, item := range resp.List {
			if item.Value.ID == nil || *item.Value.ID == "" {
				continue
			}
			// Client-side filter by label value if specified.
			if labelValue != "" && item.Value.Labels != nil {
				if v, ok := item.Value.Labels[labelKey]; !ok || v != labelValue {
					continue
				}
			}
			ids = append(ids, *item.Value.ID)
		}

		if len(resp.List) == 0 || page*pageSize >= resp.Total {
			break
		}
		page++
	}

	return ids, nil
}

func selectRoute(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/routes", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch routes: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Route]
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
		return "", fmt.Errorf("no routes found")
	}

	return selector.SelectOne("Select a route", items)
}
