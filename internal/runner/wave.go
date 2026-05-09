package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apiqube/engine/internal/wire"
	"github.com/apiqube/engine/internal/assertion"
	"github.com/apiqube/engine/internal/dataflow"
	"github.com/apiqube/engine/internal/graph"
	"github.com/apiqube/engine/internal/manifest"
	"github.com/apiqube/engine/internal/plugin"
	"github.com/apiqube/engine/internal/template"
)

// runTest executes a single test case end-to-end: target/protocol resolution,
// template resolution, plugin call (with retry), assertions, save extraction,
// and prev-snapshot update for scenario mode.
func (r *Runner) runTest(ctx context.Context, ref *graph.TestRef, store *dataflow.Store) TestOutcome {
	tc := ref.Test
	target := resolveTarget(ref.File, tc)

	if tc.Skip {
		out := TestOutcome{Ref: ref, Target: target, Status: wire.StatusSkipped, Skipped: true}
		r.fireTestStart(ref, "", target)
		r.fireTestComplete(out)
		return out
	}

	proto, err := detectProtocol(target)
	if err != nil {
		out := TestOutcome{Ref: ref, Target: target, Status: wire.StatusErrored, Error: err.Error()}
		r.fireTestStart(ref, proto, target)
		r.fireTestComplete(out)
		return out
	}

	plug := r.registry.Lookup(proto)
	if plug == nil {
		out := noPlugin(ref, proto, target)
		r.fireTestStart(ref, proto, target)
		r.fireTestComplete(out)
		return out
	}

	r.fireTestStart(ref, proto, target)

	resolver := r.makeResolver(store)
	input, err := buildTestInput(resolver, tc, target)
	if err != nil {
		out := TestOutcome{Ref: ref, Protocol: proto, Target: target, Status: wire.StatusErrored, Error: err.Error()}
		r.fireTestComplete(out)
		return out
	}

	var output *plugin.TestOutput
	var assertions []AssertionOutcome
	var execErr error
	start := time.Now()

	emit := r.makeEventSink(tc.Alias, store)
	output, assertions, execErr = r.runWithRetry(ctx, plug, input, tc, emit)

	duration := time.Since(start)

	out := TestOutcome{
		Ref:        ref,
		Protocol:   proto,
		Target:     target,
		StartTime:  start,
		Duration:   duration,
		Request:    &input,
		Response:   output,
		Assertions: assertions,
	}

	switch {
	case execErr != nil:
		out.Status = wire.StatusErrored
		out.Error = execErr.Error()
	case anyAssertionFailed(assertions):
		out.Status = wire.StatusFailed
	default:
		out.Status = wire.StatusPassed
	}

	if output != nil {
		r.persistResults(tc, output, store)
		if ref.File != nil && ref.File.Mode == manifest.ModeScenario {
			store.SetPrev(toSnapshot(tc.Name, output))
		}
	}

	r.fireTestComplete(out)
	return out
}

// runWithRetry runs Execute, optionally retrying per RetryConfig until either
// the Until expectation passes or attempts are exhausted.
func (r *Runner) runWithRetry(
	ctx context.Context,
	plug plugin.Plugin,
	input plugin.TestInput,
	tc *manifest.TestCase,
	emit plugin.EventSink,
) (*plugin.TestOutput, []AssertionOutcome, error) {
	if tc.Retry == nil || tc.Retry.MaxAttempts <= 1 {
		return r.runOnce(ctx, plug, input, tc, emit)
	}

	interval, _ := time.ParseDuration(tc.Retry.Interval)
	var lastOutput *plugin.TestOutput
	var lastAssertions []AssertionOutcome
	var lastErr error

	for attempt := range tc.Retry.MaxAttempts {
		if err := ctx.Err(); err != nil {
			return lastOutput, lastAssertions, err
		}
		out, assertions, err := r.runOnce(ctx, plug, input, tc, emit)
		lastOutput, lastAssertions, lastErr = out, assertions, err

		if err == nil && tc.Retry.Until != nil {
			untilResults := r.runAssertions(*tc.Retry.Until, out)
			if !anyAssertionFailed(untilResults) {
				return out, assertions, nil
			}
		}

		if attempt < tc.Retry.MaxAttempts-1 && interval > 0 {
			select {
			case <-time.After(interval):
			case <-ctx.Done():
				return lastOutput, lastAssertions, ctx.Err()
			}
		}
	}
	return lastOutput, lastAssertions, lastErr
}

