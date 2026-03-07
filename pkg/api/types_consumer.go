package api

// Consumer represents an APISIX consumer resource.
type Consumer struct {
	Username   *string                `json:"username,omitempty" yaml:"username,omitempty"`
	Desc       *string                `json:"desc,omitempty" yaml:"desc,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	GroupID    *string                `json:"group_id,omitempty" yaml:"group_id,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty" yaml:"update_time,omitempty"`
}
