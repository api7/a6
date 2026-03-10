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

	Username string
	Force    bool
	All      bool
	Label    string
}

func NewCmdDelete(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "delete [username]",
		Short: "Delete a consumer",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Username = args[0]
			}
			if opts.All && opts.Label != "" {
				return fmt.Errorf("--all and --label are mutually exclusive")
			}
			if opts.All && opts.Username != "" {
				return fmt.Errorf("--all cannot be used with a specific ID")
			}
			if opts.Label != "" && opts.Username != "" {
				return fmt.Errorf("--label cannot be used with a specific ID")
			}
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Delete all consumers")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Delete consumers matching label (key=value)")

	return cmd
}

func deleteRun(opts *Options) error {
	if opts.All || opts.Label != "" {
		return bulkDeleteConsumers(opts)
	}

	if opts.Username == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("username argument is required (or run interactively in a terminal)")
		}
		id, err := selectConsumer(opts)
		if err != nil {
			return err
		}
		opts.Username = id
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete consumer %s? (y/N): ", opts.Username)
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

	_, err = client.Delete(fmt.Sprintf("/apisix/admin/consumers/%s", opts.Username), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	fmt.Fprintf(opts.IO.Out, "✓ Consumer %s deleted.\n", opts.Username)
	return nil
}

func bulkDeleteConsumers(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	usernames, err := listAllConsumerUsernames(client, opts.Label)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	if len(usernames) == 0 {
		fmt.Fprintln(opts.IO.ErrOut, "No consumers found.")
		return nil
	}

	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Delete %d consumers? (y/N): ", len(usernames))
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

	for _, username := range usernames {
		_, err := client.Delete(fmt.Sprintf("/apisix/admin/consumers/%s", username), nil)
		if err != nil {
			return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
		}
		fmt.Fprintf(opts.IO.Out, "✓ Consumer %s deleted.\n", username)
	}

	fmt.Fprintf(opts.IO.Out, "✓ %d consumers deleted.\n", len(usernames))
	return nil
}

func listAllConsumerUsernames(client *api.Client, label string) ([]string, error) {
	page := 1
	pageSize := 500
	usernames := make([]string, 0)

	for {
		query := map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		}
		if label != "" {
			query["label"] = cmdutil.NormalizeLabel(label)
		}

		body, err := client.Get("/apisix/admin/consumers", query)
		if err != nil {
			return nil, err
		}

		var resp api.ListResponse[api.Consumer]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, item := range resp.List {
			if item.Value.Username != nil && *item.Value.Username != "" {
				usernames = append(usernames, *item.Value.Username)
			}
		}

		if len(resp.List) == 0 || len(usernames) >= resp.Total {
			break
		}
		page++
	}

	return usernames, nil
}

func selectConsumer(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/consumers", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch consumers: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Consumer]
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]selector.Item, 0, len(resp.List))
	for _, item := range resp.List {
		id := ""
		if item.Value.Username != nil {
			id = *item.Value.Username
		}
		if id == "" {
			continue
		}
		label := id
		items = append(items, selector.Item{ID: id, Label: label})
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no consumers found")
	}

	return selector.SelectOne("Select a consumer", items)
}
