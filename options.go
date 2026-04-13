package engine

import "maps"

type Option func(*Engine)

func WithEventHandler(h EventHandler) Option {
	return func(e *Engine) {
		if h != nil {
			e.handler = h
		}
	}
}

func WithSignals(ch chan Signal) Option {
	return func(e *Engine) { e.signals = ch }
}

func WithConfigPath(path string) Option {
	return func(e *Engine) { e.configPath = path }
}

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

func WithEnv(env map[string]string) Option {
	return func(e *Engine) { maps.Copy(e.env, env) }
}
