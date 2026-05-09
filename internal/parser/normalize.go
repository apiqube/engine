package parser

import (
	"fmt"
	"regexp"

	"github.com/goccy/go-yaml"

	"github.com/apiqube/engine/internal/manifest"
)

// oneLinerRE matches `METHOD RESOURCE -> EXPECTED`.
var oneLinerRE = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)\s+(\S+)\s*->\s*(\S+)$`)

// methodResourceRE matches `METHOD RESOURCE` for compact-form keys.
var methodResourceRE = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)\s+(\S+)$`)

// knownTestCaseFields are TestCase fields handled directly by struct unmarshal.
// Anything else encountered in a test entry is captured into TestCase.Extra.
var knownTestCaseFields = map[string]bool{
	"name": true, "alias": true, "target": true,
	"method": true, "resource": true, "tags": true,
	"skip": true, "headers": true, "timeout": true,
	"expect": true, "save": true, "when": true,
	"retry": true, "matrix": true, "depends": true,
}

// normalizeTests converts each raw entry into a manifest.TestCase, handling
// the three syntax levels: one-liner string, compact-form map with
// method-resource key, and full-form map.
func normalizeTests(raw []any) ([]manifest.TestCase, error) {
	out := make([]manifest.TestCase, 0, len(raw))
	for i, item := range raw {
		tc, err := normalizeOne(item)
		if err != nil {
			return nil, fmt.Errorf("test[%d]: %w", i, err)
		}
		out = append(out, tc)
	}
	return out, nil
}

func normalizeOne(item any) (manifest.TestCase, error) {
	switch v := item.(type) {
	case string:
		return parseOneLiner(v)
	case map[string]any:
		// Compact form: a single key matching METHOD RESOURCE.
		if len(v) == 1 {
			for k, val := range v {
				if m := methodResourceRE.FindStringSubmatch(k); m != nil {
					sub, _ := val.(map[string]any)
					if sub == nil {
						sub = map[string]any{}
					}
					return parseFullForm(sub, m[1], m[2])
				}
			}
		}
		return parseFullForm(v, "", "")
	}
	return manifest.TestCase{}, fmt.Errorf("unsupported test entry type %T", item)
}

func parseOneLiner(s string) (manifest.TestCase, error) {
	m := oneLinerRE.FindStringSubmatch(s)
	if m == nil {
		return manifest.TestCase{}, fmt.Errorf("invalid one-liner: %q", s)
	}
	return manifest.TestCase{
		Method:   m[1],
		Resource: m[2],
		Expect:   manifest.Expect{Status: m[3]},
	}, nil
}

func parseFullForm(m map[string]any, method, resource string) (manifest.TestCase, error) {
	known := make(map[string]any, len(m))
	extra := make(map[string]any)

	for k, v := range m {
		if knownTestCaseFields[k] {
			known[k] = v
		} else {
			extra[k] = v
		}
	}

	var tc manifest.TestCase
	if len(known) > 0 {
		data, err := yaml.Marshal(known)
		if err != nil {
			return tc, fmt.Errorf("re-marshal known fields: %w", err)
		}
		if err := yaml.Unmarshal(data, &tc); err != nil {
			return tc, fmt.Errorf("unmarshal test case: %w", err)
		}
	}

	if method != "" {
		tc.Method = method
	}
	if resource != "" {
		tc.Resource = resource
	}

	if len(extra) > 0 {
		tc.Extra = extra
	}
	return tc, nil
}
