package template

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolve_NoTemplate(t *testing.T) {
	r := New(nil, nil)
	got, err := r.Resolve("plain string")
	if err != nil {
		t.Fatal(err)
	}
	if got != "plain string" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_LookupReference(t *testing.T) {
	r := New(func(name, path string) (any, bool) {
		if name == "user" && path == "name" {
			return "alice", true
		}
		return nil, false
	}, nil)

	got, _ := r.Resolve("Hi {{ user.name }}!")
	if got != "Hi alice!" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_BareReference(t *testing.T) {
	r := New(func(name, path string) (any, bool) {
		if name == "token" && path == "" {
			return "abc123", true
		}
		return nil, false
	}, nil)

	got, _ := r.Resolve("Bearer {{ token }}")
	if got != "Bearer abc123" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_LookupMiss_LeavesExpression(t *testing.T) {
	r := New(func(string, string) (any, bool) { return nil, false }, nil)

	src := "x = {{ unknown.thing }}"
	got, _ := r.Resolve(src)
	if got != src {
		t.Errorf("missing reference should be left in place; got %q", got)
	}
}

func TestResolve_EnvVar(t *testing.T) {
	r := New(nil, map[string]string{"API_HOST": "example.com"})
	got, _ := r.Resolve("https://{{ env.API_HOST }}/")
	if got != "https://example.com/" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_FakeGenerator(t *testing.T) {
	r := New(nil, nil)
	got, _ := r.Resolve("{{ fake.email }}")
	if !strings.Contains(got, "@") {
		t.Errorf("fake.email should contain @, got %q", got)
	}
}

func TestResolve_FakeWithArgs(t *testing.T) {
	r := New(nil, nil)
	got, _ := r.Resolve("{{ fake.int.10.20 }}")
	// Result is a stringified int in [10, 20].
	if got == "" {
		t.Fatal("empty result")
	}
}

func TestResolve_Regex(t *testing.T) {
	r := New(nil, nil)
	got, _ := r.Resolve(`{{ regex('^[a-z]{3}$') }}`)
	if len(got) != 3 {
		t.Errorf("regex result should be 3 chars, got %q", got)
	}
}

func TestResolve_MethodChain(t *testing.T) {
	r := New(func(string, string) (any, bool) { return "  Hello  ", true }, nil)
	got, _ := r.Resolve("{{ greeting.TrimSpace().ToLower() }}")
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_MethodWithArgs(t *testing.T) {
	r := New(func(string, string) (any, bool) { return "alice@old.com", true }, nil)
	got, _ := r.Resolve("{{ email.Replace('@old.com', '@new.com') }}")
	if got != "alice@new.com" {
		t.Errorf("got %q", got)
	}
}

func TestResolveValue_Map(t *testing.T) {
	r := New(func(name, path string) (any, bool) {
		if name == "user" && path == "id" {
			return 42, true
		}
		return nil, false
	}, nil)

	in := map[string]any{
		"name":   "alice",
		"userID": "{{ user.id }}",
		"nested": map[string]any{
			"path": "{{ user.id }}",
		},
		"list": []any{"plain", "{{ user.id }}"},
		"int":  100,
	}
	got, err := r.ResolveValue(in)
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if out["userID"] != "42" {
		t.Errorf("userID resolved wrong: %v", out["userID"])
	}
	if nested := out["nested"].(map[string]any); nested["path"] != "42" {
		t.Errorf("nested resolved wrong: %v", nested)
	}
	if list := out["list"].([]any); list[1] != "42" {
		t.Errorf("list resolved wrong: %v", list)
	}
	if out["int"] != 100 {
		t.Errorf("non-string passed through wrong: %v", out["int"])
	}
}

func TestResolveValue_Nil(t *testing.T) {
	r := New(nil, nil)
	got, err := r.ResolveValue(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("nil should pass through, got %v", got)
	}
}

func TestSplitPrimaryAndMethods(t *testing.T) {
	cases := []struct {
		in            string
		wantPrimary   string
		wantMethods   []string
	}{
		{"fake.email", "fake.email", nil},
		{"fake.email.ToLower()", "fake.email", []string{"ToLower"}},
		{"x.Replace('a', 'b').ToUpper()", "x", []string{"Replace", "ToUpper"}},
		{"alias.path", "alias.path", nil},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			primary, methods := splitPrimaryAndMethods(c.in)
			if primary != c.wantPrimary {
				t.Errorf("primary = %q, want %q", primary, c.wantPrimary)
			}
			if len(methods) != len(c.wantMethods) {
				t.Errorf("got %d methods, want %d", len(methods), len(c.wantMethods))
			}
			for i, name := range c.wantMethods {
				if methods[i].name != name {
					t.Errorf("method[%d] = %q, want %q", i, methods[i].name, name)
				}
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"'hello'", []string{"hello"}},
		{`"hello"`, []string{"hello"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"'a, b', c", []string{"a, b", "c"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := parseArgs(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("parseArgs(%q) = %v; want %v", c.in, got, c.want)
			}
		})
	}
}
