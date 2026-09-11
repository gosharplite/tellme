package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

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


func TestProviderYAMLUnmarshal(t *testing.T) {
	raw := `
TYPE: deepseek
MODEL: deepseek-v4-pro
URL: https://api.deepseek.com
API_KEY: secret-123
MAX_TOKENS: 32768
HEADERS:
  X-Custom: custom-value
THINKING_BUDGET: 16384
THINKING_LEVEL: HIGH
`
	var p Provider
	if err := yaml.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if p.Type != "deepseek" {
		t.Errorf("p.Type = %q, want deepseek", p.Type)
	}
	if p.Model != "deepseek-v4-pro" {
		t.Errorf("p.Model = %q, want deepseek-v4-pro", p.Model)
	}
	if p.URL != "https://api.deepseek.com" {
		t.Errorf("p.URL = %q, want https://api.deepseek.com", p.URL)
	}
	if p.APIKey != "secret-123" {
		t.Errorf("p.APIKey = %q, want secret-123", p.APIKey)
	}
	if p.MaxTokens != 32768 {
		t.Errorf("p.MaxTokens = %d, want 32768", p.MaxTokens)
	}
	if p.Headers["X-Custom"] != "custom-value" {
		t.Errorf("p.Headers[X-Custom] = %q, want custom-value", p.Headers["X-Custom"])
	}
	if p.ThinkingBudget != 16384 {
		t.Errorf("p.ThinkingBudget = %d, want 16384", p.ThinkingBudget)
	}
	if p.ThinkingLevel != "HIGH" {
		t.Errorf("p.ThinkingLevel = %q, want HIGH", p.ThinkingLevel)
	}
}

func TestValidateProvider(t *testing.T) {
	tests := []struct {
		name    string
		p       Provider
		wantErr bool
	}{
		{
			name: "valid full provider",
			p: Provider{
				Type:           "openai",
				Model:          "gpt-5.5",
				URL:            "https://api.openai.com/v1",
				APIKey:         "key",
				MaxTokens:      1000,
				Headers:        map[string]string{"h": "v"},
				ThinkingBudget: 500,
				ThinkingLevel:  "HIGH",
			},
			wantErr: false,
		},
		{
			name: "valid minimal provider",
			p: Provider{
				Type:  "gemini",
				Model: "gemini-3-flash",
				URL:   "https://api.google.com",
			},
			wantErr: false,
		},
		{
			name: "missing TYPE fails",
			p: Provider{
				Model: "gemini-3-flash",
				URL:   "https://api.google.com",
			},
			wantErr: true,
		},
		{
			name: "missing MODEL fails",
			p: Provider{
				Type: "gemini",
				URL:  "https://api.google.com",
			},
			wantErr: true,
		},
		{
			name: "missing URL fails",
			p: Provider{
				Type:  "gemini",
				Model: "gemini-3-flash",
			},
			wantErr: true,
		},
		{
			name: "negative MAX_TOKENS fails",
			p: Provider{
				Type:      "gemini",
				Model:     "gemini-3-flash",
				URL:       "https://api.google.com",
				MaxTokens: -1,
			},
			wantErr: true,
		},
		{
			name: "negative THINKING_BUDGET fails",
			p: Provider{
				Type:           "gemini",
				Model:          "gemini-3-flash",
				URL:            "https://api.google.com",
				ThinkingBudget: -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("p.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
