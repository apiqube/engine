package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/apiqube/engine/internal/manifest"
)

// Parser loads and normalizes test manifests. A Parser is stateless and safe
// for concurrent use.
type Parser struct{}

// New returns a new Parser.
func New() *Parser {
	return &Parser{}
}

// ParsePaths loads test files from one or more file or directory paths.
// Directories are walked recursively for .yaml and .yml files.
func (p *Parser) ParsePaths(paths ...string) ([]*manifest.TestFile, error) {
	var out []*manifest.TestFile
	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				if !isYAMLFile(path) {
					return nil
				}
				files, err := p.parseFile(path)
				if err != nil {
					return err
				}
				out = append(out, files...)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			files, err := p.parseFile(root)
			if err != nil {
				return nil, err
			}
			out = append(out, files...)
		}
	}
	return out, nil
}

// ParseBytes parses raw YAML bytes into one or more TestFiles. Multi-document
// YAML is split before parsing.
func (p *Parser) ParseBytes(data []byte) ([]*manifest.TestFile, error) {
	return p.parse(bytes.NewReader(data), "")
}

// ParseReader reads YAML from r and parses it.
func (p *Parser) ParseReader(r io.Reader) ([]*manifest.TestFile, error) {
	return p.parse(r, "")
}

func (p *Parser) parseFile(path string) ([]*manifest.TestFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.parse(f, path)
}

func (p *Parser) parse(r io.Reader, path string) ([]*manifest.TestFile, error) {
	decoder := yaml.NewDecoder(r)
	var out []*manifest.TestFile
	for {
		var raw rawTestFile
		err := decoder.Decode(&raw)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, p.errAt(path, err)
		}
		f, err := buildTestFile(raw)
		if err != nil {
			return nil, p.errAt(path, err)
		}
		f.Path = path
		out = append(out, f)
	}
	return out, nil
}

func (p *Parser) errAt(path string, err error) error {
	if path == "" {
		return err
	}
	return fmt.Errorf("%s: %w", path, err)
}

func isYAMLFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return true
	}
	return false
}

// rawTestFile mirrors manifest.TestFile but holds Tests as []any so we can
// normalize each entry through the multiple-syntax pipeline.
type rawTestFile struct {
	Mode     manifest.TestMode      `yaml:"mode,omitempty"`
	Target   string                 `yaml:"target,omitempty"`
	Targets  map[string]string      `yaml:"targets,omitempty"`
	Defaults *manifest.FileDefaults `yaml:"defaults,omitempty"`
	Parallel *bool                  `yaml:"parallel,omitempty"`
	Depends  []string               `yaml:"depends,omitempty"`
	Load     *manifest.LoadConfig   `yaml:"load,omitempty"`
	Tests    []any                  `yaml:"tests"`
}

func buildTestFile(raw rawTestFile) (*manifest.TestFile, error) {
	cases, err := normalizeTests(raw.Tests)
	if err != nil {
		return nil, err
	}
	return &manifest.TestFile{
		Mode:     raw.Mode,
		Target:   raw.Target,
		Targets:  raw.Targets,
		Defaults: raw.Defaults,
		Parallel: raw.Parallel,
		Depends:  raw.Depends,
		Load:     raw.Load,
		Tests:    cases,
	}, nil
}
