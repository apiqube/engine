package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/apiqube/engine/internal/wire"
	"github.com/apiqube/engine/internal/graph"
	"github.com/apiqube/engine/internal/manifest"
	"github.com/apiqube/engine/internal/plugin"
)

// fakePlugin implements plugin.Plugin for runner tests.
type fakePlugin struct {
	info       plugin.PluginInfo
	executeFn  func(context.Context, plugin.TestInput, plugin.EventSink) (*plugin.TestOutput, error)
	validateFn func(plugin.TestInput) []plugin.FieldError
}

func (p *fakePlugin) Info() plugin.PluginInfo { return p.info }
func (p *fakePlugin) Init(context.Context, map[string]any) error { return nil }
func (p *fakePlugin) Validate(in plugin.TestInput) []plugin.FieldError {
	if p.validateFn != nil {
		return p.validateFn(in)
	}
	return nil
}
func (p *fakePlugin) Execute(ctx context.Context, in plugin.TestInput, emit plugin.EventSink) (*plugin.TestOutput, error) {
	if p.executeFn != nil {
		return p.executeFn(ctx, in, emit)
	}
	return &plugin.TestOutput{Status: 200}, nil
}
func (p *fakePlugin) Close() error { return nil }

func newEchoPlugin() *fakePlugin {
	return &fakePlugin{
		info: plugin.PluginInfo{Name: "echo", Protocols: []wire.Protocol{"echo"}},
		executeFn: func(_ context.Context, in plugin.TestInput, emit plugin.EventSink) (*plugin.TestOutput, error) {
			if emit != nil {
				emit(wire.PluginEvent{Plugin: "echo", Kind: "Replied", Data: map[string]any{"target": in.Target}})
			}
			body := map[string]any{
				"method":   in.Method,
				"resource": in.Resource,
				"target":   in.Target,
				"fields":   in.Fields,
			}
			return &plugin.TestOutput{Status: 200, Body: body, DurationMs: 5}, nil
		},
	}
}

func mustRegistry(t *testing.T, plugins ...plugin.Plugin) *plugin.Registry {
	t.Helper()
	reg := plugin.NewRegistry()
	for _, p := range plugins {
		if err := reg.Register(p); err != nil {
			t.Fatalf("Register: %v", err)
		}
	}
	return reg
}

func makeFile(path string, mode manifest.TestMode, tests ...manifest.TestCase) *manifest.TestFile {
	return &manifest.TestFile{Path: path, Mode: mode, Tests: tests}
}

func buildPlan(t *testing.T, files ...*manifest.TestFile) *graph.Plan {
	t.Helper()
	plan, err := graph.New().Build(files)
	if err != nil {
		t.Fatalf("graph.Build: %v", err)
	}
	return plan
}

func TestExecute_SingleTest(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "ping", Method: "GET", Resource: "/x", Target: "echo://t",
			Expect: manifest.Expect{Status: 200},
		},
	)
	plan := buildPlan(t, file)
	r := New(mustRegistry(t, newEchoPlugin()), Options{Parallel: true})

	res, err := r.Execute(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tests) != 1 {
		t.Fatalf("got %d tests", len(res.Tests))
	}
	if res.Tests[0].Status != wire.StatusPassed {
		t.Errorf("got %v: %+v", res.Tests[0].Status, res.Tests[0])
	}
}

func TestExecute_AssertionFailure(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "x", Method: "GET", Resource: "/", Target: "echo://t",
			Expect: manifest.Expect{Status: 999},
		},
	)
	plan := buildPlan(t, file)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})

	res, _ := r.Execute(context.Background(), plan)
	if res.Tests[0].Status != wire.StatusFailed {
		t.Errorf("expected failed, got %v", res.Tests[0].Status)
	}
}

func TestExecute_PluginError(t *testing.T) {
	bad := &fakePlugin{
		info: plugin.PluginInfo{Name: "bad", Protocols: []wire.Protocol{"echo"}},
		executeFn: func(context.Context, plugin.TestInput, plugin.EventSink) (*plugin.TestOutput, error) {
			return nil, errors.New("boom")
		},
	}
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Method: "GET", Target: "echo://t"},
	)
	r := New(mustRegistry(t, bad), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if res.Tests[0].Status != wire.StatusErrored {
		t.Errorf("expected errored, got %v", res.Tests[0].Status)
	}
	if res.Tests[0].Error != "boom" {
		t.Errorf("error wrong: %q", res.Tests[0].Error)
	}
}

func TestExecute_Skip(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "skipped", Skip: true, Target: "echo://t"},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if !res.Tests[0].Skipped {
		t.Errorf("expected skipped")
	}
}

func TestExecute_NoPlugin(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Target: "grpc://nowhere"},
	)
	r := New(mustRegistry(t /* no plugins */), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if res.Tests[0].Status != wire.StatusErrored {
		t.Errorf("expected errored, got %v", res.Tests[0].Status)
	}
}

