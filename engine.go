package engine

import "context"

type Engine struct {
	pluginDir     string
	parallel      bool
	maxConcurrent int
	failFast      bool
}

func New(opts ...Option) *Engine {
	e := &Engine{
		parallel:      true,
		maxConcurrent: 10,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) Run(ctx context.Context, opts ...RunOption) (*Results, error) {
	rc := e.buildRunConfig(opts)
	_ = rc // TODO: Phase 1-2
	return &Results{}, nil
}

func (e *Engine) Check(ctx context.Context, opts ...RunOption) ([]ValidationError, error) {
	rc := e.buildRunConfig(opts)
	_ = rc // TODO: Phase 1-2
	return nil, nil
}

func (e *Engine) buildRunConfig(opts []RunOption) *runConfig {
	rc := &runConfig{
		handler:   NopHandler{},
		pluginDir: e.pluginDir,
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

func (rc *runConfig) emit(event Event) {
	rc.handler.Handle(event)
}
