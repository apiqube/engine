package parser

import (
	"io"

	"github.com/apiqube/engine/internal/manifest"
)

// Parser loads and normalizes test manifests from various sources.
// A single Parser instance is stateless and safe for concurrent use.
type Parser struct{}

// New creates a new Parser.
func New() *Parser {
	return &Parser{}
}

// ParsePaths loads test files from one or more file or directory paths.
// Directories are walked recursively for .yaml and .yml files.
func (p *Parser) ParsePaths(paths ...string) ([]*manifest.TestFile, error) {
	// TODO: implementation
	// 1. Walk each path, collecting .yaml/.yml files
	// 2. For each file, call ParseBytes with its contents
	// 3. Return all parsed files or the first error
	return nil, nil
}

// ParseBytes parses raw YAML bytes into one or more TestFiles.
// Multi-document YAML is split before parsing.
func (p *Parser) ParseBytes(data []byte) ([]*manifest.TestFile, error) {
	// TODO: implementation
	// 1. Split YAML into documents
	// 2. For each document, unmarshal into TestFile
	// 3. Normalize tests (compact/one-liner → full)
	// 4. Populate Extra fields from unknown keys
	return nil, nil
}

// ParseReader reads YAML from an io.Reader and parses it.
func (p *Parser) ParseReader(r io.Reader) ([]*manifest.TestFile, error) {
	// TODO: implementation
	// 1. Read all bytes from reader
	// 2. Delegate to ParseBytes
	return nil, nil
}
