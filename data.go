package engine

import "time"

// RequestData captures what was sent to the target during a test.
// Protocol-specific fields go in Metadata.
type RequestData struct {
	Method   string            `json:"method,omitempty"`
	URL      string            `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     any               `json:"body,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// ResponseData captures what the target returned.
// Status is any because HTTP uses int, gRPC uses string, etc.
type ResponseData struct {
	Status   any               `json:"status"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     any               `json:"body,omitempty"`
	Duration time.Duration     `json:"duration"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// AssertionResult captures the outcome of a single assertion check.
type AssertionResult struct {
	Expression string `json:"expression"`
	Passed     bool   `json:"passed"`
	Expected   any    `json:"expected,omitempty"`
	Actual     any    `json:"actual,omitempty"`
	Message    string `json:"message,omitempty"`
}
