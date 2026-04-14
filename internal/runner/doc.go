// Package runner executes parsed test plans by coordinating all engine subsystems.
//
// The runner is the orchestrator — it does not itself send requests (that's
// plugins), resolve templates (that's template), extract data (that's dataflow),
// or check assertions (that's assertion). It wires everything together and
// ensures each subsystem gets what it needs at the right time.
//
// # Execution flow
//
//  1. Take a graph.Plan produced by the graph builder
//  2. For each Wave in the plan:
//     a. Start all tests in the wave (parallel or sequential based on wave type)
//     b. For each test:
//        - Resolve templates in all fields via template.Resolver
//        - Build plugin.TestInput with resolved data
//        - Emit TestStarted event
//        - Call plugin.Execute(input) → output
//        - Run assertions via assertion.Engine
//        - Extract save fields via dataflow.Store
//        - Update prev snapshot
//        - Emit TestCompleted event
//     c. Wait for all wave tests to finish
//     d. Emit WaveCompleted event
//  3. Assemble final Results
package runner
