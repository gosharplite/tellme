package openai

import (
	"strings"
	"testing"
)

// T011 [UNIT] — the round-030 OpenAI-compatible finish-reason truncation guard.
// A `choices[0].finish_reason == "length"` response must fail; a healthy finish
// reason (stop / tool_calls / absent) must succeed; the truncation error must
// take precedence over the generic "no usable answer" error.

func TestParseResponse_FinishReasonLengthFails(t *testing.T) {
	raw := []byte(`{"choices":[{"message":{"role":"assistant","content":"partial answer"},"finish_reason":"length"}]}`)
	_, err := parseResponse(raw)
	if err == nil {
		t.Fatal("expected an error for finish_reason=length, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "truncat") {
		t.Errorf("expected a truncation-specific error, got %v", err)
	}
}

func TestParseResponse_FinishReasonHealthySucceeds(t *testing.T) {
	cases := map[string]string{
		"stop":       `{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`,
		"tool_calls": `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"read_files","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`,
		"absent":     `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`,
	}
	for name, body := range cases {
		if _, err := parseResponse([]byte(body)); err != nil {
			t.Errorf("%s: expected no error, got %v", name, err)
		}
	}
}

func TestParseResponse_TruncationPrecedesNoUsableAnswer(t *testing.T) {
	// A truncated response that is ALSO empty must be reported as a truncation,
	// not the generic "no usable answer".
	raw := []byte(`{"choices":[{"message":{"role":"assistant","content":""},"finish_reason":"length"}]}`)
	_, err := parseResponse(raw)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "truncat") {
		t.Errorf("expected the truncation error to take precedence, got %v", err)
	}
}
