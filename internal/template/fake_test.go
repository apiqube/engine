package template

import (
	"strings"
	"testing"
)

func TestApplyGenerator_Unknown(t *testing.T) {
	if v, ok := applyGenerator("nope", nil); ok {
		t.Errorf("unknown generator should fail, got %v", v)
	}
}

func TestGenerators_Smoke(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"name", nil},
		{"email", nil},
		{"uuid", nil},
		{"phone", nil},
		{"url", nil},
		{"company", nil},
		{"city", nil},
		{"country", nil},
		{"color", nil},
		{"word", nil},
		{"sentence", []string{"5"}},
		{"address", nil},
		{"password", []string{"16"}},
		{"bool", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v, ok := applyGenerator(c.name, c.args)
			if !ok {
				t.Fatalf("%s returned not-ok", c.name)
			}
			if v == nil {
				t.Errorf("%s returned nil", c.name)
			}
		})
	}
}

func TestEmail_ContainsAt(t *testing.T) {
	v, _ := applyGenerator("email", nil)
	s, _ := v.(string)
	if !strings.Contains(s, "@") {
		t.Errorf("email lacks @: %q", s)
	}
}

func TestIntRange(t *testing.T) {
	v, ok := applyGenerator("int", []string{"10", "20"})
	if !ok {
		t.Fatal("int range failed")
	}
	n, ok := v.(int)
	if !ok {
		t.Fatalf("int returned %T", v)
	}
	if n < 10 || n > 20 {
		t.Errorf("out of range: %d", n)
	}
}

func TestIntRange_Reversed(t *testing.T) {
	// Min/max reversed should still produce a value in range.
	v, _ := applyGenerator("int", []string{"50", "10"})
	n := v.(int)
	if n < 10 || n > 50 {
		t.Errorf("out of range: %d", n)
	}
}

func TestUintRange(t *testing.T) {
	v, _ := applyGenerator("uint", []string{"1", "5"})
	n := v.(uint)
	if n < 1 || n > 5 {
		t.Errorf("out of range: %d", n)
	}
}

func TestFloatRange(t *testing.T) {
	v, _ := applyGenerator("float", []string{"0.5", "1.5"})
	f := v.(float64)
	if f < 0.5 || f > 1.5 {
		t.Errorf("out of range: %f", f)
	}
}

func TestDateGenerator(t *testing.T) {
	v, ok := applyGenerator("date", nil)
	if !ok {
		t.Fatal("date generator failed")
	}
	s := v.(string)
	if len(s) < 4 {
		t.Errorf("unexpected date: %q", s)
	}
}

func TestRegex(t *testing.T) {
	v, ok := generateRegex("^[a-z]{5}$")
	if !ok {
		t.Fatal("regex failed")
	}
	s := v.(string)
	if len(s) != 5 {
		t.Errorf("expected 5 chars, got %q", s)
	}
}

func TestRegex_Empty(t *testing.T) {
	if _, ok := generateRegex(""); ok {
		t.Error("empty pattern should fail")
	}
}

func TestArgHelpers(t *testing.T) {
	if got := intArg([]string{"10"}, 0, 5); got != 10 {
		t.Errorf("intArg ok: got %d", got)
	}
	if got := intArg([]string{"x"}, 0, 5); got != 5 {
		t.Errorf("intArg fallback: got %d", got)
	}
	if got := intArg(nil, 0, 5); got != 5 {
		t.Errorf("intArg empty: got %d", got)
	}
	if got := uintArg([]string{"x"}, 0, 5); got != 5 {
		t.Errorf("uintArg fallback: got %d", got)
	}
	if got := floatArg([]string{"x"}, 0, 5.5); got != 5.5 {
		t.Errorf("floatArg fallback: got %f", got)
	}
}
