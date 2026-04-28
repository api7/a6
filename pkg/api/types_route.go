package api

// Route represents an APISIX route resource.
type Route struct {
	ID              *string                `json:"id,omitempty" yaml:"id,omitempty"`
	URI             *string                `json:"uri,omitempty" yaml:"uri,omitempty"`
	URIs            []string               `json:"uris,omitempty" yaml:"uris,omitempty"`
	Name            *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Desc            *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	Methods         []string               `json:"methods,omitempty" yaml:"methods,omitempty"`
	Host            *string                `json:"host,omitempty" yaml:"host,omitempty"`
	Hosts           []string               `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	RemoteAddr      *string                `json:"remote_addr,omitempty" yaml:"remote_addr,omitempty"`
	RemoteAddrs     []string               `json:"remote_addrs,omitempty" yaml:"remote_addrs,omitempty"`
	Priority        *int                   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Status          *int                   `json:"status,omitempty" yaml:"status,omitempty"`
	Plugins         map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Upstream        *RouteUpstream         `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	UpstreamID      *string                `json:"upstream_id,omitempty" yaml:"upstream_id,omitempty"`
	ServiceID       *string                `json:"service_id,omitempty" yaml:"service_id,omitempty"`
	PluginConfigID  *string                `json:"plugin_config_id,omitempty" yaml:"plugin_config_id,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	Timeout         *RouteTimeout          `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	EnableWebsocket *bool                  `json:"enable_websocket,omitempty" yaml:"enable_websocket,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}

// RouteUpstream defines an inline upstream for a route.
type RouteUpstream struct {
	Type    string        `json:"type,omitempty" yaml:"type,omitempty"`
	Nodes   UpstreamNodes `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Scheme  string        `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Retries *int          `json:"retries,omitempty" yaml:"retries,omitempty"`
	Timeout *RouteTimeout `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// RouteTimeout defines timeout settings.
type RouteTimeout struct {
	Connect *float64 `json:"connect,omitempty" yaml:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty" yaml:"send,omitempty"`
	Read    *float64 `json:"read,omitempty" yaml:"read,omitempty"`
}
