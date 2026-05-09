package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"time"

	"github.com/apiqube/engine/internal/graph"
	"github.com/apiqube/engine/internal/manifest"
	"github.com/apiqube/engine/internal/parser"
	"github.com/apiqube/engine/internal/plugin"
	"github.com/apiqube/engine/internal/runner"
	"github.com/apiqube/engine/internal/wire"
)

// Engine is the core of the ApiQube testing platform.
//
// Engine is stateless after construction — safe for concurrent Run() calls.
// Configuration set via Options applies to all runs. Per-run configuration
// (event handler, signals, env overrides) is passed via RunOptions.
type Engine struct {
	pluginDir     string
	parallel      bool
	maxConcurrent int
	failFast      bool

	loader *plugin.Loader
}

// New creates a new Engine with the given Options.
// Defaults: parallel=true, maxConcurrent=GOMAXPROCS, failFast=false.
func New(opts ...Option) *Engine {
	e := &Engine{
		parallel:      true,
		maxConcurrent: runtime.GOMAXPROCS(0),
		loader:        plugin.NewLoader(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Close releases plugin resources held by the engine. Safe to call multiple
// times. The engine cannot be used after Close.
func (e *Engine) Close() error {
	if e.loader == nil {
		return nil
	}
	return e.loader.Close(context.Background())
}

// ErrNoInput is returned when Run or Check is called without a valid Input source.
var ErrNoInput = errors.New("engine: no input source provided")

// Run executes tests from the given Input source.
//
// Input is mandatory. Additional RunOptions configure the event handler,
// signals, environment, etc. Run returns a non-nil *Results in all paths;
// the error is non-nil only on engine-level failures (parse error, dependency
// cycle, plugin load failure, ctx cancelled). Individual test failures are
// reported within Results.
func (e *Engine) Run(ctx context.Context, input Input, opts ...RunOption) (*Results, error) {
	if input == nil {
		return &Results{}, ErrNoInput
	}
	rc := e.buildRunConfig(input, opts)

	registry, err := e.loader.LoadDir(ctx, rc.pluginDir)
	if err != nil {
		return &Results{}, fmt.Errorf("engine: load plugins: %w", err)
	}

	files, err := loadInput(rc.input)
	if err != nil {
		return &Results{}, fmt.Errorf("engine: parse input: %w", err)
	}

	plan, err := graph.New().Build(files)
	if err != nil {
		return &Results{}, fmt.Errorf("engine: build graph: %w", err)
	}

	rc.handler.Handle(RunStarted{
		Files:      filePaths(files),
		TotalTests: countTests(plan),
		TotalWaves: len(plan.Waves),
	})
	rc.handler.Handle(GraphResolved{
		TotalWaves:       len(plan.Waves),
		Dependencies:     len(plan.Dependencies),
		SaveRequirements: len(plan.SaveRequirements),
	})

	runStart := time.Now()
	r := runner.New(registry, runner.Options{
		Parallel:      e.parallel,
		MaxConcurrent: e.maxConcurrent,
		FailFast:      e.failFast,
		Env:           rc.env,
		Hooks:         buildHooks(rc.handler),
	})

	runResult, err := r.Execute(ctx, plan)
	if err != nil {
		return convertResults(runResult, time.Since(runStart)), err
	}

	results := convertResults(runResult, time.Since(runStart))
	rc.handler.Handle(buildRunCompleted(results))
	return results, nil
}

// Check validates manifests from the given Input source without executing them.
//
// Returns a list of validation errors (parse errors, unresolved alias references,
// dependency cycles). An empty slice means everything validated.
func (e *Engine) Check(ctx context.Context, input Input, opts ...CheckOption) ([]ValidationError, error) {
	if input == nil {
		return nil, ErrNoInput
	}
	cc := e.buildCheckConfig(input, opts)
	_ = ctx

	files, err := loadInput(cc.input)
	if err != nil {
		return []ValidationError{{Message: err.Error(), Severity: SeverityError}}, nil
	}
	if _, err := graph.New().Build(files); err != nil {
		return []ValidationError{{Message: err.Error(), Severity: SeverityError}}, nil
	}
	return nil, nil
}

// buildRunConfig assembles a runConfig from engine defaults and user options.
func (e *Engine) buildRunConfig(input Input, opts []RunOption) *runConfig {
	rc := &runConfig{
		input:     input,
		handler:   NopHandler{},
		pluginDir: e.pluginDir,
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

// buildCheckConfig assembles a checkConfig from engine defaults and user options.
func (e *Engine) buildCheckConfig(input Input, opts []CheckOption) *checkConfig {
	cc := &checkConfig{
		input:     input,
		pluginDir: e.pluginDir,
	}
	for _, opt := range opts {
		opt(cc)
	}
	return cc
}

// loadInput dispatches to the parser based on the concrete Input form.
func loadInput(input Input) ([]*manifest.TestFile, error) {
	p := parser.New()
	switch in := input.(type) {
	case inputPaths:
		return p.ParsePaths(in.paths...)
	case inputBytes:
		return p.ParseBytes(in.data)
	case inputReader:
		data, err := io.ReadAll(in.reader)
		if err != nil {
			return nil, err
		}
		return p.ParseBytes(data)
	}
	return nil, fmt.Errorf("engine: unknown input form %T", input)
}

// buildHooks wires runner hooks to translate runner events into engine.Event.
func buildHooks(handler EventHandler) runner.Hooks {
	return runner.Hooks{
		OnWaveStart: func(ws runner.WaveStart) {
			names := make([]string, len(ws.Tests))
			for i, ref := range ws.Tests {
				names[i] = ref.Test.Name
			}
			handler.Handle(WaveStarted{Index: ws.Index, TestNames: names, Parallel: ws.Parallel})
		},
		OnWaveComplete: func(wo runner.WaveOutcome) {
			handler.Handle(WaveCompleted{WaveResult: convertWave(wo)})
		},
		OnTestStart: func(ts runner.TestStart) {
			file := ""
			if ts.Ref.File != nil {
				file = filepath.ToSlash(ts.Ref.File.Path)
			}
			handler.Handle(TestStarted{
				Name:     ts.Ref.Test.Name,
				File:     file,
				Protocol: ts.Protocol,
				Target:   ts.Target,
			})
		},
		OnTestComplete: func(to runner.TestOutcome) {
			handler.Handle(TestCompleted{TestResult: convertTest(to)})
		},
		OnPluginEvent: func(ev wire.PluginEvent) {
			handler.Handle(PluginEvent{Plugin: ev.Plugin, Kind: ev.Kind, Data: ev.Data})
		},
	}
}

func convertResults(rr *runner.Result, total time.Duration) *Results {
	if rr == nil {
		return &Results{}
	}
	out := &Results{Duration: total}
	for _, t := range rr.Tests {
		tr := convertTest(t)
		out.Tests = append(out.Tests, tr)
		out.Total++
		switch t.Status {
		case wire.StatusPassed:
			out.Passed++
		case wire.StatusFailed:
			out.Failed++
		case wire.StatusSkipped:
			out.Skipped++
		case wire.StatusErrored:
			out.Errored++
		}
	}
	for _, w := range rr.Waves {
		out.Waves = append(out.Waves, convertWave(w))
	}
	return out
}

func convertWave(wo runner.WaveOutcome) WaveResult {
	tests := make([]TestResult, len(wo.Tests))
	var passed, failed, skipped int
	for i, t := range wo.Tests {
		tests[i] = convertTest(t)
		switch t.Status {
		case wire.StatusPassed:
			passed++
		case wire.StatusFailed:
			failed++
		case wire.StatusSkipped:
			skipped++
		}
	}
	return WaveResult{
		Index:    wo.Index,
		Parallel: wo.Parallel,
		Tests:    tests,
		Duration: wo.Duration,
		Passed:   passed,
		Failed:   failed,
		Skipped:  skipped,
	}
}

func convertTest(t runner.TestOutcome) TestResult {
	res := TestResult{
		Protocol: t.Protocol,
		Target:   t.Target,
		Status:   t.Status,
		Duration: t.Duration,
		Error:    t.Error,
	}
	if t.Ref != nil && t.Ref.Test != nil {
		res.Name = t.Ref.Test.Name
	}
	if t.Ref != nil && t.Ref.File != nil {
		res.File = filepath.ToSlash(t.Ref.File.Path)
	}
	if t.Request != nil {
		res.Request = &RequestData{
			Method:  t.Request.Method,
			URL:     t.Request.Resource,
			Headers: t.Request.Headers,
		}
	}
	if t.Response != nil {
		res.Response = &ResponseData{
			Status:   t.Response.Status,
			Headers:  t.Response.Headers,
			Body:     t.Response.Body,
			Duration: time.Duration(t.Response.DurationMs) * time.Millisecond,
		}
	}
	for _, a := range t.Assertions {
		res.Assertions = append(res.Assertions, AssertionResult{
			Expression: a.Expression,
			Passed:     a.Passed,
			Expected:   a.Expected,
			Actual:     a.Actual,
			Message:    a.Message,
		})
	}
	return res
}

func buildRunCompleted(r *Results) RunCompleted {
	return RunCompleted{
		Total:    r.Total,
		Passed:   r.Passed,
		Failed:   r.Failed,
		Skipped:  r.Skipped,
		Errored:  r.Errored,
		Duration: r.Duration,
	}
}

func filePaths(files []*manifest.TestFile) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if f.Path != "" {
			out = append(out, filepath.ToSlash(f.Path))
		}
	}
	return out
}

func countTests(plan *graph.Plan) int {
	n := 0
	for _, w := range plan.Waves {
		n += len(w.Tests)
	}
	return n
}
