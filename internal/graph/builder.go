package graph

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/apiqube/engine/internal/manifest"
)

// Builder constructs a dependency graph from a set of test files.
type Builder struct{}

// New returns a new Builder. Builder is stateless and safe for concurrent use.
func New() *Builder {
	return &Builder{}
}

// Build analyzes every test across the given files and produces a Plan with
// ordered waves, dependencies, and the per-producer save requirements.
//
// Build returns an error on duplicate aliases, missing alias references in
// explicit `depends:`, or dependency cycles (returned as *CycleError).
//
// Matrix expansion and load-test replication are deferred — see TODOs.
func (b *Builder) Build(files []*manifest.TestFile) (*Plan, error) {
	refs, aliasMap, err := collectRefs(files)
	if err != nil {
		return nil, err
	}

	deps := make([]Dependency, 0)

	// Auto-detected template references.
	for _, ref := range refs {
		// TODO(matrix): when matrix is implemented, expand each row before this loop.
		for _, r := range extractReferencesFromTest(ref.Test) {
			target := aliasMap[r.Name]
			if target == nil || target == ref {
				continue
			}
			deps = mergeDep(deps, Dependency{
				From: ref, To: target, Type: DepTemplate,
				Alias: r.Name, Paths: trimNonEmpty([]string{r.Path}),
			})
		}
	}

	// Explicit depends: list.
	for _, ref := range refs {
		for _, alias := range ref.Test.Depends {
			target := aliasMap[alias]
			if target == nil {
				return nil, fmt.Errorf("test %s: depends on unknown alias %q", ref.ID(), alias)
			}
			if target == ref {
				continue
			}
			deps = mergeDep(deps, Dependency{
				From: ref, To: target, Type: DepExplicit, Alias: alias,
			})
		}
	}

	// Scenario mode: each test depends on the previous test in the same file.
	for _, file := range files {
		if file.Mode != manifest.ModeScenario {
			continue
		}
		fileRefs := refsForFile(refs, file)
		for i := 1; i < len(fileRefs); i++ {
			deps = mergeDep(deps, Dependency{
				From: fileRefs[i], To: fileRefs[i-1], Type: DepPrev,
			})
		}
	}

	waves, err := toposort(refs, deps)
	if err != nil {
		return nil, err
	}

	return &Plan{
		Waves:            waves,
		Dependencies:     deps,
		SaveRequirements: computeSaveRequirements(deps),
	}, nil
}

// Plan is the execution plan produced by the graph builder.
type Plan struct {
	Waves            []Wave
	Dependencies     []Dependency
	SaveRequirements map[string]SaveRequirement
}

// Wave is a group of tests that can run in parallel (no dependencies between them).
type Wave struct {
	Index    int
	Parallel bool
	Tests    []*TestRef
}

// TestRef points to a specific test within a specific file.
type TestRef struct {
	File  *manifest.TestFile
	Test  *manifest.TestCase
	Mode  manifest.TestMode
	Index int // position within the file
}

// ID returns a stable identifier for the test based on file path and index.
func (r *TestRef) ID() string {
	path := ""
	if r.File != nil {
		path = r.File.Path
	}
	return path + "#" + strconv.Itoa(r.Index)
}

func (r *TestRef) testName() string {
	if r == nil || r.Test == nil {
		return ""
	}
	return r.Test.Name
}

// Dependency describes a "must run after" relationship.
type Dependency struct {
	From  *TestRef
	To    *TestRef
	Type  DependencyType
	Alias string
	Paths []string
}

// DependencyType classifies how a dependency was discovered.
type DependencyType int

const (
	DepTemplate DependencyType = iota
	DepExplicit
	DepPrev
)

// SaveRequirement describes what a producer test must persist for its consumers.
type SaveRequirement struct {
	Required  bool
	Paths     []string
	Consumers []string
}

func collectRefs(files []*manifest.TestFile) ([]*TestRef, map[string]*TestRef, error) {
	var refs []*TestRef
	aliasMap := make(map[string]*TestRef)
	for _, f := range files {
		for i := range f.Tests {
			ref := &TestRef{File: f, Test: &f.Tests[i], Mode: f.Mode, Index: i}
			refs = append(refs, ref)
			alias := f.Tests[i].Alias
			if alias == "" {
				continue
			}
			if existing := aliasMap[alias]; existing != nil {
				return nil, nil, fmt.Errorf("duplicate alias %q (in %s and %s)",
					alias, existing.ID(), ref.ID())
			}
			aliasMap[alias] = ref
		}
	}
	return refs, aliasMap, nil
}

func refsForFile(all []*TestRef, file *manifest.TestFile) []*TestRef {
	var out []*TestRef
	for _, r := range all {
		if r.File == file {
			out = append(out, r)
		}
	}
	return out
}

// mergeDep merges a new dependency into the slice, accumulating Paths when an
// edge of the same (From, To, Type) already exists.
func mergeDep(deps []Dependency, d Dependency) []Dependency {
	for i := range deps {
		if deps[i].From == d.From && deps[i].To == d.To && deps[i].Type == d.Type {
			deps[i].Paths = mergePaths(deps[i].Paths, d.Paths)
			return deps
		}
	}
	return append(deps, d)
}

func mergePaths(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, p := range a {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, p := range b {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func trimNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func computeSaveRequirements(deps []Dependency) map[string]SaveRequirement {
	out := make(map[string]SaveRequirement)
	for _, d := range deps {
		if d.Type != DepTemplate {
			continue
		}
		producer := d.To.ID()
		req := out[producer]
		req.Required = true
		req.Paths = mergePaths(req.Paths, d.Paths)
		req.Consumers = appendUnique(req.Consumers, d.From.ID())
		out[producer] = req
	}
	return out
}

func appendUnique(slice []string, s string) []string {
	if slices.Contains(slice, s) {
		return slice
	}
	return append(slice, s)
}
