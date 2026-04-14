package assertion

// Engine evaluates assertion expressions against response data.
type Engine struct{}

// New creates a new assertion Engine.
func New() *Engine {
	return &Engine{}
}

// Result is the outcome of evaluating one assertion.
// Mirrors engine.AssertionResult but lives here to avoid circular imports.
type Result struct {
	Expression string
	Passed     bool
	Expected   any
	Actual     any
	Message    string
}

// Check evaluates a single assertion expression against the given actual value.
// The expression can be a raw value (equality), a string with operator (">", "<",
// "contains", "exists", "is integer"), or a map (one_of, not_exists, etc.).
func (e *Engine) Check(path string, expected any, actual any) Result {
	// TODO: implementation
	// 1. Inspect expected:
	//    - raw value → equality check
	//    - string starting with ">", ">=", "<", "<=" → numeric
	//    - "contains X" → substring
	//    - "exists" → presence check
	//    - "is TYPE" → type check
	//    - map with "one_of" → any-of check
	//    - map with "not_exists" → absence check
	// 2. Build Result with Expression, Passed, Expected, Actual, Message
	return Result{Expression: path}
}
