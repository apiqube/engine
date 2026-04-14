// Package parser loads and normalizes test manifests from YAML sources.
//
// The parser handles three syntax levels — full, compact, and one-liner —
// and normalizes all of them into the canonical TestFile structure.
//
// Parser is a pure transformation layer: it does not execute tests,
// resolve templates, or validate plugin fields. Its sole job is producing
// valid manifest structures from raw bytes.
//
// # Responsibilities
//
//   - Read files/directories and load YAML documents
//   - Split multi-document YAML files
//   - Detect and expand compact syntax ("POST /users": {...})
//   - Detect and expand one-liner syntax ("GET /users -> 200")
//   - Populate manifest.TestCase.Extra with plugin-specific fields
//   - Report parse errors with line numbers
package parser
