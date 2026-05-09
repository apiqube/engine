package plugin

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/apiqube/engine/internal/wire"
)

// Registry holds loaded plugins indexed by protocol. A Registry is safe for
// concurrent reads after Register; mutating operations (Register, Close) take
// a write lock.
type Registry struct {
	mu        sync.RWMutex
	byName    map[string]*entry
	byProto   map[wire.Protocol]*entry
	closed    bool
}

type entry struct {
	plugin Plugin
	info   PluginInfo
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		byName:  make(map[string]*entry),
		byProto: make(map[wire.Protocol]*entry),
	}
}

// Register installs the plugin in the registry. The plugin's Info() is cached
// and indexed by name and by every protocol it claims.
//
// Returns an error on duplicate name, duplicate protocol, or after Close.
func (r *Registry) Register(p Plugin) error {
	if p == nil {
		return fmt.Errorf("plugin: cannot register nil plugin")
	}
	info := p.Info()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("plugin: registry closed")
	}
	if _, exists := r.byName[info.Name]; exists {
		return fmt.Errorf("plugin: duplicate plugin name %q", info.Name)
	}
	for _, proto := range info.Protocols {
		if other, exists := r.byProto[proto]; exists {
			return fmt.Errorf("plugin: protocol %q already provided by %q", proto, other.info.Name)
		}
	}
	e := &entry{plugin: p, info: info}
	r.byName[info.Name] = e
	for _, proto := range info.Protocols {
		r.byProto[proto] = e
	}
	return nil
}

// Lookup returns the plugin for the given protocol, or nil if none.
func (r *Registry) Lookup(proto wire.Protocol) Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if e, ok := r.byProto[proto]; ok {
		return e.plugin
	}
	return nil
}

// Snapshot returns a stable, sorted slice of plugin metadata for introspection.
func (r *Registry) Snapshot() []wire.PluginSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]wire.PluginSchema, 0, len(r.byName))
	for _, e := range r.byName {
		out = append(out, e.info.Snapshot())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// EventSchema looks up an event by fully-qualified name "<plugin>.<kind>".
// Returns nil if not found.
func (r *Registry) EventSchema(fqn string) *wire.EventSpec {
	for i := len(fqn) - 1; i >= 0; i-- {
		if fqn[i] != '.' {
			continue
		}
		pluginName, kind := fqn[:i], fqn[i+1:]
		r.mu.RLock()
		e, ok := r.byName[pluginName]
		r.mu.RUnlock()
		if !ok {
			return nil
		}
		spec, ok := e.info.Events[kind]
		if !ok {
			return nil
		}
		return &spec
	}
	return nil
}

// Close shuts down every registered plugin and marks the registry closed.
// Returns the first error encountered; remaining plugins are still closed.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true

	var firstErr error
	for _, e := range r.byName {
		if err := e.plugin.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// InitAll runs Init on every registered plugin with its config from
// pluginConfigs[name] (nil-safe lookup yields nil config).
func (r *Registry) InitAll(ctx context.Context, pluginConfigs map[string]map[string]any) error {
	r.mu.RLock()
	plugins := make([]*entry, 0, len(r.byName))
	for _, e := range r.byName {
		plugins = append(plugins, e)
	}
	r.mu.RUnlock()

	for _, e := range plugins {
		if err := e.plugin.Init(ctx, pluginConfigs[e.info.Name]); err != nil {
			return fmt.Errorf("init %s: %w", e.info.Name, err)
		}
	}
	return nil
}