func (r *Runner) runOnce(
	ctx context.Context,
	plug plugin.Plugin,
	input plugin.TestInput,
	tc *manifest.TestCase,
	emit plugin.EventSink,
) (*plugin.TestOutput, []AssertionOutcome, error) {
	callCtx := ctx
	if tc.Timeout != "" {
		if d, err := time.ParseDuration(tc.Timeout); err == nil && d > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, d)
			defer cancel()
		}
	}

	output, err := plug.Execute(callCtx, input, emit)
	if err != nil {
		return output, nil, err
	}
	if output == nil {
		return nil, nil, errors.New("plugin returned nil output")
	}
	return output, r.runAssertions(tc.Expect, output), nil
}

// runAssertions evaluates every entry in expect against the plugin output.
func (r *Runner) runAssertions(expect manifest.Expect, output *plugin.TestOutput) []AssertionOutcome {
	eng := assertion.New()
	var results []AssertionOutcome

	if expect.Status != nil {
		results = append(results, intoOutcome(eng.Check("status", expect.Status, output.Status)))
	}

	for path, expected := range expect.Body {
		if path == "." {
			results = append(results, intoOutcome(eng.Check("body", expected, output.Body)))
			continue
		}
		actual, _ := dataflow.Extract(output.Body, path)
		results = append(results, intoOutcome(eng.Check("body."+path, expected, actual)))
	}

	for path, expected := range expect.Headers {
		actual, _ := dataflow.Extract(stringMapToAny(output.Headers), path)
		results = append(results, intoOutcome(eng.Check("headers."+path, expected, actual)))
	}

	if expect.Duration != "" {
		// Compare output.DurationMs (int64 ms) against the expected expression.
		// The expected string may contain a duration suffix (e.g. "< 500ms");
		// strip the suffix into pure ms before passing to the assertion wire.
		expr := normalizeDurationExpr(expect.Duration)
		results = append(results, intoOutcome(eng.Check("duration", expr, output.DurationMs)))
	}

	return results
}

// makeResolver builds a template Resolver wired to the dataflow store and the
// runner's environment map. Lookup tries a flat key first, then falls back
// to extracting a path from the structured value stored under the alias name.
func (r *Runner) makeResolver(store *dataflow.Store) *template.Resolver {
	return template.New(func(name, path string) (any, bool) {
		if name == "prev" {
			snap := store.GetPrev()
			if snap == nil {
				return nil, false
			}
			if path == "" {
				return snapshotToMap(snap), true
			}
			return dataflow.Extract(snapshotToMap(snap), path)
		}

		// Flat saved key: store.Set("savedKey", value)
		if v, ok := store.Get(name + "." + path); ok && path != "" {
			return v, true
		}
		if path == "" {
			if v, ok := store.Get(name); ok {
				return v, true
			}
			return nil, false
		}

		// Structured alias: store.Set(alias, testData) — extract path from it.
		if v, ok := store.Get(name); ok {
			return dataflow.Extract(v, path)
		}
		return nil, false
	}, r.options.Env)
}

// makeEventSink returns an EventSink that forwards plugin events both to the
// engine layer (via runner Hooks if any) and into the dataflow store under
// the test's alias so subsequent tests can reference them.
func (r *Runner) makeEventSink(alias string, store *dataflow.Store) plugin.EventSink {
	return func(ev wire.PluginEvent) {
		if alias != "" {
			store.AppendEvent(alias, ev)
		}
		if h := r.options.Hooks.OnPluginEvent; h != nil {
			h(ev)
		}
	}
}

// persistResults stores the plugin output (and any save: entries) into the
// dataflow store so subsequent tests can reference it.
func (r *Runner) persistResults(tc *manifest.TestCase, output *plugin.TestOutput, store *dataflow.Store) {
	data := outputToMap(output)
	if tc.Alias != "" {
		store.Set(tc.Alias, data)
	}
	for saveKey, path := range tc.Save {
		val, ok := dataflow.Extract(data, path)
		if !ok {
			continue
		}
		store.Set(saveKey, val)
	}
}

// fireTestStart and fireTestComplete invoke the optional hooks if set.
func (r *Runner) fireTestStart(ref *graph.TestRef, proto wire.Protocol, target string) {
	if h := r.options.Hooks.OnTestStart; h != nil {
		h(TestStart{Ref: ref, Protocol: proto, Target: target})
	}
}

func (r *Runner) fireTestComplete(out TestOutcome) {
	if h := r.options.Hooks.OnTestComplete; h != nil {
		h(out)
	}
}

func intoOutcome(r assertion.Result) AssertionOutcome {
	return AssertionOutcome{
		Expression: r.Expression,
		Passed:     r.Passed,
		Expected:   r.Expected,
		Actual:     r.Actual,
		Message:    r.Message,
	}
}

