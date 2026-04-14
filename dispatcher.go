package engine

import (
	"encoding/json"
	"reflect"
	"sync"
)

// Dispatcher routes events to handlers registered by type or by plugin event name.
// It implements EventHandler, so you can pass it to WithHandler() just like any
// other handler — frontends that need typed subscriptions use Dispatcher,
// simpler frontends can implement EventHandler directly with a type switch.
//
// Three subscription styles coexist:
//
//   - Subscribe[T]              — typed handler for a core engine event (T must be Event)
//   - SubscribePluginEvent      — raw handler for a plugin event by fully-qualified name
//   - SubscribePluginTyped[T]   — decodes a plugin event's Data into a user-defined T
//
// Dispatcher is safe for concurrent use — subscriptions and dispatches can
// happen from any goroutine.
type Dispatcher struct {
	mu sync.RWMutex

	// typedHandlers routes core Event values to typed handlers by their
	// concrete Go type. Key is reflect.Type of the event struct.
	typedHandlers map[reflect.Type][]func(Event)

	// rawPluginHandlers routes PluginEvent values to raw handlers by
	// fully-qualified name ("plugin.kind"). Handlers see the full PluginEvent
	// with its map[string]any Data.
	rawPluginHandlers map[string][]func(PluginEvent)

	// typedPluginHandlers routes PluginEvent values to decoder closures by
	// fully-qualified name. Each decoder JSON-unmarshals Data into a
	// user-defined Go struct and invokes the consumer's handler.
	typedPluginHandlers map[string][]func(PluginEvent)
}

// NewDispatcher creates an empty Dispatcher with no registered handlers.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		typedHandlers:       make(map[reflect.Type][]func(Event)),
		rawPluginHandlers:   make(map[string][]func(PluginEvent)),
		typedPluginHandlers: make(map[string][]func(PluginEvent)),
	}
}

// Handle implements EventHandler. It dispatches the event to all registered
// handlers matching either its Go type or (for plugin events) its fully-qualified name.
func (d *Dispatcher) Handle(event Event) {
	// Core path: handlers registered with Subscribe[T] for the exact Go type.
	d.mu.RLock()
	typed := d.typedHandlers[reflect.TypeOf(event)]
	d.mu.RUnlock()

	for _, fn := range typed {
		fn(event)
	}

	// Plugin path: additional handlers routed by fully-qualified name.
	if pe, ok := event.(PluginEvent); ok {
		name := pe.FullName()

		d.mu.RLock()
		raw := d.rawPluginHandlers[name]
		decoders := d.typedPluginHandlers[name]
		d.mu.RUnlock()

		for _, fn := range raw {
			fn(pe)
		}
		for _, dec := range decoders {
			dec(pe)
		}
	}
}

// Subscribe registers a handler for a core engine event type.
// T must be one of the sealed Event types declared in this package
// (RunStarted, TestCompleted, WaveCompleted, Progress, ...).
//
// Multiple handlers can subscribe to the same event type — all of them will
// be invoked in registration order when the event is dispatched.
//
// Example:
//
//	d := engine.NewDispatcher()
//	engine.Subscribe(d, func(e engine.TestCompleted) {
//	    fmt.Println(e.Name, e.Status)
//	})
func Subscribe[T Event](d *Dispatcher, fn func(T)) {
	var zero T
	t := reflect.TypeOf(zero)

	// The wrapped closure accepts Event and performs a safe cast because
	// the dispatcher guarantees it only calls this slot with values of type T.
	wrapped := func(e Event) {
		fn(e.(T))
	}

	d.mu.Lock()
	d.typedHandlers[t] = append(d.typedHandlers[t], wrapped)
	d.mu.Unlock()
}

// SubscribePluginEvent registers a raw handler for a plugin event identified
// by its fully-qualified name ("<plugin>.<kind>", e.g. "grpc.StreamError").
//
// Use this when you want the full PluginEvent including its Plugin, Kind,
// and raw Data map — for example, to route events dynamically, log them
// generically, or handle events whose shape you do not want to model as a
// Go struct.
//
// Example:
//
//	engine.SubscribePluginEvent(d, "grpc.StreamError", func(e engine.PluginEvent) {
//	    log.Printf("%s on stream %v", e.Kind, e.Data["stream_id"])
//	})
func SubscribePluginEvent(d *Dispatcher, name string, fn func(PluginEvent)) {
	d.mu.Lock()
	d.rawPluginHandlers[name] = append(d.rawPluginHandlers[name], fn)
	d.mu.Unlock()
}

// SubscribePluginTyped registers a typed handler for a plugin event.
// The consumer defines a Go struct matching the fields declared in the
// plugin's EventSpec schema; Dispatcher decodes PluginEvent.Data into that
// struct via JSON round-trip and invokes the handler with the typed value.
//
// Decoding errors are silently dropped — the handler simply does not fire.
// If you need strict error handling, use SubscribePluginEvent and decode
// manually, or query the plugin schema via Engine.Plugins() at startup to
// validate your struct against the declared fields.
//
// Example:
//
//	type StreamMessage struct {
//	    StreamID string `json:"stream_id"`
//	    Count    int    `json:"count"`
//	    Bytes    int64  `json:"bytes"`
//	}
//
//	engine.SubscribePluginTyped(d, "grpc.StreamMessageReceived",
//	    func(e StreamMessage) {
//	        metrics.Record(e.StreamID, e.Count)
//	    })
func SubscribePluginTyped[T any](d *Dispatcher, name string, fn func(T)) {
	decoder := func(pe PluginEvent) {
		raw, err := json.Marshal(pe.Data)
		if err != nil {
			return
		}
		var target T
		if err := json.Unmarshal(raw, &target); err != nil {
			return
		}
		fn(target)
	}

	d.mu.Lock()
	d.typedPluginHandlers[name] = append(d.typedPluginHandlers[name], decoder)
	d.mu.Unlock()
}
