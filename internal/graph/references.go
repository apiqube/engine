package graph

import "regexp"

// templateRegex matches {{ name.path.segments }} or {{ name }} references.
// First capture group is the alias/variable name, second is the optional path.
var templateRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][\w-]*)(?:\.([\w.-]+))?\s*\}\}`)

// Reference is a parsed template reference found in a test field.
type Reference struct {
	Name string // alias name or saved variable name
	Path string // dot-separated path after the name (may be empty)
}

// extractReferences walks any value (string, map, slice) and collects all
// template references found within. Used to discover implicit dependencies.
func extractReferences(value any) []Reference {
	// TODO: implementation
	//
	// 1. Switch on type:
	//    - string: regex-match templateRegex, collect groups
	//    - map[string]any: recurse on each value
	//    - []any: recurse on each element
	//    - other: no references
	// 2. Deduplicate by Name+Path
	return nil
}
