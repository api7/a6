package dump

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

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

	Output string
	File   string
}

func NewCmdDump(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
		Output: "yaml",
	}

	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump APISIX resources as declarative configuration",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return dumpRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "yaml", "Output format: yaml, json")
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Write output to file")

	return cmd
}

func dumpRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())

	routeItems, err := fetchPaginated[api.Route](client, "/apisix/admin/routes")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	serviceItems, err := fetchPaginated[api.Service](client, "/apisix/admin/services")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	upstreamItems, err := fetchPaginated[api.Upstream](client, "/apisix/admin/upstreams")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	consumerItems, err := fetchPaginated[api.Consumer](client, "/apisix/admin/consumers")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	sslItems, err := fetchPaginated[api.SSL](client, "/apisix/admin/ssl")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	globalRuleItems, err := fetchPaginated[api.GlobalRule](client, "/apisix/admin/global_rules")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	pluginConfigItems, err := fetchPaginated[api.PluginConfig](client, "/apisix/admin/plugin_configs")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	consumerGroupItems, err := fetchPaginated[api.ConsumerGroup](client, "/apisix/admin/consumer_groups")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	streamRouteItems, err := fetchPaginated[api.StreamRoute](client, "/apisix/admin/stream_routes")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	protoItems, err := fetchPaginated[api.Proto](client, "/apisix/admin/protos")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}
	secretItems, err := fetchPaginated[api.Secret](client, "/apisix/admin/secrets")
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	routes := make([]api.Route, 0, len(routeItems))
	for _, item := range routeItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		routes = append(routes, v)
	}

	services := make([]api.Service, 0, len(serviceItems))
	for _, item := range serviceItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		services = append(services, v)
	}

	upstreams := make([]api.Upstream, 0, len(upstreamItems))
	for _, item := range upstreamItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		upstreams = append(upstreams, v)
	}

	consumers := make([]api.Consumer, 0, len(consumerItems))
	for _, item := range consumerItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		consumers = append(consumers, v)
	}

	ssl := make([]api.SSL, 0, len(sslItems))
	for _, item := range sslItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		ssl = append(ssl, v)
	}

	globalRules := make([]api.GlobalRule, 0, len(globalRuleItems))
	for _, item := range globalRuleItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		globalRules = append(globalRules, v)
	}

	pluginConfigs := make([]api.PluginConfig, 0, len(pluginConfigItems))
	for _, item := range pluginConfigItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		pluginConfigs = append(pluginConfigs, v)
	}

	consumerGroups := make([]api.ConsumerGroup, 0, len(consumerGroupItems))
	for _, item := range consumerGroupItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		consumerGroups = append(consumerGroups, v)
	}

	streamRoutes := make([]api.StreamRoute, 0, len(streamRouteItems))
	for _, item := range streamRouteItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		streamRoutes = append(streamRoutes, v)
	}

	protos := make([]api.Proto, 0, len(protoItems))
	for _, item := range protoItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		protos = append(protos, v)
	}

	secrets := make([]api.Secret, 0, len(secretItems))
	for _, item := range secretItems {
		v := item.Value
		v.CreateTime = nil
		v.UpdateTime = nil
		if id := extractSecretID(item.Key); id != "" {
			idCopy := id
			v.ID = &idCopy
		}
		secrets = append(secrets, v)
	}

	pluginMetadata, err := fetchPluginMetadata(client)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	configFile := api.ConfigFile{
		Version:        "1",
		Routes:         routes,
		Services:       services,
		Upstreams:      upstreams,
		Consumers:      consumers,
		SSL:            ssl,
		GlobalRules:    globalRules,
		PluginConfigs:  pluginConfigs,
		ConsumerGroups: consumerGroups,
		StreamRoutes:   streamRoutes,
		Protos:         protos,
		Secrets:        secrets,
		PluginMetadata: pluginMetadata,
	}

	format := opts.Output
	if format == "" {
		format = "yaml"
	}

	var out io.Writer = opts.IO.Out
	if opts.File != "" {
		f, err := os.Create(opts.File)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer f.Close()
		out = f
	}

	return cmdutil.NewExporter(format, out).Write(configFile)
}

func fetchPaginated[T any](client *api.Client, path string) ([]api.ListItem[T], error) {
	page := 1
	pageSize := 500
	items := make([]api.ListItem[T], 0)

	for {
		body, err := client.Get(path, map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		})
		if err != nil {
			return nil, err
		}

		var resp api.ListResponse[T]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		items = append(items, resp.List...)
		if len(resp.List) == 0 || len(items) >= resp.Total {
			break
		}
		page++
	}

	return items, nil
}

func fetchPluginMetadata(client *api.Client) ([]api.PluginMetadataEntry, error) {
	body, err := client.Get("/apisix/admin/plugins/list", nil)
	if err != nil {
		return nil, err
	}

	var plugins []string
	if err := json.Unmarshal(body, &plugins); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]api.PluginMetadataEntry, 0, len(plugins))
	for _, pluginName := range plugins {
		metadataBody, err := client.Get(fmt.Sprintf("/apisix/admin/plugin_metadata/%s", pluginName), nil)
		if err != nil {
			if cmdutil.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		var resp api.SingleResponse[map[string]interface{}]
		if err := json.Unmarshal(metadataBody, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		entry := api.PluginMetadataEntry{}
		for k, v := range resp.Value {
			entry[k] = v
		}
		entry["plugin_name"] = pluginName
		result = append(result, entry)
	}

	return result, nil
}

func extractSecretID(key string) string {
	parts := strings.Split(strings.Trim(key, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
