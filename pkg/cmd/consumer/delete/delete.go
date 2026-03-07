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
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")

	return cmd
}

func deleteRun(opts *Options) error {
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
