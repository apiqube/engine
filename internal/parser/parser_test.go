package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apiqube/engine/internal/manifest"
)

func TestParseBytes_FullForm(t *testing.T) {
	src := `
target: http://api.example.com
mode: test
tests:
  - name: get user
    method: GET
    resource: /users/1
    expect:
      status: 200
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	f := files[0]
	if f.Target != "http://api.example.com" {
		t.Errorf("target wrong: %q", f.Target)
	}
	if f.Mode != manifest.ModeTest {
		t.Errorf("mode wrong: %q", f.Mode)
	}
	if len(f.Tests) != 1 {
		t.Fatalf("got %d tests", len(f.Tests))
	}
	tc := f.Tests[0]
	if tc.Name != "get user" || tc.Method != "GET" || tc.Resource != "/users/1" {
		t.Errorf("test parsed wrong: %+v", tc)
	}
	if tc.Expect.Status != uint64(200) && tc.Expect.Status != 200 {
		t.Errorf("status wrong: %v (%T)", tc.Expect.Status, tc.Expect.Status)
	}
}

func TestParseBytes_CompactForm(t *testing.T) {
	src := `
tests:
  - "POST /users":
      body:
        name: alice
      expect:
        status: 201
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	tc := files[0].Tests[0]
	if tc.Method != "POST" || tc.Resource != "/users" {
		t.Errorf("compact-form parsed wrong: method=%q resource=%q", tc.Method, tc.Resource)
	}
	if tc.Extra == nil || tc.Extra["body"] == nil {
		t.Errorf("body should be in Extra: %v", tc.Extra)
	}
}

func TestParseBytes_OneLiner(t *testing.T) {
	src := `
tests:
  - "GET /health -> 200"
  - "POST /users -> 201"
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	tests := files[0].Tests
	if len(tests) != 2 {
		t.Fatalf("got %d tests", len(tests))
	}
	if tests[0].Method != "GET" || tests[0].Resource != "/health" || tests[0].Expect.Status != "200" {
		t.Errorf("one-liner 0: %+v", tests[0])
	}
	if tests[1].Method != "POST" || tests[1].Resource != "/users" {
		t.Errorf("one-liner 1: %+v", tests[1])
	}
}

func TestParseBytes_OneLinerInvalid(t *testing.T) {
	src := `
tests:
  - "garbage"
`
	if _, err := New().ParseBytes([]byte(src)); err == nil {
		t.Error("expected parse error on garbage one-liner")
	}
}

func TestParseBytes_PluginExtraFields(t *testing.T) {
	src := `
tests:
  - name: query users
    method: GET
    resource: /users
    query:
      page: 1
      limit: 50
    followRedirects: true
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	tc := files[0].Tests[0]
	if tc.Extra == nil {
		t.Fatal("extra should be populated")
	}
	if tc.Extra["query"] == nil || tc.Extra["followRedirects"] != true {
		t.Errorf("extra wrong: %v", tc.Extra)
	}
}

func TestParseBytes_MultiDoc(t *testing.T) {
	src := `
tests:
  - "GET /a -> 200"
---
tests:
  - "GET /b -> 200"
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
}

func TestParseBytes_MalformedYAML(t *testing.T) {
	src := `tests:
  - { invalid yaml`
	if _, err := New().ParseBytes([]byte(src)); err == nil {
		t.Error("expected error on malformed YAML")
	}
}

func TestParseBytes_EmptyDoc(t *testing.T) {
	src := `# comment only
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("empty doc should produce no files, got %d", len(files))
	}
}

func TestParseBytes_RetryUntil(t *testing.T) {
	src := `
tests:
  - name: poll
    method: GET
    resource: /status
    retry:
      maxAttempts: 5
      interval: 100ms
      until:
        body.processed: true
`
	files, err := New().ParseBytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	tc := files[0].Tests[0]
	if tc.Retry == nil || tc.Retry.MaxAttempts != 5 {
		t.Errorf("retry config wrong: %+v", tc.Retry)
	}
	if tc.Retry.Until == nil {
		t.Fatal("retry.until should be parsed")
	}
}

func TestParseReader(t *testing.T) {
	src := `
tests:
  - "GET / -> 200"
`
	files, err := New().ParseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files", len(files))
	}
}

func TestParsePaths_FileAndDir(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yml")
	c := filepath.Join(dir, "ignored.txt")
	if err := os.WriteFile(a, []byte("tests:\n  - \"GET /a -> 200\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("tests:\n  - \"GET /b -> 200\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c, []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := New().ParsePaths(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("dir walk: got %d, want 2", len(files))
	}
	for _, f := range files {
		if f.Path == "" {
			t.Errorf("Path should be populated for file-loaded TestFile")
		}
	}

	// Single-file path also works.
	files, err = New().ParsePaths(a)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("single file: got %d, want 1", len(files))
	}
}

func TestParsePaths_Missing(t *testing.T) {
	if _, err := New().ParsePaths("/no/such/path"); err == nil {
		t.Error("missing path should error")
	}
}

func TestIsYAMLFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"a.yaml", true},
		{"a.yml", true},
		{"a.YAML", true},
		{"a.json", false},
		{"a.txt", false},
		{"a", false},
	}
	for _, c := range cases {
		if got := isYAMLFile(c.path); got != c.want {
			t.Errorf("isYAMLFile(%q) = %v; want %v", c.path, got, c.want)
		}
	}
}
