package config

import (
	"errors"
	"testing"
)

// T013 — table-driven unit tests for environment variable expansion
// (research Decision 2 & 4).
//
// Review finding #1: the tests drive the injected lookup port
// (ExpandStringWithLookup) rather than mutating process-global environment
// state, so the table runs fully in memory and in parallel — no os.Setenv, no
// cross-test pollution, no serialization.
func TestExpandStringWithLookup(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		switch key {
		case "EXPAND_FOO":
			return "bar", true
		case "EXPAND_HOST":
			return "api.example.com", true
		case "EXPAND_PORT":
			return "8080", true
		case "EXPAND_EMPTY":
			return "", true
		default:
			return "", false
		}
	}

	tests := []struct {
		name      string
		input     string
		want      string
		wantErrIs error
	}{
		{
			name:      "no variables",
			input:     "https://api.deepseek.com",
			want:      "https://api.deepseek.com",
			wantErrIs: nil,
		},
		{
			name:      "simple variable expansion",
			input:     "${EXPAND_FOO}",
			want:      "bar",
			wantErrIs: nil,
		},
		{
			name:      "variable expansion embedded in text",
			input:     "Bearer ${EXPAND_FOO}",
			want:      "Bearer bar",
			wantErrIs: nil,
		},
		{
			name:      "multiple variables expansion",
			input:     "https://${EXPAND_HOST}:${EXPAND_PORT}/v1",
			want:      "https://api.example.com:8080/v1",
			wantErrIs: nil,
		},
		{
			name:      "variable with default when variable is set",
			input:     "${EXPAND_FOO:-default_val}",
			want:      "bar",
			wantErrIs: nil,
		},
		{
			name:      "variable with default when variable is unset",
			input:     "${EXPAND_UNSET:-default_val}",
			want:      "default_val",
			wantErrIs: nil,
		},
		{
			name:      "variable with empty default when variable is unset",
			input:     "${EXPAND_UNSET:-}",
			want:      "",
			wantErrIs: nil,
		},
		{
			name:      "empty variable value falls back to default",
			input:     "${EXPAND_EMPTY:-fallback}",
			want:      "fallback",
			wantErrIs: nil,
		},
		{
			name:      "unset variable without default fails",
			input:     "${EXPAND_UNSET}",
			want:      "",
			wantErrIs: ErrUnsetVariable,
		},
		{
			name:      "unclosed expression fails",
			input:     "https://${EXPAND_HOST/v1",
			want:      "",
			wantErrIs: ErrMalformedVariable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ExpandStringWithLookup(tt.input, lookup)
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("ExpandStringWithLookup(%q) expected error %v, got nil", tt.input, tt.wantErrIs)
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("ExpandStringWithLookup(%q) error = %v, want %v", tt.input, err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandStringWithLookup(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ExpandStringWithLookup(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestExpandString_ProcessEnvFallback verifies the delegating entry point
// ExpandString still resolves against the process environment. It is
// deliberately NOT parallel: t.Setenv mutates process-global state.
func TestExpandString_ProcessEnvFallback(t *testing.T) {
	t.Setenv("TELLME_EXPAND_DELEGATE", "from-process-env")

	got, err := ExpandString("x=${TELLME_EXPAND_DELEGATE}")
	if err != nil {
		t.Fatalf("ExpandString unexpected error: %v", err)
	}
	if got != "x=from-process-env" {
		t.Errorf("ExpandString = %q, want %q", got, "x=from-process-env")
	}
}
