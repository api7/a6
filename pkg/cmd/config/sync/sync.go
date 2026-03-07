package sync

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/config/configutil"
	"github.com/api7/a6/pkg/cmd/config/validate"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	File   string
	DryRun bool
	Delete bool
}

func NewCmdSync(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
		Delete: true,
	}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize APISIX with local declarative config",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if opts.File == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("required flag \"file\" not set")}
			}
			return syncRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to declarative config file (required)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would change without applying")
	cmd.Flags().BoolVar(&opts.Delete, "delete", true, "Delete remote resources not present in local config")

	return cmd
}

func syncRun(opts *Options) error {
	local, err := configutil.ReadConfigFile(opts.File)
	if err != nil {
		return err
	}

	errs := validate.ValidateConfigFile(local)
	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
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

	remote, err := configutil.FetchRemoteConfig(client)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	result, err := configutil.ComputeDiff(local, *remote)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprint(opts.IO.Out, configutil.FormatDiffSummary(result))
		return nil
	}

	if err := applyAllCreates(client, result); err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	if err := applyAllUpdates(client, result); err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	if opts.Delete {
		if err := applyAllDeletes(client, result); err != nil {
			return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
		}
	}

	fmt.Fprint(opts.IO.Out, formatSyncSummary(result, opts.Delete))
	return nil
}

func applyAllCreates(client *api.Client, result *configutil.DiffResult) error {
	for _, section := range result.Sections() {
		for _, item := range section.Diff.Create {
			if err := putResource(client, section.Name, item.Key, item.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyAllUpdates(client *api.Client, result *configutil.DiffResult) error {
	for _, section := range result.Sections() {
		for _, item := range section.Diff.Update {
			if err := putResource(client, section.Name, item.Key, item.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyAllDeletes(client *api.Client, result *configutil.DiffResult) error {
	sections := result.Sections()
	for i := len(sections) - 1; i >= 0; i-- {
		for _, item := range sections[i].Diff.Delete {
			if err := deleteResource(client, sections[i].Name, item.Key); err != nil {
				return err
			}
		}
	}
	return nil
}

func putResource(client *api.Client, resourceType, key string, payload map[string]interface{}) error {
	body := cloneMap(payload)
	path, cleanBody, err := putPathAndBody(resourceType, key, body)
	if err != nil {
		return err
	}
	_, err = client.Put(path, cleanBody)
	return err
}

func deleteResource(client *api.Client, resourceType, key string) error {
	path, err := deletePath(resourceType, key)
	if err != nil {
		return err
	}
	_, err = client.Delete(path, nil)
	return err
}

func putPathAndBody(resourceType, key string, payload map[string]interface{}) (string, map[string]interface{}, error) {
	switch resourceType {
	case "routes":
		return fmt.Sprintf("/apisix/admin/routes/%s", key), payload, nil
	case "services":
		return fmt.Sprintf("/apisix/admin/services/%s", key), payload, nil
	case "upstreams":
		return fmt.Sprintf("/apisix/admin/upstreams/%s", key), payload, nil
	case "consumers":
		return "/apisix/admin/consumers", payload, nil
	case "ssl":
		return fmt.Sprintf("/apisix/admin/ssl/%s", key), payload, nil
	case "global_rules":
		return fmt.Sprintf("/apisix/admin/global_rules/%s", key), payload, nil
	case "plugin_configs":
		return fmt.Sprintf("/apisix/admin/plugin_configs/%s", key), payload, nil
	case "consumer_groups":
		return fmt.Sprintf("/apisix/admin/consumer_groups/%s", key), payload, nil
	case "stream_routes":
		return fmt.Sprintf("/apisix/admin/stream_routes/%s", key), payload, nil
	case "protos":
		return fmt.Sprintf("/apisix/admin/protos/%s", key), payload, nil
	case "secrets":
		return fmt.Sprintf("/apisix/admin/secrets/%s", key), payload, nil
	case "plugin_metadata":
		delete(payload, "plugin_name")
		return fmt.Sprintf("/apisix/admin/plugin_metadata/%s", key), payload, nil
	default:
		return "", nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func deletePath(resourceType, key string) (string, error) {
	switch resourceType {
	case "routes":
		return fmt.Sprintf("/apisix/admin/routes/%s", key), nil
	case "services":
		return fmt.Sprintf("/apisix/admin/services/%s", key), nil
	case "upstreams":
		return fmt.Sprintf("/apisix/admin/upstreams/%s", key), nil
	case "consumers":
		return fmt.Sprintf("/apisix/admin/consumers/%s", key), nil
	case "ssl":
		return fmt.Sprintf("/apisix/admin/ssl/%s", key), nil
	case "global_rules":
		return fmt.Sprintf("/apisix/admin/global_rules/%s", key), nil
	case "plugin_configs":
		return fmt.Sprintf("/apisix/admin/plugin_configs/%s", key), nil
	case "consumer_groups":
		return fmt.Sprintf("/apisix/admin/consumer_groups/%s", key), nil
	case "stream_routes":
		return fmt.Sprintf("/apisix/admin/stream_routes/%s", key), nil
	case "protos":
		return fmt.Sprintf("/apisix/admin/protos/%s", key), nil
	case "secrets":
		return fmt.Sprintf("/apisix/admin/secrets/%s", key), nil
	case "plugin_metadata":
		return fmt.Sprintf("/apisix/admin/plugin_metadata/%s", key), nil
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func formatSyncSummary(result *configutil.DiffResult, deleteEnabled bool) string {
	var b strings.Builder
	b.WriteString("Sync completed:\n")
	for _, section := range result.Sections() {
		if deleteEnabled {
			fmt.Fprintf(&b, "%s: created=%d updated=%d deleted=%d\n", section.Name, section.Diff.CreateCount(), section.Diff.UpdateCount(), section.Diff.DeleteCount())
		} else {
			fmt.Fprintf(&b, "%s: created=%d updated=%d deleted=0\n", section.Name, section.Diff.CreateCount(), section.Diff.UpdateCount())
		}
	}
	return b.String()
}
