package api

// GlobalRule represents an APISIX global rule resource.
type GlobalRule struct {
	ID         *string                `json:"id,omitempty" yaml:"id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}
