package template

import (
	"reflect"
	"testing"
)

func TestApplyMethod_Unknown(t *testing.T) {
	got := applyMethod("Nope", "input", nil)
	if got != "input" {
		t.Errorf("unknown method should pass through: got %v", got)
	}
}

func TestMethodCases(t *testing.T) {
	cases := []struct {
		name   string
		method string
		input  any
		args   []string
		want   any
	}{
		{"upper", "ToUpper", "hello", nil, "HELLO"},
		{"lower", "ToLower", "HELLO", nil, "hello"},
		{"trim", "TrimSpace", "  abc  ", nil, "abc"},
		{"replace", "Replace", "a-b-c", []string{"-", "_"}, "a_b_c"},
		{"replace not enough args", "Replace", "abc", []string{"x"}, "abc"},
		{"capitalize empty", "Capitalize", "", nil, ""},
		{"capitalize", "Capitalize", "hello", nil, "Hello"},
		{"capitalize mixed", "Capitalize", "hELlO", nil, "Hello"},
		{"reverse ascii", "Reverse", "hello", nil, "olleh"},
		{"reverse unicode", "Reverse", "Привет", nil, "тевирП"},
		{"snakeCase camelCase", "SnakeCase", "fooBarBaz", nil, "foo_bar_baz"},
		{"snakeCase mixed", "SnakeCase", "Foo Bar-baz", nil, "foo_bar_baz"},
		{"camelCase from snake", "CamelCase", "foo_bar_baz", nil, "fooBarBaz"},
		{"camelCase from kebab", "CamelCase", "foo-bar", nil, "fooBar"},
		{"split default", "Split", "a b c", nil, []string{"a", "b", "c"}},
		{"split sep", "Split", "a-b-c", []string{"-"}, []string{"a", "b", "c"}},
		{"join strings", "Join", []string{"a", "b", "c"}, []string{"-"}, "a-b-c"},
		{"join []any", "Join", []any{"a", 2, "c"}, []string{":"}, "a:2:c"},
		{"join no input", "Join", "not-a-slice", []string{":"}, "not-a-slice"},
		{"toInt", "ToInt", "42", nil, int64(42)},
		{"toInt bad", "ToInt", "x", nil, "x"},
		{"toFloat", "ToFloat", "3.14", nil, 3.14},
		{"toFloat bad", "ToFloat", "x", nil, "x"},
		{"toBool true", "ToBool", "true", nil, true},
		{"toBool from bool", "ToBool", false, nil, false},
		{"toBool bad", "ToBool", "x", nil, "x"},
		{"toString from int", "ToString", 42, nil, "42"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := applyMethod(c.method, c.input, c.args)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s(%v, %v) = %v; want %v", c.method, c.input, c.args, got, c.want)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	cases := []struct {
		in   any
		args []string
		want any
	}{
		{"abc", []string{"5"}, "  abc"},
		{"abc", []string{"5", "0"}, "00abc"},
		{"abc", []string{"3"}, "abc"},
		{"abc", nil, "abc"}, // not enough args → input
	}
	for _, c := range cases {
		t.Run("padLeft", func(t *testing.T) {
			got := applyMethod("PadLeft", c.in, c.args)
			if got != c.want {
				t.Errorf("PadLeft(%v, %v) = %v; want %v", c.in, c.args, got, c.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	cases := []struct {
		in   any
		args []string
		want any
	}{
		{"abc", []string{"5"}, "abc  "},
		{"abc", []string{"5", "x"}, "abcxx"},
		{"abc", []string{"3"}, "abc"},
		{"abc", nil, "abc"},
	}
	for _, c := range cases {
		t.Run("padRight", func(t *testing.T) {
			got := applyMethod("PadRight", c.in, c.args)
			if got != c.want {
				t.Errorf("PadRight(%v, %v) = %v; want %v", c.in, c.args, got, c.want)
			}
		})
	}
}

func TestAsString_Boolean(t *testing.T) {
	if s, ok := asString(true); !ok || s != "true" {
		t.Errorf("asString(true) = (%q, %v)", s, ok)
	}
	if _, ok := asString(nil); ok {
		t.Error("asString(nil) should fail")
	}
}
