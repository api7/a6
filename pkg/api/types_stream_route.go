package api

// StreamRoute represents an APISIX stream route resource for L4 (TCP/UDP) traffic.
type StreamRoute struct {
	ID         *string                `json:"id,omitempty"`
	Name       *string                `json:"name,omitempty"`
	Desc       *string                `json:"desc,omitempty"`
	RemoteAddr *string                `json:"remote_addr,omitempty"`
	ServerAddr *string                `json:"server_addr,omitempty"`
	ServerPort *int                   `json:"server_port,omitempty"`
	SNI        *string                `json:"sni,omitempty"`
	Upstream   *StreamRouteUpstream   `json:"upstream,omitempty"`
	UpstreamID *string                `json:"upstream_id,omitempty"`
	ServiceID  *string                `json:"service_id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Protocol   map[string]interface{} `json:"protocol,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty"`
}

// StreamRouteUpstream defines an inline upstream for a stream route.
type StreamRouteUpstream struct {
	Type    string                 `json:"type,omitempty"`
	Nodes   map[string]interface{} `json:"nodes,omitempty"`
	Retries *int                   `json:"retries,omitempty"`
}
