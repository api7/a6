package update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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
	File     string
	Output   string
}

func NewCmdUpdate(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "update [username]",
		Short: "Update a consumer",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Username = args[0]
			}
			if opts.File == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("required flag \"file\" not set")}
			}
			return updateRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to JSON/YAML file (required)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml")

	return cmd
}

func updateRun(opts *Options) error {
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

	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	baseURL := cfg.BaseURL()

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var payload map[string]interface{}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		if err := json.Unmarshal(trimmed, &payload); err != nil {
			return fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(trimmed, &payload); err != nil {
			return fmt.Errorf("failed to parse YAML file: %w", err)
		}
	}

	payload["username"] = opts.Username

	client := api.NewClient(httpClient, baseURL)

	body, err := client.Put("/apisix/admin/consumers", payload)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.SingleResponse[api.Consumer]
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
