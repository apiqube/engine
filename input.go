package engine

import "io"

// Input is a source of test manifests for Run and Check.
// Create instances via FromPaths, FromBytes, or FromReader.
//
// Input is a sealed interface — only this package can implement it.
type Input interface {
	input()
}

type inputPaths struct {
	paths []string
}

type inputBytes struct {
	data []byte
}

type inputReader struct {
	reader io.Reader
}

func (inputPaths) input()  {}
func (inputBytes) input()  {}
func (inputReader) input() {}

// FromPaths loads test manifests from the given file or directory paths.
// Directories are walked recursively for .yaml and .yml files.
func FromPaths(paths ...string) Input {
	return inputPaths{paths: paths}
}

// FromBytes loads test manifests from raw YAML bytes.
// Useful for web UIs where tests are authored in-memory.
func FromBytes(data []byte) Input {
	return inputBytes{data: data}
}

// FromReader loads test manifests from any io.Reader.
// Useful for piped input or streaming sources.
func FromReader(r io.Reader) Input {
	return inputReader{reader: r}
}
