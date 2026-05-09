package graph

import (
	"errors"
	"testing"

	"github.com/apiqube/engine/internal/manifest"
)

func newRef(name string, idx int, path string) *TestRef {
	return &TestRef{
		File:  &manifest.TestFile{Path: path},
		Test:  &manifest.TestCase{Name: name},
		Index: idx,
	}
}

func TestToposort_Empty(t *testing.T) {
	got, err := toposort(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("empty input should produce nil waves, got %v", got)
	}
}

func TestToposort_Linear(t *testing.T) {
	a := newRef("a", 0, "f.yaml")
	b := newRef("b", 1, "f.yaml")
	c := newRef("c", 2, "f.yaml")
	deps := []Dependency{
		{From: b, To: a, Type: DepTemplate},
		{From: c, To: b, Type: DepTemplate},
	}
	waves, err := toposort([]*TestRef{a, b, c}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 3 {
		t.Fatalf("got %d waves, want 3", len(waves))
	}
	if waves[0].Tests[0] != a || waves[1].Tests[0] != b || waves[2].Tests[0] != c {
		t.Errorf("wrong order: %+v", waves)
	}
}

func TestToposort_Parallel(t *testing.T) {
	a := newRef("a", 0, "f.yaml")
	b := newRef("b", 1, "f.yaml")
	c := newRef("c", 2, "f.yaml")
	// b and c independent, both before nothing.
	waves, err := toposort([]*TestRef{a, b, c}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 1 {
		t.Fatalf("expected single wave, got %d", len(waves))
	}
	if !waves[0].Parallel {
		t.Errorf("multi-test wave should be marked Parallel")
	}
}

func TestToposort_Diamond(t *testing.T) {
	root := newRef("root", 0, "f.yaml")
	left := newRef("left", 1, "f.yaml")
	right := newRef("right", 2, "f.yaml")
	end := newRef("end", 3, "f.yaml")
	deps := []Dependency{
		{From: left, To: root, Type: DepTemplate},
		{From: right, To: root, Type: DepTemplate},
		{From: end, To: left, Type: DepTemplate},
		{From: end, To: right, Type: DepTemplate},
	}
	waves, err := toposort([]*TestRef{root, left, right, end}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 3 {
		t.Fatalf("got %d waves, want 3", len(waves))
	}
	if waves[0].Tests[0] != root {
		t.Error("first wave must be root")
	}
	if len(waves[1].Tests) != 2 {
		t.Errorf("middle wave must have 2 tests, got %d", len(waves[1].Tests))
	}
	if waves[2].Tests[0] != end {
		t.Error("last wave must be end")
	}
}

func TestToposort_Cycle(t *testing.T) {
	a := newRef("a", 0, "f.yaml")
	b := newRef("b", 1, "f.yaml")
	deps := []Dependency{
		{From: a, To: b, Type: DepTemplate},
		{From: b, To: a, Type: DepTemplate},
	}
	_, err := toposort([]*TestRef{a, b}, deps)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *CycleError, got %T", err)
	}
	if len(ce.Cycle) != 2 {
		t.Errorf("cycle should include both tests, got %d", len(ce.Cycle))
	}
}

func TestToposort_DeterministicOrder(t *testing.T) {
	// Two independent tests: order in waves must be deterministic by file/index/name.
	a := newRef("zeta", 0, "x.yaml")
	b := newRef("alpha", 1, "x.yaml")
	for range 5 {
		waves, err := toposort([]*TestRef{b, a}, nil)
		if err != nil {
			t.Fatal(err)
		}
		// File equal; lower index wins.
		if waves[0].Tests[0] != a {
			t.Errorf("ordering not deterministic: %v", waves[0].Tests)
		}
	}
}

func TestToposort_SelfEdgeIgnored(t *testing.T) {
	a := newRef("a", 0, "f.yaml")
	deps := []Dependency{{From: a, To: a, Type: DepTemplate}}
	waves, err := toposort([]*TestRef{a}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 1 {
		t.Errorf("self-edge should not block, got %d waves", len(waves))
	}
}

func TestCycleError_Message(t *testing.T) {
	a := newRef("a", 0, "f.yaml")
	b := newRef("b", 1, "f.yaml")
	err := &CycleError{Cycle: []*TestRef{a, b}}
	if err.Error() == "" {
		t.Error("error should have a message")
	}
}
