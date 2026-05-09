package dataflow

import (
	"context"
	"errors"
	"sync"

	"github.com/apiqube/engine/internal/wire"
)

// ErrStoreClosed is returned by WaitFor after Close has been called.
var ErrStoreClosed = errors.New("dataflow: store closed")

// Store holds runtime data for cross-test communication during a single Run.
// One Store per Run — not shared across concurrent runs.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry
	prev    *Snapshot
	events  map[string][]wire.PluginEvent

	closeOnce sync.Once
	done      chan struct{}
}

// entry holds a value and a per-key "done" channel that is closed when the
// value first becomes available. Closing the channel wakes every waiter on
// that key — supports many concurrent WaitFor callers per key.
type entry struct {
	set   bool
	value any
	ready chan struct{}
}

// Snapshot captures a test response for use as `prev` in scenario mode.
type Snapshot struct {
	Name     string
	Status   any
	Headers  map[string]string
	Body     any
	Metadata map[string]any
}

// New creates an empty Store ready for a single Run.
func New() *Store {
	return &Store{
		entries: make(map[string]*entry),
		events:  make(map[string][]wire.PluginEvent),
		done:    make(chan struct{}),
	}
}

// Get retrieves a named value synchronously.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok || !e.set {
		return nil, false
	}
	return e.value, true
}

// Set stores a named value and wakes any WaitFor callers blocked on the key.
// Subsequent Sets overwrite the value but do not re-notify (waiters have moved on).
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	e := s.entryLocked(key)
	first := !e.set
	e.value = value
	e.set = true
	s.mu.Unlock()
	if first {
		close(e.ready)
	}
}

// SetPrev updates the previous-test snapshot (scenario mode).
func (s *Store) SetPrev(snap *Snapshot) {
	s.mu.Lock()
	s.prev = snap
	s.mu.Unlock()
}

// GetPrev returns the most recent prev snapshot, or nil.
func (s *Store) GetPrev() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prev
}

// WaitFor blocks until a value is available for key, ctx is cancelled, or the
// Store is closed. Multiple goroutines may wait on the same key.
func (s *Store) WaitFor(ctx context.Context, key string) (any, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	e := s.entryLocked(key)
	if e.set {
		v := e.value
		s.mu.Unlock()
		return v, nil
	}
	ready := e.ready
	s.mu.Unlock()

	select {
	case <-ready:
		s.mu.RLock()
		v := e.value
		s.mu.RUnlock()
		return v, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		return nil, ErrStoreClosed
	}
}

// AppendEvent appends a streaming event to the per-alias event list. Templates
// reference these via `{{ alias.events.N.field }}`.
func (s *Store) AppendEvent(alias string, event wire.PluginEvent) {
	s.mu.Lock()
	s.events[alias] = append(s.events[alias], event)
	s.mu.Unlock()
}

// Events returns a copy of the events list for the given alias. Returns an
// empty slice if no events are recorded.
func (s *Store) Events(alias string) []wire.PluginEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.events[alias]
	if len(src) == 0 {
		return nil
	}
	out := make([]wire.PluginEvent, len(src))
	copy(out, src)
	return out
}

// Close releases every waiter and marks the store closed. Idempotent.
func (s *Store) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
	})
	return nil
}

func (s *Store) checkClosed() error {
	select {
	case <-s.done:
		return ErrStoreClosed
	default:
		return nil
	}
}

// entryLocked returns the entry for key, creating it if absent.
// Caller must hold s.mu (write lock).
func (s *Store) entryLocked(key string) *entry {
	e, ok := s.entries[key]
	if ok {
		return e
	}
	e = &entry{ready: make(chan struct{})}
	s.entries[key] = e
	return e
}