func anyAssertionFailed(results []AssertionOutcome) bool {
	for _, r := range results {
		if !r.Passed {
			return true
		}
	}
	return false
}

// resolveTarget picks the effective target for a test, resolving aliases
// against the file-level Targets map.
func resolveTarget(file *manifest.TestFile, tc *manifest.TestCase) string {
	target := tc.Target
	if target == "" && file != nil {
		target = file.Target
	}
	if file != nil && file.Targets != nil {
		if full, ok := file.Targets[target]; ok {
			return full
		}
	}
	return target
}

// detectProtocol reads the URL scheme from a target string. Returns an error
// if no scheme is present.
func detectProtocol(target string) (wire.Protocol, error) {
	if target == "" {
		return "", errors.New("target is empty")
	}
	if i := strings.Index(target, "://"); i > 0 {
		return wire.Protocol(target[:i]), nil
	}
	return "", fmt.Errorf("target %q has no protocol scheme", target)
}

// buildTestInput resolves every templated string in the test case and
// produces the plugin.TestInput.
func buildTestInput(resolver *template.Resolver, tc *manifest.TestCase, target string) (plugin.TestInput, error) {
	method, _ := resolver.Resolve(tc.Method)
	resource, _ := resolver.Resolve(tc.Resource)
	timeout, _ := resolver.Resolve(tc.Timeout)

	headers := make(map[string]string, len(tc.Headers))
	for k, v := range tc.Headers {
		rv, _ := resolver.Resolve(v)
		headers[k] = rv
	}

	fields := make(map[string]any, len(tc.Extra))
	for k, v := range tc.Extra {
		rv, err := resolver.ResolveValue(v)
		if err != nil {
			return plugin.TestInput{}, fmt.Errorf("resolve field %q: %w", k, err)
		}
		fields[k] = rv
	}

	return plugin.TestInput{
		Method:   method,
		Resource: resource,
		Target:   target,
		Headers:  headers,
		Timeout:  timeout,
		Fields:   fields,
	}, nil
}

// toSnapshot converts a plugin output into the prev-snapshot shape.
func toSnapshot(name string, output *plugin.TestOutput) *dataflow.Snapshot {
	return &dataflow.Snapshot{
		Name:     name,
		Status:   output.Status,
		Headers:  output.Headers,
		Body:     output.Body,
		Metadata: output.Metadata,
	}
}

// snapshotToMap renders a Snapshot in the same shape as outputToMap so
// templates referencing `prev.response.body` and `alias.response.body` look
// alike.
func snapshotToMap(s *dataflow.Snapshot) map[string]any {
	return map[string]any{
		"name":     s.Name,
		"response": responseFields(s.Status, s.Headers, s.Body, s.Metadata),
	}
}

// outputToMap renders a TestOutput as a nested map for path-based templates.
func outputToMap(o *plugin.TestOutput) map[string]any {
	return map[string]any{
		"response": responseFields(o.Status, o.Headers, o.Body, o.Metadata),
		"events":   eventsToAny(o.Events),
	}
}

func responseFields(status any, headers map[string]string, body any, meta map[string]any) map[string]any {
	out := map[string]any{
		"status":  status,
		"body":    body,
		"headers": stringMapToAny(headers),
	}
	if meta != nil {
		out["metadata"] = meta
	}
	return out
}

func stringMapToAny(in map[string]string) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func eventsToAny(events []wire.PluginEvent) []any {
	if len(events) == 0 {
		return nil
	}
	out := make([]any, len(events))
	for i, ev := range events {
		out[i] = map[string]any{
			"plugin": ev.Plugin,
			"kind":   ev.Kind,
			"data":   ev.Data,
		}
	}
	return out
}

// normalizeDurationExpr turns a duration assertion like "< 500ms" into a form
// the assertion engine can evaluate against an int64 millisecond value.
func normalizeDurationExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	for _, suffix := range []string{"ms", "s", "m", "h"} {
		idx := strings.LastIndex(expr, suffix)
		if idx < 0 {
			continue
		}
		numPart := strings.TrimSpace(expr[:idx])
		// Find where the number starts in numPart.
		var prefix string
		i := len(numPart)
		for i > 0 && (numPart[i-1] >= '0' && numPart[i-1] <= '9' || numPart[i-1] == '.') {
			i--
		}
		prefix, num := strings.TrimSpace(numPart[:i]), strings.TrimSpace(numPart[i:])
		d, err := time.ParseDuration(num + suffix)
		if err != nil {
			continue
		}
		ms := d.Milliseconds()
		if prefix == "" {
			return fmt.Sprintf("%d", ms)
		}
		return fmt.Sprintf("%s %d", prefix, ms)
	}
	return expr
}
