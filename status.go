package engine

import "github.com/apiqube/engine/internal/wire"

// TestStatus represents the outcome of a test case.
type TestStatus = wire.TestStatus

const (
	StatusPassed  = wire.StatusPassed
	StatusFailed  = wire.StatusFailed
	StatusSkipped = wire.StatusSkipped
	StatusErrored = wire.StatusErrored
)

// Signal is a control command from frontend to engine during a run.
type Signal = wire.Signal

const (
	SignalPause    = wire.SignalPause
	SignalResume   = wire.SignalResume
	SignalSkipTest = wire.SignalSkipTest
)
