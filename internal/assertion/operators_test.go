package assertion

import "testing"

func TestParseOperator(t *testing.T) {
	tests := []struct {
		in        string
		wantOp    string
		wantValue string
	}{
		{"200", "eq", "200"},
		{">= 18", "gte", "18"},
		{"<=18", "lte", "18"},
		{"> 0", "gt", "0"},
		{"<3.14", "lt", "3.14"},
		{"== abc", "eq", "abc"},
		{"!= abc", "ne", "abc"},
		{"contains @", "contains", "@"},
		{"matches ^foo$", "matches", "^foo$"},
		{"is integer", "is", "integer"},
		{"exists", "exists", ""},
		{"  spaced  ", "eq", "spaced"},
		{"", "eq", ""},
		{"contains", "eq", "contains"}, // bare word, no value → not parsed
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			op, val := parseOperator(tt.in)
			if op != tt.wantOp || val != tt.wantValue {
				t.Errorf("parseOperator(%q) = (%q, %q); want (%q, %q)",
					tt.in, op, val, tt.wantOp, tt.wantValue)
			}
		})
	}
}

func TestOperatorFromMap(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]any
		wantOp  string
		wantVal any
		wantOK  bool
	}{
		{"single op", map[string]any{"gt": 18}, "gt", 18, true},
		{"oneOf", map[string]any{"oneOf": []any{200, 201}}, "oneOf", []any{200, 201}, true},
		{"plain map", map[string]any{"name": "alice"}, "", nil, false},
		{"two keys", map[string]any{"gt": 18, "lt": 65}, "", nil, false},
		{"empty", map[string]any{}, "", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, val, ok := operatorFromMap(tt.in)
			if ok != tt.wantOK || op != tt.wantOp {
				t.Errorf("operatorFromMap = (%q, %v, %v); want (%q, %v, %v)",
					op, val, ok, tt.wantOp, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name           string
		actual, expect any
		want           bool
	}{
		{"int int eq", 200, 200, true},
		{"int int ne", 200, 201, false},
		{"int float eq", 200, 200.0, true},
		{"int string eq via coercion", 200, "200", true},
		{"string int eq via coercion", "200", 200, true},
		{"string string eq", "alice", "alice", true},
		{"bool bool eq", true, true, true},
		{"bool bool ne", true, false, false},
		{"bool int rejected", true, 1, false},
		{"int bool rejected", 1, true, false},
		{"nil nil eq", nil, nil, true},
		{"nil int ne", nil, 0, false},
		{"slice slice eq", []any{1, 2}, []any{1, 2}, true},
		{"slice slice ne", []any{1, 2}, []any{1, 3}, false},
		{"map map eq", map[string]any{"a": 1}, map[string]any{"a": 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equal(tt.actual, tt.expect); got != tt.want {
				t.Errorf("equal(%v, %v) = %v; want %v", tt.actual, tt.expect, got, tt.want)
			}
		})
	}
}

func TestNumericOperators(t *testing.T) {
	tests := []struct {
		op             string
		actual, expect any
		want           bool
	}{
		{"gt", 10, 5, true},
		{"gt", 5, 10, false},
		{"gt", 5, 5, false},
		{"gte", 5, 5, true},
		{"gte", 4, 5, false},
		{"lt", 5, 10, true},
		{"lt", 10, 5, false},
		{"lte", 5, 5, true},
		{"lte", 6, 5, false},
		{"gt", 200, "100", true},   // numeric coercion from string
		{"gt", "200", 100, true},
		{"gt", 200, "abc", false},  // unparseable → fail
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			passed, _ := runOperator(tt.op, tt.actual, tt.expect)
			if passed != tt.want {
				t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.actual, tt.expect, passed, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name           string
		actual, expect any
		want           bool
	}{
		{"string substring hit", "hello@world", "@", true},
		{"string substring miss", "hello", "@", false},
		{"slice membership hit", []any{"a", "b", "c"}, "b", true},
		{"slice membership miss", []any{"a", "b"}, "z", false},
		{"int in array", []any{1, 2, 3}, 2, true},
		{"int in array via coercion", []any{1, 2, 3}, "2", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, _ := opContains(tt.actual, tt.expect)
			if passed != tt.want {
				t.Errorf("contains(%v, %v) = %v; want %v", tt.actual, tt.expect, passed, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name           string
		actual, expect any
		want           bool
	}{
		{"hit", "alice@example.com", `^[a-z]+@`, true},
		{"miss", "ALICE", `^[a-z]+$`, false},
		{"non-string actual coerced", 12345, `^\d+$`, true},
		{"invalid regex", "x", `[`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, _ := opMatches(tt.actual, tt.expect)
			if passed != tt.want {
				t.Errorf("matches(%v, %v) = %v; want %v", tt.actual, tt.expect, passed, tt.want)
			}
		})
	}
}

func TestExistsAndNotExists(t *testing.T) {
	if passed, _ := opExists("anything", nil); !passed {
		t.Error("exists on non-nil should pass")
	}
	if passed, _ := opExists(nil, nil); passed {
		t.Error("exists on nil should fail")
	}
	if passed, _ := opNotExists(nil, nil); !passed {
		t.Error("notExists on nil should pass")
	}
	if passed, _ := opNotExists("anything", nil); passed {
		t.Error("notExists on non-nil should fail")
	}
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name           string
		actual, expect any
		want           bool
	}{
		{"hit int", 200, []any{200, 201}, true},
		{"miss int", 500, []any{200, 201}, false},
		{"hit via coercion", "200", []any{200, 201}, true},
		{"non-array expected", 200, "200", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, _ := opOneOf(tt.actual, tt.expect)
			if passed != tt.want {
				t.Errorf("oneOf(%v, %v) = %v; want %v", tt.actual, tt.expect, passed, tt.want)
			}
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name   string
		actual any
		typeOf string
		want   bool
	}{
		{"null nil", nil, "null", true},
		{"null non-nil", 0, "null", false},
		{"bool true", true, "bool", true},
		{"bool int", 1, "bool", false},
		{"integer int", 42, "integer", true},
		{"integer float-int", 42.0, "integer", true},
		{"integer float-fract", 42.5, "integer", false},
		{"integer string", "42", "integer", true},
		{"number float", 3.14, "number", true},
		{"number bool", true, "number", false},
		{"string string", "x", "string", true},
		{"string int", 1, "string", false},
		{"array slice", []any{1}, "array", true},
		{"array map", map[string]any{}, "array", false},
		{"object map", map[string]any{"a": 1}, "object", true},
		{"object slice", []any{}, "object", false},
		{"unknown type", 1, "duck", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, _ := opIs(tt.actual, tt.typeOf)
			if passed != tt.want {
				t.Errorf("is(%v, %q) = %v; want %v", tt.actual, tt.typeOf, passed, tt.want)
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	if passed, _ := opNe(1, 2); !passed {
		t.Error("ne(1, 2) should pass")
	}
	if passed, _ := opNe(1, 1); passed {
		t.Error("ne(1, 1) should fail")
	}
}

func TestToFloat_AllNumericKinds(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want float64
		ok   bool
	}{
		{"int", int(1), 1, true},
		{"int8", int8(2), 2, true},
		{"int16", int16(3), 3, true},
		{"int32", int32(4), 4, true},
		{"int64", int64(5), 5, true},
		{"uint", uint(6), 6, true},
		{"uint8", uint8(7), 7, true},
		{"uint16", uint16(8), 8, true},
		{"uint32", uint32(9), 9, true},
		{"uint64", uint64(10), 10, true},
		{"float32", float32(3.5), 3.5, true},
		{"float64", float64(2.5), 2.5, true},
		{"string-numeric", "42.5", 42.5, true},
		{"string-non-numeric", "abc", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
		{"struct", struct{}{}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toFloat(tc.in)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Errorf("toFloat(%v) = (%v, %v); want (%v, %v)", tc.in, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestToString_AllKinds(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
		ok   bool
	}{
		{"string", "hello", "hello", true},
		{"int", 42, "42", true},
		{"float", 3.5, "3.5", true},
		{"bool true", true, "true", true},
		{"bool false", false, "false", true},
		{"nil", nil, "", false},
		{"map (unstringifiable)", map[string]any{}, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toString(tc.in)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Errorf("toString(%v) = (%q, %v); want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestNumericOperators_Failures(t *testing.T) {
	// Cover the failing branches of gte/lt/lte for completeness.
	cases := []struct {
		op             string
		actual, expect any
		want           bool
	}{
		{"gte", 4, 5, false},
		{"lt", 10, 5, false},
		{"lte", 6, 5, false},
	}
	for _, tt := range cases {
		t.Run(tt.op+" fail", func(t *testing.T) {
			passed, msg := runOperator(tt.op, tt.actual, tt.expect)
			if passed != tt.want {
				t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.actual, tt.expect, passed, tt.want)
			}
			if !passed && msg == "" {
				t.Errorf("%s should produce failure message", tt.op)
			}
		})
	}
}

func TestUnknownOperator(t *testing.T) {
	passed, msg := runOperator("nosuch", 1, 1)
	if passed {
		t.Error("unknown operator should not pass")
	}
	if msg == "" {
		t.Error("unknown operator should produce a message")
	}
}
