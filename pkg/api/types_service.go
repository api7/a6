package api

// Service represents an APISIX service resource.
type Service struct {
	ID              *string                `json:"id,omitempty" yaml:"id,omitempty"`
	Name            *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Desc            *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	Plugins         map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Upstream        *ServiceUpstream       `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	UpstreamID      *string                `json:"upstream_id,omitempty" yaml:"upstream_id,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	EnableWebsocket *bool                  `json:"enable_websocket,omitempty" yaml:"enable_websocket,omitempty"`
	Hosts           []string               `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Status          *int                   `json:"status,omitempty" yaml:"status,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}

// ServiceUpstream defines an inline upstream for a service.
type ServiceUpstream struct {
	Type    string          `json:"type,omitempty" yaml:"type,omitempty"`
	Nodes   UpstreamNodes   `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Scheme  string          `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Retries *int            `json:"retries,omitempty" yaml:"retries,omitempty"`
	Timeout *ServiceTimeout `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// ServiceTimeout defines timeout settings for a service upstream.
type ServiceTimeout struct {
	Connect *float64 `json:"connect,omitempty" yaml:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty" yaml:"send,omitempty"`
	Read    *float64 `json:"read,omitempty" yaml:"read,omitempty"`
}
