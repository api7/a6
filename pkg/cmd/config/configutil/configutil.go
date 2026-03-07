package configutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmdutil"
)

type ResourceItem struct {
	Key   string                 `json:"key"`
	Value map[string]interface{} `json:"value,omitempty"`
}

type ResourceDiff struct {
	Create    []ResourceItem `json:"create,omitempty"`
	Update    []ResourceItem `json:"update,omitempty"`
	Delete    []ResourceItem `json:"delete,omitempty"`
	Unchanged []string       `json:"unchanged,omitempty"`
}

func (d ResourceDiff) CreateCount() int { return len(d.Create) }
func (d ResourceDiff) UpdateCount() int { return len(d.Update) }
func (d ResourceDiff) DeleteCount() int { return len(d.Delete) }

func (d ResourceDiff) HasDifferences() bool {
	return len(d.Create) > 0 || len(d.Update) > 0 || len(d.Delete) > 0
}

type DiffResult struct {
	Routes         ResourceDiff `json:"routes"`
	Services       ResourceDiff `json:"services"`
	Upstreams      ResourceDiff `json:"upstreams"`
	Consumers      ResourceDiff `json:"consumers"`
	SSL            ResourceDiff `json:"ssl"`
	GlobalRules    ResourceDiff `json:"global_rules"`
	PluginConfigs  ResourceDiff `json:"plugin_configs"`
	ConsumerGroups ResourceDiff `json:"consumer_groups"`
	StreamRoutes   ResourceDiff `json:"stream_routes"`
	Protos         ResourceDiff `json:"protos"`
	Secrets        ResourceDiff `json:"secrets"`
	PluginMetadata ResourceDiff `json:"plugin_metadata"`
}

func (r *DiffResult) HasDifferences() bool {
	if r == nil {
		return false
	}
	for _, section := range r.Sections() {
		if section.Diff.HasDifferences() {
			return true
		}
	}
	return false
}

type DiffSection struct {
	Name string
	Diff ResourceDiff
}

func (r *DiffResult) Sections() []DiffSection {
	if r == nil {
		return nil
	}
	// Order: base resources first, then dependents (routes reference upstreams/services).
	return []DiffSection{
		{Name: "upstreams", Diff: r.Upstreams},
		{Name: "services", Diff: r.Services},
		{Name: "consumers", Diff: r.Consumers},
		{Name: "consumer_groups", Diff: r.ConsumerGroups},
		{Name: "plugin_configs", Diff: r.PluginConfigs},
		{Name: "ssl", Diff: r.SSL},
		{Name: "global_rules", Diff: r.GlobalRules},
		{Name: "protos", Diff: r.Protos},
		{Name: "secrets", Diff: r.Secrets},
		{Name: "plugin_metadata", Diff: r.PluginMetadata},
		{Name: "routes", Diff: r.Routes},
		{Name: "stream_routes", Diff: r.StreamRoutes},
	}
}

func ReadConfigFile(file string) (api.ConfigFile, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return api.ConfigFile{}, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg api.ConfigFile
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		if err := json.Unmarshal(trimmed, &cfg); err != nil {
			return api.ConfigFile{}, fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(trimmed, &cfg); err != nil {
			return api.ConfigFile{}, fmt.Errorf("failed to parse YAML file: %w", err)
		}
	}

	return cfg, nil
}

