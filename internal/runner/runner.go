package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apiqube/engine/internal/wire"
	"github.com/apiqube/engine/internal/dataflow"
	"github.com/apiqube/engine/internal/graph"
	"github.com/apiqube/engine/internal/plugin"
)

// Options control runner behaviour for one Execute call.
type Options struct {
	Parallel      bool
	MaxConcurrent int
	FailFast      bool
	Env           map[string]string
	Hooks         Hooks
}

// Hooks let the engine layer observe runner progress without coupling runner
// to wire.Event types (which would create an import cycle). Each hook is
// optional; nil hooks are no-ops.
type Hooks struct {
	OnWaveStart    func(WaveStart)
	OnWaveComplete func(WaveOutcome)
	OnTestStart    func(TestStart)
	OnTestComplete func(TestOutcome)
	OnPluginEvent  func(wire.PluginEvent)
}

// WaveStart describes a wave that is about to begin.
type WaveStart struct {
	Index    int
	Parallel bool
	Tests    []*graph.TestRef
}

// TestStart describes a test that is about to begin.
type TestStart struct {
	Ref      *graph.TestRef
	Protocol wire.Protocol
	Target   string
}

// TestOutcome is the recorded outcome of one test execution.
type TestOutcome struct {
	Ref        *graph.TestRef
	Protocol   wire.Protocol
	Target     string
	Status     wire.TestStatus
	Skipped    bool
	StartTime  time.Time
	Duration   time.Duration
	Request    *plugin.TestInput
	Response   *plugin.TestOutput
	Assertions []AssertionOutcome
	Error      string
}

// AssertionOutcome is one assertion result captured for reporting.
type AssertionOutcome struct {
	Expression string
	Passed     bool
	Expected   any
	Actual     any
	Message    string
}

// WaveOutcome aggregates one wave's test outcomes.
type WaveOutcome struct {
	Index    int
	Parallel bool
	Tests    []TestOutcome
	Duration time.Duration
}

// Result is the aggregated outcome of an Execute call.
type Result struct {
	Tests []TestOutcome
	Waves []WaveOutcome
}

// Runner orchestrates the execution of a graph.Plan via plugins from the
// registry, using dataflow.Store for cross-test data passing.
type Runner struct {
	registry *plugin.Registry
	options  Options
}

// New constructs a Runner with the given plugin registry and options.
func New(registry *plugin.Registry, options Options) *Runner {
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 8
	}
	return &Runner{registry: registry, options: options}
}

// Execute runs every wave of plan in order. Within a wave, tests run in
// parallel when both Wave.Parallel and Options.Parallel are true.
//
// Returns a non-nil Result; the error is non-nil only on a fatal engine-layer
// problem (missing plugin, ctx canceled). Test failures are reported in
// Result.Tests.
func (r *Runner) Execute(ctx context.Context, plan *graph.Plan) (*Result, error) {
	store := dataflow.New()
	defer func() { _ = store.Close() }()

	result := &Result{}
	for i, wave := range plan.Waves {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		outcome, err := r.runWave(ctx, wave, store)
		if err != nil {
			return result, err
		}
		result.Waves = append(result.Waves, outcome)
		result.Tests = append(result.Tests, outcome.Tests...)
		if r.options.FailFast && hasFailure(outcome) {
			r.skipRemaining(result, plan.Waves[i+1:])
			break
		}
	}
	return result, nil
}

// skipRemaining records every test in subsequent waves as Skipped so the
// final Result reflects the full plan even after a failFast abort.
func (r *Runner) skipRemaining(result *Result, remaining []graph.Wave) {
	for _, w := range remaining {
		skipped := make([]TestOutcome, len(w.Tests))
		for i, ref := range w.Tests {
			skipped[i] = TestOutcome{
				Ref:     ref,
				Status:  wire.StatusSkipped,
				Skipped: true,
			}
		}
		result.Waves = append(result.Waves, WaveOutcome{
			Index:    w.Index,
			Parallel: w.Parallel && r.options.Parallel,
			Tests:    skipped,
		})
		result.Tests = append(result.Tests, skipped...)
	}
}

func hasFailure(w WaveOutcome) bool {
	for _, t := range w.Tests {
		if t.Status == wire.StatusFailed || t.Status == wire.StatusErrored {
			return true
		}
	}
	return false
}

// runWave executes all tests in a wave. Parallel waves use bounded concurrency
// from options.MaxConcurrent. failFast in sequential mode stops after the
// first failed test.
func (r *Runner) runWave(ctx context.Context, wave graph.Wave, store *dataflow.Store) (WaveOutcome, error) {
	if h := r.options.Hooks.OnWaveStart; h != nil {
		h(WaveStart{Index: wave.Index, Parallel: wave.Parallel && r.options.Parallel, Tests: wave.Tests})
	}
	start := time.Now()

	outcomes := make([]TestOutcome, len(wave.Tests))
	if wave.Parallel && r.options.Parallel && len(wave.Tests) > 1 {
		r.runParallel(ctx, wave.Tests, store, outcomes)
	} else {
		r.runSequential(ctx, wave.Tests, store, outcomes)
	}

	out := WaveOutcome{
		Index:    wave.Index,
		Parallel: wave.Parallel && r.options.Parallel,
		Tests:    outcomes,
		Duration: time.Since(start),
	}
	if h := r.options.Hooks.OnWaveComplete; h != nil {
		h(out)
	}
	return out, nil
}

func (r *Runner) runParallel(ctx context.Context, refs []*graph.TestRef, store *dataflow.Store, outcomes []TestOutcome) {
	sem := make(chan struct{}, r.options.MaxConcurrent)
	var wg sync.WaitGroup
	for i, ref := range refs {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			outcomes[i] = r.runTest(ctx, ref, store)
		})
	}
	wg.Wait()
}

func (r *Runner) runSequential(ctx context.Context, refs []*graph.TestRef, store *dataflow.Store, outcomes []TestOutcome) {
	for i, ref := range refs {
		outcomes[i] = r.runTest(ctx, ref, store)
		if r.options.FailFast {
			s := outcomes[i].Status
			if s == wire.StatusFailed || s == wire.StatusErrored {
				// Mark remaining as skipped to keep result complete.
				for j := i + 1; j < len(refs); j++ {
					outcomes[j] = TestOutcome{
						Ref:     refs[j],
						Status:  wire.StatusSkipped,
						Skipped: true,
					}
				}
				return
			}
		}
	}
}

// noPlugin builds an errored outcome when no plugin handles the protocol.
func noPlugin(ref *graph.TestRef, proto wire.Protocol, target string) TestOutcome {
	return TestOutcome{
		Ref:      ref,
		Protocol: proto,
		Target:   target,
		Status:   wire.StatusErrored,
		Error:    fmt.Sprintf("no plugin registered for protocol %q", proto),
	}
}
