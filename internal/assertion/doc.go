// Package assertion evaluates expect expressions against test responses.
//
// # Expression forms
//
// An assertion is given a `path` (for error reporting), an `expected` value,
// and the `actual` value extracted from the test response. The expected side
// can take three shapes:
//
//	body.id: 42                           # primitive — equality with coercion
//	body.age: "> 18"                      # string operator form
//	body.code: { oneOf: [200, 201] }      # map operator form
//
// # Operators
//
// String form (operator prefix on a string):
//
//	> X, >= X, < X, <= X, == X, != X      # numeric comparison (with coercion)
//	contains X                            # substring (string) or membership (slice)
//	matches X                             # regular expression match on a string
//	is X                                  # type check: int, number, string, bool, array, object, null
//	exists                                # actual is not nil/missing
//
// Map form (single key):
//
//	eq, ne, gt, gte, lt, lte
//	contains, matches, is
//	exists: true, notExists: true
//	oneOf: [...]
//
// # Type coercion rules
//
//   - Both values numeric: compared as float64.
//   - Both values strings: compared as strings.
//   - One numeric, one string: the string is parsed as a number; if successful,
//     the comparison is numeric. Otherwise the assertion fails.
//   - Bool values are equal only to bool values with the same truth value.
//     "true"/"false" strings are NOT coerced to bool unless `is` is the operator.
//   - nil is equal only to nil; `exists` and `notExists` test nil-ness.
//
// # Lenient by design
//
// An unknown operator or a malformed expression yields a failed Result with a
// descriptive Message rather than a panic. Tests should fail on assertion
// errors, not on exceptions.
package assertion
