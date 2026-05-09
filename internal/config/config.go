package config

// Config is the parsed shape of a .qube.yaml file.
type Config struct {
	Version  int                       `yaml:"version"`
	Targets  map[string]string         `yaml:"targets,omitempty"`
	Defaults *Defaults                 `yaml:"defaults,omitempty"`
	Plugins  []string                  `yaml:"plugins,omitempty"`
	Runner   *Runner                   `yaml:"runner,omitempty"`
	Hooks    *Hooks                    `yaml:"hooks,omitempty"`
	Services map[string]any            `yaml:"services,omitempty"`
	Output   *Output                   `yaml:"output,omitempty"`
	Env      map[string]string         `yaml:"env,omitempty"`

	// PluginConfigs is the per-plugin configuration map sourced from the
	// `pluginConfigs:` section of .qube.yaml. Each key is a plugin name; the
	// value is passed to plugin.Init(config).
	PluginConfigs map[string]map[string]any `yaml:"pluginConfigs,omitempty"`
}

type Defaults struct {
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
}

type Runner struct {
	Parallel      bool `yaml:"parallel"`
	MaxConcurrent int  `yaml:"maxConcurrent,omitempty"`
	FailFast      bool `yaml:"failFast,omitempty"`
}

type Hooks struct {
	Before    []HookAction `yaml:"before,omitempty"`
	After     []HookAction `yaml:"after,omitempty"`
	OnFailure []HookAction `yaml:"onFailure,omitempty"`
}

type HookAction struct {
	Run    string `yaml:"run,omitempty"`
	Wait   string `yaml:"wait,omitempty"`
	Notify string `yaml:"notify,omitempty"`
}

// OutputFormat is the format for result rendering/serialization.
type OutputFormat string

const (
	OutputPretty OutputFormat = "pretty"
	OutputJSON   OutputFormat = "json"
	OutputJUnit  OutputFormat = "junit"
	OutputTAP    OutputFormat = "tap"
)

type Output struct {
	Format      OutputFormat `yaml:"format,omitempty"`
	Verbose     bool         `yaml:"verbose,omitempty"`
	SaveResults string       `yaml:"saveResults,omitempty"`
}
