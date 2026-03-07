package api

// Proto represents an APISIX proto resource for Protocol Buffer definitions.
type Proto struct {
	ID         *string           `json:"id,omitempty"`
	Content    *string           `json:"content,omitempty"`
	Name       *string           `json:"name,omitempty"`
	Desc       *string           `json:"desc,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreateTime *int64            `json:"create_time,omitempty"`
	UpdateTime *int64            `json:"update_time,omitempty"`
}
