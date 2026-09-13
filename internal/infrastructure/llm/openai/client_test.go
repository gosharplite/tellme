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
	body, err := requestBody("deepseek-v4-flash", "hello world", nil, nil, 32768, "HIGH", "")
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

	body, err = requestBody("m", "p", nil, nil, 0, "", "")
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
			resp, err := parseResponse([]byte(c.body))
			if c.wantErr {
				if err == nil {
					t.Fatalf("parseResponse = %q, nil; want error", resp.Text)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseResponse error: %v", err)
			}
			if resp.Text != c.want {
				t.Errorf("parseResponse.Text = %q, want %q", resp.Text, c.want)
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

// TestRequestBody_PriorMessages pins round-007 RF-3: prior conversation messages
// are prepended to the current user prompt in order; an empty prior slice yields
// exactly the single current user message (byte-identical to rounds 004–006).
func TestRequestBody_PriorMessages(t *testing.T) {
	prior := []llm.Message{
		{Role: "user", Content: "my name is alice"},
		{Role: "assistant", Content: "noted"},
	}
	body, err := requestBody("m", "what is my name?", prior, nil, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Messages) != 3 {
		t.Fatalf("messages = %+v, want 3", decoded.Messages)
	}
	want := []struct{ role, content string }{
		{"user", "my name is alice"},
		{"assistant", "noted"},
		{"user", "what is my name?"},
	}
	for i, m := range decoded.Messages {
		if m.Role != want[i].role || m.Content != want[i].content {
			t.Errorf("messages[%d] = %s/%q, want %s/%q", i, m.Role, m.Content, want[i].role, want[i].content)
		}
	}
}

// TestRequestBody_Tools pins the tools array assembly (round-008 research
// Decision 2): each definition emits type/function/name/description/parameters.
func TestRequestBody_Tools(t *testing.T) {
	defs := []llm.ToolDef{{
		Name:        "read_files",
		Description: "Read a file.",
		Parameters:  []byte(`{"type":"object","properties":{"path":{"type":"string"}}}`),
	}}
	body, err := requestBody("m", "read x", nil, defs, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Tools []struct {
			Type     string `json:"type"`
			Function struct {
				Name        string         `json:"name"`
				Description string         `json:"description"`
				Parameters  map[string]any `json:"parameters"`
			} `json:"function"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Tools) != 1 || decoded.Tools[0].Type != "function" || decoded.Tools[0].Function.Name != "read_files" {
		t.Fatalf("tools = %+v, want one read_files function", decoded.Tools)
	}
}

// TestParseResponse_ToolCalls pins structured tool-call extraction (round-008
// research Decision 2): a tool-only response (empty content) is valid and
// carries the id/name/arguments.
func TestParseResponse_ToolCalls(t *testing.T) {
	body := `{"choices":[{"message":{"content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"read_files","arguments":"{\"path\":\"x\"}"}}]}}]}`
	resp, err := parseResponse([]byte(body))
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls = %+v, want one", resp.ToolCalls)
	}
	if resp.ToolCalls[0].ID != "call_1" || resp.ToolCalls[0].Name != "read_files" || resp.ToolCalls[0].Arguments != `{"path":"x"}` {
		t.Errorf("ToolCall = %+v", resp.ToolCalls[0])
	}
}
