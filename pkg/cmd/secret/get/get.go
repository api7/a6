package get

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
	"github.com/api7/a6/pkg/selector"
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
		Use:   "get [manager/id]",
		Short: "Get a secret manager configuration by ID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ID = args[0]
			}
			return getRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml")

	return cmd
}

func getRun(opts *Options) error {
	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("id argument is required (or run interactively in a terminal)")
		}
		id, err := selectSecret(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())

	body, err := client.Get(fmt.Sprintf("/apisix/admin/secrets/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.SingleResponse[api.Secret]
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

func selectSecret(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/secrets", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch secrets: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Secret]
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]selector.Item, 0, len(resp.List))
	for _, item := range resp.List {
		id := secretIDFromKey(item.Key)
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
		return "", fmt.Errorf("no secrets found")
	}

	return selector.SelectOne("Select a secret manager configuration", items)
}

func secretIDFromKey(key string) string {
	parts := strings.Split(strings.Trim(key, "/"), "/")
	if len(parts) < 2 {
		return key
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
