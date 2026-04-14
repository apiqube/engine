package engine

// CheckOption configures a single Check() invocation.
// Check validates manifests without executing them, so it accepts a smaller
// set of options than Run — only input source, plugins, and config.
type CheckOption func(*checkConfig)

// checkConfig holds per-check state.
type checkConfig struct {
	input      Input
	configPath string
	pluginDir  string
}

// WithCheckConfigPath loads a specific .qube.yaml for validation context.
func WithCheckConfigPath(path string) CheckOption {
	return func(c *checkConfig) { c.configPath = path }
}

// WithCheckPlugins sets the plugin directory for this Check,
// so plugin-specific fields in manifests can be validated.
func WithCheckPlugins(dir string) CheckOption {
	return func(c *checkConfig) { c.pluginDir = dir }
}
