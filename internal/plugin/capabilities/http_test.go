package capabilities

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPRequestLogic_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"echo":"` + string(body) + `","method":"` + r.Method + `"}`))
	}))
	defer server.Close()

	req := HTTPRequest{
		Method: "POST",
		URL:    server.URL + "/foo",
		Body:   []byte("hello"),
	}
	reqJSON, _ := json.Marshal(req)
	respJSON := HTTPRequestLogic(context.Background(), nil, reqJSON)

	var resp HTTPResponse
	if err := json.Unmarshal(respJSON, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != 201 {
		t.Errorf("status = %d, want 201", resp.Status)
	}
	if !strings.Contains(string(resp.Body), "hello") {
		t.Errorf("body should contain echo of input, got %q", resp.Body)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("missing Content-Type header: %v", resp.Headers)
	}
	if resp.DurationMs < 0 {
		t.Errorf("duration should be non-negative, got %d", resp.DurationMs)
	}
}

func TestHTTPRequestLogic_NetworkError(t *testing.T) {
	req := HTTPRequest{Method: "GET", URL: "http://127.0.0.1:1/", TimeoutMs: 100}
	reqJSON, _ := json.Marshal(req)
	respJSON := HTTPRequestLogic(context.Background(), nil, reqJSON)

	var resp HTTPResponse
	_ = json.Unmarshal(respJSON, &resp)
	if resp.Error == "" {
		t.Errorf("expected error on connection refused, got %+v", resp)
	}
}

func TestHTTPRequestLogic_BadJSON(t *testing.T) {
	respJSON := HTTPRequestLogic(context.Background(), nil, []byte(`{not json}`))
	var resp HTTPResponse
	_ = json.Unmarshal(respJSON, &resp)
	if resp.Error == "" {
		t.Error("malformed JSON should produce error response")
	}
}

func TestHTTPRequestLogic_BadURL(t *testing.T) {
	req := HTTPRequest{Method: "GET", URL: "://bad"}
	reqJSON, _ := json.Marshal(req)
	respJSON := HTTPRequestLogic(context.Background(), nil, reqJSON)
	var resp HTTPResponse
	_ = json.Unmarshal(respJSON, &resp)
	if resp.Error == "" {
		t.Error("bad URL should produce error response")
	}
}

func TestHTTPRequestLogic_Headers(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := HTTPRequest{
		Method:  "GET",
		URL:     server.URL,
		Headers: map[string]string{"Authorization": "Bearer xyz"},
	}
	reqJSON, _ := json.Marshal(req)
	HTTPRequestLogic(context.Background(), nil, reqJSON)
	if gotAuth != "Bearer xyz" {
		t.Errorf("header not forwarded: got %q", gotAuth)
	}
}

func TestHTTPRequestLogic_NoFollowRedirects(t *testing.T) {
	redirected := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/dst" {
			redirected = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/dst", http.StatusFound)
	}))
	defer server.Close()

	follow := false
	req := HTTPRequest{
		Method:          "GET",
		URL:             server.URL + "/src",
		FollowRedirects: &follow,
	}
	reqJSON, _ := json.Marshal(req)
	respJSON := HTTPRequestLogic(context.Background(), nil, reqJSON)
	var resp HTTPResponse
	_ = json.Unmarshal(respJSON, &resp)
	if redirected {
		t.Error("redirect should not have been followed")
	}
	if resp.Status != 302 {
		t.Errorf("expected 302, got %d", resp.Status)
	}
}

// fakeHTTPClient lets tests override the HTTPClient.
type fakeHTTPClient struct {
	called int
	status int
}

func (c *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.called++
	return &http.Response{
		StatusCode: c.status,
		Body:       io.NopCloser(strings.NewReader("fake")),
		Header:     http.Header{},
	}, nil
}

func TestHTTPRequestLogic_InjectedClient(t *testing.T) {
	client := &fakeHTTPClient{status: 418}
	req := HTTPRequest{Method: "GET", URL: "http://anything"}
	reqJSON, _ := json.Marshal(req)
	respJSON := HTTPRequestLogic(context.Background(), client, reqJSON)
	var resp HTTPResponse
	_ = json.Unmarshal(respJSON, &resp)
	if resp.Status != 418 {
		t.Errorf("got %d, want 418", resp.Status)
	}
	if client.called != 1 {
		t.Errorf("client called %d times, want 1", client.called)
	}
}
