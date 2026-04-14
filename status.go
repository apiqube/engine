package engine

// TestStatus represents the outcome of a test case.
type TestStatus int

const (
	StatusPassed TestStatus = iota
	StatusFailed
	StatusSkipped
	StatusErrored // test could not run (plugin missing, parse error, etc.)
)

func (s TestStatus) String() string {
	switch s {
	case StatusPassed:
		return "passed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	case StatusErrored:
		return "errored"
	default:
		return "unknown"
	}
}

// Signal is a control command from frontend to engine during a run.
type Signal int

const (
	SignalPause Signal = iota
	SignalResume
	SignalSkipTest
)

func (s Signal) String() string {
	switch s {
	case SignalPause:
		return "pause"
	case SignalResume:
		return "resume"
	case SignalSkipTest:
		return "skip"
	default:
		return "unknown"
	}
}