func FetchRemoteConfig(client *api.Client) (*api.ConfigFile, error) {
	routeItems, err := fetchPaginated[api.Route](client, "/apisix/admin/routes")
	if err != nil {
		return nil, err
	}
	serviceItems, err := fetchPaginated[api.Service](client, "/apisix/admin/services")
	if err != nil {
		return nil, err
	}
	upstreamItems, err := fetchPaginated[api.Upstream](client, "/apisix/admin/upstreams")
	if err != nil {
		return nil, err
	}
	consumerItems, err := fetchPaginated[api.Consumer](client, "/apisix/admin/consumers")
	if err != nil {
		return nil, err
	}
	sslItems, err := fetchPaginated[api.SSL](client, "/apisix/admin/ssl")
	if err != nil {
		return nil, err
	}
	globalRuleItems, err := fetchPaginated[api.GlobalRule](client, "/apisix/admin/global_rules")
	if err != nil {
		return nil, err
	}
	pluginConfigItems, err := fetchPaginated[api.PluginConfig](client, "/apisix/admin/plugin_configs")
	if err != nil {
		return nil, err
	}
	consumerGroupItems, err := fetchPaginated[api.ConsumerGroup](client, "/apisix/admin/consumer_groups")
	if err != nil {
		return nil, err
	}
	streamRouteItems, err := fetchPaginated[api.StreamRoute](client, "/apisix/admin/stream_routes")
	if err != nil {
		return nil, err
	}
	protoItems, err := fetchPaginated[api.Proto](client, "/apisix/admin/protos")
	if err != nil {
		return nil, err
	}
	secretItems, err := fetchPaginated[api.Secret](client, "/apisix/admin/secrets")
	if err != nil {
		return nil, err
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
		return nil, err
	}

	remote := &api.ConfigFile{
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

	return remote, nil
}

func ComputeDiff(local, remote api.ConfigFile) (*DiffResult, error) {
	localRoutes, err := toMapSlice(local.Routes)
	if err != nil {
		return nil, err
	}
	remoteRoutes, err := toMapSlice(remote.Routes)
	if err != nil {
		return nil, err
	}

	localServices, err := toMapSlice(local.Services)
	if err != nil {
		return nil, err
	}
	remoteServices, err := toMapSlice(remote.Services)
	if err != nil {
		return nil, err
	}

	localUpstreams, err := toMapSlice(local.Upstreams)
	if err != nil {
		return nil, err
	}
	remoteUpstreams, err := toMapSlice(remote.Upstreams)
	if err != nil {
		return nil, err
	}

	localConsumers, err := toMapSlice(local.Consumers)
	if err != nil {
		return nil, err
	}
	remoteConsumers, err := toMapSlice(remote.Consumers)
	if err != nil {
		return nil, err
	}

	localSSL, err := toMapSlice(local.SSL)
	if err != nil {
		return nil, err
	}
	remoteSSL, err := toMapSlice(remote.SSL)
	if err != nil {
		return nil, err
	}

	localGlobalRules, err := toMapSlice(local.GlobalRules)
	if err != nil {
		return nil, err
	}
	remoteGlobalRules, err := toMapSlice(remote.GlobalRules)
	if err != nil {
		return nil, err
	}

	localPluginConfigs, err := toMapSlice(local.PluginConfigs)
	if err != nil {
		return nil, err
	}
	remotePluginConfigs, err := toMapSlice(remote.PluginConfigs)
	if err != nil {
		return nil, err
	}

	localConsumerGroups, err := toMapSlice(local.ConsumerGroups)
	if err != nil {
		return nil, err
	}
	remoteConsumerGroups, err := toMapSlice(remote.ConsumerGroups)
	if err != nil {
		return nil, err
	}

	localStreamRoutes, err := toMapSlice(local.StreamRoutes)
	if err != nil {
		return nil, err
	}
	remoteStreamRoutes, err := toMapSlice(remote.StreamRoutes)
	if err != nil {
		return nil, err
	}

	localProtos, err := toMapSlice(local.Protos)
	if err != nil {
		return nil, err
	}
	remoteProtos, err := toMapSlice(remote.Protos)
	if err != nil {
		return nil, err
	}

	localSecrets, err := toMapSlice(local.Secrets)
	if err != nil {
		return nil, err
	}
	remoteSecrets, err := toMapSlice(remote.Secrets)
	if err != nil {
		return nil, err
	}

	localPluginMetadata, err := toMapSlice(local.PluginMetadata)
	if err != nil {
		return nil, err
	}
	remotePluginMetadata, err := toMapSlice(remote.PluginMetadata)
	if err != nil {
		return nil, err
	}

	routes, err := diffByKey(localRoutes, remoteRoutes, "id", "routes")
	if err != nil {
		return nil, err
	}
	services, err := diffByKey(localServices, remoteServices, "id", "services")
	if err != nil {
		return nil, err
	}
	upstreams, err := diffByKey(localUpstreams, remoteUpstreams, "id", "upstreams")
	if err != nil {
		return nil, err
	}
	consumers, err := diffByKey(localConsumers, remoteConsumers, "username", "consumers")
	if err != nil {
		return nil, err
	}
	ssl, err := diffByKey(localSSL, remoteSSL, "id", "ssl")
	if err != nil {
		return nil, err
	}
	globalRules, err := diffByKey(localGlobalRules, remoteGlobalRules, "id", "global_rules")
	if err != nil {
		return nil, err
	}
	pluginConfigs, err := diffByKey(localPluginConfigs, remotePluginConfigs, "id", "plugin_configs")
	if err != nil {
		return nil, err
	}
	consumerGroups, err := diffByKey(localConsumerGroups, remoteConsumerGroups, "id", "consumer_groups")
	if err != nil {
		return nil, err
	}
	streamRoutes, err := diffByKey(localStreamRoutes, remoteStreamRoutes, "id", "stream_routes")
	if err != nil {
		return nil, err
	}
	protos, err := diffByKey(localProtos, remoteProtos, "id", "protos")
	if err != nil {
		return nil, err
	}
	secrets, err := diffByKey(localSecrets, remoteSecrets, "id", "secrets")
	if err != nil {
		return nil, err
	}
	pluginMetadata, err := diffByKey(localPluginMetadata, remotePluginMetadata, "plugin_name", "plugin_metadata")
	if err != nil {
		return nil, err
	}

	return &DiffResult{
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
	}, nil
}

func FormatDiffSummary(result *DiffResult) string {
	if result == nil || !result.HasDifferences() {
		return "No differences found.\n"
	}

	var b strings.Builder
	b.WriteString("Differences found:\n")
	for _, section := range result.Sections() {
		d := section.Diff
		fmt.Fprintf(&b, "%s: create=%d update=%d delete=%d unchanged=%d\n", section.Name, len(d.Create), len(d.Update), len(d.Delete), len(d.Unchanged))
		for _, item := range d.Create {
			fmt.Fprintf(&b, "  CREATE %s\n", item.Key)
		}
		for _, item := range d.Update {
			fmt.Fprintf(&b, "  UPDATE %s\n", item.Key)
		}
		for _, item := range d.Delete {
			fmt.Fprintf(&b, "  DELETE %s\n", item.Key)
		}
	}

	return b.String()
}

func diffByKey(localItems, remoteItems []map[string]interface{}, keyField, resourceName string) (ResourceDiff, error) {
	localByKey := make(map[string]map[string]interface{}, len(localItems))
	for i, item := range localItems {
		key, err := extractKey(item, keyField)
		if err != nil {
			return ResourceDiff{}, fmt.Errorf("%s[%d]: %w", resourceName, i, err)
		}
		localByKey[key] = normalizeMap(item)
	}

	remoteByKey := make(map[string]map[string]interface{}, len(remoteItems))
	for i, item := range remoteItems {
		key, err := extractKey(item, keyField)
		if err != nil {
			return ResourceDiff{}, fmt.Errorf("remote %s[%d]: %w", resourceName, i, err)
		}
		remoteByKey[key] = normalizeMap(item)
	}

	result := ResourceDiff{
		Create:    make([]ResourceItem, 0),
		Update:    make([]ResourceItem, 0),
		Delete:    make([]ResourceItem, 0),
		Unchanged: make([]string, 0),
	}

	for key, localItem := range localByKey {
		remoteItem, ok := remoteByKey[key]
		if !ok {
			result.Create = append(result.Create, ResourceItem{Key: key, Value: localItem})
			continue
		}

		if reflect.DeepEqual(localItem, remoteItem) {
			result.Unchanged = append(result.Unchanged, key)
		} else {
			result.Update = append(result.Update, ResourceItem{Key: key, Value: localItem})
		}
	}

	for key, remoteItem := range remoteByKey {
		if _, ok := localByKey[key]; ok {
			continue
		}
		result.Delete = append(result.Delete, ResourceItem{Key: key, Value: remoteItem})
	}

	return result, nil
}

func extractKey(item map[string]interface{}, field string) (string, error) {
	raw, ok := item[field]
	if !ok {
		return "", fmt.Errorf("missing %q field", field)
	}
	key := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if key == "" || key == "<nil>" {
		return "", fmt.Errorf("empty %q field", field)
	}
	return key, nil
}

func normalizeMap(item map[string]interface{}) map[string]interface{} {
	b, _ := json.Marshal(item)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	stripTimestamps(out)
	return out
}

func stripTimestamps(v interface{}) {
	switch typed := v.(type) {
	case map[string]interface{}:
		delete(typed, "create_time")
		delete(typed, "update_time")
		for _, vv := range typed {
			stripTimestamps(vv)
		}
	case []interface{}:
		for _, vv := range typed {
			stripTimestamps(vv)
		}
	}
}

func toMapSlice[T any](items []T) ([]map[string]interface{}, error) {
	b, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources: %w", err)
	}
	if len(b) == 0 || string(b) == "null" {
		return []map[string]interface{}{}, nil
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("failed to convert resources: %w", err)
	}
	if out == nil {
		return []map[string]interface{}{}, nil
	}
	return out, nil
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
			if cmdutil.IsOptionalResourceError(err) {
				return nil, nil
			}
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
		if cmdutil.IsOptionalResourceError(err) {
			return nil, nil
		}
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
