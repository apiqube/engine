package capabilities

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero"

	"github.com/apiqube/engine/internal/wire"
)

func TestSupportedCapabilities(t *testing.T) {
	caps := SupportedCapabilities()
	if len(caps) == 0 {
		t.Fatal("expected at least one supported capability")
	}
	if !IsSupported("http") {
		t.Error("http must be supported")
	}
	if IsSupported("grpc") {
		t.Error("grpc must NOT be supported in v1.0")
	}
}

func TestInstantiateHostModule(t *testing.T) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	if err := InstantiateHostModule(ctx, rt); err != nil {
		t.Fatalf("InstantiateHostModule: %v", err)
	}
	// Calling again would fail (same name); verifying single-instantiation.
	if err := InstantiateHostModule(ctx, rt); err == nil {
		t.Error("second InstantiateHostModule call should fail (duplicate module)")
	}
}

func TestSinks_NilSafe(t *testing.T) {
	if got := resolveLogger(nil); got == nil {
		t.Error("resolveLogger(nil) should return a non-nil sink")
	}
	if got := resolveNow(nil); got == nil {
		t.Error("resolveNow(nil) should return a non-nil func")
	}
	if got := resolveEventSink(nil); got == nil {
		t.Error("resolveEventSink(nil) should return a non-nil func")
	}
	if got := resolveHTTPClient(nil); got == nil {
		t.Error("resolveHTTPClient(nil) should return a non-nil client")
	}
}

func TestWithSinks_RoundTrip(t *testing.T) {
	want := &Sinks{Logger: nopLogSink{}}
	ctx := WithSinks(context.Background(), want)
	got := sinksFrom(ctx)
	if got != want {
		t.Errorf("sinks not preserved through context")
	}
}

func TestSinksFrom_NoValue(t *testing.T) {
	got := sinksFrom(context.Background())
	if got == nil {
		t.Fatal("sinksFrom should never return nil")
	}
}

func TestEmitEventLogic_Valid(t *testing.T) {
	var got wire.PluginEvent
	err := EmitEventLogic([]byte(`{"plugin":"grpc","kind":"Msg","data":{"x":1}}`),
		func(ev wire.PluginEvent) { got = ev })
	if err != nil {
		t.Fatal(err)
	}
	if got.Plugin != "grpc" || got.Kind != "Msg" {
		t.Errorf("got %+v", got)
	}
}

func TestEmitEventLogic_Empty(t *testing.T) {
	called := false
	err := EmitEventLogic(nil, func(wire.PluginEvent) { called = true })
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("empty bytes should not invoke sink")
	}
}

func TestEmitEventLogic_BadJSON(t *testing.T) {
	if err := EmitEventLogic([]byte(`{not json}`), nil); err == nil {
		t.Error("malformed JSON should produce error")
	}
}

func TestLogLevel_String(t *testing.T) {
	cases := map[LogLevel]string{
		LogDebug: "debug",
		LogInfo:  "info",
		LogWarn:  "warn",
		LogError: "error",
		99:       "unknown",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("%d.String() = %q; want %q", level, got, want)
		}
	}
}

func TestLogSinkFunc(t *testing.T) {
	var gotLevel LogLevel
	var gotMsg string
	sink := LogSinkFunc(func(l LogLevel, m string) {
		gotLevel = l
		gotMsg = m
	})
	sink.Log(LogWarn, "hello")
	if gotLevel != LogWarn || gotMsg != "hello" {
		t.Errorf("got (%v, %q)", gotLevel, gotMsg)
	}
}
