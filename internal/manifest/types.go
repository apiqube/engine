package manifest

// TestMode controls how tests in a file are executed.
type TestMode string

const (
	ModeTest     TestMode = "test"
	ModeScenario TestMode = "scenario"
	ModeLoad     TestMode = "load"
)

// IsValid reports whether this is a known mode.
func (m TestMode) IsValid() bool {
	switch m {
	case ModeTest, ModeScenario, ModeLoad:
		return true
	}
	return false
}

// TestFile is a single parsed test manifest file (after normalization).
type TestFile struct {
	Path     string            `yaml:"-"`
	Mode     TestMode          `yaml:"mode,omitempty"`
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

// TestCase is a single normalized test — one request-response cycle plus assertions.
//
// Method and Resource are core fields readable by every plugin (HTTP method/path,
// gRPC method-name, GraphQL operation, etc.). Plugin-specific fields land in Extra.
type TestCase struct {
	Name     string            `yaml:"name,omitempty"`
	Alias    string            `yaml:"alias,omitempty"`
	Target   string            `yaml:"target,omitempty"`
	Method   string            `yaml:"method,omitempty"`
	Resource string            `yaml:"resource,omitempty"`
	Tags     []string          `yaml:"tags,omitempty"`
	Skip     bool              `yaml:"skip,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Timeout  string            `yaml:"timeout,omitempty"`
	Expect   Expect            `yaml:"expect"`
	Save     map[string]string `yaml:"save,omitempty"`
	When     string            `yaml:"when,omitempty"`
	Retry    *RetryConfig      `yaml:"retry,omitempty"`
	Matrix   []map[string]any  `yaml:"matrix,omitempty"`
	Depends  []string          `yaml:"depends,omitempty"`
	Extra    map[string]any    `yaml:"-"`
}

// Expect describes the expected outcome of a test.
//
// Body is keyed by gjson-style path ("user.name", "items.0.id"). The special
// key "." matches the entire body against the given value (no path traversal).
type Expect struct {
	Status   any            `yaml:"status,omitempty"`
	Body     map[string]any `yaml:"body,omitempty"`
	Headers  map[string]any `yaml:"headers,omitempty"`
	Duration string         `yaml:"duration,omitempty"`
}

// RetryConfig controls automatic retry with polling for async APIs.
//
// Until reuses the Expect form: retry continues until the assertion passes,
// up to MaxAttempts, with Interval between attempts.
type RetryConfig struct {
	MaxAttempts int     `yaml:"maxAttempts"`
	Interval    string  `yaml:"interval"`
	Until       *Expect `yaml:"until,omitempty"`
}

// LoadConfig holds load testing parameters (only relevant when Mode == ModeLoad).
type LoadConfig struct {
	Users      int                          `yaml:"users,omitempty"`
	Duration   string                       `yaml:"duration,omitempty"`
	RPS        int                          `yaml:"rps,omitempty"`
	RampUp     string                       `yaml:"rampUp,omitempty"`
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
