package engine

import "time"

type Results struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Errored  int           `json:"errored"`
	Duration time.Duration `json:"duration"`
	Tests    []TestResult  `json:"tests"`
	Waves    []WaveResult  `json:"waves"`
}

// TestResult is the final snapshot of a completed test (not a transient event).
type TestResult struct {
	Name       string            `json:"name"`
	File       string            `json:"file"`
	Status     TestStatus        `json:"status"`
	Duration   time.Duration     `json:"duration"`
	Protocol   string            `json:"protocol"`
	Request    *RequestData      `json:"request,omitempty"`
	Response   *ResponseData     `json:"response,omitempty"`
	Assertions []AssertionResult `json:"assertions,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type WaveResult struct {
	Index    int           `json:"index"`
	Parallel bool          `json:"parallel"`
	Tests    []TestResult  `json:"tests"`
	Duration time.Duration `json:"duration"`
}

type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Level   string `json:"level"` // "error" or "warning"
}
