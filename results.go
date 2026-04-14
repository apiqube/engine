package engine

import "time"

// Results holds the aggregated outcome of a Run().
// Frontends use this as the final snapshot after execution.
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

// TestResult is the complete record of a single test case execution.
// The TestCompleted event embeds this type, so any field added here is
// automatically available on the event.
type TestResult struct {
	Name       string            `json:"name"`
	File       string            `json:"file"`
	Protocol   string            `json:"protocol"`
	Target     string            `json:"target,omitempty"`
	Status     TestStatus        `json:"status"`
	Duration   time.Duration     `json:"duration"`
	Request    *RequestData      `json:"request,omitempty"`
	Response   *ResponseData     `json:"response,omitempty"`
	Assertions []AssertionResult `json:"assertions,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// WaveResult is the complete record of a single execution wave.
// The WaveCompleted event embeds this type.
type WaveResult struct {
	Index    int           `json:"index"`
	Parallel bool          `json:"parallel"`
	Tests    []TestResult  `json:"tests"`
	Duration time.Duration `json:"duration"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
}

// ValidationError describes a problem found during manifest validation.
type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Level   string `json:"level"` // "error" or "warning"
}
