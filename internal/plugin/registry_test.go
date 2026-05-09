package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/apiqube/engine/internal/wire"
)

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	p := newFakePlugin("http", wire.ProtocolHTTP, wire.ProtocolHTTPS)

	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if got := reg.Lookup(wire.ProtocolHTTP); got != p {
		t.Errorf("lookup http: got %v, want %v", got, p)
	}
	if got := reg.Lookup(wire.ProtocolHTTPS); got != p {
		t.Errorf("lookup https: got %v, want %v", got, p)
	}
	if got := reg.Lookup(wire.ProtocolGRPC); got != nil {
		t.Errorf("lookup grpc: got %v, want nil", got)
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	if err := NewRegistry().Register(nil); err == nil {
		t.Error("Register(nil) should fail")
	}
}

func TestRegistry_DuplicateName(t *testing.T) {
	reg := NewRegistry()
	a := newFakePlugin("dup", wire.ProtocolHTTP)
	b := newFakePlugin("dup", wire.ProtocolGRPC)
	if err := reg.Register(a); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(b); err == nil {
		t.Error("duplicate name should fail")
	}
}

func TestRegistry_DuplicateProtocol(t *testing.T) {
	reg := NewRegistry()
	a := newFakePlugin("a", wire.ProtocolHTTP)
	b := newFakePlugin("b", wire.ProtocolHTTP)
	if err := reg.Register(a); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(b); err == nil {
		t.Error("duplicate protocol should fail")
	}
}

func TestRegistry_Snapshot(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(newFakePlugin("zeta", wire.ProtocolGRPC)); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(newFakePlugin("alpha", wire.ProtocolHTTP)); err != nil {
		t.Fatal(err)
	}

	snap := reg.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("got %d entries, want 2", len(snap))
	}
	if snap[0].Name != "alpha" || snap[1].Name != "zeta" {
		t.Errorf("not sorted by name: %+v", snap)
	}
}

func TestRegistry_EventSchema(t *testing.T) {
	reg := NewRegistry()
	p := newFakePlugin("grpc", wire.ProtocolGRPC)
	p.info.Events = map[string]wire.EventSpec{
		"StreamMessage": {Description: "one stream message"},
	}
	if err := reg.Register(p); err != nil {
		t.Fatal(err)
	}

	if got := reg.EventSchema("grpc.StreamMessage"); got == nil {
		t.Error("EventSchema should find existing event")
	}
	if got := reg.EventSchema("grpc.None"); got != nil {
		t.Error("non-existent kind should return nil")
	}
	if got := reg.EventSchema("nope.Kind"); got != nil {
		t.Error("non-existent plugin should return nil")
	}
	if got := reg.EventSchema("noPeriod"); got != nil {
		t.Error("malformed name should return nil")
	}
}

func TestRegistry_Close(t *testing.T) {
	closed := false
	p := newFakePlugin("p", wire.ProtocolHTTP)
	p.closeFn = func() error { closed = true; return nil }

	reg := NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatal(err)
	}
	if err := reg.Close(); err != nil {
		t.Fatal(err)
	}
	if !closed {
		t.Error("Close did not close plugin")
	}

	// Double close is no-op.
	if err := reg.Close(); err != nil {
		t.Errorf("second Close should not error: %v", err)
	}

	// Register after Close fails.
	if err := reg.Register(newFakePlugin("after", wire.ProtocolHTTP)); err == nil {
		t.Error("Register after Close should fail")
	}
}

func TestRegistry_CloseAggregatesErrors(t *testing.T) {
	a := newFakePlugin("a", wire.ProtocolHTTP)
	b := newFakePlugin("b", wire.ProtocolGRPC)
	a.closeFn = func() error { return errors.New("a fail") }
	bClosed := false
	b.closeFn = func() error { bClosed = true; return nil }

	reg := NewRegistry()
	_ = reg.Register(a)
	_ = reg.Register(b)
	err := reg.Close()
	if err == nil {
		t.Fatal("expected error from a")
	}
	if !bClosed {
		t.Error("b should still close even if a errored")
	}
}

func TestRegistry_InitAll(t *testing.T) {
	gotConfig := map[string]map[string]any{}
	a := newFakePlugin("a", wire.ProtocolHTTP)
	a.initFn = func(_ context.Context, cfg map[string]any) error {
		gotConfig["a"] = cfg
		return nil
	}
	b := newFakePlugin("b", wire.ProtocolGRPC)
	b.initFn = func(_ context.Context, cfg map[string]any) error {
		gotConfig["b"] = cfg
		return nil
	}

	reg := NewRegistry()
	_ = reg.Register(a)
	_ = reg.Register(b)

	configs := map[string]map[string]any{
		"a": {"k": "v"},
		"b": nil,
	}
	if err := reg.InitAll(context.Background(), configs); err != nil {
		t.Fatal(err)
	}

	if gotConfig["a"]["k"] != "v" {
		t.Errorf("a got wrong config: %v", gotConfig["a"])
	}
	if _, called := gotConfig["b"]; !called {
		t.Errorf("b should have been initialized")
	}
}

func TestRegistry_InitAll_Error(t *testing.T) {
	a := newFakePlugin("a", wire.ProtocolHTTP)
	a.initFn = func(context.Context, map[string]any) error { return errors.New("boom") }
	reg := NewRegistry()
	_ = reg.Register(a)
	if err := reg.InitAll(context.Background(), nil); err == nil {
		t.Error("expected error from failing Init")
	}
}
