package plugin

import (
	"context"

	"github.com/apiqube/engine/internal/wire"
)

// Plugin is the host-side view of a loaded plugin. Implementations include
// the wazero-backed WASM adapter (wasm.go) and in-process fakes used in tests.
type Plugin interface {
	Info() PluginInfo
	Init(ctx context.Context, config map[string]any) error
	Validate(input TestInput) []FieldError
	Execute(ctx context.Context, input TestInput, emit EventSink) (*TestOutput, error)
	Close() error
}

// EventSink receives plugin events emitted during Execute.
// The runner provides a sink that fans events out to the engine's EventHandler
// and into the dataflow store so they can be referenced by later tests.
type EventSink func(wire.PluginEvent)

// PluginInfo mirrors the metadata a WASM plugin returns from plugin_info().
// For built-in/fake plugins, the host constructs it directly.
type PluginInfo struct {
	Name         string                    `json:"name"`
	Version      string                    `json:"version"`
	Description  string                    `json:"description"`
	Protocols    []wire.Protocol           `json:"protocols"`
	Capabilities []string                  `json:"capabilities,omitempty"`
	Fields       map[string]wire.FieldSpec `json:"fields,omitempty"`
	Events       map[string]wire.EventSpec `json:"events,omitempty"`
}

// FieldError reports a validation problem on a specific manifest field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// TestInput is what the host passes to Execute.
//
// Method and Resource are core fields readable by every plugin (HTTP method,
// gRPC method-name, GraphQL operation, SQL query name, etc.). Plugin-specific
// data lives in Fields.
type TestInput struct {
	Method   string            `json:"method,omitempty"`
	Resource string            `json:"resource,omitempty"`
	Target   string            `json:"target"`
	Headers  map[string]string `json:"headers,omitempty"`
	Timeout  string            `json:"timeout,omitempty"`
	Fields   map[string]any    `json:"fields,omitempty"`
}

// TestOutput is what Execute returns.
//
// Events carries plugin events accumulated during a single Execute call;
// streaming plugins also emit via EventSink during the call (both paths feed
// the same downstream sinks).
type TestOutput struct {
	Status     any                `json:"status"`
	Headers    map[string]string  `json:"headers,omitempty"`
	Body       any                `json:"body,omitempty"`
	DurationMs int64              `json:"duration_ms"`
	Error      string             `json:"error,omitempty"`
	Metadata   map[string]any     `json:"metadata,omitempty"`
	Events     []wire.PluginEvent `json:"events,omitempty"`
}

// Snapshot converts a PluginInfo into the public introspection shape.
func (p PluginInfo) Snapshot() wire.PluginSchema {
	return wire.PluginSchema{
		Name:         p.Name,
		Version:      p.Version,
		Description:  p.Description,
		Protocols:    p.Protocols,
		Capabilities: append([]string(nil), p.Capabilities...),
		Fields:       p.Fields,
		Events:       p.Events,
	}
}
