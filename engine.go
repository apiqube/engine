package engine

import "context"

type Engine struct {
	handler       EventHandler
	signals       chan Signal
	configPath    string
	pluginDir     string
	parallel      bool
	maxConcurrent int
	failFast      bool
	env           map[string]string
}

func New(opts ...Option) *Engine {
	e := &Engine{
		handler:       NopHandler{},
		parallel:      true,
		maxConcurrent: 10,
		env:           make(map[string]string),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) Run(ctx context.Context, paths ...string) (*Results, error) {
	// TODO: Phase 1-2
	return &Results{}, nil
}

func (e *Engine) Check(paths ...string) ([]ValidationError, error) {
	// TODO: Phase 1-2
	return nil, nil
}

func (e *Engine) emit(event Event) {
	e.handler.Handle(event)
}
