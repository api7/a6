package api

// StreamRoute represents an APISIX stream route resource for L4 (TCP/UDP) traffic.
type StreamRoute struct {
	ID         *string                `json:"id,omitempty" yaml:"id,omitempty"`
	Name       *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Desc       *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	RemoteAddr *string                `json:"remote_addr,omitempty" yaml:"remote_addr,omitempty"`
	ServerAddr *string                `json:"server_addr,omitempty" yaml:"server_addr,omitempty"`
	ServerPort *int                   `json:"server_port,omitempty" yaml:"server_port,omitempty"`
	SNI        *string                `json:"sni,omitempty" yaml:"sni,omitempty"`
	Upstream   *StreamRouteUpstream   `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	UpstreamID *string                `json:"upstream_id,omitempty" yaml:"upstream_id,omitempty"`
	ServiceID  *string                `json:"service_id,omitempty" yaml:"service_id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Protocol   map[string]interface{} `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}

// StreamRouteUpstream defines an inline upstream for a stream route.
type StreamRouteUpstream struct {
	Type    string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Nodes   map[string]interface{} `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Retries *int                   `json:"retries,omitempty" yaml:"retries,omitempty"`
}
