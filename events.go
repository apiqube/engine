package engine

import "time"

// EventHandler receives events from the engine during execution.
// Frontends implement this with a type switch to handle events they care about.
//
// Example:
//
//	func (h *CLIHandler) Handle(event engine.Event) {
//	    switch e := event.(type) {
//	    case engine.TestCompleted:
//	        fmt.Println(e.Name, e.Status)
//	    case engine.Progress:
//	        updateBar(e.Completed, e.Total)
//	    }
//	}
type EventHandler interface {
	Handle(event Event)
}

// Event is a sealed interface — only types in this package can implement it.
// Use a type switch in your EventHandler to handle events.
type Event interface {
	sealed()
	Type() string
}

// NopHandler is an EventHandler that discards all events.
// Used as default when no handler is provided.
type NopHandler struct{}

func (NopHandler) Handle(Event) {}

// RunStarted is emitted once when a Run() begins, after parsing and graph building.
type RunStarted struct {
	Files      []string `json:"files"`
	TotalTests int      `json:"total_tests"`
	TotalWaves int      `json:"total_waves"`
}

// RunCompleted is emitted once when all waves have finished.
type RunCompleted struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Errored  int           `json:"errored"`
	Duration time.Duration `json:"duration"`
}

// WaveStarted is emitted before a parallel wave begins executing.
type WaveStarted struct {
	Index     int      `json:"index"`
	TestNames []string `json:"test_names"`
	Parallel  bool     `json:"parallel"`
}

// WaveCompleted is emitted after a parallel wave has finished.
// It embeds WaveResult so all wave fields are available on the event.
type WaveCompleted struct {
	WaveResult
}

// TestStarted is emitted before a single test case runs.
type TestStarted struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Protocol Protocol `json:"protocol"`
	Target   string   `json:"target"`
}

// TestCompleted is emitted after a single test case completes.
// It embeds TestResult so all result fields are available on the event.
type TestCompleted struct {
	TestResult
}

// GraphResolved is emitted after the dependency graph is built,
// before execution starts.
type GraphResolved struct {
	TotalWaves       int      `json:"total_waves"`
	Dependencies     int      `json:"dependencies"`
	SaveRequirements int      `json:"save_requirements"`
	Warnings         []string `json:"warnings,omitempty"`
}

// PluginLoaded is emitted when a plugin is successfully loaded.
type PluginLoaded struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
}

// ConfigLoaded is emitted after the .qube.yaml config is parsed.
type ConfigLoaded struct {
	Path        string   `json:"path"`
	Targets     []string `json:"targets,omitempty"`
	PluginCount int      `json:"plugin_count"`
}

// TemplateError is emitted when a template reference cannot be resolved.
// Execution may continue or halt depending on the error severity.
type TemplateError struct {
	File       string `json:"file"`
	Test       string `json:"test"`
	Expression string `json:"expression"`
	Message    string `json:"message"`
}

// Progress is emitted periodically to report execution progress.
type Progress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
	Wave      int `json:"wave"`
}

// PluginEvent is the universal wrapper for events emitted by plugins.
//
// Plugins (WASM or built-in) declare their events via PluginInfo.Events at load
// time. When a plugin emits an event, the engine builds a PluginEvent with the
// plugin name, event kind, and typed payload data (map[string]any).
//
// Frontends subscribe via Dispatcher using one of:
//   - SubscribePluginEvent — raw handler by fully-qualified name
//   - SubscribePluginTyped — decoded into a user-defined Go struct
type PluginEvent struct {
	Plugin string         `json:"plugin"`         // plugin name, e.g. "grpc"
	Kind   string         `json:"kind"`           // event kind, e.g. "StreamMessageReceived"
	Data   map[string]any `json:"data,omitempty"` // event payload matching the declared schema
}

// FullName returns the fully-qualified event name: "<plugin>.<kind>".
// Used as the key for subscription routing.
func (e PluginEvent) FullName() string {
	return e.Plugin + "." + e.Kind
}

// sealed() marker methods — restrict Event implementations to this package.

func (RunStarted) sealed()    {}
func (RunCompleted) sealed()  {}
func (WaveStarted) sealed()   {}
func (WaveCompleted) sealed() {}
func (TestStarted) sealed()   {}
func (TestCompleted) sealed() {}
func (GraphResolved) sealed() {}
func (PluginLoaded) sealed()  {}
func (ConfigLoaded) sealed()  {}
func (TemplateError) sealed() {}
func (Progress) sealed()      {}
func (PluginEvent) sealed()   {}

// Type() returns a stable string identifier for each event type.
// Used for JSON/SSE serialization where the Go type is not preserved.

func (RunStarted) Type() string    { return "RunStarted" }
func (RunCompleted) Type() string  { return "RunCompleted" }
func (WaveStarted) Type() string   { return "WaveStarted" }
func (WaveCompleted) Type() string { return "WaveCompleted" }
func (TestStarted) Type() string   { return "TestStarted" }
func (TestCompleted) Type() string { return "TestCompleted" }
func (GraphResolved) Type() string { return "GraphResolved" }
func (PluginLoaded) Type() string  { return "PluginLoaded" }
func (ConfigLoaded) Type() string  { return "ConfigLoaded" }
func (TemplateError) Type() string { return "TemplateError" }
func (Progress) Type() string      { return "Progress" }
func (PluginEvent) Type() string   { return "PluginEvent" }
