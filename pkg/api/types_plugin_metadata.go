package api

// PluginMetadata represents APISIX plugin metadata.
// Plugin metadata is keyed by plugin name, not by ID.
// The schema is defined by each plugin's metadata_schema.
type PluginMetadata map[string]interface{}
