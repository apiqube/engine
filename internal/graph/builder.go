package graph

import "github.com/apiqube/engine/internal/manifest"

// Builder constructs a dependency graph from a set of test files.
type Builder struct{}

// New creates a new graph Builder.
func New() *Builder {
	return &Builder{}
}

// Build analyzes all tests across the given files and produces an execution Plan.
// The Plan contains ordered waves, dependencies, and save requirements.
func (b *Builder) Build(files []*manifest.TestFile) (*Plan, error) {
	// TODO: implementation
	// 1. Collect all aliases across all files (alias → test pointer)
	// 2. For each test, extract template references from all string fields
	// 3. Match references to aliases, build Dependency edges
	// 4. Run topological sort with cycle detection
	// 5. Group into Waves (independent tests → same wave)
	// 6. Compute SaveRequirements (what each producer must save)
	return nil, nil
}

// Plan is the execution plan produced by the graph builder.
type Plan struct {
	Waves            []Wave
	Dependencies     []Dependency
	SaveRequirements map[string]SaveRequirement // test ID → what to save
}

// Wave is a group of tests that can run in parallel (no dependencies between them).
type Wave struct {
	Index    int
	Parallel bool
	Tests    []*TestRef
}

// TestRef is a pointer to a specific test within a specific file.
type TestRef struct {
	File *manifest.TestFile
	Test *manifest.TestCase
	Mode string // "test", "scenario", "load"
}

// Dependency describes a "must run after" relationship.
type Dependency struct {
	From     *TestRef // consumer
	To       *TestRef // producer
	Type     DependencyType
	Alias    string   // alias being referenced
	Paths    []string // gjson paths consumer needs (e.g., ["response.body.id"])
}

// DependencyType classifies how a dependency was discovered.
type DependencyType int

const (
	DepTemplate DependencyType = iota // auto-detected from {{ alias.path }}
	DepExplicit                       // from depends: field
	DepPrev                           // from {{ prev.* }} in scenario mode
)

// SaveRequirement tells a producer test what fields it must save to support consumers.
type SaveRequirement struct {
	Required  bool
	Paths     []string
	Consumers []string // test IDs that read these paths
}
