package assertion

import "testing"

func TestEngineCheck_Primitive(t *testing.T) {
	e := New()

	r := e.Check("status", 200, 200)
	if !r.Passed {
		t.Errorf("expected pass for primitive equality, got %+v", r)
	}

	r = e.Check("status", 201, 200)
	if r.Passed {
		t.Errorf("expected fail for unequal primitives")
	}
	if r.Message == "" {
		t.Errorf("failure must include message")
	}
}

func TestEngineCheck_StringOperator(t *testing.T) {
	e := New()

	cases := []struct {
		name           string
		expr           string
		actual         any
		want           bool
	}{
		{"gt pass", "> 18", 25, true},
		{"gt fail", "> 18", 17, false},
		{"contains pass", "contains @", "x@y", true},
		{"contains fail", "contains @", "xyz", false},
		{"matches pass", "matches ^\\d+$", "12345", true},
		{"matches fail", "matches ^\\d+$", "12abc", false},
		{"is integer pass", "is integer", 42, true},
		{"is integer fail", "is integer", "abc", false},
		{"exists pass", "exists", "anything", true},
		{"exists fail", "exists", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := e.Check("path", tc.expr, tc.actual)
			if r.Passed != tc.want {
				t.Errorf("Check(%q, %v) passed=%v want=%v msg=%q",
					tc.expr, tc.actual, r.Passed, tc.want, r.Message)
			}
		})
	}
}

func TestEngineCheck_MapOperator(t *testing.T) {
	e := New()

	cases := []struct {
		name           string
		expr           map[string]any
		actual         any
		want           bool
	}{
		{"oneOf pass", map[string]any{"oneOf": []any{200, 201}}, 200, true},
		{"oneOf fail", map[string]any{"oneOf": []any{200, 201}}, 500, false},
		{"gt pass", map[string]any{"gt": 0}, 5, true},
		{"notExists pass", map[string]any{"notExists": true}, nil, true},
		{"notExists fail", map[string]any{"notExists": true}, "x", false},
		{"is string pass", map[string]any{"is": "string"}, "hello", true},
		{"is string fail", map[string]any{"is": "string"}, 42, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := e.Check("p", tc.expr, tc.actual)
			if r.Passed != tc.want {
				t.Errorf("Check(%v, %v) passed=%v want=%v msg=%q",
					tc.expr, tc.actual, r.Passed, tc.want, r.Message)
			}
		})
	}
}

func TestEngineCheck_StructuralEquality(t *testing.T) {
	e := New()

	expected := map[string]any{"name": "alice", "age": 30}
	r := e.Check(".", expected, map[string]any{"name": "alice", "age": 30})
	if !r.Passed {
		t.Errorf("expected pass on equal maps, got %+v", r)
	}

	r = e.Check(".", expected, map[string]any{"name": "alice"})
	if r.Passed {
		t.Errorf("expected fail when keys differ")
	}
}

func TestEngineCheck_TypeCoercionAcrossYAMLForms(t *testing.T) {
	e := New()

	// YAML "200" → string; HTTP response status → int.
	r := e.Check("status", "200", 200)
	if !r.Passed {
		t.Errorf("string-int coercion should pass: %+v", r)
	}

	// Float vs int.
	r = e.Check("amount", 100, 100.0)
	if !r.Passed {
		t.Errorf("int-float coercion should pass: %+v", r)
	}
}

func TestPathExpr_EmptyPath(t *testing.T) {
	// Path-less assertion (used when only a value is involved).
	r := New().Check("", 1, 1)
	if r.Expression == "" {
		t.Errorf("empty path should still produce a non-empty expression")
	}
}

func TestEngineCheck_IsTypePathSelectsBoolBranch(t *testing.T) {
	// is bool branch — exercise the bool path of opIs explicitly.
	r := New().Check("flag", "is bool", true)
	if !r.Passed {
		t.Errorf("is bool on true should pass: %+v", r)
	}
}

func TestEngineCheck_ResultExpectedField(t *testing.T) {
	e := New()

	// Equality fallback: original expected preserved.
	r := e.Check("status", 200, 200)
	if r.Expected != 200 {
		t.Errorf("equality fallback should preserve expected; got %v", r.Expected)
	}

	// Operator path: expected reflects operator value (e.g. parsed string).
	r = e.Check("status", "> 100", 200)
	if r.Expected != "100" {
		t.Errorf("operator form should expose operator value; got %v", r.Expected)
	}
}
