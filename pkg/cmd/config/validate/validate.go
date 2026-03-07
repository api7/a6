package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	File string
}

func NewCmdValidate(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a declarative configuration file",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if opts.File == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("required flag \"file\" not set")}
			}
			return validateRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to declarative config file (required)")

	return cmd
}

func validateRun(opts *Options) error {
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var cfg api.ConfigFile
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		if err := json.Unmarshal(trimmed, &cfg); err != nil {
			return fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(trimmed, &cfg); err != nil {
			return fmt.Errorf("failed to parse YAML file: %w", err)
		}
	}

	errs := validateConfigFile(cfg)
	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
	}

	fmt.Fprintln(opts.IO.Out, "Config is valid")
	return nil
}

func validateConfigFile(cfg api.ConfigFile) []string {
	errs := make([]string, 0)

	if cfg.Version == "" {
		errs = append(errs, "version is required")
	} else if cfg.Version != "1" {
		errs = append(errs, "version must be \"1\"")
	}

	seenRouteIDs := map[string]struct{}{}
	for i, r := range cfg.Routes {
		if !hasRouteURI(r) {
			errs = append(errs, fmt.Sprintf("routes[%d]: either uri or uris is required", i))
		}
		if r.ID != nil {
			if err := checkID(*r.ID, "routes", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenRouteIDs[*r.ID]; ok {
				errs = append(errs, fmt.Sprintf("routes[%d]: duplicate id %q", i, *r.ID))
			} else {
				seenRouteIDs[*r.ID] = struct{}{}
			}
		}
	}

	seenServiceIDs := map[string]struct{}{}
	for i, item := range cfg.Services {
		if item.ID != nil {
			if err := checkID(*item.ID, "services", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenServiceIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("services[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenServiceIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenUpstreamIDs := map[string]struct{}{}
	for i, item := range cfg.Upstreams {
		if item.ID != nil {
			if err := checkID(*item.ID, "upstreams", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenUpstreamIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("upstreams[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenUpstreamIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenConsumerUsernames := map[string]struct{}{}
	for i, c := range cfg.Consumers {
		if c.Username == nil || strings.TrimSpace(*c.Username) == "" {
			errs = append(errs, fmt.Sprintf("consumers[%d]: username is required", i))
			continue
		}
		username := strings.TrimSpace(*c.Username)
		if !idPattern.MatchString(username) {
			errs = append(errs, fmt.Sprintf("consumers[%d]: invalid username %q", i, username))
		} else if _, ok := seenConsumerUsernames[username]; ok {
			errs = append(errs, fmt.Sprintf("consumers[%d]: duplicate username %q", i, username))
		} else {
			seenConsumerUsernames[username] = struct{}{}
		}
	}

	seenSSLIDs := map[string]struct{}{}
	for i, item := range cfg.SSL {
		if item.ID != nil {
			if err := checkID(*item.ID, "ssl", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenSSLIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("ssl[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenSSLIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenGlobalRuleIDs := map[string]struct{}{}
	for i, item := range cfg.GlobalRules {
		if item.ID != nil {
			if err := checkID(*item.ID, "global_rules", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenGlobalRuleIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("global_rules[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenGlobalRuleIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenPluginConfigIDs := map[string]struct{}{}
	for i, item := range cfg.PluginConfigs {
		if item.ID != nil {
			if err := checkID(*item.ID, "plugin_configs", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenPluginConfigIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("plugin_configs[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenPluginConfigIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenConsumerGroupIDs := map[string]struct{}{}
	for i, item := range cfg.ConsumerGroups {
		if item.ID != nil {
			if err := checkID(*item.ID, "consumer_groups", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenConsumerGroupIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("consumer_groups[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenConsumerGroupIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenStreamRouteIDs := map[string]struct{}{}
	for i, item := range cfg.StreamRoutes {
		if item.ID != nil {
			if err := checkID(*item.ID, "stream_routes", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenStreamRouteIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("stream_routes[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenStreamRouteIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenProtoIDs := map[string]struct{}{}
	for i, item := range cfg.Protos {
		if item.ID != nil {
			if err := checkID(*item.ID, "protos", i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenProtoIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("protos[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenProtoIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenSecretIDs := map[string]struct{}{}
	for i, item := range cfg.Secrets {
		if item.ID != nil {
			if err := checkSecretID(*item.ID, i); err != "" {
				errs = append(errs, err)
			} else if _, ok := seenSecretIDs[*item.ID]; ok {
				errs = append(errs, fmt.Sprintf("secrets[%d]: duplicate id %q", i, *item.ID))
			} else {
				seenSecretIDs[*item.ID] = struct{}{}
			}
		}
	}

	seenPluginMetadataNames := map[string]struct{}{}
	for i, item := range cfg.PluginMetadata {
		raw, ok := item["plugin_name"]
		if !ok {
			errs = append(errs, fmt.Sprintf("plugin_metadata[%d]: plugin_name is required", i))
			continue
		}
		name, ok := raw.(string)
		if !ok || strings.TrimSpace(name) == "" {
			errs = append(errs, fmt.Sprintf("plugin_metadata[%d]: plugin_name must be a non-empty string", i))
			continue
		}
		if !idPattern.MatchString(name) {
			errs = append(errs, fmt.Sprintf("plugin_metadata[%d]: invalid plugin_name %q", i, name))
		} else if _, ok := seenPluginMetadataNames[name]; ok {
			errs = append(errs, fmt.Sprintf("plugin_metadata[%d]: duplicate plugin_name %q", i, name))
		} else {
			seenPluginMetadataNames[name] = struct{}{}
		}
	}

	return errs
}

func hasRouteURI(r api.Route) bool {
	if r.URI != nil && strings.TrimSpace(*r.URI) != "" {
		return true
	}
	for _, uri := range r.URIs {
		if strings.TrimSpace(uri) != "" {
			return true
		}
	}
	return false
}

func checkID(id, section string, idx int) string {
	if !idPattern.MatchString(id) {
		return fmt.Sprintf("%s[%d]: invalid id %q", section, idx, id)
	}
	return ""
}

func checkSecretID(id string, idx int) string {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return fmt.Sprintf("secrets[%d]: invalid id %q", idx, id)
	}
	if !idPattern.MatchString(parts[0]) || !idPattern.MatchString(parts[1]) {
		return fmt.Sprintf("secrets[%d]: invalid id %q", idx, id)
	}
	return ""
}
