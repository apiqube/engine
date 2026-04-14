package runner

import (
	"context"

	"github.com/apiqube/engine/internal/graph"
)

// runWave executes all tests in a single wave.
// Tests run in parallel if wave.Parallel is true, otherwise sequentially.
// Respects options.MaxConcurrent to limit concurrency.
func (r *Runner) runWave(ctx context.Context, wave graph.Wave) (WaveOutcome, error) {
	// TODO: implementation
	//
	// Parallel path:
	//   1. Create a semaphore with capacity min(options.MaxConcurrent, len(wave.Tests))
	//   2. For each test, spawn goroutine that acquires semaphore, calls runTest, releases
	//   3. Collect all TestOutcomes, sort by test order for determinism
	//
	// Sequential path:
	//   1. For each test in order, call runTest
	//   2. If failFast and test failed, stop iteration
	return WaveOutcome{Index: wave.Index, Parallel: wave.Parallel}, nil
}

// runTest executes a single test case through the full pipeline:
// resolve templates → invoke plugin → check assertions → save data.
func (r *Runner) runTest(ctx context.Context, ref *graph.TestRef) (TestOutcome, error) {
	// TODO: implementation
	//
	// 1. Determine protocol from target URL
	// 2. Look up plugin for protocol
	// 3. Resolve templates in all test fields (including Extra)
	// 4. Build plugin.TestInput with resolved data
	// 5. Emit TestStarted (via r.emit)
	// 6. Call plugin.Execute(input) → output
	// 7. Run assertions against output
	// 8. Extract save fields into dataflow.Store
	// 9. Update prev snapshot if scenario mode
	// 10. Emit TestCompleted
	// 11. Return TestOutcome
	return TestOutcome{}, nil
}
