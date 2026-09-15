// Package config loads and validates the tellme boot-time YAML configuration
// and resolves its effective values (TELL_ME_* precedence over the file).
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrInvalidValue marks a configuration value that is present but invalid (e.g.
// a negative rendered width). It distinguishes value errors from parse errors
// so the CLI can emit the general `the configuration is invalid` class phrase
// (round-006 FR-006) rather than the parse phrase.
var ErrInvalidValue = errors.New("invalid configuration value")

// DefaultMaxToolLoop is the default bound on tool-iteration rounds within one
// prompt run (round-008 research Decision 3 / FR-006).
const DefaultMaxToolLoop = 1000

// DefaultMaxHistoryTokens is the default payload budget the status line measures
// against (round-009 research Decision 5; the user-locked default is 1000000).
const DefaultMaxHistoryTokens = 1000000

// Config is the boot-time YAML configuration input (FR-005).
type Config struct {
	Mode             string              `yaml:"MODE"`
	Person           string              `yaml:"PERSON"`
	SelectedProvider string              `yaml:"SELECTED_PROVIDER"`
	WrapWidth        int                 `yaml:"WRAP_WIDTH"`
	MaxToolLoop      int                 `yaml:"MAX_TOOL_LOOP"`
	MaxHistoryTokens int                 `yaml:"MAX_HISTORY_TOKENS"`
	UseTUIPrompt     bool                `yaml:"USE_TUI_PROMPT"`
	Providers        map[string]Provider `yaml:"PROVIDERS"`
	// Models carries the round-018 config-only pricing table: a `MODELS` entry
	// keyed by model name, each holding the per-million-token HIT/MISS/COMP
	// rates the post-turn cost is computed from. It is NOT env-overrideable (the
	// standing TELL_ME_* precedence applies to scalar keys only) and there are NO
	// built-in rates — an un-priced model renders `$0.0000` (round-018 D2).
	Models map[string]ModelPricing `yaml:"MODELS"`
}

// ModelPricing is one `MODELS` entry: the model's optional `CONTEXT_WINDOW`
// (round-024) and its `PRICING` rates.
type ModelPricing struct {
	// ContextWindow is the model's token window (round-024), used — capped by
	// MAX_HISTORY_TOKENS — as the effective budget for the tool resource
	// contract. 0/absent → no configured window (the effective budget falls back
	// to MAX_HISTORY_TOKENS).
	ContextWindow int          `yaml:"CONTEXT_WINDOW"`
	Pricing       PricingRates `yaml:"PRICING"`
}

// PricingRates are the USD-per-million-token cost rates for a model (round-018).
type PricingRates struct {
	HIT  float64 `yaml:"HIT"`
	MISS float64 `yaml:"MISS"`
	COMP float64 `yaml:"COMP"`
}

// PricingFor returns the configured pricing for a model, and whether one exists
// (round-018 D2 — config-only, no built-in rates).
func (c *Config) PricingFor(model string) (PricingRates, bool) {
	mp, ok := c.Models[model]
	if !ok {
		return PricingRates{}, false
	}
	return mp.Pricing, true
}

// ContextWindowFor returns the configured context window for a model, and
// whether one is set (> 0). It is deliberately SEPARATE from PricingFor (a model
// may set a window without PRICING, or vice versa); the nested table stays
// file-only — the TELL_ME_* env-over-file precedence is not extended to it
// (round-024 FR-015).
func (c *Config) ContextWindowFor(model string) (int, bool) {
	mp, ok := c.Models[model]
	if !ok || mp.ContextWindow <= 0 {
		return 0, false
	}
	return mp.ContextWindow, true
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

// EffectiveWrapWidth resolves AND validates the rendered width: the environment
// override (TELL_ME_WRAP_WIDTH) when non-empty, else the file WRAP_WIDTH
// (round-006 FR-006 / FR-015). A non-integer override, or a negative value from
// either source, is an invalid configuration value (wrapping ErrInvalidValue).
// The helper owns both resolve and validate so the caller cannot forget either.
func (c *Config) EffectiveWrapWidth(override string) (int, error) {
	width := c.WrapWidth
	if override != "" {
		n, err := strconv.Atoi(strings.TrimSpace(override))
		if err != nil {
			return 0, fmt.Errorf("%w: WRAP_WIDTH %q is not an integer", ErrInvalidValue, override)
		}
		width = n
	}
	if width < 0 {
		return 0, fmt.Errorf("%w: WRAP_WIDTH cannot be negative (%d)", ErrInvalidValue, width)
	}
	return width, nil
}

// EffectiveMaxToolLoop resolves the tool-loop bound: the environment override
// (MAX_TOOL_LOOP) when non-empty, else the file MAX_TOOL_LOOP, else the default
// DefaultMaxToolLoop (round-008 research Decision 3 / FR-006). A non-integer
// override, or a negative value from either source, is an invalid configuration
// value (wrapping ErrInvalidValue).
func (c *Config) EffectiveMaxToolLoop(override string) (int, error) {
	limit := c.MaxToolLoop
	if override != "" {
		n, err := strconv.Atoi(strings.TrimSpace(override))
		if err != nil {
			return 0, fmt.Errorf("%w: MAX_TOOL_LOOP %q is not an integer", ErrInvalidValue, override)
		}
		limit = n
	}
	if limit == 0 {
		return DefaultMaxToolLoop, nil
	}
	if limit < 0 {
		return 0, fmt.Errorf("%w: MAX_TOOL_LOOP cannot be negative (%d)", ErrInvalidValue, limit)
	}
	return limit, nil
}

// EffectiveMaxHistoryTokens resolves the payload budget the status line measures
// against: the environment override (MAX_HISTORY_TOKENS) when non-empty, else
// the file MAX_HISTORY_TOKENS, else the default DefaultMaxHistoryTokens
// (round-009 research Decision 5). A zero (unset) value falls back to the
// default; a non-integer override or a negative value from either source is an
// invalid configuration value (wrapping ErrInvalidValue).
func (c *Config) EffectiveMaxHistoryTokens(override string) (int, error) {
	limit := c.MaxHistoryTokens
	if override != "" {
		n, err := strconv.Atoi(strings.TrimSpace(override))
		if err != nil {
			return 0, fmt.Errorf("%w: MAX_HISTORY_TOKENS %q is not an integer", ErrInvalidValue, override)
		}
		limit = n
	}
	if limit == 0 {
		return DefaultMaxHistoryTokens, nil
	}
	if limit < 0 {
		return 0, fmt.Errorf("%w: MAX_HISTORY_TOKENS cannot be negative (%d)", ErrInvalidValue, limit)
	}
	return limit, nil
}

// EffectiveUseTUIPrompt resolves whether the opt-in interactive TUI prompt is
// enabled: the -i/--interactive flag OR the config USE_TUI_PROMPT key
// (round-015 FR-001). The TUI also requires a terminal stdin; that second gate
// lives in internal/cli (the real-isatty seam), not here.
func (c *Config) EffectiveUseTUIPrompt(flag bool) bool {
	return flag || c.UseTUIPrompt
}
