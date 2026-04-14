package engine

import "maps"

// RunOption configures a single Run() invocation.
// RunOptions are isolated per-execution — safe to use with concurrent Run() calls.
type RunOption func(*runConfig)

// runConfig holds all per-run state assembled from RunOptions.
// This is built fresh for each Run() call, so concurrent runs do not share state.
type runConfig struct {
	input      Input
	handler    EventHandler
	signals    chan Signal
	env        map[string]string
	configPath string
	pluginDir  string
}

// WithHandler sets the EventHandler for this Run.
// If not set, events are discarded via NopHandler.
func WithHandler(h EventHandler) RunOption {
	return func(c *runConfig) {
		if h != nil {
			c.handler = h
		}
	}
}

// WithSignals provides a channel for sending control signals (pause, resume, skip)
// to the engine during execution.
func WithSignals(ch chan Signal) RunOption {
	return func(c *runConfig) { c.signals = ch }
}

// WithEnv overrides environment variables for this Run.
// Values are accessible in templates via {{ env.KEY }}.
func WithEnv(env map[string]string) RunOption {
	return func(c *runConfig) {
		if c.env == nil {
			c.env = make(map[string]string)
		}
		maps.Copy(c.env, env)
	}
}

// WithConfigPath loads a specific .qube.yaml config file for this Run,
// overriding any engine-level default.
func WithConfigPath(path string) RunOption {
	return func(c *runConfig) { c.configPath = path }
}

// WithPlugins sets the plugin directory for this Run,
// overriding any engine-level default. Useful for multi-tenant cloud deployments.
func WithPlugins(dir string) RunOption {
	return func(c *runConfig) { c.pluginDir = dir }
}
