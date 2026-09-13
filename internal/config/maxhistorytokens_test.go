package config

import (
	"errors"
	"testing"
)

// TestEffectiveMaxHistoryTokens pins the payload-budget resolver (round-009
// research Decision 5): env-over-file, default 1000000, zero → default,
// negative/non-integer → ErrInvalidValue.
func TestEffectiveMaxHistoryTokens(t *testing.T) {
	tests := []struct {
		name     string
		file     int
		override string
		want     int
		wantErr  bool
	}{
		{"default when unset", 0, "", DefaultMaxHistoryTokens, false},
		{"file value", 5000, "", 5000, false},
		{"env override wins", 5000, "2000000", 2000000, false},
		{"zero override falls back to default", 5000, "0", DefaultMaxHistoryTokens, false},
		{"non-integer override", 0, "abc", 0, true},
		{"negative override", 0, "-1", 0, true},
		{"negative file", -1, "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{MaxHistoryTokens: tt.file}
			got, err := c.EffectiveMaxHistoryTokens(tt.override)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidValue) {
					t.Fatalf("err = %v, want ErrInvalidValue", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
