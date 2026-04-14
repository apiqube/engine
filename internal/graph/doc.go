// Package graph builds and analyzes the dependency graph between test cases.
//
// The graph builder inspects template references in test fields ({{ alias.path }},
// {{ savedVar }}, {{ prev.body.id }}) and constructs a directed acyclic graph
// where edges represent "must run after" relationships.
//
// After the graph is built, it is topologically sorted and grouped into waves:
// sets of tests that have no dependencies on each other and can run in parallel.
//
// # Responsibilities
//
//   - Extract template references via regex
//   - Match references to alias-defining tests
//   - Build a directed graph of dependencies
//   - Detect cycles and report them with the participating tests
//   - Topologically sort with deterministic tie-breaking
//   - Group into parallel waves
//   - Compute SaveRequirements (which fields each producer must persist)
package graph
