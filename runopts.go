package engine

import (
	"io"
	"maps"
)

// RunOption configures a single Run() invocation (isolated per-execution).
type RunOption func(*runConfig)

type runConfig struct {
	// Input source (one of these must be set)
	paths  []string
	data   []byte
	reader io.Reader

	// Per-run overrides
	handler    EventHandler
	signals    chan Signal
	env        map[string]string
	configPath string
	pluginDir  string
}

// Input sources

func FromPaths(paths ...string) RunOption {
	return func(c *runConfig) { c.paths = paths }
}

func FromBytes(data []byte) RunOption {
	return func(c *runConfig) { c.data = data }
}

func FromReader(r io.Reader) RunOption {
	return func(c *runConfig) { c.reader = r }
}

// Per-run configuration

func WithHandler(h EventHandler) RunOption {
	return func(c *runConfig) {
		if h != nil {
			c.handler = h
		}
	}
}

func WithSignals(ch chan Signal) RunOption {
	return func(c *runConfig) { c.signals = ch }
}

func WithEnv(env map[string]string) RunOption {
	return func(c *runConfig) {
		if c.env == nil {
			c.env = make(map[string]string)
		}
		maps.Copy(c.env, env)
	}
}

func WithConfigPath(path string) RunOption {
	return func(c *runConfig) { c.configPath = path }
}

func WithPlugins(dir string) RunOption {
	return func(c *runConfig) { c.pluginDir = dir }
}
