package plugin

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Loader manages plugin discovery and caching for an Engine.
//
// The cache is keyed by (pluginDir, fileName, mtime). A second LoadDir() call
// for the same directory whose files have not changed reuses the existing
// runtime and Registry; otherwise the runtime is rebuilt.
type Loader struct {
	mu      sync.Mutex
	current *cached
}

type cached struct {
	dir      string
	stamps   map[string]int64 // filename → mtime UnixNano
	runtime  *WASMRuntime
	registry *Registry
}

// NewLoader returns an empty Loader.
func NewLoader() *Loader { return &Loader{} }

// Current returns the most-recently loaded Registry, or nil if LoadDir has
// never been called. Used by Engine for introspection without requiring a
// context or pluginDir.
func (l *Loader) Current() *Registry {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.current == nil {
		return nil
	}
	return l.current.registry
}

// LoadDir scans dir for *.wasm files, instantiates each plugin, and returns
// the populated Registry. If dir is unchanged since the previous call, the
// cached Registry is reused.
//
// dir == "" returns an empty Registry without scanning. Missing directory
// returns an error.
func (l *Loader) LoadDir(ctx context.Context, dir string) (*Registry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if dir == "" {
		if l.current != nil {
			return l.current.registry, nil
		}
		empty := &cached{registry: NewRegistry()}
		l.current = empty
		return empty.registry, nil
	}

	stamps, err := collectStamps(dir)
	if err != nil {
		return nil, err
	}

	if l.current != nil && l.current.dir == dir && stampsEqual(l.current.stamps, stamps) {
		return l.current.registry, nil
	}

	if l.current != nil {
		_ = l.current.registry.Close()
		_ = l.current.runtime.Close(ctx)
	}

	rt, err := NewWASMRuntime(ctx)
	if err != nil {
		return nil, err
	}
	registry := NewRegistry()

	for fname := range stamps {
		fullPath := filepath.Join(dir, fname)
		bytes, err := os.ReadFile(fullPath)
		if err != nil {
			_ = rt.Close(ctx)
			return nil, fmt.Errorf("read %s: %w", fullPath, err)
		}
		p, err := rt.LoadPlugin(ctx, fname, bytes)
		if err != nil {
			_ = rt.Close(ctx)
			return nil, fmt.Errorf("load %s: %w", fullPath, err)
		}
		if err := registry.Register(p); err != nil {
			_ = p.Close()
			_ = rt.Close(ctx)
			return nil, err
		}
	}

	l.current = &cached{dir: dir, stamps: stamps, runtime: rt, registry: registry}
	return registry, nil
}

// Close releases the cached registry and runtime, if any.
func (l *Loader) Close(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.current == nil {
		return nil
	}
	c := l.current
	l.current = nil
	if c.registry != nil {
		_ = c.registry.Close()
	}
	if c.runtime != nil {
		return c.runtime.Close(ctx)
	}
	return nil
}

// collectStamps walks dir non-recursively and returns the modtime of each
// *.wasm file.
func collectStamps(dir string) (map[string]int64, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("plugin: %s is not a directory", dir)
	}
	stamps := make(map[string]int64)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, ent := range entries {
		if !validWASMEntry(ent) {
			continue
		}
		fi, err := ent.Info()
		if err != nil {
			return nil, err
		}
		stamps[ent.Name()] = fi.ModTime().UnixNano()
	}
	return stamps, nil
}

func validWASMEntry(ent fs.DirEntry) bool {
	if ent.IsDir() {
		return false
	}
	return strings.HasSuffix(strings.ToLower(ent.Name()), ".wasm")
}

func stampsEqual(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || vb != va {
			return false
		}
	}
	return true
}
