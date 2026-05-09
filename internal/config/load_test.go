package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Errorf("empty path should return nil config, got %+v", cfg)
	}
}

func TestLoad_FullConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".qube.yaml")
	contents := `version: 1
targets:
  api: http://api.example.com
defaults:
  headers:
    Accept: application/json
  timeout: 30s
plugins:
  - http
runner:
  parallel: true
  maxConcurrent: 8
  failFast: false
hooks:
  before:
    - run: setup.sh
output:
  format: json
  verbose: true
  saveResults: results.json
env:
  API_KEY: secret
pluginConfigs:
  http:
    timeout: 5s
`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("config should not be nil")
	}
	if cfg.Version != 1 {
		t.Errorf("version wrong: %d", cfg.Version)
	}
	if cfg.Targets["api"] != "http://api.example.com" {
		t.Errorf("targets wrong: %v", cfg.Targets)
	}
	if cfg.Runner == nil || cfg.Runner.MaxConcurrent != 8 {
		t.Errorf("runner wrong: %+v", cfg.Runner)
	}
	if cfg.Output == nil || cfg.Output.Format != OutputJSON {
		t.Errorf("output wrong: %+v", cfg.Output)
	}
	if cfg.PluginConfigs["http"]["timeout"] != "5s" {
		t.Errorf("plugin config wrong: %v", cfg.PluginConfigs)
	}
}

func TestLoad_Missing(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "no-such-file.yaml"))
	if err == nil {
		t.Fatal("missing file should error")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should wrap fs.ErrNotExist, got %v", err)
	}
}

func TestLoad_Malformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(`version: [`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("malformed YAML should error")
	}
}

func TestLoadReader(t *testing.T) {
	cfg, err := LoadReader(strings.NewReader(`version: 2`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 2 {
		t.Errorf("got version %d", cfg.Version)
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	cfg, err := LoadReader(strings.NewReader(`version: 1`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Runner == nil || !cfg.Runner.Parallel {
		t.Errorf("default runner should be parallel: %+v", cfg.Runner)
	}
	if cfg.Output == nil || cfg.Output.Format != OutputPretty {
		t.Errorf("default output should be pretty: %+v", cfg.Output)
	}
}
