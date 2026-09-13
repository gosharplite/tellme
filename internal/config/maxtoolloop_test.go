package config

import (
	"errors"
	"testing"
)

// T024 (round 008) — MAX_TOOL_LOOP resolution + validation
// (research Decision 3 / FR-006: env/config, default 1000).

func TestEffectiveMaxToolLoop(t *testing.T) {
	c := &Config{MaxToolLoop: 25}
	tests := []struct {
		name     string
		override string
		want     int
		wantErr  bool
	}{
		{name: "no override uses the file value", override: "", want: 25},
		{name: "override wins over the file value", override: "3", want: 3},
		{name: "override is trimmed", override: "  7 ", want: 7},
		{name: "non-integer override is an invalid value", override: "abc", wantErr: true},
		{name: "negative override is an invalid value", override: "-1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.EffectiveMaxToolLoop(tt.override)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("EffectiveMaxToolLoop(%q) = %d, want an error", tt.override, got)
				}
				if !errors.Is(err, ErrInvalidValue) {
					t.Errorf("EffectiveMaxToolLoop(%q) error = %v, want to wrap ErrInvalidValue", tt.override, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("EffectiveMaxToolLoop(%q) error: %v", tt.override, err)
			}
			if got != tt.want {
				t.Fatalf("EffectiveMaxToolLoop(%q) = %d, want %d", tt.override, got, tt.want)
			}
		})
	}
}

func TestEffectiveMaxToolLoopDefault(t *testing.T) {
	got, err := (&Config{}).EffectiveMaxToolLoop("")
	if err != nil {
		t.Fatalf("EffectiveMaxToolLoop(default) error: %v", err)
	}
	if got != DefaultMaxToolLoop {
		t.Fatalf("default = %d, want DefaultMaxToolLoop (%d)", got, DefaultMaxToolLoop)
	}
	if DefaultMaxToolLoop != 1000 {
		t.Fatalf("DefaultMaxToolLoop = %d, want the pinned 1000", DefaultMaxToolLoop)
	}
}

func TestEffectiveMaxToolLoopNegativeFileValue(t *testing.T) {
	_, err := (&Config{MaxToolLoop: -3}).EffectiveMaxToolLoop("")
	if err == nil || !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("EffectiveMaxToolLoop(file=-3) error = %v, want to wrap ErrInvalidValue", err)
	}
}
