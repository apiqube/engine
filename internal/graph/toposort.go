package graph

// toposort performs Kahn's algorithm on the dependency graph.
// Tests with no unsatisfied dependencies move into the "ready" set each iteration.
// Each iteration produces one Wave.
//
// Cycle detection: if after processing all reachable nodes, any node remains
// in the pending set, a cycle exists. The remaining nodes participate in it.
func toposort(tests []*TestRef, deps []Dependency) ([]Wave, error) {
	// TODO: implementation
	//
	// Algorithm:
	// 1. Build in-degree map: for each test, count incoming dependencies
	// 2. Start with all tests that have in-degree 0
	// 3. For each iteration:
	//    a. Collect all tests with in-degree 0 → this is the current wave
	//    b. Sort wave deterministically (by file path, then test name)
	//    c. For each test in wave, decrement in-degree of dependents
	//    d. Remove processed tests from in-degree map
	// 4. Repeat until in-degree map is empty
	// 5. If map is not empty but no zero-degree tests → cycle
	return nil, nil
}

// findCycle returns the set of tests participating in a dependency cycle.
// Used for error reporting when toposort fails.
func findCycle(tests []*TestRef, deps []Dependency) []*TestRef {
	// TODO: implementation via DFS with grey/black coloring
	return nil
}
