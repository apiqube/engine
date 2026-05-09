package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader()
	defer loader.Close(context.Background())

	reg, err := loader.LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if reg == nil {
		t.Fatal("registry should not be nil")
	}
	if len(reg.Snapshot()) != 0 {
		t.Errorf("empty dir should yield empty registry")
	}
}

func TestLoader_NoDir(t *testing.T) {
	loader := NewLoader()
	defer loader.Close(context.Background())

	reg, err := loader.LoadDir(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if reg == nil {
		t.Fatal("empty dir argument should return empty registry, not nil")
	}
}

func TestLoader_NonexistentDir(t *testing.T) {
	loader := NewLoader()
	defer loader.Close(context.Background())

	_, err := loader.LoadDir(context.Background(), filepath.Join(os.TempDir(), "no-such-dir-xyz"))
	if err == nil {
		t.Error("missing dir should return error")
	}
}

func TestLoader_NotADirectory(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	loader := NewLoader()
	defer loader.Close(context.Background())

	if _, err := loader.LoadDir(context.Background(), file); err == nil {
		t.Error("file path should error")
	}
}

func TestLoader_IgnoresNonWASMFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "ignored.txt"), "")
	mustWrite(t, filepath.Join(dir, "ignored.json"), "{}")

	loader := NewLoader()
	defer loader.Close(context.Background())

	reg, err := loader.LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Snapshot()) != 0 {
		t.Errorf("non-wasm files should be ignored")
	}
}

func TestLoader_InvalidWASM(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "broken.wasm"), "not a wasm module")

	loader := NewLoader()
	defer loader.Close(context.Background())

	if _, err := loader.LoadDir(context.Background(), dir); err == nil {
		t.Error("invalid WASM should produce error")
	}
}

func TestLoader_CacheReuse(t *testing.T) {
	dir := t.TempDir() // empty
	loader := NewLoader()
	defer loader.Close(context.Background())

	reg1, err := loader.LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	reg2, err := loader.LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if reg1 != reg2 {
		t.Error("identical dir should return cached registry")
	}
}

func TestLoader_CacheInvalidatedOnFileAdded(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader()
	defer loader.Close(context.Background())

	reg1, err := loader.LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}

	mustWrite(t, filepath.Join(dir, "new.wasm"), "still bad")
	_, err = loader.LoadDir(context.Background(), dir)
	if err == nil {
		t.Error("adding broken wasm should cause re-scan and load error")
	}
	_ = reg1 // keep alive
}

func TestValidWASMEntry(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.wasm"), "")
	mustWrite(t, filepath.Join(dir, "B.WASM"), "")
	mustWrite(t, filepath.Join(dir, "c.txt"), "")

	stamps, err := collectStamps(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(stamps) != 2 {
		t.Errorf("expected 2 wasm files (case-insensitive), got %d", len(stamps))
	}
}

func TestStampsEqual(t *testing.T) {
	a := map[string]int64{"x": 1, "y": 2}
	b := map[string]int64{"x": 1, "y": 2}
	if !stampsEqual(a, b) {
		t.Error("equal maps should match")
	}
	if stampsEqual(a, map[string]int64{"x": 1}) {
		t.Error("different sizes should not match")
	}
	if stampsEqual(a, map[string]int64{"x": 1, "y": 99}) {
		t.Error("different values should not match")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
