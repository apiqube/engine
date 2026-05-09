package graph

import (
	"reflect"
	"testing"

	"github.com/apiqube/engine/internal/manifest"
)

func TestExtractReferences_Strings(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want []Reference
	}{
		{"single ref", "Bearer {{ token.value }}",
			[]Reference{{Name: "token", Path: "value"}}},
		{"bare", "{{ token }}",
			[]Reference{{Name: "token", Path: ""}}},
		{"two refs same string", "{{ a.x }} and {{ b.y }}",
			[]Reference{{Name: "a", Path: "x"}, {Name: "b", Path: "y"}}},
		{"deduped", "{{ a.x }} and {{ a.x }}",
			[]Reference{{Name: "a", Path: "x"}}},
		{"with method chain", "{{ user.name.ToUpper() }}",
			[]Reference{{Name: "user", Path: "name"}}},
		{"nested method chain", "{{ a.b.Replace('x', 'y').ToLower() }}",
			[]Reference{{Name: "a", Path: "b"}}},
		{"fake skipped", "{{ fake.email }}", nil},
		{"env skipped", "{{ env.HOST }}", nil},
		{"regex skipped", `{{ regex('^abc$') }}`, nil},
		{"plain string", "no templates here", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractReferences(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v; want %v", got, c.want)
			}
		})
	}
}

func TestExtractReferences_NestedStructures(t *testing.T) {
	source := map[string]any{
		"top":    "{{ a.x }}",
		"nested": map[string]any{"k": "{{ b.y }}"},
		"list":   []any{"plain", "{{ c.z }}"},
		"int":    42,
		"nilv":   nil,
	}
	got := extractReferences(source)
	want := map[Reference]bool{
		{Name: "a", Path: "x"}: true,
		{Name: "b", Path: "y"}: true,
		{Name: "c", Path: "z"}: true,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d refs, want %d", len(got), len(want))
	}
	for _, r := range got {
		if !want[r] {
			t.Errorf("unexpected ref %v", r)
		}
	}
}

func TestExtractReferencesFromTest(t *testing.T) {
	tc := &manifest.TestCase{
		Method:   "POST",
		Resource: "/users/{{ parent.id }}",
		Headers:  map[string]string{"X-Trace": "{{ ctx.trace }}"},
		Save:     map[string]string{"newId": "body.id"},
		Expect: manifest.Expect{
			Status: "{{ expected.status }}",
		},
		Extra: map[string]any{
			"body": map[string]any{
				"name": "{{ fake.name }}",      // skipped
				"hint": "{{ parent.value }}",   // counted; same alias name "parent"
			},
		},
	}
	got := extractReferencesFromTest(tc)
	want := map[Reference]bool{
		{Name: "parent", Path: "id"}:    true,
		{Name: "ctx", Path: "trace"}:    true,
		{Name: "expected", Path: "status"}: true,
		{Name: "parent", Path: "value"}: true,
	}
	if len(got) != len(want) {
		t.Errorf("got %d refs, want %d: %v", len(got), len(want), got)
	}
	for _, r := range got {
		if !want[r] {
			t.Errorf("unexpected ref %v", r)
		}
	}
}

func TestExtractReferencesFromTest_NilSafe(t *testing.T) {
	if got := extractReferencesFromTest(nil); got != nil {
		t.Errorf("nil test should return nil refs, got %v", got)
	}
}

func TestParseRef_TrailingMethods(t *testing.T) {
	r, ok := parseRef("user.name.ToUpper().Replace('a', 'b')")
	if !ok || r.Name != "user" || r.Path != "name" {
		t.Errorf("got (%v, %v)", r, ok)
	}
}

func TestParseRef_Empty(t *testing.T) {
	if _, ok := parseRef(""); ok {
		t.Error("empty body should not parse")
	}
}
