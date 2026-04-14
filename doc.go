// Package engine is the core of the ApiQube testing platform.
//
// The engine parses declarative test manifests, builds dependency graphs from
// template references, resolves templates, executes tests through protocol
// plugins, and emits events to frontends (CLI, desktop, web, SDK).
//
// Engine is stateless after construction — a single Engine instance is safe
// for concurrent Run() calls. Each Run() creates an isolated execution
// context with its own event handler, signals, and data flow store.
//
// # Basic Usage
//
//	eng := engine.New()
//	results, err := eng.Run(ctx, engine.FromPaths("./tests/"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("%d passed, %d failed\n", results.Passed, results.Failed)
//
// # With Event Handler
//
//	type myHandler struct{}
//
//	func (myHandler) Handle(event engine.Event) {
//	    switch e := event.(type) {
//	    case engine.TestCompleted:
//	        fmt.Printf("%s: %s\n", e.Name, e.Status)
//	    }
//	}
//
//	eng := engine.New()
//	results, err := eng.Run(ctx,
//	    engine.FromPaths("./tests/"),
//	    engine.WithHandler(myHandler{}),
//	)
//
// # Input Sources
//
// Tests can come from multiple sources:
//
//	engine.FromPaths("./tests/")          // files and directories
//	engine.FromBytes(yamlBytes)           // raw YAML bytes
//	engine.FromReader(os.Stdin)           // any io.Reader
//
// # Concurrent Runs
//
// A single Engine can handle multiple concurrent Run() calls safely:
//
//	eng := engine.New()
//	go eng.Run(ctx1, engine.FromPaths("./suite-a/"))
//	go eng.Run(ctx2, engine.FromPaths("./suite-b/"))
//
// Each run has its own isolated EventHandler, Signals channel, and data flow.
package engine
