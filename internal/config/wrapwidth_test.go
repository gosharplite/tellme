package config

import "testing"

// T013 — rendered-width resolution (round-006 research Decision 7).

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.EffectiveWrapWidth(tt.override)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("EffectiveWrapWidth(%q) = %d, want an error", tt.override, got)
				}
				if _, ok := err.(interface{ Unwrap() error }); !ok {
					t.Errorf("error %v is not wrapped for errors.Is", err)
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

func TestEffectiveWrapWidthNonIntegerIsErrInvalidValue(t *testing.T) {
	if _, err := (&Config{}).EffectiveWrapWidth("-x"); err == nil || !isErrInvalidValue(err) {
		t.Fatalf("EffectiveWrapWidth(-x) error = %v, want ErrInvalidValue", err)
	}
}

// isErrInvalidValue reports whether err wraps ErrInvalidValue.
func isErrInvalidValue(err error) bool { return err != nil && errIs(err) }

func errIs(err error) bool {
	for err != nil {
		if err == ErrInvalidValue {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
