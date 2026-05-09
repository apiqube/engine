package dataflow

import (
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Extract pulls a value at the given dot-separated path from a nested value.
//
// Supported shapes:
//   - map[string]any: segment is a key
//   - map[any]any: segment is matched as string key
//   - []any / arrays / slices: segment is parsed as a non-negative integer index
//   - any: the pseudo-segment "length" yields len() of slice/array/map/string
//
// An empty path returns the source unchanged. A missing key, out-of-range index,
// or unsupported shape returns (nil, false).
func Extract(source any, path string) (any, bool) {
	if path == "" {
		return source, true
	}

	current := source
	for seg := range strings.SplitSeq(path, ".") {
		if seg == "" {
			return nil, false
		}
		if seg == "length" {
			n, ok := lengthOf(current)
			if !ok {
				return nil, false
			}
			current = n
			continue
		}

		next, ok := step(current, seg)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func step(current any, seg string) (any, bool) {
	switch v := current.(type) {
	case nil:
		return nil, false
	case map[string]any:
		val, ok := v[seg]
		return val, ok
	case map[any]any:
		val, ok := v[seg]
		return val, ok
	case []any:
		i, err := strconv.Atoi(seg)
		if err != nil || i < 0 || i >= len(v) {
			return nil, false
		}
		return v[i], true
	}

	rv := reflect.ValueOf(current)
	if !rv.IsValid() {
		return nil, false
	}
	switch rv.Kind() {
	case reflect.Map:
		mv := rv.MapIndex(reflect.ValueOf(seg))
		if !mv.IsValid() {
			return nil, false
		}
		return mv.Interface(), true
	case reflect.Slice, reflect.Array:
		i, err := strconv.Atoi(seg)
		if err != nil || i < 0 || i >= rv.Len() {
			return nil, false
		}
		return rv.Index(i).Interface(), true
	}
	return nil, false
}

func lengthOf(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	if s, ok := v.(string); ok {
		return utf8.RuneCountInString(s), true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len(), true
	}
	return 0, false
}