func TestExecute_NoProtocol(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Target: "no-scheme"},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if res.Tests[0].Status != wire.StatusErrored {
		t.Errorf("expected errored, got %v", res.Tests[0].Status)
	}
}

func TestExecute_DataFlow(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "produce", Alias: "user", Method: "GET",
			Target: "echo://t",
			Save:   map[string]string{"id": "response.body.fields.id"},
			Extra:  map[string]any{"id": 42},
		},
		manifest.TestCase{
			Name: "consume", Method: "POST", Target: "echo://t",
			Headers: map[string]string{"X-User-ID": "{{ id }}"},
			Expect:  manifest.Expect{Status: 200},
		},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})
	res, err := r.Execute(context.Background(), buildPlan(t, file))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tests) != 2 {
		t.Fatalf("got %d tests", len(res.Tests))
	}
	consumer := res.Tests[1]
	if got := consumer.Request.Headers["X-User-ID"]; got != "42" {
		t.Errorf("save+template chain failed: header = %q", got)
	}
}

func TestExecute_AliasResponseReference(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "first", Alias: "first", Method: "GET", Target: "echo://t",
		},
		manifest.TestCase{
			Name: "second", Method: "POST", Target: "echo://t",
			Extra: map[string]any{"echo": "{{ first.response.body.target }}"},
		},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})
	res, err := r.Execute(context.Background(), buildPlan(t, file))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tests) != 2 {
		t.Fatalf("got %d tests", len(res.Tests))
	}
	consumer := res.Tests[1]
	echo, _ := consumer.Request.Fields["echo"].(string)
	if echo != "echo://t" {
		t.Errorf("alias.response.body.target should resolve to producer's target; got %q", echo)
	}
}

func TestExecute_ScenarioPrev(t *testing.T) {
	file := makeFile("s.yaml", manifest.ModeScenario,
		manifest.TestCase{Name: "first", Method: "GET", Target: "echo://t"},
		manifest.TestCase{
			Name:  "second",
			Extra: map[string]any{"prev_target": "{{ prev.response.body.target }}"},
			Method: "POST", Target: "echo://t",
		},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if len(res.Tests) != 2 {
		t.Fatalf("got %d tests", len(res.Tests))
	}
	prev, _ := res.Tests[1].Request.Fields["prev_target"].(string)
	if prev != "echo://t" {
		t.Errorf("prev should give first response; got %q", prev)
	}
}

func TestExecute_Hooks(t *testing.T) {
	var startedTests, completedTests int
	var startedWaves, completedWaves int

	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Method: "GET", Target: "echo://t"},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{
		Hooks: Hooks{
			OnTestStart:    func(TestStart) { startedTests++ },
			OnTestComplete: func(TestOutcome) { completedTests++ },
			OnWaveStart:    func(WaveStart) { startedWaves++ },
			OnWaveComplete: func(WaveOutcome) { completedWaves++ },
		},
	})
	_, err := r.Execute(context.Background(), buildPlan(t, file))
	if err != nil {
		t.Fatal(err)
	}
	if startedTests != 1 || completedTests != 1 {
		t.Errorf("test hooks called wrong: %d started, %d completed", startedTests, completedTests)
	}
	if startedWaves != 1 || completedWaves != 1 {
		t.Errorf("wave hooks called wrong: %d started, %d completed", startedWaves, completedWaves)
	}
}

func TestExecute_FailFastSequential(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeScenario,
		manifest.TestCase{Name: "fails", Method: "GET", Target: "echo://t",
			Expect: manifest.Expect{Status: 999}},
		manifest.TestCase{Name: "skipped", Method: "GET", Target: "echo://t"},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{FailFast: true})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if len(res.Tests) != 2 {
		t.Fatalf("expected 2 tests in result, got %d", len(res.Tests))
	}
	if res.Tests[1].Status != wire.StatusSkipped {
		t.Errorf("second test should be skipped under failFast: %v", res.Tests[1].Status)
	}
}

func TestExecute_ParallelWave(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "a", Method: "GET", Target: "echo://t"},
		manifest.TestCase{Name: "b", Method: "GET", Target: "echo://t"},
		manifest.TestCase{Name: "c", Method: "GET", Target: "echo://t"},
	)
	r := New(mustRegistry(t, newEchoPlugin()), Options{Parallel: true, MaxConcurrent: 4})
	res, err := r.Execute(context.Background(), buildPlan(t, file))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Waves) != 1 || !res.Waves[0].Parallel {
		t.Errorf("expected single parallel wave, got %+v", res.Waves)
	}
	for _, tc := range res.Tests {
		if tc.Status != wire.StatusPassed {
			t.Errorf("test %s failed: %+v", tc.Ref.Test.Name, tc)
		}
	}
}

