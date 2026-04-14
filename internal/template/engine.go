package template

// Resolver resolves templates in string values using a context of available data.
// Unknown references return the original value unchanged (lenient mode).
type Resolver struct {
	// Lookup is called when the resolver encounters a {{ name.path }} reference
	// that is not a built-in generator (like fake.* or env.*).
	// Implementations typically query the dataflow.Store.
	Lookup func(name, path string) (any, bool)

	// Env is the environment variable map for {{ env.KEY }} references.
	Env map[string]string
}

// New creates a new Resolver with the given data lookup function.
func New(lookup func(name, path string) (any, bool), env map[string]string) *Resolver {
	return &Resolver{Lookup: lookup, Env: env}
}

// Resolve processes a single string, replacing all {{ ... }} expressions.
// Returns the resolved string. If any expression fails to resolve, it is
// left in place (lenient).
func (r *Resolver) Resolve(s string) (string, error) {
	// TODO: implementation
	// 1. Find all {{ ... }} matches
	// 2. For each match:
	//    a. Parse expression (generator / method chain / reference)
	//    b. Evaluate
	//    c. Substitute result (or leave if failed)
	return s, nil
}

// ResolveValue resolves templates in any value — string, map, slice, nested.
// Used for request bodies where templates can be buried deep in the structure.
func (r *Resolver) ResolveValue(value any) (any, error) {
	// TODO: implementation
	// Walk the value, applying Resolve() to every string encountered
	return value, nil
}
