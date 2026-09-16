package gemini

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
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

// N-2 (review follow-up) — parseResponse returns a *plain* truncation error; the
// *llm.ProviderError wrapping that establishes the frozen `the provider request
// failed` + exit-6 contract happens in Complete. Pin the type at that layer too,
// on the function-call-aware path (FR-003).
func TestComplete_TruncationIsProviderError(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok-abc","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	genSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":"write_file","args":{}}}]},"finishReason":"MAX_TOKENS"}]}`))
	}))
	defer genSrv.Close()

	c, err := New(Config{
		ProviderName: "vertex-flash",
		BaseURL:      genSrv.URL + "/models",
		APIKey:       writeTestKey(t, tokenSrv.URL),
		Model:        "gemini-3-flash",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Complete(context.Background(), llm.Request{Prompt: "go"})
	var perr *llm.ProviderError
	if err == nil || !errors.As(err, &perr) {
		t.Fatalf("error = %v, want *llm.ProviderError", err)
	}
	if perr.Provider != "vertex-flash" {
		t.Errorf("ProviderError.Provider = %q, want vertex-flash", perr.Provider)
	}
	if !strings.Contains(perr.Err.Error(), "write_file") {
		t.Errorf("ProviderError.Err = %q, want the function-call-aware detail", perr.Err.Error())
	}
}
