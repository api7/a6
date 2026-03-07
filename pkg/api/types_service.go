package api

// Service represents an APISIX service resource.
type Service struct {
	ID              *string                `json:"id,omitempty"`
	Name            *string                `json:"name,omitempty"`
	Desc            *string                `json:"desc,omitempty"`
	Plugins         map[string]interface{} `json:"plugins,omitempty"`
	Upstream        *ServiceUpstream       `json:"upstream,omitempty"`
	UpstreamID      *string                `json:"upstream_id,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	EnableWebsocket *bool                  `json:"enable_websocket,omitempty"`
	Hosts           []string               `json:"hosts,omitempty"`
	Status          *int                   `json:"status,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty"`
}

// ServiceUpstream defines an inline upstream for a service.
type ServiceUpstream struct {
	Type    string                 `json:"type,omitempty"`
	Nodes   map[string]interface{} `json:"nodes,omitempty"`
	Scheme  string                 `json:"scheme,omitempty"`
	Retries *int                   `json:"retries,omitempty"`
	Timeout *ServiceTimeout        `json:"timeout,omitempty"`
}

// ServiceTimeout defines timeout settings for a service upstream.
type ServiceTimeout struct {
	Connect *float64 `json:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty"`
	Read    *float64 `json:"read,omitempty"`
}
