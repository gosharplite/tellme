package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// T033 — the per-model CONTEXT_WINDOW accessor (round 024).

func TestContextWindowFor(t *testing.T) {
	c := &Config{Models: map[string]ModelPricing{
		"with-window": {ContextWindow: 200000},
		"no-window":   {},
	}}
	if w, ok := c.ContextWindowFor("with-window"); !ok || w != 200000 {
		t.Fatalf("ContextWindowFor(with-window) = (%d, %v); want (200000, true)", w, ok)
	}
	if _, ok := c.ContextWindowFor("no-window"); ok {
		t.Error("a zero window must report false")
	}
	if _, ok := c.ContextWindowFor("absent"); ok {
		t.Error("an absent model must report false")
	}
}

func TestContextWindowParsesFromYAML(t *testing.T) {
	var c Config
	if err := yaml.Unmarshal([]byte("MODELS:\n  m:\n    CONTEXT_WINDOW: 12345\n    PRICING:\n      HIT: 0.1\n"), &c); err != nil {
		t.Fatal(err)
	}
	if w, ok := c.ContextWindowFor("m"); !ok || w != 12345 {
		t.Fatalf("window = (%d, %v); want (12345, true)", w, ok)
	}
}
