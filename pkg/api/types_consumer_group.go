package api

// ConsumerGroup represents an APISIX consumer group resource.
type ConsumerGroup struct {
	ID         *string                `json:"id,omitempty" yaml:"id,omitempty"`
	Name       *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Desc       *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}
