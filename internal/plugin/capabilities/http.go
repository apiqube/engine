package capabilities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// HTTPClient is the abstraction host_http_request uses to perform HTTP calls.
// It is satisfied by *http.Client; tests inject fakes.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRequest is the wire-level shape host_http_request consumes.
type HTTPRequest struct {
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            []byte            `json:"body,omitempty"`
	TimeoutMs       int64             `json:"timeout_ms,omitempty"`
	FollowRedirects *bool             `json:"follow_redirects,omitempty"`
	MaxRedirects    int               `json:"max_redirects,omitempty"`
}

// HTTPResponse is the wire-level shape returned to the plugin.
type HTTPResponse struct {
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
}

func resolveHTTPClient(s *Sinks) HTTPClient {
	if s == nil || s.HTTP == nil {
		return defaultClient()
	}
	return s.HTTP
}

// HTTPRequestLogic performs the HTTP request described by reqJSON and returns
// the JSON bytes of the response. Pure function — no WASM dependency.
func HTTPRequestLogic(ctx context.Context, client HTTPClient, reqJSON []byte) []byte {
	resp := doHTTPRequest(ctx, client, reqJSON)
	out, err := json.Marshal(resp)
	if err != nil {
		return fmt.Appendf(nil, `{"status":0,"error":%q}`, err.Error())
	}
	return out
}

func doHTTPRequest(ctx context.Context, client HTTPClient, reqJSON []byte) HTTPResponse {
	var req HTTPRequest
	if err := json.Unmarshal(reqJSON, &req); err != nil {
		return HTTPResponse{Error: "decode: " + err.Error()}
	}
	if client == nil {
		client = configureClient(&req)
	}

	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return HTTPResponse{Error: "build request: " + err.Error()}
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	dur := time.Since(start).Milliseconds()
	if err != nil {
		return HTTPResponse{Error: err.Error(), DurationMs: dur}
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	headers := make(map[string]string, len(httpResp.Header))
	for k, vs := range httpResp.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}
	return HTTPResponse{
		Status:     httpResp.StatusCode,
		Headers:    headers,
		Body:       respBody,
		DurationMs: dur,
	}
}

func configureClient(req *HTTPRequest) *http.Client {
	c := &http.Client{}
	if req.TimeoutMs > 0 {
		c.Timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}
	if req.FollowRedirects != nil && !*req.FollowRedirects {
		c.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if req.MaxRedirects > 0 {
		max := req.MaxRedirects
		c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			if len(via) >= max {
				return errors.New("too many redirects")
			}
			return nil
		}
	}
	return c
}

func addHTTP(builder wazero.HostModuleBuilder) {
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr, length uint32) uint64 {
			data, ok := mod.Memory().Read(ptr, length)
			if !ok {
				return 0
			}
			client := resolveHTTPClient(sinksFrom(ctx))
			respJSON := HTTPRequestLogic(ctx, client, data)
			out, err := writeBackToPlugin(ctx, mod, respJSON)
			if err != nil {
				return 0
			}
			return out
		}).
		Export("host_http_request")
}

// writeBackToPlugin allocates plugin memory and writes data there, returning
// the packed (ptr<<32)|len pointer.
func writeBackToPlugin(ctx context.Context, mod api.Module, data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	alloc := mod.ExportedFunction("allocate")
	if alloc == nil {
		return 0, errors.New("plugin: 'allocate' export missing")
	}
	res, err := alloc.Call(ctx, uint64(len(data)))
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, errors.New("plugin: allocate returned no value")
	}
	packed := res[0]
	ptr := uint32(packed >> 32)
	if !mod.Memory().Write(ptr, data) {
		return 0, errors.New("plugin: memory write failed")
	}
	return packed, nil
}
