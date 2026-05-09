package assertion

import "fmt"

// Engine evaluates assertion expressions against extracted response values.
type Engine struct{}

// New returns a new assertion Engine. Engine is stateless and safe for
// concurrent use.
func New() *Engine {
	return &Engine{}
}

// Result is the outcome of evaluating one assertion. Mirrors
// engine.AssertionResult; lives here to keep the package import-cycle-free.
type Result struct {
	Expression string
	Passed     bool
	Expected   any
	Actual     any
	Message    string
}

// Check evaluates the expected value against the actual value at the given path.
//
// Dispatch:
//   - expected is a string: parsed for an operator prefix; falls back to equality.
//   - expected is a map with a single recognized operator key: that operator runs.
//   - expected is any other map or slice: structural equality.
//   - otherwise: equality with type coercion.
func (e *Engine) Check(path string, expected, actual any) Result {
	op, opVal, used := classify(expected)
	passed, msg := runOperator(op, actual, opVal)
	if !used {
		// Equality fallback — keep the original expected in the Result for clarity.
		return Result{
			Expression: pathExpr(path, expected),
			Passed:     passed,
			Expected:   expected,
			Actual:     actual,
			Message:    msg,
		}
	}
	return Result{
		Expression: pathExpr(path, expected),
		Passed:     passed,
		Expected:   opVal,
		Actual:     actual,
		Message:    msg,
	}
}

// classify maps the expected value to (operator, operatorValue, wasOperator).
// wasOperator is false for plain equality fallback so callers can preserve the
// original expected value in the Result.
func classify(expected any) (string, any, bool) {
	switch v := expected.(type) {
	case string:
		op, val := parseOperator(v)
		if op == "exists" {
			return op, nil, true
		}
		if op == "eq" {
			return "eq", val, false
		}
		return op, val, true
	case map[string]any:
		if op, val, ok := operatorFromMap(v); ok {
			return op, val, true
		}
	}
	return "eq", expected, false
}

func pathExpr(path string, expected any) string {
	if path == "" {
		return fmt.Sprintf("%v", expected)
	}
	return fmt.Sprintf("%s = %v", path, expected)
}
