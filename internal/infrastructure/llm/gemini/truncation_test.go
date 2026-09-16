package gemini

import (
	"strings"
	"testing"
)

// T011 [UNIT] — the round-030 Vertex/Gemini finish-reason truncation guard. A
// `candidates[0].finishReason == "MAX_TOKENS"` response must fail; a healthy
// finish reason (STOP / absent) must succeed; the truncation error must take
// precedence over the generic "no usable answer" error; and the message must
// name the tool on a function-call truncation site (FR-003).

func TestParseResponse_FinishReasonMaxTokensFails(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"partial answer"}]},"finishReason":"MAX_TOKENS"}]}`)
	_, err := parseResponse(raw)
	if err == nil {
		t.Fatal("expected an error for finishReason=MAX_TOKENS, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "truncat") {
		t.Errorf("expected a truncation-specific error, got %v", err)
	}
}

func TestParseResponse_FinishReasonHealthySucceeds(t *testing.T) {
	cases := map[string]string{
		"stop":   `{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]},"finishReason":"STOP"}]}`,
		"absent": `{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]}}]}`,
	}
	for name, body := range cases {
		if _, err := parseResponse([]byte(body)); err != nil {
			t.Errorf("%s: expected no error, got %v", name, err)
		}
	}
}

func TestParseResponse_TruncationPrecedesNoUsableAnswer(t *testing.T) {
	// An empty + MAX_TOKENS response must be reported as a truncation, not the
	// generic "no usable answer".
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[]},"finishReason":"MAX_TOKENS"}]}`)
	_, err := parseResponse(raw)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "truncat") {
		t.Errorf("expected the truncation error to take precedence, got %v", err)
	}
}

func TestParseResponse_FinishReasonMaxTokensOnFunctionCallNamesTool(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":"write_file","args":{}}}]},"finishReason":"MAX_TOKENS"}]}`)
	_, err := parseResponse(raw)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "write_file") {
		t.Errorf("expected the Gemini truncation error to name the tool, got %v", err)
	}
}
