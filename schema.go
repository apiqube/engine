package engine

// FieldSpec describes one field in a plugin contract — used both for manifest
// fields (what the plugin reads from test cases) and event fields (what the
// plugin writes into events it emits).
//
// Plugins declare these at plugin_info() time. Frontends can introspect them
// to build UI, generate documentation, or validate input.
type FieldSpec struct {
	Type        FieldType `json:"type"`
	Required    bool      `json:"required"`
	Description string    `json:"description"`
}

// FieldType enumerates the YAML/JSON types a plugin field may hold.
type FieldType string

const (
	FieldString FieldType = "string"
	FieldNumber FieldType = "number"
	FieldBool   FieldType = "bool"
	FieldObject FieldType = "object"
	FieldArray  FieldType = "array"
	FieldMap    FieldType = "map"
	FieldAny    FieldType = "any"
)

// EventSpec describes a single event that a plugin can emit at runtime.
// The schema is returned as part of PluginInfo during plugin_info().
//
// Frontends read these schemas to know what events exist, what fields they
// carry, and how to display them in dashboards, logs, or debug panels.
type EventSpec struct {
	Description string               `json:"description"`
	Fields      map[string]FieldSpec `json:"fields"`
}

// PluginSchema is the subset of PluginInfo exposed to consumers of the engine.
// Frontends use it to introspect loaded plugins — their protocols, manifest
// fields, and emitted events.
type PluginSchema struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Protocols   []Protocol           `json:"protocols"`
	Fields      map[string]FieldSpec `json:"fields,omitempty"`
	Events      map[string]EventSpec `json:"events,omitempty"`
}
