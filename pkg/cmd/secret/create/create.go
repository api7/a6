package create

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
)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	ID     string
	File   string
	Output string
}

func NewCmdCreate(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "create <manager/id>",
		Short: "Create a secret manager configuration from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.ID = args[0]
			if opts.File == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("required flag \"file\" not set")}
			}
			return createRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to JSON/YAML file (required)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml")

	return cmd
}

func createRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

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

	payload = stripConflictingID(payload, opts.ID)

	client := api.NewClient(httpClient, cfg.BaseURL())

	body, err := client.Put(fmt.Sprintf("/apisix/admin/secrets/%s", opts.ID), payload)
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

// stripConflictingID removes the body's `id` field when it would conflict
// with the manager-prefixed positional id (e.g. file says `id: my-vault`
// while the positional says `vault/my-vault`). APISIX rejects the mismatch
// with `wrong secret id`, so drop the body id and let the URL path win.
func stripConflictingID(payload map[string]interface{}, positional string) map[string]interface{} {
	if payload == nil {
		return payload
	}
	bodyID, ok := payload["id"].(string)
	if !ok {
		return payload
	}
	if bodyID == positional {
		return payload
	}
	delete(payload, "id")
	return payload
}
