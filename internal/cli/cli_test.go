package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// T003 materialised as a real test — the CLI package's one pure helper (default
// path discovery) is covered here (research Decision 1 & 3).
func TestDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		unset bool
		want  string
	}{
		{name: "mode taken from TELL_ME_MODE", mode: "coder", want: filepath.Join("/home/x", "configs", "coder.yaml")},
		{name: "defaults to butler when unset", unset: true, want: filepath.Join("/home/x", "configs", "butler.yaml")},
		{name: "defaults to butler when blank", mode: "", want: filepath.Join("/home/x", "configs", "butler.yaml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TELL_ME_MODE", "placeholder")
			if tt.unset {
				if err := os.Unsetenv("TELL_ME_MODE"); err != nil {
					t.Fatalf("unset TELL_ME_MODE: %v", err)
				}
			} else {
				t.Setenv("TELL_ME_MODE", tt.mode)
			}
			if got := defaultConfigPath("/home/x"); got != tt.want {
				t.Fatalf("defaultConfigPath = %q, want %q", got, tt.want)
			}
		})
	}
}
