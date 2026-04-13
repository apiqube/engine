package engine

import "time"

// EventHandler receives events from the engine during execution.
// Frontends implement this with a type switch to handle events they care about.
type EventHandler interface {
	Handle(event Event)
}

// Event is a sealed interface — only types in this package can implement it.
type Event interface {
	sealed()
}

// Signal represents a control command from frontend to engine.
type Signal int

const (
	SignalPause    Signal = iota
	SignalResume
	SignalSkipTest
)

// TestStatus represents the outcome of a test case.
type TestStatus int

const (
	StatusPassed  TestStatus = iota
	StatusFailed
	StatusSkipped
	StatusErrored
)

func (s TestStatus) String() string {
	switch s {
	case StatusPassed:
		return "passed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	case StatusErrored:
		return "errored"
	default:
		return "unknown"
	}
}

// Run events

type RunStarted struct {
	Files      []string `json:"files"`
	TotalTests int      `json:"total_tests"`
	TotalWaves int      `json:"total_waves"`
}

type RunCompleted struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Errored  int           `json:"errored"`
	Duration time.Duration `json:"duration"`
}

// Wave events

type WaveStarted struct {
	Index     int      `json:"index"`
	TestNames []string `json:"test_names"`
	Parallel  bool     `json:"parallel"`
}

type WaveCompleted struct {
	Index    int           `json:"index"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Duration time.Duration `json:"duration"`
}

// Test events

type TestStarted struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Protocol string `json:"protocol"`
	Target   string `json:"target"`
}

type TestCompleted struct {
	Name       string            `json:"name"`
	File       string            `json:"file"`
	Status     TestStatus        `json:"status"`
	Duration   time.Duration     `json:"duration"`
	Protocol   string            `json:"protocol"`
	Request    *RequestData      `json:"request,omitempty"`
	Response   *ResponseData     `json:"response,omitempty"`
	Assertions []AssertionResult `json:"assertions,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// Informational events (replace OnLog)

type GraphResolved struct {
	TotalWaves       int      `json:"total_waves"`
	Dependencies     int      `json:"dependencies"`
	SaveRequirements int      `json:"save_requirements"`
	Warnings         []string `json:"warnings,omitempty"`
}

type PluginLoaded struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
}

type ConfigLoaded struct {
	Path        string   `json:"path"`
	Targets     []string `json:"targets,omitempty"`
	PluginCount int      `json:"plugin_count"`
}

type TemplateError struct {
	File       string `json:"file"`
	Test       string `json:"test"`
	Expression string `json:"expression"`
	Message    string `json:"message"`
}

type Progress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
	Wave      int `json:"wave"`
}

// PluginEvent allows plugins to emit custom events without changing core.
type PluginEvent struct {
	Plugin string         `json:"plugin"`
	Kind   string         `json:"kind"`
	Data   map[string]any `json:"data,omitempty"`
}

// Seal all event types

func (RunStarted) sealed()     {}
func (RunCompleted) sealed()   {}
func (WaveStarted) sealed()    {}
func (WaveCompleted) sealed()  {}
func (TestStarted) sealed()    {}
func (TestCompleted) sealed()  {}
func (GraphResolved) sealed()  {}
func (PluginLoaded) sealed()   {}
func (ConfigLoaded) sealed()   {}
func (TemplateError) sealed()  {}
func (Progress) sealed()       {}
func (PluginEvent) sealed()    {}

// Supporting data types

type RequestData struct {
	Method   string            `json:"method,omitempty"`
	URL      string            `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     any               `json:"body,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type ResponseData struct {
	Status   any               `json:"status"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     any               `json:"body,omitempty"`
	Duration time.Duration     `json:"duration"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type AssertionResult struct {
	Expression string `json:"expression"`
	Passed     bool   `json:"passed"`
	Expected   any    `json:"expected,omitempty"`
	Actual     any    `json:"actual,omitempty"`
	Message    string `json:"message,omitempty"`
}

// NopHandler is a no-op EventHandler that discards all events.
type NopHandler struct{}

func (NopHandler) Handle(Event) {}