func TestNormalizeDurationExpr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"< 500ms", "< 500"},
		{"<= 1s", "<= 1000"},
		{"> 100ms", "> 100"},
		{"500ms", "500"},
		{"plain", "plain"},
	}
	for _, c := range cases {
		got := normalizeDurationExpr(c.in)
		if got != c.want {
			t.Errorf("normalizeDurationExpr(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestDetectProtocol(t *testing.T) {
	cases := []struct {
		in   string
		want wire.Protocol
		ok   bool
	}{
		{"http://x", "http", true},
		{"echo://t", "echo", true},
		{"grpc://service", "grpc", true},
		{"", "", false},
		{"no-scheme", "", false},
	}
	for _, c := range cases {
		got, err := detectProtocol(c.in)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("detectProtocol(%q) = (%v, err=%v); want (%v, ok=%v)",
				c.in, got, err, c.want, c.ok)
		}
	}
}

func TestExecute_RetryUntilPasses(t *testing.T) {
	attempts := 0
	flaky := &fakePlugin{
		info: plugin.PluginInfo{Name: "echo", Protocols: []wire.Protocol{"echo"}},
		executeFn: func(_ context.Context, in plugin.TestInput, _ plugin.EventSink) (*plugin.TestOutput, error) {
			attempts++
			status := 500
			if attempts >= 3 {
				status = 200
			}
			return &plugin.TestOutput{Status: status}, nil
		},
	}
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "poll", Method: "GET", Target: "echo://t",
			Retry: &manifest.RetryConfig{
				MaxAttempts: 5,
				Interval:    "1ms",
				Until:       &manifest.Expect{Status: 200},
			},
		},
	)
	r := New(mustRegistry(t, flaky), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if res.Tests[0].Status != wire.StatusPassed {
		t.Errorf("retry should eventually succeed, got %v", res.Tests[0].Status)
	}
}

func TestExecute_StringMapToAny(t *testing.T) {
	got := stringMapToAny(map[string]string{"a": "1", "b": "2"})
	if got["a"] != "1" || got["b"] != "2" {
		t.Errorf("conversion wrong: %v", got)
	}
	if stringMapToAny(nil) != nil {
		t.Error("nil input should yield nil output")
	}
}

func TestExecute_EventsToAny(t *testing.T) {
	in := []wire.PluginEvent{
		{Plugin: "p", Kind: "K", Data: map[string]any{"x": 1}},
	}
	got := eventsToAny(in)
	if len(got) != 1 {
		t.Fatalf("got %d events", len(got))
	}
	m := got[0].(map[string]any)
	if m["plugin"] != "p" || m["kind"] != "K" {
		t.Errorf("conversion wrong: %v", m)
	}
	if eventsToAny(nil) != nil {
		t.Error("nil input should yield nil")
	}
}

func TestExecute_DurationAssertion(t *testing.T) {
	slow := &fakePlugin{
		info: plugin.PluginInfo{Name: "echo", Protocols: []wire.Protocol{"echo"}},
		executeFn: func(context.Context, plugin.TestInput, plugin.EventSink) (*plugin.TestOutput, error) {
			return &plugin.TestOutput{Status: 200, DurationMs: 250}, nil
		},
	}
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{
			Name: "x", Method: "GET", Target: "echo://t",
			Expect: manifest.Expect{Status: 200, Duration: "< 500ms"},
		},
	)
	r := New(mustRegistry(t, slow), Options{})
	res, _ := r.Execute(context.Background(), buildPlan(t, file))
	if res.Tests[0].Status != wire.StatusPassed {
		t.Errorf("250ms < 500ms should pass; got %v: %+v", res.Tests[0].Status, res.Tests[0].Assertions)
	}
}

func TestExecute_StreamingEventCapture(t *testing.T) {
	streamer := &fakePlugin{
		info: plugin.PluginInfo{Name: "stream", Protocols: []wire.Protocol{"echo"}},
		executeFn: func(_ context.Context, in plugin.TestInput, emit plugin.EventSink) (*plugin.TestOutput, error) {
			emit(wire.PluginEvent{Plugin: "stream", Kind: "Frame", Data: map[string]any{"i": 1}})
			emit(wire.PluginEvent{Plugin: "stream", Kind: "Frame", Data: map[string]any{"i": 2}})
			return &plugin.TestOutput{Status: 200}, nil
		},
	}

	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "src", Alias: "src", Method: "GET", Target: "echo://t"},
	)
	r := New(mustRegistry(t, streamer), Options{})

	store, _ := r.Execute(context.Background(), buildPlan(t, file))
	if len(store.Tests) != 1 {
		t.Fatalf("got %d tests", len(store.Tests))
	}
}

func TestResolveTarget(t *testing.T) {
	file := &manifest.TestFile{
		Target:  "default",
		Targets: map[string]string{"default": "http://default", "named": "http://named"},
	}
	if got := resolveTarget(file, &manifest.TestCase{}); got != "http://default" {
		t.Errorf("resolved default wrong: %q", got)
	}
	if got := resolveTarget(file, &manifest.TestCase{Target: "named"}); got != "http://named" {
		t.Errorf("resolved named wrong: %q", got)
	}
	if got := resolveTarget(file, &manifest.TestCase{Target: "http://override"}); got != "http://override" {
		t.Errorf("absolute override wrong: %q", got)
	}
}
