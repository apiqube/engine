package runner

import (
	"context"

	"github.com/apiqube/engine/internal/assertion"
	"github.com/apiqube/engine/internal/dataflow"
	"github.com/apiqube/engine/internal/graph"
	"github.com/apiqube/engine/internal/plugin"
	"github.com/apiqube/engine/internal/template"
)

// Runner executes a graph.Plan using the given plugins and subsystems.
// A Runner is created fresh for each Run() call — not reused across runs.
type Runner struct {
	plugins   map[string]plugin.Plugin   // protocol name → plugin
	store     *dataflow.Store
	templates *template.Resolver
	asserts   *assertion.Engine
	emit      func(EventLike)
	signals   <-chan SignalLike
	options   Options
}

// Options control runner behavior.
type Options struct {
	Parallel      bool
	MaxConcurrent int
	FailFast      bool
}

// EventLike is a minimal interface the runner uses to emit events to the engine layer.
// The engine wraps runner events and forwards them to the user's EventHandler.
type EventLike interface {
	Type() string
}

// SignalLike represents a control signal from the frontend.
// Runner reads from this channel if provided and reacts to pause/resume/skip.
type SignalLike int

// New creates a Runner with all subsystems wired up.
func New(
	plugins map[string]plugin.Plugin,
	store *dataflow.Store,
	templates *template.Resolver,
	asserts *assertion.Engine,
	emit func(EventLike),
	signals <-chan SignalLike,
	opts Options,
) *Runner {
	return &Runner{
		plugins:   plugins,
		store:     store,
		templates: templates,
		asserts:   asserts,
		emit:      emit,
		signals:   signals,
		options:   opts,
	}
}

// Execute runs the given plan to completion or until ctx is cancelled.
// Returns aggregated results or the first fatal error encountered.
func (r *Runner) Execute(ctx context.Context, plan *graph.Plan) (*Result, error) {
	// TODO: implementation
	// 1. Emit RunStarted event
	// 2. For each wave in plan.Waves:
	//    a. Emit WaveStarted
	//    b. runWave() — parallel or sequential
	//    c. Emit WaveCompleted
	//    d. If failFast and wave has failures → stop
	// 3. Emit RunCompleted
	return nil, nil
}

// Result holds the execution outcome that the engine layer will convert to engine.Results.
type Result struct {
	Tests []TestOutcome
	Waves []WaveOutcome
}

// TestOutcome is one test's execution record, plugin-agnostic.
type TestOutcome struct {
	Name       string
	File       string
	Protocol   string
	Target     string
	Status     int
	DurationMs int64
	Request    *plugin.TestInput
	Response   *plugin.TestOutput
	Assertions []assertion.Result
	Error      string
}

// WaveOutcome is one wave's execution record.
type WaveOutcome struct {
	Index      int
	Parallel   bool
	Tests      []TestOutcome
	DurationMs int64
}
