package api

type PluginInfo struct {
	Priority int    `json:"priority" yaml:"priority"`
	Phase    string `json:"phase,omitempty" yaml:"phase,omitempty"`
	Version  string `json:"version,omitempty" yaml:"version,omitempty"`
}
