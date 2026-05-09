package assertion

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type operatorFn func(actual, expected any) (bool, string)

var operators = map[string]operatorFn{
	"eq":        opEq,
	"ne":        opNe,
	"gt":        opGt,
	"gte":       opGte,
	"lt":        opLt,
	"lte":       opLte,
	"contains":  opContains,
	"matches":   opMatches,
	"exists":    opExists,
	"notExists": opNotExists,
	"oneOf":     opOneOf,
	"is":        opIs,
}

// parseOperator splits a string assertion like "> 18" into operator name and
// value. If no recognized prefix is present, returns ("eq", expression).
func parseOperator(expression string) (op, value string) {
	s := strings.TrimSpace(expression)

	// Two-character prefixes first.
	for _, p := range []struct {
		prefix string
		op     string
	}{
		{">=", "gte"},
		{"<=", "lte"},
		{"==", "eq"},
		{"!=", "ne"},
	} {
		if strings.HasPrefix(s, p.prefix) {
			return p.op, strings.TrimSpace(s[len(p.prefix):])
		}
	}

	// One-character prefixes.
	for _, p := range []struct {
		prefix string
		op     string
	}{
		{">", "gt"},
		{"<", "lt"},
	} {
		if strings.HasPrefix(s, p.prefix) {
			return p.op, strings.TrimSpace(s[len(p.prefix):])
		}
	}

	// Word prefixes that take an argument.
	for _, p := range []string{"contains", "matches", "is"} {
		if rest, ok := wordPrefix(s, p); ok {
			return p, rest
		}
	}

	// Bare keywords.
	if s == "exists" {
		return "exists", ""
	}

	return "eq", s
}

// wordPrefix returns (rest, true) if s starts with word followed by a space.
func wordPrefix(s, word string) (string, bool) {
	if len(s) <= len(word) || !strings.HasPrefix(s, word) {
		return "", false
	}
	if s[len(word)] != ' ' && s[len(word)] != '\t' {
		return "", false
	}
	return strings.TrimSpace(s[len(word):]), true
}

// operatorFromMap inspects a map for a single recognized operator key.
// Returns (opName, opValue, true) if exactly one operator key is present.
func operatorFromMap(m map[string]any) (op string, value any, ok bool) {
	if len(m) != 1 {
		return "", nil, false
	}
	for k, v := range m {
		if _, isOp := operators[k]; isOp {
			return k, v, true
		}
	}
	return "", nil, false
}

// runOperator dispatches to the named operator. An unknown name fails.
func runOperator(op string, actual, expected any) (bool, string) {
	fn, ok := operators[op]
	if !ok {
		return false, fmt.Sprintf("unknown operator %q", op)
	}
	return fn(actual, expected)
}

// toFloat coerces a value to float64 if possible.
func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case bool:
		return 0, false
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

// toString coerces a value to a string representation suitable for
// substring/regex operations. Booleans and nil are handled explicitly.
func toString(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", false
	case string:
		return x, true
	case bool:
		return strconv.FormatBool(x), true
	case json.Number:
		return x.String(), true
	}
	if f, ok := toFloat(v); ok {
		return strconv.FormatFloat(f, 'g', -1, 64), true
	}
	return "", false
}

// equal applies the type-coercion rules from doc.go to test value equality.
func equal(actual, expected any) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}

	// Both bool: direct compare.
	if ab, aok := actual.(bool); aok {
		if eb, eok := expected.(bool); eok {
			return ab == eb
		}
		return false
	}
	if _, eok := expected.(bool); eok {
		return false
	}

	// Both numeric or one numeric and one numeric-string.
	af, aOk := toFloat(actual)
	ef, eOk := toFloat(expected)
	if aOk && eOk {
		return af == ef
	}

	// Both strings.
	if as, aok := actual.(string); aok {
		if es, eok := expected.(string); eok {
			return as == es
		}
	}

	// Fall back to deep equal for slices / maps / structs.
	return reflect.DeepEqual(actual, expected)
}

func opEq(actual, expected any) (bool, string) {
	if equal(actual, expected) {
		return true, ""
	}
	return false, fmt.Sprintf("expected %v, got %v", expected, actual)
}

func opNe(actual, expected any) (bool, string) {
	if !equal(actual, expected) {
		return true, ""
	}
	return false, fmt.Sprintf("expected != %v, got %v", expected, actual)
}

func compareNumeric(actual, expected any) (af, ef float64, ok bool, msg string) {
	af, aOk := toFloat(actual)
	ef, eOk := toFloat(expected)
	if !aOk || !eOk {
		return 0, 0, false, fmt.Sprintf("not numeric: actual=%v expected=%v", actual, expected)
	}
	return af, ef, true, ""
}

