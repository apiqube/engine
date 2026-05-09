// Package capabilities provides the host-side functions WASM plugins import
// to interact with the outside world.
//
// All host functions live in a single import namespace ("apiqube") and are
// registered once per runtime. Plugins import only what they need; unused
// imports are inert. PluginInfo.Capabilities documents which logical
// capabilities a plugin uses, but the host module itself is not partitioned.
//
// Sinks (logger, event handler, HTTP client) are passed per-call via context
// using WithSinks/sinksFrom. This keeps host-module instantiation a one-time
// cost per Engine while supporting concurrent Run() calls each with their own
// sinks.
package capabilities

import (
	"context"
	"net/http"
	"slices"

	"github.com/tetratelabs/wazero"

	"github.com/apiqube/engine/internal/wire"
)

// HostNamespace is the WASM import namespace used by all host functions.
// Plugins target it via //go:wasmimport "apiqube" host_X.
const HostNamespace = "apiqube"

// Sinks holds dynamic per-call dependencies a host function needs.
type Sinks struct {
	Logger LogSink
	Now    func() int64
	Events func(wire.PluginEvent)
	HTTP   HTTPClient
}

type sinksKey struct{}

// WithSinks returns a context carrying the given Sinks. Capability functions
// read the sinks from this context value.
func WithSinks(ctx context.Context, s *Sinks) context.Context {
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, sinksKey{}, s)
}

func sinksFrom(ctx context.Context) *Sinks {
	if s, ok := ctx.Value(sinksKey{}).(*Sinks); ok && s != nil {
		return s
	}
	return &Sinks{}
}

// InstantiateHostModule registers the universal host module on rt with every
// supported host function. Call once per runtime, before instantiating any
// plugin module.
func InstantiateHostModule(ctx context.Context, rt wazero.Runtime) error {
	builder := rt.NewHostModuleBuilder(HostNamespace)
	addLog(builder)
	addTime(builder)
	addEvents(builder)
	addHTTP(builder)
	_, err := builder.Instantiate(ctx)
	return err
}

// SupportedCapabilities lists the logical capability names this build provides.
// The plugin loader uses this to validate PluginInfo.Capabilities before
// running a plugin.
func SupportedCapabilities() []string {
	return []string{"http"}
}

// IsSupported reports whether the named capability is implemented.
func IsSupported(name string) bool {
	return slices.Contains(SupportedCapabilities(), name)
}

// defaultClient returns a zero-config http.Client for cases where the caller
// did not provide one.
func defaultClient() HTTPClient {
	return &http.Client{}
}
