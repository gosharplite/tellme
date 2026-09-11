package config

import "testing"

// T013 — table-driven unit tests for the pure resolver helpers (research
// Decision 1 & 3): env-over-file precedence and registry membership.
func TestEffectiveSelectedProvider(t *testing.T) {
	c := &Config{SelectedProvider: "file-provider"}
	tests := []struct {
		name     string
		override string
		want     string
	}{
		{"override wins over the file", "env-provider", "env-provider"},
		{"file value used when override empty", "", "file-provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.EffectiveSelectedProvider(tt.override); got != tt.want {
				t.Fatalf("EffectiveSelectedProvider(%q) = %q, want %q", tt.override, got, tt.want)
			}
		})
	}
}

func TestEffectiveMode(t *testing.T) {
	tests := []struct {
		name     string
		fileMode string
		override string
		want     string
	}{
		{"override wins over the file", "butler", "architect", "architect"},
		{"file value used when override empty", "coder", "", "coder"},
		{"defaults to butler when both empty", "", "", "butler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Mode: tt.fileMode}
			if got := c.EffectiveMode(tt.override); got != tt.want {
				t.Fatalf("EffectiveMode(%q) = %q, want %q", tt.override, got, tt.want)
			}
		})
	}
}

func TestProviderInRegistry(t *testing.T) {
	c := &Config{Providers: map[string]Provider{"deepseek-flash": {Type: "deepseek"}}}
	if !c.ProviderInRegistry("deepseek-flash") {
		t.Error("ProviderInRegistry(deepseek-flash) = false, want true")
	}
	if c.ProviderInRegistry("ghost") {
		t.Error("ProviderInRegistry(ghost) = true, want false")
	}
	if c.ProviderInRegistry("") {
		t.Error("ProviderInRegistry(empty) = true, want false")
	}
}
