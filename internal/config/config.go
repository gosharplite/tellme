// Package config loads and validates the tellme boot-time YAML configuration
// and resolves its effective values (TELL_ME_* precedence over the file).
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the boot-time YAML configuration input (FR-005).
type Config struct {
	Mode             string              `yaml:"MODE"`
	Person           string              `yaml:"PERSON"`
	SelectedProvider string              `yaml:"SELECTED_PROVIDER"`
	Providers        map[string]Provider `yaml:"PROVIDERS"`
}

// Provider is a single entry in the PROVIDERS registry.
type Provider struct {
	Type      string `yaml:"TYPE"`
	Model     string `yaml:"MODEL"`
	URL       string `yaml:"URL"`
	MaxTokens int    `yaml:"MAX_TOKENS"`
}

// Load reads and parses the YAML configuration at path. A missing file yields an
// error satisfying errors.Is(err, os.ErrNotExist); a malformed file yields a
// parser error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// EffectiveSelectedProvider resolves the effective selected provider: the
// override (TELL_ME_SELECTED_PROVIDER) when non-empty, else the file value
// (FR-003 / FR-015).
func (c *Config) EffectiveSelectedProvider(override string) string {
	if override != "" {
		return override
	}
	return c.SelectedProvider
}

// ProviderInRegistry reports whether name is present in the PROVIDERS registry.
func (c *Config) ProviderInRegistry(name string) bool {
	_, ok := c.Providers[name]
	return ok
}

// EffectiveMode resolves the effective mode: the override (TELL_ME_MODE) when
// non-empty, else the file MODE, else "butler" (FR-007 / FR-015).
func (c *Config) EffectiveMode(override string) string {
	if override != "" {
		return override
	}
	if c.Mode != "" {
		return c.Mode
	}
	return "butler"
}
