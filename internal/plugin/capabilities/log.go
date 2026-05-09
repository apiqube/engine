package capabilities

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// LogSink receives log records emitted by plugins via host_log.
type LogSink interface {
	Log(level LogLevel, message string)
}

// LogLevel classifies a host_log message.
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// String returns the canonical name for a log level.
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "debug"
	case LogInfo:
		return "info"
	case LogWarn:
		return "warn"
	case LogError:
		return "error"
	}
	return "unknown"
}

// LogSinkFunc adapts a function to LogSink.
type LogSinkFunc func(LogLevel, string)

// Log implements LogSink.
func (f LogSinkFunc) Log(l LogLevel, m string) { f(l, m) }

type nopLogSink struct{}

func (nopLogSink) Log(LogLevel, string) {}

func resolveLogger(s *Sinks) LogSink {
	if s == nil || s.Logger == nil {
		return nopLogSink{}
	}
	return s.Logger
}

func addLog(builder wazero.HostModuleBuilder) {
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, level, ptr, length uint32) {
			if length == 0 {
				return
			}
			data, ok := mod.Memory().Read(ptr, length)
			if !ok {
				return
			}
			resolveLogger(sinksFrom(ctx)).Log(LogLevel(level), string(data))
		}).
		Export("host_log")
}
