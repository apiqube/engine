package plugin

// Plugin is the interface for all protocol executors (built-in and WASM).
type Plugin interface {
	Info() PluginInfo
	Init(config map[string]any) error
	Validate(test TestInput) []FieldError
	Execute(test TestInput) (*TestOutput, error)
	Destroy()
}

type PluginInfo struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Protocols   []string             `json:"protocols"`
	Fields      map[string]FieldSpec `json:"fields"`
}

type FieldSpec struct {
	Type        string `json:"type"` // string, object, array, map, bool, number
	Required    bool   `json:"required"`
	Description string `json:"description"`
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
	Status     any               `json:"status"` // int for HTTP, string for gRPC
	Headers    map[string]string `json:"headers,omitempty"`
	Body       any               `json:"body,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}
