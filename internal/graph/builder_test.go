package graph

import (
	"errors"
	"testing"

	"github.com/apiqube/engine/internal/manifest"
)

func makeFile(path string, mode manifest.TestMode, tests ...manifest.TestCase) *manifest.TestFile {
	return &manifest.TestFile{Path: path, Mode: mode, Tests: tests}
}

func TestBuild_AutoDependencyFromTemplate(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "fetch", Alias: "user", Method: "GET", Resource: "/users/1"},
		manifest.TestCase{Name: "use", Method: "POST", Resource: "/posts",
			Extra: map[string]any{"body": map[string]any{"author": "{{ user.id }}"}}},
	)

	plan, err := New().Build([]*manifest.TestFile{file})
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.Dependencies) != 1 {
		t.Fatalf("got %d deps, want 1", len(plan.Dependencies))
	}
	dep := plan.Dependencies[0]
	if dep.Type != DepTemplate || dep.Alias != "user" {
		t.Errorf("dep wrong: %+v", dep)
	}
	if len(dep.Paths) != 1 || dep.Paths[0] != "id" {
		t.Errorf("paths wrong: %v", dep.Paths)
	}

	// Save requirements: producer must save "id".
	req := plan.SaveRequirements[plan.Dependencies[0].To.ID()]
	if !req.Required {
		t.Error("producer should be flagged Required")
	}
	if len(req.Paths) != 1 || req.Paths[0] != "id" {
		t.Errorf("saveRequirement paths wrong: %v", req.Paths)
	}
}

func TestBuild_ExplicitDepends(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "first", Alias: "x"},
		manifest.TestCase{Name: "second", Depends: []string{"x"}},
	)
	plan, err := New().Build([]*manifest.TestFile{file})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Dependencies) != 1 {
		t.Fatalf("got %d deps, want 1", len(plan.Dependencies))
	}
	if plan.Dependencies[0].Type != DepExplicit {
		t.Errorf("expected DepExplicit, got %v", plan.Dependencies[0].Type)
	}
}

func TestBuild_ExplicitDependsUnknown(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Depends: []string{"missing"}},
	)
	if _, err := New().Build([]*manifest.TestFile{file}); err == nil {
		t.Error("expected error on unknown alias")
	}
}

func TestBuild_DuplicateAlias(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "a", Alias: "x"},
		manifest.TestCase{Name: "b", Alias: "x"},
	)
	_, err := New().Build([]*manifest.TestFile{file})
	if err == nil {
		t.Error("expected duplicate-alias error")
	}
}

func TestBuild_ScenarioMode(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeScenario,
		manifest.TestCase{Name: "step1"},
		manifest.TestCase{Name: "step2"},
		manifest.TestCase{Name: "step3"},
	)
	plan, err := New().Build([]*manifest.TestFile{file})
	if err != nil {
		t.Fatal(err)
	}
	// Scenario mode: each step depends on previous → 2 DepPrev edges.
	prev := 0
	for _, d := range plan.Dependencies {
		if d.Type == DepPrev {
			prev++
		}
	}
	if prev != 2 {
		t.Errorf("got %d DepPrev edges, want 2", prev)
	}

	// Three sequential waves.
	if len(plan.Waves) != 3 {
		t.Errorf("got %d waves, want 3", len(plan.Waves))
	}
}

func TestBuild_CrossFileAliases(t *testing.T) {
	a := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "produce", Alias: "shared"},
	)
	b := makeFile("b.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "consume", Headers: map[string]string{"X-Id": "{{ shared.id }}"}},
	)
	plan, err := New().Build([]*manifest.TestFile{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Dependencies) != 1 {
		t.Fatalf("got %d deps, want 1", len(plan.Dependencies))
	}
	if len(plan.Waves) != 2 {
		t.Errorf("got %d waves, want 2", len(plan.Waves))
	}
}

func TestBuild_CycleError(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "x", Alias: "x", Depends: []string{"y"}},
		manifest.TestCase{Name: "y", Alias: "y", Depends: []string{"x"}},
	)
	_, err := New().Build([]*manifest.TestFile{file})
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Errorf("expected *CycleError, got %T", err)
	}
}

func TestBuild_PathDeduplication(t *testing.T) {
	file := makeFile("a.yaml", manifest.ModeTest,
		manifest.TestCase{Name: "p", Alias: "p"},
		manifest.TestCase{Name: "c", Headers: map[string]string{
			"H1": "{{ p.x }}",
			"H2": "{{ p.x }}",
			"H3": "{{ p.y }}",
		}},
	)
	plan, err := New().Build([]*manifest.TestFile{file})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Dependencies) != 1 {
		t.Fatalf("got %d deps, want 1 (merged)", len(plan.Dependencies))
	}
	if len(plan.Dependencies[0].Paths) != 2 {
		t.Errorf("paths should dedup to 2, got %v", plan.Dependencies[0].Paths)
	}
}

func TestBuild_NoFiles(t *testing.T) {
	plan, err := New().Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan == nil {
		t.Fatal("plan should not be nil")
	}
	if len(plan.Waves) != 0 || len(plan.Dependencies) != 0 {
		t.Errorf("empty input should produce empty plan, got %+v", plan)
	}
}

func TestTestRef_ID(t *testing.T) {
	r := newRef("x", 5, "path/to/file.yaml")
	if got := r.ID(); got != "path/to/file.yaml#5" {
		t.Errorf("ID() = %q", got)
	}
}
