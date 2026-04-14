package assertion

// Operator is a comparison function that evaluates actual against expected.
type Operator func(actual, expected any) (bool, string)

// operators registers all built-in comparison operators.
var operators = map[string]Operator{
	// TODO: implementation
	// "eq":       opEqual,
	// "gt":       opGreater,
	// "gte":      opGreaterEqual,
	// "lt":       opLess,
	// "lte":      opLessEqual,
	// "contains": opContains,
	// "exists":   opExists,
	// "matches":  opMatches,
	// "one_of":   opOneOf,
	// "is":       opIsType,
}

// parseOperator splits an assertion string like "> 18" into operator and value.
// Returns ("gt", "18") or ("eq", "18") if no operator prefix.
func parseOperator(expression string) (op string, value string) {
	// TODO: implementation
	// 1. Trim whitespace
	// 2. Check for prefixes: ">=", "<=", ">", "<", "==", "!="
	// 3. If no prefix, assume equality
	return "eq", expression
}
