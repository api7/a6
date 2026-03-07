package api

// Consumer represents an APISIX consumer resource.
type Consumer struct {
	Username   *string                `json:"username,omitempty"`
	Desc       *string                `json:"desc,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	GroupID    *string                `json:"group_id,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty"`
}
