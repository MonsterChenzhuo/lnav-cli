// Package source manages lnav-cli named log sources (sources.yaml).
package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Paths        []string `yaml:"paths,omitempty"`
	Command      string   `yaml:"command,omitempty"`
	DefaultLevel string   `yaml:"default_level,omitempty"`
}

type Config struct {
	Sources map[string]Source `yaml:"sources"`
}

type Resolved struct {
	Files    []string
	StdinCmd string
}

// DefaultPath returns the standard sources.yaml location (~/.lnav-cli/sources.yaml).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "sources.yaml"
	}
	return filepath.Join(home, ".lnav-cli", "sources.yaml")
}

// Load reads the yaml at path. A missing file is treated as an empty config.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Sources: map[string]Source{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Sources == nil {
		cfg.Sources = map[string]Source{}
	}
	return &cfg, nil
}

// Save writes cfg to path, creating parent directories as needed.
func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// Resolve expands a mix of alias names and raw paths/globs into a Resolved set.
func (c *Config) Resolve(input []string) (Resolved, error) {
	var out Resolved
	for _, item := range input {
		if src, ok := c.Sources[item]; ok {
			switch {
			case src.Command != "":
				if out.StdinCmd != "" || len(out.Files) > 0 {
					return Resolved{}, fmt.Errorf("command-backed source %q cannot be combined with other sources", item)
				}
				out.StdinCmd = src.Command
			default:
				out.Files = append(out.Files, src.Paths...)
			}
			continue
		}
		if strings.ContainsAny(item, "\n\r") {
			return Resolved{}, fmt.Errorf("invalid source entry (contains newline)")
		}
		out.Files = append(out.Files, item)
	}
	if out.StdinCmd != "" && len(out.Files) > 0 {
		return Resolved{}, fmt.Errorf("command-backed source cannot be combined with file sources")
	}
	return out, nil
}
