package config

import (
	"errors"
	"os"
	"testing"
)

// T013 — table-driven unit tests for environment variable expansion (research Decision 2 & 4).
func TestExpandString(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_EXPAND_FOO", "bar")
	_ = os.Setenv("TEST_EXPAND_HOST", "api.example.com")
	_ = os.Setenv("TEST_EXPAND_PORT", "8080")
	_ = os.Unsetenv("TEST_EXPAND_UNSET")

	defer func() {
		_ = os.Unsetenv("TEST_EXPAND_FOO")
		_ = os.Unsetenv("TEST_EXPAND_HOST")
		_ = os.Unsetenv("TEST_EXPAND_PORT")
	}()

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
			input:     "${TEST_EXPAND_FOO}",
			want:      "bar",
			wantErrIs: nil,
		},
		{
			name:      "variable expansion embedded in text",
			input:     "Bearer ${TEST_EXPAND_FOO}",
			want:      "Bearer bar",
			wantErrIs: nil,
		},
		{
			name:      "multiple variables expansion",
			input:     "https://${TEST_EXPAND_HOST}:${TEST_EXPAND_PORT}/v1",
			want:      "https://api.example.com:8080/v1",
			wantErrIs: nil,
		},
		{
			name:      "variable with default when variable is set",
			input:     "${TEST_EXPAND_FOO:-default_val}",
			want:      "bar",
			wantErrIs: nil,
		},
		{
			name:      "variable with default when variable is unset",
			input:     "${TEST_EXPAND_UNSET:-default_val}",
			want:      "default_val",
			wantErrIs: nil,
		},
		{
			name:      "variable with empty default when variable is unset",
			input:     "${TEST_EXPAND_UNSET:-}",
			want:      "",
			wantErrIs: nil,
		},
		{
			name:      "unset variable without default fails",
			input:     "${TEST_EXPAND_UNSET}",
			want:      "",
			wantErrIs: ErrUnsetVariable,
		},
		{
			name:      "unclosed expression fails",
			input:     "https://${TEST_EXPAND_HOST/v1",
			want:      "",
			wantErrIs: ErrMalformedVariable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandString(tt.input)
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("ExpandString(%q) expected error %v, got nil", tt.input, tt.wantErrIs)
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("ExpandString(%q) error = %v, want %v", tt.input, err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandString(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ExpandString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
