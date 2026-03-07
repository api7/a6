package api

// Secret represents an APISIX secret manager configuration.
type Secret struct {
	ID              *string                `json:"id,omitempty" yaml:"id,omitempty"`
	URI             *string                `json:"uri,omitempty" yaml:"uri,omitempty"`
	Prefix          *string                `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Token           *string                `json:"token,omitempty" yaml:"token,omitempty"`
	Namespace       *string                `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	AccessKeyID     *string                `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty"`
	SecretAccessKey *string                `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty"`
	Region          *string                `json:"region,omitempty" yaml:"region,omitempty"`
	EndpointURL     *string                `json:"endpoint_url,omitempty" yaml:"endpoint_url,omitempty"`
	AuthConfig      map[string]interface{} `json:"auth_config,omitempty" yaml:"auth_config,omitempty"`
	AuthFile        *string                `json:"auth_file,omitempty" yaml:"auth_file,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}
