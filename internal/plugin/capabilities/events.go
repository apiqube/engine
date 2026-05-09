package capabilities

import (
	"context"
	"encoding/json"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/apiqube/engine/internal/wire"
)

func resolveEventSink(s *Sinks) func(wire.PluginEvent) {
	if s == nil || s.Events == nil {
		return func(wire.PluginEvent) {}
	}
	return s.Events
}

// EmitEventLogic decodes a single event from JSON bytes and forwards it.
// Exposed for unit testing without a WASM module.
func EmitEventLogic(data []byte, sink func(wire.PluginEvent)) error {
	if len(data) == 0 {
		return nil
	}
	var ev wire.PluginEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return err
	}
	if sink != nil {
		sink(ev)
	}
	return nil
}

func addEvents(builder wazero.HostModuleBuilder) {
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr, length uint32) {
			if length == 0 {
				return
			}
			data, ok := mod.Memory().Read(ptr, length)
			if !ok {
				return
			}
			_ = EmitEventLogic(data, resolveEventSink(sinksFrom(ctx)))
		}).
		Export("host_emit_event")
}
