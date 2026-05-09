package template

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// methodFn applies a chained method to a value with optional arguments.
type methodFn func(input any, args []string) (any, bool)

var methods = map[string]methodFn{
	"ToUpper":    methodToUpper,
	"ToLower":    methodToLower,
	"TrimSpace":  methodTrimSpace,
	"Replace":    methodReplace,
	"PadLeft":    methodPadLeft,
	"PadRight":   methodPadRight,
	"Capitalize": methodCapitalize,
	"Reverse":    methodReverse,
	"SnakeCase":  methodSnakeCase,
	"CamelCase":  methodCamelCase,
	"Split":      methodSplit,
	"Join":       methodJoin,
	"ToInt":      methodToInt,
	"ToFloat":    methodToFloat,
	"ToBool":     methodToBool,
	"ToString":   methodToString,
}

// applyMethod invokes a named method on the current value.
// Lenient: if the method does not exist or fails, returns the input unchanged.
func applyMethod(name string, input any, args []string) any {
	fn, ok := methods[name]
	if !ok {
		return input
	}
	out, ok := fn(input, args)
	if !ok {
		return input
	}
	return out
}

func asString(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", false
	case string:
		return x, true
	case bool:
		return strconv.FormatBool(x), true
	}
	return fmt.Sprintf("%v", v), true
}

func methodToUpper(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return strings.ToUpper(s), true
}

func methodToLower(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return strings.ToLower(s), true
}

func methodTrimSpace(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return strings.TrimSpace(s), true
}

func methodReplace(in any, args []string) (any, bool) {
	if len(args) < 2 {
		return nil, false
	}
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return strings.ReplaceAll(s, args[0], args[1]), true
}

func methodPadLeft(in any, args []string) (any, bool) {
	if len(args) < 1 {
		return nil, false
	}
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n <= len(s) {
		return s, true
	}
	pad := " "
	if len(args) >= 2 && args[1] != "" {
		pad = args[1]
	}
	return strings.Repeat(pad, n-len(s)) + s, true
}

func methodPadRight(in any, args []string) (any, bool) {
	if len(args) < 1 {
		return nil, false
	}
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n <= len(s) {
		return s, true
	}
	pad := " "
	if len(args) >= 2 && args[1] != "" {
		pad = args[1]
	}
	return s + strings.Repeat(pad, n-len(s)), true
}

func methodCapitalize(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	if s == "" {
		return s, true
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r), true
}

func methodReverse(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), true
}

func methodSnakeCase(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return toSnakeCase(s), true
}

func methodCamelCase(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return toCamelCase(s), true
}

func methodSplit(in any, args []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	sep := " "
	if len(args) >= 1 {
		sep = args[0]
	}
	return strings.Split(s, sep), true
}

func methodJoin(in any, args []string) (any, bool) {
	parts, ok := in.([]string)
	if !ok {
		// Try []any → coerce each to string.
		generic, ok2 := in.([]any)
		if !ok2 {
			return nil, false
		}
		parts = make([]string, len(generic))
		for i, v := range generic {
			s, _ := asString(v)
			parts[i] = s
		}
	}
	sep := ""
	if len(args) >= 1 {
		sep = args[0]
	}
	return strings.Join(parts, sep), true
}

func methodToInt(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return nil, false
	}
	return v, true
}

func methodToFloat(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil, false
	}
	return v, true
}

func methodToBool(in any, _ []string) (any, bool) {
	if b, ok := in.(bool); ok {
		return b, true
	}
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	v, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return nil, false
	}
	return v, true
}

func methodToString(in any, _ []string) (any, bool) {
	s, ok := asString(in)
	if !ok {
		return nil, false
	}
	return s, true
}

func toSnakeCase(s string) string {
	var b strings.Builder
	prevLower := false
	for _, r := range s {
		switch {
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('_')
			prevLower = false
		case unicode.IsUpper(r):
			if prevLower {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			prevLower = false
		default:
			b.WriteRune(r)
			prevLower = unicode.IsLower(r) || unicode.IsDigit(r)
		}
	}
	return b.String()
}

func toCamelCase(s string) string {
	var b strings.Builder
	upperNext := false
	for i, r := range s {
		switch r {
		case ' ', '-', '_':
			upperNext = true
		default:
			switch {
			case upperNext:
				b.WriteRune(unicode.ToUpper(r))
				upperNext = false
			case i == 0:
				b.WriteRune(unicode.ToLower(r))
			default:
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
