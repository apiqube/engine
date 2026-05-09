package graph

import (
	"fmt"
	"sort"
)

// toposort runs Kahn's algorithm over the dependency graph, producing waves
// of tests that can run in parallel. Each wave contains tests with all of
// their dependencies satisfied by earlier waves. Within a wave, tests are
// sorted deterministically by (file path, in-file index, name).
func toposort(tests []*TestRef, deps []Dependency) ([]Wave, error) {
	if len(tests) == 0 {
		return nil, nil
	}

	inDeg := make(map[*TestRef]int, len(tests))
	for _, ref := range tests {
		inDeg[ref] = 0
	}
	dependents := make(map[*TestRef][]*TestRef, len(tests))
	for _, d := range deps {
		if d.From == d.To {
			continue
		}
		dependents[d.To] = append(dependents[d.To], d.From)
		inDeg[d.From]++
	}

	processed := make(map[*TestRef]bool, len(tests))
	var waves []Wave
	for waveIdx := 0; len(processed) < len(tests); waveIdx++ {
		var ready []*TestRef
		for _, ref := range tests {
			if processed[ref] {
				continue
			}
			if inDeg[ref] == 0 {
				ready = append(ready, ref)
			}
		}
		if len(ready) == 0 {
			cycle := findCycle(tests, deps, processed)
			return nil, &CycleError{Cycle: cycle}
		}
		sort.SliceStable(ready, func(i, j int) bool {
			return less(ready[i], ready[j])
		})
		waves = append(waves, Wave{Index: waveIdx, Parallel: len(ready) > 1, Tests: ready})

		for _, ref := range ready {
			processed[ref] = true
			for _, dep := range dependents[ref] {
				inDeg[dep]--
			}
		}
	}
	return waves, nil
}

// findCycle returns the tests that participate in a dependency cycle. Used
// when toposort cannot make progress.
func findCycle(tests []*TestRef, _ []Dependency, processed map[*TestRef]bool) []*TestRef {
	var remaining []*TestRef
	for _, ref := range tests {
		if !processed[ref] {
			remaining = append(remaining, ref)
		}
	}
	sort.SliceStable(remaining, func(i, j int) bool {
		return less(remaining[i], remaining[j])
	})
	return remaining
}

// CycleError indicates a dependency cycle. Cycle lists the tests involved.
type CycleError struct {
	Cycle []*TestRef
}

func (e *CycleError) Error() string {
	names := make([]string, len(e.Cycle))
	for i, r := range e.Cycle {
		names[i] = r.ID()
	}
	return fmt.Sprintf("dependency cycle: %v", names)
}

func less(a, b *TestRef) bool {
	if a.File != nil && b.File != nil && a.File.Path != b.File.Path {
		return a.File.Path < b.File.Path
	}
	if a.Index != b.Index {
		return a.Index < b.Index
	}
	return a.testName() < b.testName()
}
