package engine

import (
	"time"

	"github.com/apiqube/engine/internal/wire"
)

// EventHandler receives events from the engine during execution.
// Frontends implement this with a type switch to handle events they care about.
type EventHandler interface {
	Handle(event Event)
}

// Event is a sealed interface — only types in this package can implement it.
type Event interface {
	sealed()
	Type() string
}

// NopHandler is an EventHandler that discards all events.
type NopHandler struct{}

// Handle implements EventHandler.
func (NopHandler) Handle(Event) {}

// PluginEvent is the universal wrapper for events emitted by plugins.
//
// Plugins (WASM or built-in) declare their events via PluginInfo.Events at
// load time. When a plugin emits an event, the engine builds a PluginEvent
// with the plugin name, event kind, and typed payload data.
//
// Frontends subscribe via Dispatcher using SubscribePluginEvent (raw) or
// SubscribePluginTyped (decoded into a user-defined Go struct).
type PluginEvent struct {
	Plugin string         `json:"plugin"`
	Kind   string         `json:"kind"`
	Data   map[string]any `json:"data,omitempty"`
}

// FullName returns the fully-qualified event name "<plugin>.<kind>".
func (e PluginEvent) FullName() string {
	return e.Plugin + "." + e.Kind
}

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
type TestCompleted struct {
	TestResult
}

// GraphResolved is emitted after the dependency graph is built, before execution starts.
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

// sealed() marker methods.

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

// fromWirePluginEvent converts the internal-wire form to the public PluginEvent.
func fromWirePluginEvent(e wire.PluginEvent) PluginEvent {
	return PluginEvent{Plugin: e.Plugin, Kind: e.Kind, Data: e.Data}
}
