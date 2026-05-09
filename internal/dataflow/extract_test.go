package dataflow

import (
	"reflect"
	"testing"
)

func TestExtract(t *testing.T) {
	source := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  30,
			"tags": []any{"admin", "user"},
		},
		"items": []any{
			map[string]any{"id": 1, "name": "a"},
			map[string]any{"id": 2, "name": "b"},
		},
		"empty":   "",
		"unicode": "Привет",
	}

	cases := []struct {
		name string
		path string
		want any
		ok   bool
	}{
		{"empty path returns source", "", source, true},
		{"top-level key", "user", source["user"], true},
		{"nested key", "user.name", "alice", true},
		{"nested int", "user.age", 30, true},
		{"slice index", "items.0.id", 1, true},
		{"slice index 2", "items.1.name", "b", true},
		{"slice element", "user.tags.0", "admin", true},
		{"length of slice", "user.tags.length", 2, true},
		{"length of map", "user.length", 3, true},
		{"length of string", "unicode.length", 6, true},
		{"length of empty string", "empty.length", 0, true},
		{"missing key", "user.email", nil, false},
		{"out of range", "items.99", nil, false},
		{"index on map", "user.0", nil, false},
		{"key on slice", "items.foo", nil, false},
		{"length on int", "user.age.length", nil, false},
		{"trailing dot", "user.", nil, false},
		{"leading dot", ".user", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Extract(source, tc.path)
			if ok != tc.ok {
				t.Fatalf("Extract(_, %q) ok=%v want=%v", tc.path, ok, tc.ok)
			}
			if ok && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Extract(_, %q) = %v; want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestExtract_TypedMap(t *testing.T) {
	// reflect-based path for map[string]int and friends.
	source := map[string]int{"a": 1, "b": 2}
	got, ok := Extract(source, "b")
	if !ok || got != 2 {
		t.Errorf("typed-map lookup failed: got=%v ok=%v", got, ok)
	}
}

func TestExtract_AnyKeyMap(t *testing.T) {
	source := map[any]any{"k": "v"}
	got, ok := Extract(source, "k")
	if !ok || got != "v" {
		t.Errorf("any-key map lookup failed: got=%v ok=%v", got, ok)
	}
}

func TestExtract_NilSource(t *testing.T) {
	if _, ok := Extract(nil, "x"); ok {
		t.Error("nil source should not match any path")
	}
}
