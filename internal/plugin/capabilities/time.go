package capabilities

import (
	"context"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func resolveNow(s *Sinks) func() int64 {
	if s == nil || s.Now == nil {
		return func() int64 { return time.Now().UnixMilli() }
	}
	return s.Now
}

func addTime(builder wazero.HostModuleBuilder) {
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, _ api.Module) uint64 {
			now := resolveNow(sinksFrom(ctx))()
			if now < 0 {
				return 0
			}
			return uint64(now)
		}).
		Export("host_now")
}
