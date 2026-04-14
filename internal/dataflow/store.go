package dataflow

import (
	"context"
	"sync"
)

// Store holds runtime data for cross-test communication during a single Run().
// One Store instance per Run — not shared across concurrent runs.
type Store struct {
	mu     sync.RWMutex
	values map[string]any       // save: variables, alias results
	prev   *Snapshot            // previous test response (scenario mode)

	chMu     sync.Mutex
	channels map[string]chan any // async waiters for values not yet produced
}

// New creates an empty Store ready for a single Run.
func New() *Store {
	return &Store{
		values:   make(map[string]any),
		channels: make(map[string]chan any),
	}
}

// Snapshot captures a test response for use as `prev` in scenario mode.
type Snapshot struct {
	Name     string
	Status   any
	Headers  map[string]string
	Body     any
	Metadata map[string]any
}

// Get retrieves a named variable synchronously.
// Returns (value, true) if present, (nil, false) otherwise.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.values[key]
	return v, ok
}

// Set stores a named variable and notifies any async waiters on that key.
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	s.values[key] = value
	s.mu.Unlock()
	s.notify(key, value)
}

// SetPrev updates the `prev` snapshot. Called by runner after each test in scenario mode.
func (s *Store) SetPrev(snap *Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prev = snap
}

// GetPrev returns the most recent test response, or nil if no prior test.
func (s *Store) GetPrev() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prev
}

// WaitFor blocks until a value is available for the given key, or ctx is cancelled.
// Used when a consumer in a parallel wave needs data from a producer that hasn't finished yet.
func (s *Store) WaitFor(ctx context.Context, key string) (any, error) {
	// TODO: implementation
	// 1. Fast path: check if value already exists → return immediately
	// 2. Slow path: get or create channel for this key, select on channel and ctx.Done()
	return nil, nil
}

// notify wakes up all waiters on a given key with the supplied value.
func (s *Store) notify(key string, value any) {
	s.chMu.Lock()
	defer s.chMu.Unlock()
	if ch, ok := s.channels[key]; ok {
		select {
		case ch <- value:
		default:
			// buffer full, consumer will get value via Get()
		}
	}
}
