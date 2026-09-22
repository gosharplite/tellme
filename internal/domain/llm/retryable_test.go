package llm

import (
	"errors"
	"testing"
)

func TestRetryable_TransportAndStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"transport failure", &ProviderError{Provider: "p", Err: errors.New("connection reset"), Transport: true}, true},
		{"429", &ProviderError{Provider: "p", Err: errors.New("rate"), Status: 429}, true},
		{"500", &ProviderError{Provider: "p", Err: errors.New("boom"), Status: 500}, true},
		{"503", &ProviderError{Provider: "p", Err: errors.New("boom"), Status: 503}, true},
		{"599", &ProviderError{Provider: "p", Err: errors.New("boom"), Status: 599}, true},
		{"400", &ProviderError{Provider: "p", Err: errors.New("bad"), Status: 400}, false},
		{"401", &ProviderError{Provider: "p", Err: errors.New("auth"), Status: 401}, false},
		{"404", &ProviderError{Provider: "p", Err: errors.New("nf"), Status: 404}, false},
		{"decode (no facts)", &ProviderError{Provider: "p", Err: errors.New("decode")}, false},
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Retryable(tc.err); got != tc.want {
				t.Fatalf("Retryable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRetryable_Unwraps(t *testing.T) {
	wrapped := errors.Join(errors.New("ctx"), &ProviderError{Provider: "p", Transport: true})
	if !Retryable(wrapped) {
		t.Fatal("Retryable must find a *ProviderError through errors.As")
	}
}
