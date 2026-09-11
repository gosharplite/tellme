package cli

import (
	"path/filepath"
	"testing"
)

// T003 skeleton materialised as a real test — the CLI package's one pure helper
// (default path discovery) is covered here (research Decision 1 & 3).
func TestDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"mode taken from TELL_ME_MODE", "coder", filepath.Join("/home/x", "configs", "coder.yaml")},
		{"defaults to butler when unset", "", filepath.Join("/home/x", "configs", "butler.yaml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TELL_ME_MODE", tt.mode)
			if got := defaultConfigPath("/home/x"); got != tt.want {
				t.Fatalf("defaultConfigPath = %q, want %q", got, tt.want)
			}
		})
	}
}
