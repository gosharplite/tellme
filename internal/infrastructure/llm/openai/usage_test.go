package openai

import "testing"

// TestParseResponseUsage pins the adapter's usage extraction (round-009 research
// Decision 2): a response carrying a `usage` block reports it; one without does
// not.
func TestParseResponseUsage(t *testing.T) {
	withUsage := `{"choices":[{"message":{"role":"assistant","content":"hi"}}],"usage":{"prompt_tokens":42,"completion_tokens":7,"total_tokens":49}}`
	resp, err := parseResponse([]byte(withUsage))
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if !resp.Usage.Reported || resp.Usage.PromptTokens != 42 || resp.Usage.CompletionTokens != 7 || resp.Usage.TotalTokens != 49 {
		t.Fatalf("usage = %+v, want reported 42/7/49", resp.Usage)
	}

	noUsage := `{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`
	resp2, err := parseResponse([]byte(noUsage))
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if resp2.Usage.Reported {
		t.Fatalf("usage = %+v, want not reported", resp2.Usage)
	}
}
