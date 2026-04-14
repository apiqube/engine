// Package assertion evaluates expect expressions against test responses.
//
// Assertion syntax supports simple comparisons, type checks, and complex expressions:
//
//   status: 200                      # exact match
//   body.users.length: "> 0"         # numeric comparison
//   body.email: contains "@"         # substring
//   body.id: exists                  # presence check
//   body.age: is integer             # type check
//   status: { one_of: [200, 201] }   # one of several values
//
// The assertion engine is responsible for parsing assertion expressions,
// evaluating them against response data, and producing AssertionResult records
// for reporting.
package assertion
