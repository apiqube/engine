package dataflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/apiqube/engine/internal/wire"
)

func TestStore_GetSet(t *testing.T) {
	s := New()
	if _, ok := s.Get("k"); ok {
		t.Error("expected miss before Set")
	}
	s.Set("k", 42)
	v, ok := s.Get("k")
	if !ok || v != 42 {
		t.Errorf("Get after Set = (%v, %v); want (42, true)", v, ok)
	}
}

func TestStore_PrevSnapshot(t *testing.T) {
	s := New()
	if got := s.GetPrev(); got != nil {
		t.Error("prev should start nil")
	}
	snap := &Snapshot{Name: "first", Status: 200}
	s.SetPrev(snap)
	if got := s.GetPrev(); got != snap {
		t.Errorf("GetPrev returned wrong snapshot")
	}
}

func TestStore_WaitFor_AlreadySet(t *testing.T) {
	s := New()
	s.Set("k", "v")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, err := s.WaitFor(ctx, "k")
	if err != nil {
		t.Fatalf("WaitFor: %v", err)
	}
	if got != "v" {
		t.Errorf("got %v want v", got)
	}
}

func TestStore_WaitFor_BlocksUntilSet(t *testing.T) {
	s := New()
	ctx := context.Background()

	done := make(chan any, 1)
	go func() {
		v, err := s.WaitFor(ctx, "later")
		if err != nil {
			t.Errorf("WaitFor: %v", err)
		}
		done <- v
	}()

	// Give the waiter a chance to register.
	time.Sleep(20 * time.Millisecond)
	s.Set("later", 99)

	select {
	case v := <-done:
		if v != 99 {
			t.Errorf("got %v want 99", v)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitFor did not return after Set")
	}
}

func TestStore_WaitFor_MultipleWaiters(t *testing.T) {
	s := New()
	ctx := context.Background()

	const n = 10
	results := make(chan any, n)

	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := s.WaitFor(ctx, "shared")
			if err != nil {
				t.Errorf("WaitFor: %v", err)
				return
			}
			results <- v
		})
	}

	time.Sleep(20 * time.Millisecond)
	s.Set("shared", "broadcast")

	wg.Wait()
	close(results)

	count := 0
	for v := range results {
		if v != "broadcast" {
			t.Errorf("got %v want broadcast", v)
		}
		count++
	}
	if count != n {
		t.Errorf("got %d results, want %d", count, n)
	}
}

func TestStore_WaitFor_CtxCancel(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := s.WaitFor(ctx, "never")
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitFor did not return after cancel")
	}
}

func TestStore_WaitFor_StoreClose(t *testing.T) {
	s := New()
	ctx := context.Background()

	done := make(chan error, 1)
	go func() {
		_, err := s.WaitFor(ctx, "never")
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}

	select {
	case err := <-done:
		if !errors.Is(err, ErrStoreClosed) {
			t.Errorf("expected ErrStoreClosed, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitFor did not return after Close")
	}
}

func TestStore_WaitFor_AfterClose(t *testing.T) {
	s := New()
	_ = s.Close()
	_, err := s.WaitFor(context.Background(), "k")
	if !errors.Is(err, ErrStoreClosed) {
		t.Errorf("expected ErrStoreClosed, got %v", err)
	}
}

func TestStore_CloseIsIdempotent(t *testing.T) {
	s := New()
	if err := s.Close(); err != nil {
		t.Errorf("Close 1: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close 2: %v", err)
	}
}

func TestStore_AppendEvent(t *testing.T) {
	s := New()

	if got := s.Events("none"); got != nil {
		t.Errorf("absent alias should return nil, got %v", got)
	}

	s.AppendEvent("stream", wire.PluginEvent{Plugin: "grpc", Kind: "Msg", Data: map[string]any{"i": 1}})
	s.AppendEvent("stream", wire.PluginEvent{Plugin: "grpc", Kind: "Msg", Data: map[string]any{"i": 2}})

	got := s.Events("stream")
	if len(got) != 2 {
		t.Fatalf("got %d events; want 2", len(got))
	}
	if got[0].Data["i"] != 1 || got[1].Data["i"] != 2 {
		t.Errorf("events out of order: %v", got)
	}

	// Returned slice is a copy — mutating it must not affect the store.
	got[0].Kind = "Mutated"
	again := s.Events("stream")
	if again[0].Kind == "Mutated" {
		t.Error("Events() must return a defensive copy")
	}
}

func TestStore_ConcurrentSetGet(t *testing.T) {
	s := New()
	const writers = 50
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Set("k", i)
			_, _ = s.Get("k")
		}(i)
	}
	wg.Wait()

	// After all writes, key must be set with one of the written values.
	v, ok := s.Get("k")
	if !ok {
		t.Fatal("key not set after concurrent writes")
	}
	if iv, ok := v.(int); !ok || iv < 0 || iv >= writers {
		t.Errorf("unexpected value %v", v)
	}
}

func TestStore_OverwriteSet(t *testing.T) {
	s := New()
	s.Set("k", 1)
	s.Set("k", 2)
	v, _ := s.Get("k")
	if v != 2 {
		t.Errorf("overwrite failed: got %v", v)
	}
}
