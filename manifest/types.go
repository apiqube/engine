package manifest

type TestFile struct {
	Path     string            `yaml:"-"`
	Mode     string            `yaml:"mode,omitempty"`     // test (default), scenario, load
	Target   string            `yaml:"target,omitempty"`
	Targets  map[string]string `yaml:"targets,omitempty"`
	Defaults *FileDefaults     `yaml:"defaults,omitempty"`
	Parallel *bool             `yaml:"parallel,omitempty"`
	Depends  []string          `yaml:"depends,omitempty"`
	Load     *LoadConfig       `yaml:"load,omitempty"`
	Tests    []TestCase        `yaml:"tests"`
}

type FileDefaults struct {
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
}

type TestCase struct {
	Name    string            `yaml:"name,omitempty"`
	Alias   string            `yaml:"alias,omitempty"`
	Target  string            `yaml:"target,omitempty"`
	Tags    []string          `yaml:"tags,omitempty"`
	Skip    bool              `yaml:"skip,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
	Expect  Expect            `yaml:"expect"`
	Save    map[string]string `yaml:"save,omitempty"`
	When    string            `yaml:"when,omitempty"`
	Retry   *RetryConfig      `yaml:"retry,omitempty"`
	Matrix  []map[string]any  `yaml:"matrix,omitempty"`
	Depends []string          `yaml:"depends,omitempty"`
	Extra   map[string]any    `yaml:"-"` // plugin-specific fields
}

type Expect struct {
	Status   any            `yaml:"status,omitempty"`   // int, string, or {one_of: [...]}
	Body     map[string]any `yaml:"body,omitempty"`     // gjson path → assertion
	Headers  map[string]any `yaml:"headers,omitempty"`
	Duration string         `yaml:"duration,omitempty"` // e.g. "< 500ms"
}

type RetryConfig struct {
	MaxAttempts int            `yaml:"max_attempts"`
	Interval    string         `yaml:"interval"`
	Until       map[string]any `yaml:"until,omitempty"`
}

type LoadConfig struct {
	Users      int                          `yaml:"users,omitempty"`
	Duration   string                       `yaml:"duration,omitempty"`
	RPS        int                          `yaml:"rps,omitempty"`
	RampUp     string                       `yaml:"ramp_up,omitempty"`
	Stages     []LoadStage                  `yaml:"stages,omitempty"`
	Thresholds map[string]map[string]string `yaml:"thresholds,omitempty"`
	Scenarios  map[string]LoadScenario      `yaml:"scenarios,omitempty"`
}

type LoadStage struct {
	Duration string `yaml:"duration"`
	Users    int    `yaml:"users"`
}

type LoadScenario struct {
	Weight int      `yaml:"weight"`
	Tests  []string `yaml:"tests"`
}
