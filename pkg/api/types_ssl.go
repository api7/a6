package api

// SSL represents an APISIX SSL certificate resource.
type SSL struct {
	ID           *string           `json:"id,omitempty" yaml:"id,omitempty"`
	Cert         *string           `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key          *string           `json:"key,omitempty" yaml:"key,omitempty"`
	Certs        []string          `json:"certs,omitempty" yaml:"certs,omitempty"`
	Keys         []string          `json:"keys,omitempty" yaml:"keys,omitempty"`
	SNI          *string           `json:"sni,omitempty" yaml:"sni,omitempty"`
	SNIs         []string          `json:"snis,omitempty" yaml:"snis,omitempty"`
	Client       *SSLClient        `json:"client,omitempty" yaml:"client,omitempty"`
	Type         *string           `json:"type,omitempty" yaml:"type,omitempty"`
	Status       *int              `json:"status,omitempty" yaml:"status,omitempty"`
	SSLProtocols []string          `json:"ssl_protocols,omitempty" yaml:"ssl_protocols,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreateTime   *int64            `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime   *int64            `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}

// SSLClient defines mTLS client verification settings.
type SSLClient struct {
	CA    *string `json:"ca,omitempty" yaml:"ca,omitempty"`
	Depth *int    `json:"depth,omitempty" yaml:"depth,omitempty"`
}
