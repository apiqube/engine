package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/goccy/go-yaml"
)

// Load reads and parses a .qube.yaml file from path. Returns (nil, nil) when
// path is empty so callers can pass an unset configPath without special-casing.
//
// Returns os.ErrNotExist (wrapped) when the path is set but the file does
// not exist; callers may treat that as a hard error or fall back to defaults.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, path)
		}
		return nil, err
	}
	return parse(data)
}

// LoadReader parses .qube.yaml content from any io.Reader.
func LoadReader(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&c)
	return &c, nil
}

// applyDefaults fills in sensible defaults so callers never see nil sections
// for the common cases.
func applyDefaults(c *Config) {
	if c.Runner == nil {
		c.Runner = &Runner{Parallel: true}
	}
	if c.Output == nil {
		c.Output = &Output{Format: OutputPretty}
	}
}
