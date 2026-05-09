package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/apiqube/engine/internal/wire"
)

const httpFixtureName = "plugin-http.wasm"

// loadHTTPFixture loads the plugin-http fixture if its `.wasm` is present.
// Skips with an informative message otherwise — the fixture is produced by
// plugin-http's CI (TinyGo build) and committed manually into _testdata/.
//
// See plugin-http/README.md for build instructions.
func loadHTTPFixture(t *testing.T) (*WASMRuntime, *WASMPlugin) {
	t.Helper()
	wasmPath := filepath.Join("_testdata", httpFixtureName)
	bytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Skipf("plugin-http fixture not present at %s: %v", wasmPath, err)
	}

	ctx := context.Background()
	rt, err := NewWASMRuntime(ctx)
	if err != nil {
		t.Fatalf("NewWASMRuntime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(ctx) })

	p, err := rt.LoadPlugin(ctx, "http", bytes)
	if err != nil {
		t.Fatalf("LoadPlugin: %v", err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return rt, p
}

func TestWASM_LoadAndInfo(t *testing.T) {
	_, p := loadHTTPFixture(t)

	info := p.Info()
	if info.Name != "http" {
		t.Errorf("info.Name = %q, want http", info.Name)
	}
	if info.Version == "" {
		t.Error("version should be populated")
	}
	if len(info.Capabilities) == 0 || info.Capabilities[0] != "http" {
		t.Errorf("capabilities should declare 'http', got %v", info.Capabilities)
	}
}

func TestWASM_Init(t *testing.T) {
	_, p := loadHTTPFixture(t)
	if err := p.Init(context.Background(), map[string]any{}); err != nil {
		t.Errorf("Init: %v", err)
	}
}

func TestWASM_Validate(t *testing.T) {
	_, p := loadHTTPFixture(t)
	errs := p.Validate(TestInput{Method: "GET", Resource: "/x", Target: "http://example.test"})
	if len(errs) != 0 {
		t.Errorf("plugin returned validation errors: %v", errs)
	}
}

func TestWASM_Validate_MissingTarget(t *testing.T) {
	_, p := loadHTTPFixture(t)
	errs := p.Validate(TestInput{Method: "GET"})
	if len(errs) == 0 {
		t.Error("missing target should produce a FieldError")
	}
}

// TestWASM_Execute asserts a full round-trip: plugin builds an HTTP request,
// host_http_request runs it via httptest.Server, plugin assembles a TestOutput.
//
// Implemented as a separate function in wasm_e2e_test.go so the lighter unit
// tests above don't depend on httptest.

func TestWASM_LoadInvalidWASM(t *testing.T) {
	ctx := context.Background()
	rt, err := NewWASMRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	if _, err := rt.LoadPlugin(ctx, "broken", []byte("not a wasm module")); err == nil {
		t.Error("invalid wasm bytes should fail to load")
	}
}

func TestRegistry_WithRealWASM(t *testing.T) {
	_, p := loadHTTPFixture(t)

	reg := NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if reg.Lookup("http") != p {
		t.Error("Lookup did not find http plugin")
	}
	snap := reg.Snapshot()
	if len(snap) != 1 || snap[0].Name != "http" {
		t.Errorf("Snapshot wrong: %+v", snap)
	}
}

// silence unused warning for wire when only loadHTTPFixture references it
// indirectly through Plugin types.
var _ = wire.Protocol("http")
