package template

import (
	"fmt"
	"regexp"
	"strings"
)

// Resolver resolves templates in string values using a context of available data.
type Resolver struct {
	Lookup func(name, path string) (any, bool)
	Env    map[string]string
}

// New returns a Resolver with the supplied lookup callback and environment.
func New(lookup func(name, path string) (any, bool), env map[string]string) *Resolver {
	return &Resolver{Lookup: lookup, Env: env}
}

// templateRegex matches one {{ ... }} expression. The expression body capture
// group is non-greedy so adjacent expressions on one line do not merge.
var templateRegex = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

// Resolve replaces every {{ expression }} in s with its evaluated value.
//
// Lenient: an unresolvable expression is left in place. The returned error is
// reserved for catastrophic failures (none today) and is always nil.
func (r *Resolver) Resolve(s string) (string, error) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}
	out := templateRegex.ReplaceAllStringFunc(s, func(match string) string {
		body := templateRegex.FindStringSubmatch(match)[1]
		val, ok := r.evaluate(body)
		if !ok {
			return match
		}
		return stringify(val)
	})
	return out, nil
}

// ResolveValue recursively resolves any value containing strings.
//
// Maps and slices are walked; strings are passed through Resolve. Other types
// pass through unchanged. The original input is not mutated.
func (r *Resolver) ResolveValue(value any) (any, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		return r.Resolve(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, child := range v {
			resolved, err := r.ResolveValue(child)
			if err != nil {
				return nil, err
			}
			out[k] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			resolved, err := r.ResolveValue(child)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	}
	return value, nil
}

// evaluate parses and evaluates one expression body. Returns (value, true) on
// success; (nil, false) means the expression could not be resolved.
func (r *Resolver) evaluate(body string) (any, bool) {
	primary, methods := splitPrimaryAndMethods(body)
	val, ok := r.evalPrimary(primary)
	if !ok {
		return nil, false
	}
	for _, m := range methods {
		val = applyMethod(m.name, val, m.args)
	}
	return val, true
}

func (r *Resolver) evalPrimary(s string) (any, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	// regex('pattern') — pattern is everything between the outermost parens,
	// with optional surrounding quotes.
	if strings.HasPrefix(s, "regex(") && strings.HasSuffix(s, ")") {
		pat := s[len("regex(") : len(s)-1]
		pat = trimQuotes(strings.TrimSpace(pat))
		return generateRegex(pat)
	}

	parts := strings.Split(s, ".")
	if len(parts) == 0 || parts[0] == "" {
		return nil, false
	}
	head := parts[0]
	rest := parts[1:]

	switch head {
	case "fake":
		if len(rest) == 0 {
			return nil, false
		}
		return applyGenerator(rest[0], rest[1:])
	case "env":
		if len(rest) == 0 || r.Env == nil {
			return nil, false
		}
		v, ok := r.Env[strings.Join(rest, ".")]
		return v, ok
	}

	if r.Lookup == nil {
		return nil, false
	}
	return r.Lookup(head, strings.Join(rest, "."))
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '\'' || first == '"') && first == last {
			return s[1 : len(s)-1]
		}
	}
	return s
}
