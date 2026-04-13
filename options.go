package engine

// Option configures the Engine (set once, reused across runs).
type Option func(*Engine)

func WithPluginDir(dir string) Option {
	return func(e *Engine) { e.pluginDir = dir }
}

func WithParallel(enabled bool) Option {
	return func(e *Engine) { e.parallel = enabled }
}

func WithMaxConcurrent(n int) Option {
	return func(e *Engine) {
		if n > 0 {
			e.maxConcurrent = n
		}
	}
}

func WithFailFast(enabled bool) Option {
	return func(e *Engine) { e.failFast = enabled }
}
