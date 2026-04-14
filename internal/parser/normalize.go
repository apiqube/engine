package parser

import "github.com/apiqube/engine/internal/manifest"

// normalize converts compact and one-liner test syntax into the full TestCase format.
// After normalization, every test in the slice has Name, Method, Endpoint, etc.
// populated consistently, regardless of how it was originally written.
func normalize(raw []any) ([]manifest.TestCase, error) {
	// TODO: implementation
	//
	// For each element in raw:
	//
	// 1. If string: one-liner form ("GET /users -> 200")
	//    Parse with regex: ^(METHOD)\s+(/path)\s*->\s*(\d+)$
	//    Produce TestCase with method, endpoint in Extra, expect.status
	//
	// 2. If map with HTTP-method key: compact form
	//    Example: { "POST /users": { body: {...}, expect: 201 } }
	//    Extract method + endpoint from key, merge value fields
	//
	// 3. If map with "name" or "method": full form
	//    Parse directly into TestCase
	return nil, nil
}

// extractExtra collects unknown fields from a test case into Extra map.
// This is how plugin-specific fields (call, query, connect, etc.) are captured.
func extractExtra(raw map[string]any, tc *manifest.TestCase) {
	// TODO: implementation
	// 1. Walk raw map
	// 2. Skip known core fields (name, alias, target, method, endpoint, ...)
	// 3. Put remaining fields into tc.Extra
}
