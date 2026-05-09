// Package wire holds the value types shared between the public engine
// package and internal subsystems (plugin, runner). Engine re-exports each
// type as an alias so the public API surface is unchanged.
//
// The package exists solely to break the would-be import cycle:
//   engine ⇄ internal/{plugin,runner}
//
// Subsystems import wire; engine imports both wire and the subsystems.
package wire

// Protocol identifies a target protocol, derived from the target URL scheme
// or declared explicitly by a plugin.
type Protocol string

const (
	ProtocolHTTP    Protocol = "http"
	ProtocolHTTPS   Protocol = "https"
	ProtocolGRPC    Protocol = "grpc"
	ProtocolGRPCS   Protocol = "grpcs"
	ProtocolWS      Protocol = "ws"
	ProtocolWSS     Protocol = "wss"
	ProtocolGraphQL Protocol = "graphql"
	ProtocolSQL     Protocol = "sql"
	ProtocolKafka   Protocol = "kafka"
	ProtocolAMQP    Protocol = "amqp"
	ProtocolRedis   Protocol = "redis"
)

// String returns the protocol as a plain string.
func (p Protocol) String() string { return string(p) }

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

// FieldSpec describes one field in a plugin contract.
type FieldSpec struct {
	Type        FieldType `json:"type"`
	Required    bool      `json:"required"`
	Description string    `json:"description"`
}

// EventSpec describes a single event a plugin can emit at runtime.
type EventSpec struct {
	Description string               `json:"description"`
	Fields      map[string]FieldSpec `json:"fields"`
}

// PluginSchema is the public introspection shape for a loaded plugin.
type PluginSchema struct {
	Name         string               `json:"name"`
	Version      string               `json:"version"`
	Description  string               `json:"description"`
	Protocols    []Protocol           `json:"protocols"`
	Capabilities []string             `json:"capabilities,omitempty"`
	Fields       map[string]FieldSpec `json:"fields,omitempty"`
	Events       map[string]EventSpec `json:"events,omitempty"`
}

// PluginEvent is the universal wrapper for events emitted by plugins.
type PluginEvent struct {
	Plugin string         `json:"plugin"`
	Kind   string         `json:"kind"`
	Data   map[string]any `json:"data,omitempty"`
}

// FullName returns the fully-qualified event name "<plugin>.<kind>".
func (e PluginEvent) FullName() string {
	return e.Plugin + "." + e.Kind
}

// TestStatus represents the outcome of a test case.
type TestStatus int

const (
	StatusPassed TestStatus = iota
	StatusFailed
	StatusSkipped
	StatusErrored
)

// String returns the canonical name for a status.
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
	}
	return "unknown"
}

// Signal is a control command from frontend to engine.
type Signal int

const (
	SignalPause Signal = iota
	SignalResume
	SignalSkipTest
)

// String returns the canonical signal name.
func (s Signal) String() string {
	switch s {
	case SignalPause:
		return "pause"
	case SignalResume:
		return "resume"
	case SignalSkipTest:
		return "skip"
	}
	return "unknown"
}

// Severity classifies the importance of a validation issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// ValidationError describes a problem found during manifest validation.
type ValidationError struct {
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Field    string   `json:"field,omitempty"`
	Message  string   `json:"message"`
	Severity Severity `json:"severity"`
}
