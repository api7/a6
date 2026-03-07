package api

type ConfigFile struct {
	Version        string                `json:"version" yaml:"version"`
	Routes         []Route               `json:"routes,omitempty" yaml:"routes,omitempty"`
	Services       []Service             `json:"services,omitempty" yaml:"services,omitempty"`
	Upstreams      []Upstream            `json:"upstreams,omitempty" yaml:"upstreams,omitempty"`
	Consumers      []Consumer            `json:"consumers,omitempty" yaml:"consumers,omitempty"`
	SSL            []SSL                 `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	GlobalRules    []GlobalRule          `json:"global_rules,omitempty" yaml:"global_rules,omitempty"`
	PluginConfigs  []PluginConfig        `json:"plugin_configs,omitempty" yaml:"plugin_configs,omitempty"`
	ConsumerGroups []ConsumerGroup       `json:"consumer_groups,omitempty" yaml:"consumer_groups,omitempty"`
	StreamRoutes   []StreamRoute         `json:"stream_routes,omitempty" yaml:"stream_routes,omitempty"`
	Protos         []Proto               `json:"protos,omitempty" yaml:"protos,omitempty"`
	Secrets        []Secret              `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	PluginMetadata []PluginMetadataEntry `json:"plugin_metadata,omitempty" yaml:"plugin_metadata,omitempty"`
}

type PluginMetadataEntry map[string]interface{}
