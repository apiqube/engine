package engine

import (
	"time"

	"github.com/apiqube/engine/internal/wire"
)

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
	Protocol   Protocol          `json:"protocol"`
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

// Severity classifies the importance of a validation issue.
type Severity = wire.Severity

const (
	SeverityError   = wire.SeverityError
	SeverityWarning = wire.SeverityWarning
)

// ValidationError describes a problem found during manifest validation.
type ValidationError = wire.ValidationError
