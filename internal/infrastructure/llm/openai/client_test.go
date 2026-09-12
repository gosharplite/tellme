package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// TestRequestURL pins the endpoint convention (research Decision 3).
func TestRequestURL(t *testing.T) {
	cases := []struct{ base, want string }{
		{"https://api.deepseek.com", "https://api.deepseek.com/chat/completions"},
		{"https://api.openai.com/v1", "https://api.openai.com/v1/chat/completions"},
		{"https://api.openai.com/v1/", "https://api.openai.com/v1/chat/completions"},
	}
	for _, c := range cases {
		if got := requestURL(c.base); got != c.want {
			t.Errorf("requestURL(%q) = %q, want %q", c.base, got, c.want)
		}
	}
}

// TestRequestBody pins the request assembly: model, single user message,
// max_tokens only when positive, reasoning_effort only when set.
func TestRequestBody(t *testing.T) {
	body, err := requestBody("deepseek-v4-flash", "hello world", 32768, "HIGH")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if decoded["model"] != "deepseek-v4-flash" {
		t.Errorf("model = %v", decoded["model"])
	}
	if decoded["max_tokens"] != float64(32768) {
		t.Errorf("max_tokens = %v", decoded["max_tokens"])
	}
	if decoded["reasoning_effort"] != "HIGH" {
		t.Errorf("reasoning_effort = %v", decoded["reasoning_effort"])
	}
	msgs, _ := decoded["messages"].([]any)
	if len(msgs) != 1 {
		t.Fatalf("messages = %v, want one", msgs)
	}
	first, _ := msgs[0].(map[string]any)
	if first["role"] != "user" || first["content"] != "hello world" {
		t.Errorf("message = %v, want user/hello world", first)
	}

	body, err = requestBody("m", "p", 0, "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	decoded = map[string]any{}
	_ = json.Unmarshal(body, &decoded)
	if _, ok := decoded["max_tokens"]; ok {
		t.Error("max_tokens emitted for 0")
	}
	if _, ok := decoded["reasoning_effort"]; ok {
		t.Error("reasoning_effort emitted for empty level")
	}
}

// TestRequestHeaders pins auth + merged headers.
func TestRequestHeaders(t *testing.T) {
	h := requestHeaders("secret", map[string]string{"X-Trace": "abc"})
	if h["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q", h["Content-Type"])
	}
	if h["Authorization"] != "Bearer secret" {
		t.Errorf("Authorization = %q", h["Authorization"])
	}
	if h["X-Trace"] != "abc" {
		t.Errorf("X-Trace = %q", h["X-Trace"])
	}
	if got := requestHeaders("", nil); got["Authorization"] != "" {
		t.Errorf("Authorization set without a key: %q", got["Authorization"])
	}
}

// TestExtractErrorMessage pins the actionable non-2xx detail (review finding #3).
func TestExtractErrorMessage(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"openai shape", `{"error":{"message":"Incorrect API key provided"}}`, "Incorrect API key provided"},
		{"trimmed", `{"error":{"message":"  rate limited  "}}`, "rate limited"},
		{"no error object", `{"foo":"bar"}`, ""},
		{"not json", `oops`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractErrorMessage([]byte(c.body)); got != c.want {
				t.Errorf("extractErrorMessage = %q, want %q", got, c.want)
			}
		})
	}
}

// TestParseAnswer pins response normalization.
func TestParseAnswer(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{"valid", `{"choices":[{"message":{"content":"the answer"}}]}`, "the answer", false},
		{"empty choices", `{"choices":[]}`, "", true},
		{"missing content", `{"choices":[{"message":{}}]}`, "", true},
		{"blank content", `{"choices":[{"message":{"content":"  "}}]}`, "", true},
		{"invalid json", `not json`, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseAnswer([]byte(c.body))
			if c.wantErr {
				if err == nil {
					t.Fatalf("parseAnswer = %q, nil; want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAnswer error: %v", err)
			}
			if got != c.want {
				t.Errorf("parseAnswer = %q, want %q", got, c.want)
			}
		})
	}
}

// TestCompleteRoundTrip drives Complete against a local server for the success
// and failure paths, asserting the wire request, the typed error, and the
// actionable non-2xx detail (review finding #3).
func TestCompleteRoundTrip(t *testing.T) {
	var gotPath, gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hi"}}]}`))
	}))
	defer srv.Close()

	c := New(Config{ProviderName: "prov", BaseURL: srv.URL, APIKey: "k", Model: "m"})
	resp, err := c.Complete(context.Background(), llm.Request{Prompt: "ping"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != "hi" {
		t.Errorf("Text = %q, want hi", resp.Text)
	}
	if gotPath != "/chat/completions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer k" {
		t.Errorf("auth = %q", gotAuth)
	}
	if !strings.Contains(gotBody, "ping") {
		t.Errorf("body %q missing prompt", gotBody)
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Incorrect API key provided"}}`))
	}))
	defer bad.Close()
	c = New(Config{ProviderName: "prov", BaseURL: bad.URL})
	_, err = c.Complete(context.Background(), llm.Request{Prompt: "ping"})
	var perr *llm.ProviderError
	if err == nil || !errors.As(err, &perr) {
		t.Fatalf("error = %v, want *llm.ProviderError", err)
	}
	if perr.Provider != "prov" {
		t.Errorf("ProviderError.Provider = %q, want prov", perr.Provider)
	}
	if !strings.Contains(perr.Err.Error(), "Incorrect API key provided") {
		t.Errorf("ProviderError.Err = %q, wants the actionable body detail", perr.Err.Error())
	}
}
