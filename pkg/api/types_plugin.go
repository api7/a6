package api

type PluginInfo struct {
	Priority int    `json:"priority"`
	Phase    string `json:"phase,omitempty"`
	Version  string `json:"version,omitempty"`
}
