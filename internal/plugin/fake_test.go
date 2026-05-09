package plugin

import (
	"context"

	"github.com/apiqube/engine/internal/wire"
)

// fakePlugin is a Plugin implementation backed by Go closures, used to test
// runner/registry/loader without compiling a WASM fixture.
type fakePlugin struct {
	info        PluginInfo
	initFn      func(context.Context, map[string]any) error
	validateFn  func(TestInput) []FieldError
	executeFn   func(context.Context, TestInput, EventSink) (*TestOutput, error)
	closeFn     func() error
}

func (p *fakePlugin) Info() PluginInfo { return p.info }

func (p *fakePlugin) Init(ctx context.Context, cfg map[string]any) error {
	if p.initFn != nil {
		return p.initFn(ctx, cfg)
	}
	return nil
}

func (p *fakePlugin) Validate(input TestInput) []FieldError {
	if p.validateFn != nil {
		return p.validateFn(input)
	}
	return nil
}

func (p *fakePlugin) Execute(ctx context.Context, input TestInput, emit EventSink) (*TestOutput, error) {
	if p.executeFn != nil {
		return p.executeFn(ctx, input, emit)
	}
	return &TestOutput{Status: 200}, nil
}

func (p *fakePlugin) Close() error {
	if p.closeFn != nil {
		return p.closeFn()
	}
	return nil
}

func newFakePlugin(name string, protocols ...wire.Protocol) *fakePlugin {
	return &fakePlugin{info: PluginInfo{Name: name, Version: "0.0.1", Protocols: protocols}}
}
