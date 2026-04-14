package engine

import (
	"context"
	"errors"
	"runtime"
)

// Engine is the core of the ApiQube testing platform.
//
// Engine is stateless after construction — safe for concurrent Run() calls.
// Configuration set via Options applies to all runs. Per-run configuration
// (event handler, signals, env overrides) is passed via RunOptions.
type Engine struct {
	pluginDir     string
	parallel      bool
	maxConcurrent int
	failFast      bool
}

// New creates a new Engine with the given Options.
// Defaults: parallel=true, maxConcurrent=GOMAXPROCS, failFast=false.
func New(opts ...Option) *Engine {
	e := &Engine{
		parallel:      true,
		maxConcurrent: runtime.GOMAXPROCS(0),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// ErrNoInput is returned when Run or Check is called without a valid Input source.
var ErrNoInput = errors.New("engine: no input source provided")

// Run executes tests from the given Input source.
//
// Input is mandatory and must be one of FromPaths, FromBytes, or FromReader.
// Additional RunOptions configure the event handler, signals, environment, etc.
//
// A single Engine instance can serve many concurrent Run() calls safely —
// each call has its own isolated runConfig.
func (e *Engine) Run(ctx context.Context, input Input, opts ...RunOption) (*Results, error) {
	if input == nil {
		return nil, ErrNoInput
	}
	rc := e.buildRunConfig(input, opts)
	_ = rc // TODO: Phase 1-2 — pass to internal/runner
	_ = ctx
	return &Results{}, nil
}

// Check validates manifests from the given Input source without executing them.
//
// Returns a list of validation errors (syntax issues, unknown fields,
// missing plugins, unresolved references, dependency cycles).
// Returns an empty slice if everything is valid.
func (e *Engine) Check(ctx context.Context, input Input, opts ...CheckOption) ([]ValidationError, error) {
	if input == nil {
		return nil, ErrNoInput
	}
	cc := e.buildCheckConfig(input, opts)
	_ = cc // TODO: Phase 1-2 — pass to internal/parser for validation
	_ = ctx
	return nil, nil
}

// buildRunConfig assembles a runConfig from engine defaults and user options.
// Called once per Run() to create isolated per-execution state.
func (e *Engine) buildRunConfig(input Input, opts []RunOption) *runConfig {
	rc := &runConfig{
		input:     input,
		handler:   NopHandler{},
		pluginDir: e.pluginDir,
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

// buildCheckConfig assembles a checkConfig from engine defaults and user options.
func (e *Engine) buildCheckConfig(input Input, opts []CheckOption) *checkConfig {
	cc := &checkConfig{
		input:     input,
		pluginDir: e.pluginDir,
	}
	for _, opt := range opts {
		opt(cc)
	}
	return cc
}
