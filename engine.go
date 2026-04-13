package engine

import "context"

type Engine struct {
	handler       EventHandler
	signals       chan Signal
	pluginDir     string
	config        *Config
	parallel      bool
	maxConcurrent int
	env           map[string]string
}

func New(opts ...Option) *Engine {
	e := &Engine{
		handler:       NopHandler{},
		parallel:      true,
		maxConcurrent: 10,
		env:           make(map[string]string),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Run executes tests from the given file or directory paths.
func (e *Engine) Run(ctx context.Context, paths ...string) (*Results, error) {
	// TODO: implementation
	return &Results{}, nil
}

// Check validates test files without executing them.
func (e *Engine) Check(paths ...string) ([]ValidationError, error) {
	// TODO: implementation
	return nil, nil
}

type Results struct {
	Total    int            `json:"total"`
	Passed   int            `json:"passed"`
	Failed   int            `json:"failed"`
	Skipped  int            `json:"skipped"`
	Errored  int            `json:"errored"`
	Duration string         `json:"duration"`
	Tests    []TestCompleted `json:"tests"`
	Waves    []WaveResult   `json:"waves"`
}

type WaveResult struct {
	Index    int             `json:"index"`
	Parallel bool            `json:"parallel"`
	Tests    []TestCompleted `json:"tests"`
	Duration string          `json:"duration"`
}

type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type Config struct {
	Version  int               `yaml:"version"`
	Targets  map[string]string `yaml:"targets,omitempty"`
	Defaults *DefaultsConfig   `yaml:"defaults,omitempty"`
	Plugins  []string          `yaml:"plugins,omitempty"`
	Runner   *RunnerConfig     `yaml:"runner,omitempty"`
	Hooks    *HooksConfig      `yaml:"hooks,omitempty"`
	Services map[string]any    `yaml:"services,omitempty"`
	Output   *OutputConfig     `yaml:"output,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
}

type DefaultsConfig struct {
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
}

type RunnerConfig struct {
	Parallel      bool `yaml:"parallel"`
	MaxConcurrent int  `yaml:"max_concurrent,omitempty"`
	FailFast      bool `yaml:"fail_fast,omitempty"`
}

type HooksConfig struct {
	Before    []HookAction `yaml:"before,omitempty"`
	After     []HookAction `yaml:"after,omitempty"`
	OnFailure []HookAction `yaml:"on_failure,omitempty"`
}

type HookAction struct {
	Run    string `yaml:"run,omitempty"`
	Wait   string `yaml:"wait,omitempty"`
	Notify string `yaml:"notify,omitempty"`
}

type OutputConfig struct {
	Format      string `yaml:"format,omitempty"`
	Verbose     bool   `yaml:"verbose,omitempty"`
	SaveResults string `yaml:"save_results,omitempty"`
}
