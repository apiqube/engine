package engine

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	e := New()
	defer e.Close()
	if !e.parallel {
		t.Error("default parallel should be true")
	}
	if e.maxConcurrent <= 0 {
		t.Error("default maxConcurrent should be > 0")
	}
}

func TestRun_NilInput(t *testing.T) {
	e := New()
	defer e.Close()
	res, err := e.Run(context.Background(), nil)
	if err == nil {
		t.Error("nil input should error")
	}
	if res == nil {
		t.Error("Results should always be non-nil")
	}
}

func TestCheck_NilInput(t *testing.T) {
	e := New()
	defer e.Close()
	if _, err := e.Check(context.Background(), nil); err == nil {
		t.Error("nil input should error")
	}
}

func TestRun_ParseError(t *testing.T) {
	e := New()
	defer e.Close()
	res, err := e.Run(context.Background(), FromBytes([]byte("not valid yaml :{")))
	if err == nil {
		t.Error("parse error should propagate")
	}
	if res == nil {
		t.Fatal("Results should be non-nil on error")
	}
}

func TestRun_NoTests(t *testing.T) {
	e := New()
	defer e.Close()

	yaml := `# Empty file with no tests`
	res, err := e.Run(context.Background(), FromBytes([]byte(yaml)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 0 {
		t.Errorf("empty input should yield 0 tests, got %d", res.Total)
	}
}

func TestRun_NoPluginForProtocol(t *testing.T) {
	e := New()
	defer e.Close()

	yaml := `tests:
  - name: x
    method: GET
    target: http://example.com
    expect:
      status: 200
`
	res, err := e.Run(context.Background(), FromBytes([]byte(yaml)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 1 {
		t.Fatalf("got %d tests", res.Total)
	}
	if res.Errored != 1 {
		t.Errorf("test should be errored (no plugin), got: %+v", res.Tests[0])
	}
}

func TestCheck_ValidatesGraph(t *testing.T) {
	e := New()
	defer e.Close()

	// Cyclic depends should produce a validation error.
	yaml := `tests:
  - name: a
    alias: a
    depends: [b]
    method: GET
    target: x://t
  - name: b
    alias: b
    depends: [a]
    method: GET
    target: x://t
`
	errs, err := e.Check(context.Background(), FromBytes([]byte(yaml)))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Error("cycle should produce validation error")
	}
	if len(errs) > 0 && !strings.Contains(errs[0].Message, "cycle") {
		t.Errorf("error message should mention cycle: %q", errs[0].Message)
	}
}

func TestCheck_Valid(t *testing.T) {
	e := New()
	defer e.Close()

	yaml := `tests:
  - "GET / -> 200"
`
	errs, err := e.Check(context.Background(), FromBytes([]byte(yaml)))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Errorf("valid manifest should produce no errors, got %v", errs)
	}
}

func TestRun_EventsEmitted(t *testing.T) {
	e := New()
	defer e.Close()

	var gotRunStarted, gotRunCompleted bool
	handler := handlerFunc(func(ev Event) {
		switch ev.(type) {
		case RunStarted:
			gotRunStarted = true
		case RunCompleted:
			gotRunCompleted = true
		}
	})

	yaml := `tests: []`
	_, _ = e.Run(context.Background(), FromBytes([]byte(yaml)), WithHandler(handler))
	if !gotRunStarted {
		t.Error("RunStarted should be emitted")
	}
	if !gotRunCompleted {
		t.Error("RunCompleted should be emitted")
	}
}

func TestRun_FromReader(t *testing.T) {
	e := New()
	defer e.Close()

	r := strings.NewReader(`tests: []`)
	if _, err := e.Run(context.Background(), FromReader(r)); err != nil {
		t.Fatal(err)
	}
}

func TestRun_ConcurrentSafe(t *testing.T) {
	e := New()
	defer e.Close()

	yaml := `tests:
  - "GET / -> 200"
`
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			_, _ = e.Run(context.Background(), FromBytes([]byte(yaml)))
		})
	}
	wg.Wait()
}

func TestPlugins_BeforeRun(t *testing.T) {
	e := New()
	defer e.Close()
	if got := e.Plugins(); got != nil {
		t.Errorf("Plugins() should be nil before any Run, got %v", got)
	}
	if got := e.EventSchema("x.y"); got != nil {
		t.Errorf("EventSchema before Run should return nil, got %v", got)
	}
}

func TestEngine_Close_IsIdempotent(t *testing.T) {
	e := New()
	if err := e.Close(); err != nil {
		t.Errorf("first Close: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestRun_GraphError(t *testing.T) {
	e := New()
	defer e.Close()
	yaml := `tests:
  - name: a
    alias: a
    depends: [a]
    method: GET
    target: x://t
`
	_, err := e.Run(context.Background(), FromBytes([]byte(yaml)))
	// Self-cycle is ignored by toposort but we can produce a real cycle below.
	_ = err

	yamlCycle := `tests:
  - name: a
    alias: a
    depends: [b]
    method: GET
    target: x://t
  - name: b
    alias: b
    depends: [a]
    method: GET
    target: x://t
`
	_, err = e.Run(context.Background(), FromBytes([]byte(yamlCycle)))
	if err == nil {
		t.Error("cycle should propagate as run-level error")
	}
}

func TestRun_PluginEventEmitted(t *testing.T) {
	// Tests that the OnPluginEvent hook wires through; without a real plugin
	// we can only assert that the handler dispatcher works for PluginEvent
	// values.
	got := false
	handler := handlerFunc(func(ev Event) {
		if _, ok := ev.(PluginEvent); ok {
			got = true
		}
	})
	pe := PluginEvent{Plugin: "p", Kind: "K"}
	handler.Handle(pe)
	if !got {
		t.Error("handler should receive PluginEvent")
	}
	if pe.FullName() != "p.K" {
		t.Errorf("FullName wrong: %q", pe.FullName())
	}
}

func TestRun_FailFast(t *testing.T) {
	e := New(WithFailFast(true))
	defer e.Close()

	// Without a plugin, all tests will error; failFast just skips remaining.
	yaml := `tests:
  - name: a
    method: GET
    target: x://t
  - name: b
    method: GET
    target: x://t
`
	res, err := e.Run(context.Background(), FromBytes([]byte(yaml)))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tests) != 2 {
		t.Errorf("got %d tests; want 2", len(res.Tests))
	}
}

func TestErrNoInput(t *testing.T) {
	if ErrNoInput == nil {
		t.Fatal("ErrNoInput must be defined")
	}
	if ErrNoInput.Error() == "" {
		t.Error("ErrNoInput should have a message")
	}
}

// handlerFunc adapts a function into the EventHandler interface for tests.
type handlerFunc func(Event)

func (f handlerFunc) Handle(e Event) { f(e) }