func opGt(actual, expected any) (bool, string) {
	af, ef, ok, msg := compareNumeric(actual, expected)
	if !ok {
		return false, msg
	}
	if af > ef {
		return true, ""
	}
	return false, fmt.Sprintf("%v not > %v", af, ef)
}

func opGte(actual, expected any) (bool, string) {
	af, ef, ok, msg := compareNumeric(actual, expected)
	if !ok {
		return false, msg
	}
	if af >= ef {
		return true, ""
	}
	return false, fmt.Sprintf("%v not >= %v", af, ef)
}

func opLt(actual, expected any) (bool, string) {
	af, ef, ok, msg := compareNumeric(actual, expected)
	if !ok {
		return false, msg
	}
	if af < ef {
		return true, ""
	}
	return false, fmt.Sprintf("%v not < %v", af, ef)
}

func opLte(actual, expected any) (bool, string) {
	af, ef, ok, msg := compareNumeric(actual, expected)
	if !ok {
		return false, msg
	}
	if af <= ef {
		return true, ""
	}
	return false, fmt.Sprintf("%v not <= %v", af, ef)
}

func opContains(actual, expected any) (bool, string) {
	// Slice membership.
	if rv := reflect.ValueOf(actual); rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		for i := 0; i < rv.Len(); i++ {
			if equal(rv.Index(i).Interface(), expected) {
				return true, ""
			}
		}
		return false, fmt.Sprintf("array does not contain %v", expected)
	}

	// String substring.
	as, aOk := toString(actual)
	es, eOk := toString(expected)
	if !aOk || !eOk {
		return false, fmt.Sprintf("contains: not stringifiable: actual=%v expected=%v", actual, expected)
	}
	if strings.Contains(as, es) {
		return true, ""
	}
	return false, fmt.Sprintf("%q does not contain %q", as, es)
}

func opMatches(actual, expected any) (bool, string) {
	as, aOk := toString(actual)
	pattern, eOk := toString(expected)
	if !aOk || !eOk {
		return false, fmt.Sprintf("matches: not stringifiable: actual=%v pattern=%v", actual, expected)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Sprintf("invalid regex %q: %v", pattern, err)
	}
	if re.MatchString(as) {
		return true, ""
	}
	return false, fmt.Sprintf("%q does not match /%s/", as, pattern)
}

func opExists(actual, _ any) (bool, string) {
	if actual == nil {
		return false, "value does not exist"
	}
	return true, ""
}

func opNotExists(actual, _ any) (bool, string) {
	if actual == nil {
		return true, ""
	}
	return false, fmt.Sprintf("value unexpectedly exists: %v", actual)
}

func opOneOf(actual, expected any) (bool, string) {
	rv := reflect.ValueOf(expected)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return false, fmt.Sprintf("oneOf: expected an array, got %T", expected)
	}
	for i := 0; i < rv.Len(); i++ {
		if equal(actual, rv.Index(i).Interface()) {
			return true, ""
		}
	}
	return false, fmt.Sprintf("%v not in %v", actual, expected)
}

func opIs(actual, expected any) (bool, string) {
	want, ok := toString(expected)
	if !ok {
		return false, fmt.Sprintf("is: type name must be a string, got %T", expected)
	}
	want = strings.ToLower(strings.TrimSpace(want))

	switch want {
	case "null", "nil":
		return actual == nil, fail("null", actual)
	case "bool", "boolean":
		_, ok := actual.(bool)
		return ok, fail("bool", actual)
	case "int", "integer":
		f, ok := toFloat(actual)
		if !ok {
			return false, fail("integer", actual)
		}
		if f != float64(int64(f)) {
			return false, fail("integer", actual)
		}
		return true, ""
	case "number", "float":
		_, ok := toFloat(actual)
		return ok, fail("number", actual)
	case "string":
		_, ok := actual.(string)
		return ok, fail("string", actual)
	case "array", "list", "slice":
		rv := reflect.ValueOf(actual)
		if !rv.IsValid() {
			return false, fail("array", actual)
		}
		ok := rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
		return ok, fail("array", actual)
	case "object", "map":
		rv := reflect.ValueOf(actual)
		if !rv.IsValid() {
			return false, fail("object", actual)
		}
		return rv.Kind() == reflect.Map, fail("object", actual)
	}
	return false, fmt.Sprintf("is: unknown type %q", want)
}

func fail(want string, actual any) string {
	if actual == nil {
		return fmt.Sprintf("expected %s, got nil", want)
	}
	return fmt.Sprintf("expected %s, got %T", want, actual)
}
