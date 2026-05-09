package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apiqube/engine/internal/plugin/capabilities"
	"github.com/apiqube/engine/internal/wire"
)

// TestWASM_Execute_HTTPRoundTrip drives plugin-http through a real HTTP call:
// engine→plugin→host_http_request→httptest.Server→back. Skipped when the
// plugin-http fixture isn't present.
func TestWASM_Execute_HTTPRoundTrip(t *testing.T) {
	_, p := loadHTTPFixture(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"echoed":` + escapeJSONString(string(body)) + `,"method":"` + r.Method + `","path":"` + r.URL.Path + `"}`))
	}))
	defer srv.Close()

	ctx := capabilities.WithSinks(context.Background(), &capabilities.Sinks{})

	in := TestInput{
		Method:   "POST",
		Resource: "/users",
		Target:   srv.URL,
		Headers:  map[string]string{"X-Trace": "abc"},
		Fields: map[string]any{
			"body": map[string]any{"name": "alice"},
		},
	}

	var emitted []wire.PluginEvent
	emit := func(e wire.PluginEvent) { emitted = append(emitted, e) }

	out, err := p.Execute(ctx, in, emit)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Error != "" {
		t.Fatalf("plugin reported error: %q", out.Error)
	}

	gotStatus, _ := toFloat(out.Status)
	if gotStatus != 201 {
		t.Errorf("status = %v, want 201", out.Status)
	}

	body, ok := out.Body.(map[string]any)
	if !ok {
		t.Fatalf("body should decode as object, got %T: %v", out.Body, out.Body)
	}
	if body["method"] != "POST" {
		t.Errorf("method round-trip wrong: %v", body["method"])
	}
	if body["path"] != "/users" {
		t.Errorf("path round-trip wrong: %v", body["path"])
	}
}

func TestWASM_Execute_HostError(t *testing.T) {
	_, p := loadHTTPFixture(t)

	in := TestInput{
		Method:   "GET",
		Resource: "/",
		Target:   "http://127.0.0.1:1", // refused
	}
	out, err := p.Execute(context.Background(), in, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Error == "" {
		t.Errorf("expected non-empty Error, got %+v", out)
	}
}

// toFloat handles JSON-decoded numbers (which are float64 in Go's encoding/json)
// and any int variants returned directly.
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// escapeJSONString JSON-encodes a string into the embedded form, including
// the surrounding quotes. Used by the in-test handler to assemble responses.
func escapeJSONString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
