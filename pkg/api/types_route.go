package api

// Route represents an APISIX route resource.
type Route struct {
	ID              *string                `json:"id,omitempty"`
	URI             *string                `json:"uri,omitempty"`
	URIs            []string               `json:"uris,omitempty"`
	Name            *string                `json:"name,omitempty"`
	Desc            *string                `json:"desc,omitempty"`
	Methods         []string               `json:"methods,omitempty"`
	Host            *string                `json:"host,omitempty"`
	Hosts           []string               `json:"hosts,omitempty"`
	RemoteAddr      *string                `json:"remote_addr,omitempty"`
	RemoteAddrs     []string               `json:"remote_addrs,omitempty"`
	Priority        *int                   `json:"priority,omitempty"`
	Status          *int                   `json:"status,omitempty"`
	Plugins         map[string]interface{} `json:"plugins,omitempty"`
	Upstream        *RouteUpstream         `json:"upstream,omitempty"`
	UpstreamID      *string                `json:"upstream_id,omitempty"`
	ServiceID       *string                `json:"service_id,omitempty"`
	PluginConfigID  *string                `json:"plugin_config_id,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Timeout         *RouteTimeout          `json:"timeout,omitempty"`
	EnableWebsocket *bool                  `json:"enable_websocket,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty"`
}

// RouteUpstream defines an inline upstream for a route.
type RouteUpstream struct {
	Type    string                 `json:"type,omitempty"`
	Nodes   map[string]interface{} `json:"nodes,omitempty"`
	Scheme  string                 `json:"scheme,omitempty"`
	Retries *int                   `json:"retries,omitempty"`
	Timeout *RouteTimeout          `json:"timeout,omitempty"`
}

// RouteTimeout defines timeout settings.
type RouteTimeout struct {
	Connect *float64 `json:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty"`
	Read    *float64 `json:"read,omitempty"`
}
