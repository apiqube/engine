// Package dataflow manages data passing between test cases at runtime.
//
// Three mechanisms coexist:
//
//   - prev:  implicit previous response, available in scenario mode
//   - save:  named variables extracted from responses via save: { key: path }
//   - alias: cross-file references via {{ alias-name.response.body.field }}
//
// The Store is thread-safe and holds all runtime data for a single Run().
// It supports both synchronous lookup (fast path) and asynchronous waiting
// (when a consumer runs in a parallel wave before its producer finishes).
package dataflow
