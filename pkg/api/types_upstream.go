package api

// Upstream represents an APISIX upstream resource.
type Upstream struct {
	ID            *string                `json:"id,omitempty" yaml:"id,omitempty"`
	Name          *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Desc          *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	Type          *string                `json:"type,omitempty" yaml:"type,omitempty"`
	Nodes         UpstreamNodes          `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	ServiceName   *string                `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	DiscoveryType *string                `json:"discovery_type,omitempty" yaml:"discovery_type,omitempty"`
	HashOn        *string                `json:"hash_on,omitempty" yaml:"hash_on,omitempty"`
	Key           *string                `json:"key,omitempty" yaml:"key,omitempty"`
	Checks        map[string]interface{} `json:"checks,omitempty" yaml:"checks,omitempty"`
	Retries       *int                   `json:"retries,omitempty" yaml:"retries,omitempty"`
	RetryTimeout  *float64               `json:"retry_timeout,omitempty" yaml:"retry_timeout,omitempty"`
	Timeout       *UpstreamTimeout       `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	PassHost      *string                `json:"pass_host,omitempty" yaml:"pass_host,omitempty"`
	UpstreamHost  *string                `json:"upstream_host,omitempty" yaml:"upstream_host,omitempty"`
	Scheme        *string                `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	KeepalivePool map[string]interface{} `json:"keepalive_pool,omitempty" yaml:"keepalive_pool,omitempty"`
	TLS           map[string]interface{} `json:"tls,omitempty" yaml:"tls,omitempty"`
	Status        *int                   `json:"status,omitempty" yaml:"status,omitempty"`
	CreateTime    *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime    *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}

// UpstreamTimeout defines timeout settings for an upstream.
type UpstreamTimeout struct {
	Connect *float64 `json:"connect,omitempty" yaml:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty" yaml:"send,omitempty"`
	Read    *float64 `json:"read,omitempty" yaml:"read,omitempty"`
}
