package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/apiqube/engine/internal/plugin/capabilities"
	"github.com/apiqube/engine/internal/wire"
)

// WASM export names used by the plugin contract.
const (
	exportInfo     = "plugin_info"
	exportInit     = "plugin_init"
	exportValidate = "validate"
	exportExecute  = "execute"
)

// WASMRuntime owns the wazero runtime and host module for one or more plugins.
// Multiple plugins share the same host module, instantiated once at runtime
// creation.
type WASMRuntime struct {
	rt wazero.Runtime
}

// NewWASMRuntime returns a runtime preloaded with WASI Preview 1 and the
// universal "apiqube" host module containing every supported host function.
func NewWASMRuntime(ctx context.Context) (*WASMRuntime, error) {
	rt := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	if err := capabilities.InstantiateHostModule(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("instantiate host module: %w", err)
	}
	return &WASMRuntime{rt: rt}, nil
}

// Close releases all plugins and the underlying runtime.
func (w *WASMRuntime) Close(ctx context.Context) error {
	return w.rt.Close(ctx)
}

// ErrMissingCapability is returned when a plugin declares a capability the
// engine does not provide.
type ErrMissingCapability struct {
	Plugin     string
	Capability string
}

func (e *ErrMissingCapability) Error() string {
	return fmt.Sprintf("plugin %q requires unsupported capability %q", e.Plugin, e.Capability)
}

// LoadPlugin compiles and instantiates one .wasm module on this runtime.
// Capabilities declared by the plugin are validated against the engine's
// supported set; an unsupported capability returns *ErrMissingCapability.
//
// Plugins are expected to be WASI reactor modules (e.g. TinyGo's
// `-buildmode=c-shared`) that export `_initialize` for runtime setup but
// do not call proc_exit. We try `_initialize` first; if absent, fall back
// to the default `_start` (command mode).
func (w *WASMRuntime) LoadPlugin(ctx context.Context, name string, wasmBytes []byte) (*WASMPlugin, error) {
	compiled, err := w.rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", name, err)
	}

	config := wazero.NewModuleConfig().WithName(name)
	if _, ok := compiled.ExportedFunctions()["_initialize"]; ok {
		config = config.WithStartFunctions("_initialize")
	}

	mod, err := w.rt.InstantiateModule(ctx, compiled, config)
	if err != nil {
		_ = compiled.Close(ctx)
		return nil, fmt.Errorf("instantiate %s: %w", name, err)
	}

	p := &WASMPlugin{name: name, mod: mod, compiled: compiled, rt: w}
	info, err := p.fetchInfo(ctx)
	if err != nil {
		_ = p.Close()
		return nil, err
	}
	for _, capName := range info.Capabilities {
		if !capabilities.IsSupported(capName) {
			_ = p.Close()
			return nil, &ErrMissingCapability{Plugin: name, Capability: capName}
		}
	}
	p.info = info
	return p, nil
}

// WASMPlugin implements Plugin against a wazero-instantiated WASM module.
type WASMPlugin struct {
	name     string
	mod      api.Module
	compiled wazero.CompiledModule
	rt       *WASMRuntime
	info     PluginInfo
}

// Info returns the cached PluginInfo from plugin_info.
func (p *WASMPlugin) Info() PluginInfo { return p.info }

// Init calls plugin_init with the JSON-encoded config.
func (p *WASMPlugin) Init(ctx context.Context, config map[string]any) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal init config: %w", err)
	}
	_, err = p.callRoundTrip(ctx, exportInit, data)
	return err
}

// Validate calls validate and decodes the result as []FieldError.
func (p *WASMPlugin) Validate(input TestInput) []FieldError {
	ctx := context.Background()
	data, err := json.Marshal(input)
	if err != nil {
		return []FieldError{{Message: "marshal input: " + err.Error()}}
	}
	resp, err := p.callRoundTrip(ctx, exportValidate, data)
	if err != nil {
		return []FieldError{{Message: err.Error()}}
	}
	if len(resp) == 0 {
		return nil
	}
	var out []FieldError
	if err := json.Unmarshal(resp, &out); err != nil {
		return []FieldError{{Message: "decode validate result: " + err.Error()}}
	}
	return out
}

// Execute calls execute with the input. Events emitted via host_emit_event
// during the call are routed through emit; events embedded in the returned
// TestOutput are also forwarded for parity.
func (p *WASMPlugin) Execute(ctx context.Context, input TestInput, emit EventSink) (*TestOutput, error) {
	sinks := &capabilities.Sinks{Events: func(ev wire.PluginEvent) {
		if emit != nil {
			emit(ev)
		}
	}}
	ctx = capabilities.WithSinks(ctx, sinks)

	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w", err)
	}
	resp, err := p.callRoundTrip(ctx, exportExecute, data)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return &TestOutput{}, nil
	}
	var out TestOutput
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, fmt.Errorf("decode execute result: %w", err)
	}
	if emit != nil {
		for _, ev := range out.Events {
			emit(ev)
		}
	}
	return &out, nil
}

// Close releases the underlying module. The runtime is closed by WASMRuntime.
func (p *WASMPlugin) Close() error {
	ctx := context.Background()
	if p.mod != nil {
		_ = p.mod.Close(ctx)
		p.mod = nil
	}
	if p.compiled != nil {
		_ = p.compiled.Close(ctx)
		p.compiled = nil
	}
	return nil
}

// fetchInfo invokes plugin_info and decodes the returned bytes as PluginInfo.
func (p *WASMPlugin) fetchInfo(ctx context.Context) (PluginInfo, error) {
	infoFn := p.mod.ExportedFunction(exportInfo)
	if infoFn == nil {
		return PluginInfo{}, fmt.Errorf("plugin %q missing export %q", p.name, exportInfo)
	}
	res, err := infoFn.Call(ctx)
	if err != nil {
		return PluginInfo{}, fmt.Errorf("plugin_info: %w", err)
	}
	if len(res) == 0 {
		return PluginInfo{}, errors.New("plugin_info returned no value")
	}
	data, err := readPacked(p.mod.Memory(), res[0])
	if err != nil {
		return PluginInfo{}, err
	}
	var info PluginInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return PluginInfo{}, fmt.Errorf("decode plugin_info: %w", err)
	}
	return info, nil
}

// callRoundTrip writes data into plugin memory, invokes the named export, and
// returns the bytes that the plugin allocated for the response.
func (p *WASMPlugin) callRoundTrip(ctx context.Context, exportName string, data []byte) ([]byte, error) {
	fn := p.mod.ExportedFunction(exportName)
	if fn == nil {
		return nil, fmt.Errorf("plugin %q missing export %q", p.name, exportName)
	}

	var inputPacked uint64
	if len(data) > 0 {
		var err error
		inputPacked, err = writePacked(ctx, p.mod, data)
		if err != nil {
			return nil, fmt.Errorf("%s: write input: %w", exportName, err)
		}
	}

	res, err := fn.Call(ctx, inputPacked)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", exportName, err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	if res[0] == 0 {
		return nil, nil
	}
	return readPacked(p.mod.Memory(), res[0])
}
