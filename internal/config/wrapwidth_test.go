package config

import (
	"errors"
	"testing"
)

// T013 — rendered-width resolution + validation (round-006 research Decision 7).

func TestEffectiveWrapWidth(t *testing.T) {
	c := &Config{WrapWidth: 120}
	tests := []struct {
		name     string
		override string
		want     int
		wantErr  bool
	}{
		{name: "no override uses the file value", override: "", want: 120},
		{name: "override wins over the file value", override: "40", want: 40},
		{name: "override is trimmed", override: "  80 ", want: 80},
		{name: "zero is a valid value (renderer default)", override: "0", want: 0},
		{name: "non-integer override is an invalid value", override: "abc", wantErr: true},
		{name: "negative override is an invalid value", override: "-5", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.EffectiveWrapWidth(tt.override)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("EffectiveWrapWidth(%q) = %d, want an error", tt.override, got)
				}
				if !errors.Is(err, ErrInvalidValue) {
					t.Errorf("EffectiveWrapWidth(%q) error = %v, want to wrap ErrInvalidValue", tt.override, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("EffectiveWrapWidth(%q) error: %v", tt.override, err)
			}
			if got != tt.want {
				t.Fatalf("EffectiveWrapWidth(%q) = %d, want %d", tt.override, got, tt.want)
			}
		})
	}
}

func TestEffectiveWrapWidthRejectsNegativeFileValue(t *testing.T) {
	if _, err := (&Config{WrapWidth: -5}).EffectiveWrapWidth(""); err == nil || !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("EffectiveWrapWidth(file=-5) error = %v, want to wrap ErrInvalidValue", err)
	}
}
