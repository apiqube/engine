package graph

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/apiqube/engine/internal/manifest"
)

// templateBodyRegex captures the body of one {{ ... }} expression.
var templateBodyRegex = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

// methodTailRegex strips a trailing .Name(...) chain so dependency analysis
// looks at the primary expression only.
var methodTailRegex = regexp.MustCompile(`\.[A-Z][a-zA-Z0-9]*\([^)]*\)$`)

// Reference is a parsed template reference found in a test field.
type Reference struct {
	Name string
	Path string
}

// extractReferences walks any value (string, map, slice) and collects every
// `{{ alias.path... }}` reference. Generator and env-var expressions are
// excluded because they do not create cross-test dependencies.
func extractReferences(value any) []Reference {
	seen := make(map[Reference]bool)
	var refs []Reference

	walkStrings(value, func(s string) {
		for _, m := range templateBodyRegex.FindAllStringSubmatch(s, -1) {
			ref, ok := parseRef(m[1])
			if !ok {
				continue
			}
			if !seen[ref] {
				seen[ref] = true
				refs = append(refs, ref)
			}
		}
	})
	return refs
}

// extractReferencesFromTest walks a TestCase across all its string-bearing
// fields and returns every cross-test reference found.
func extractReferencesFromTest(tc *manifest.TestCase) []Reference {
	if tc == nil {
		return nil
	}
	collected := []any{
		tc.Method, tc.Resource, tc.Target, tc.Timeout, tc.When,
		tc.Headers, tc.Save, tc.Expect, tc.Extra, tc.Matrix,
	}
	if tc.Retry != nil {
		collected = append(collected, tc.Retry.Until, tc.Retry.Interval)
	}
	return extractReferences(collected)
}

func parseRef(body string) (Reference, bool) {
	body = strings.TrimSpace(body)

	// Strip trailing method chain.
	for {
		stripped := methodTailRegex.ReplaceAllString(body, "")
		if stripped == body {
			break
		}
		body = strings.TrimRightFunc(stripped, isSpace)
	}

	// Skip regex(...).
	if strings.HasPrefix(body, "regex(") {
		return Reference{}, false
	}

	parts := strings.SplitN(body, ".", 2)
	if parts[0] == "" {
		return Reference{}, false
	}

	name := parts[0]
	switch name {
	case "fake", "env":
		return Reference{}, false
	}

	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}
	return Reference{Name: name, Path: path}, true
}

// walkStrings invokes fn on every string found in v, recursing through maps
// and slices/arrays. Other types are ignored.
func walkStrings(v any, fn func(string)) {
	if v == nil {
		return
	}
	switch x := v.(type) {
	case string:
		fn(x)
		return
	case map[string]any:
		for _, val := range x {
			walkStrings(val, fn)
		}
		return
	case map[string]string:
		for _, val := range x {
			fn(val)
		}
		return
	case []any:
		for _, val := range x {
			walkStrings(val, fn)
		}
		return
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return
	}
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !rv.IsNil() {
			walkStrings(rv.Elem().Interface(), fn)
		}
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			walkStrings(iter.Value().Interface(), fn)
		}
	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			walkStrings(rv.Index(i).Interface(), fn)
		}
	case reflect.Struct:
		for i := range rv.NumField() {
			f := rv.Field(i)
			if !f.CanInterface() {
				continue
			}
			walkStrings(f.Interface(), fn)
		}
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
