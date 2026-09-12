// Package config loads and validates the tellme boot-time YAML configuration
// and resolves its effective values (TELL_ME_* precedence over the file).
package config

import (
	"errors"
	"fmt"
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
	Type           string            `yaml:"TYPE"`
	Model          string            `yaml:"MODEL"`
	URL            string            `yaml:"URL"`
	APIKey         string            `yaml:"API_KEY"`
	MaxTokens      int               `yaml:"MAX_TOKENS"`
	Headers        map[string]string `yaml:"HEADERS"`
	ThinkingBudget int               `yaml:"THINKING_BUDGET"`
	ThinkingLevel  string            `yaml:"THINKING_LEVEL"`
}

// Validate validates that mandatory fields (TYPE, MODEL, URL) are non-empty and
// numeric limits (MAX_TOKENS, THINKING_BUDGET) are non-negative.
func (p *Provider) Validate() error {
	if p.Type == "" {
		return errors.New(`missing required field "TYPE"`)
	}
	if p.Model == "" {
		return errors.New(`missing required field "MODEL"`)
	}
	if p.URL == "" {
		return errors.New(`missing required field "URL"`)
	}
	if p.MaxTokens < 0 {
		return errors.New(`field "MAX_TOKENS" cannot be negative`)
	}
	if p.ThinkingBudget < 0 {
		return errors.New(`field "THINKING_BUDGET" cannot be negative`)
	}
	return nil
}

// Expand expands ${VAR} and ${VAR:-default} expressions in APIKey, URL, and Headers values.
func (p *Provider) Expand() error {
	expandedURL, err := ExpandString(p.URL)
	if err != nil {
		return fmt.Errorf("URL: %w", err)
	}
	p.URL = expandedURL

	if p.APIKey != "" {
		expandedKey, err := ExpandString(p.APIKey)
		if err != nil {
			return fmt.Errorf("API_KEY: %w", err)
		}
		p.APIKey = expandedKey
	}

	if len(p.Headers) > 0 {
		expandedHeaders := make(map[string]string, len(p.Headers))
		for k, v := range p.Headers {
			expandedVal, err := ExpandString(v)
			if err != nil {
				return fmt.Errorf("header %q: %w", k, err)
			}
			expandedHeaders[k] = expandedVal
		}
		p.Headers = expandedHeaders
	}
	return nil
}

// Load reads and parses the YAML configuration at path. A missing file yields an
// error satisfying errors.Is(err, os.ErrNotExist); a malformed file yields a
// parser error.
//
// Decode strictness — DELIBERATE DECISION (review F6, round 001): decoding is
// intentionally non-strict (unknown keys are tolerated). Round-001 Config models
// only the boot subset (MODE / PERSON / SELECTED_PROVIDER / PROVIDERS), while
// real tell-me-go configurations carry many more keys that later slices will add
// — so rejecting unknown keys would reject valid configs. Validation here is
// limited to selected-provider membership (performed in internal/cli); a typo
// that empties PROVIDERS therefore classifies as provider-mismatch, not
// config-invalid. Enabling yaml Decoder.KnownFields(true) is revisited when the
// full config schema lands.
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
