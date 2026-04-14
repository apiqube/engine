package plugin

import "github.com/apiqube/engine"

type Plugin interface {
	Info() PluginInfo
	Init(config map[string]any) error
	Validate(test TestInput) []FieldError
	Execute(test TestInput) (*TestOutput, error)
	Destroy()
}

// PluginInfo is the in-process mirror of a plugin's declared metadata.
// The host populates this from the JSON returned by plugin_info() for WASM
// plugins, or constructs it directly for built-in Go plugins.
type PluginInfo struct {
	Name        string                      `json:"name"`
	Version     string                      `json:"version"`
	Description string                      `json:"description"`
	Protocols   []engine.Protocol           `json:"protocols"`
	Fields      map[string]engine.FieldSpec `json:"fields,omitempty"`
	Events      map[string]engine.EventSpec `json:"events,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type TestInput struct {
	Target  string            `json:"target"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout string            `json:"timeout,omitempty"`
	Fields  map[string]any    `json:"fields"`
}

type TestOutput struct {
	Status     any               `json:"status"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       any               `json:"body,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

// Snapshot converts a PluginInfo into its public schema form for introspection.
func (p PluginInfo) Snapshot() engine.PluginSchema {
	return engine.PluginSchema{
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Protocols:   p.Protocols,
		Fields:      p.Fields,
		Events:      p.Events,
	}
}
