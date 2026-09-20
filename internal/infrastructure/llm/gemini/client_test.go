package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

func tsNow() time.Time { return time.Now() }

func b64urlDecode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func TestRequestURL_FromConfiguredURLNoHostnameCheck(t *testing.T) {
	got := requestURL("http://127.0.0.1:9/v1/projects/e2e/locations/global/publishers/google/models/", "gemini-3.8-flash")
	want := "http://127.0.0.1:9/v1/projects/e2e/locations/global/publishers/google/models/gemini-3.8-flash:generateContent"
	if got != want {
		t.Errorf("requestURL = %q, want %q", got, want)
	}
}

func TestRequestBody_TextPersonaBudgetTools(t *testing.T) {
	body, err := requestBody(
		"hi", nil,
		[]llm.ToolDef{{Name: "read_files", Description: "read a file", Parameters: json.RawMessage(`{"type":"object"}`)}},
		40960, 32768, "HIGH", "be terse",
	)
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	si, _ := decoded["systemInstruction"].(map[string]any)
	if si == nil {
		t.Fatalf("systemInstruction missing: %s", body)
	}
	gc, _ := decoded["generationConfig"].(map[string]any)
	if gc == nil {
		t.Fatalf("generationConfig missing: %s", body)
	}
	if gc["maxOutputTokens"].(float64) != 40960 {
		t.Errorf("maxOutputTokens = %v, want 40960", gc["maxOutputTokens"])
	}
	tcfg, _ := gc["thinkingConfig"].(map[string]any)
	if tcfg == nil || tcfg["thinkingLevel"] != "HIGH" {
		t.Errorf("thinkingConfig = %v, want thinkingLevel HIGH", gc["thinkingConfig"])
	}
	// Vertex rejects thinkingBudget + thinkingLevel together, so only the level
	// is sent when both are configured.
	if _, present := tcfg["thinkingBudget"]; present {
		t.Errorf("thinkingConfig must not carry thinkingBudget alongside thinkingLevel: %v", tcfg)
	}
	tools, _ := decoded["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools = %v", decoded["tools"])
	}
	tool0, _ := tools[0].(map[string]any)
	decls, _ := tool0["functionDeclarations"].([]any)
	if len(decls) != 1 {
		t.Fatalf("functionDeclarations = %v", tool0["functionDeclarations"])
	}
	d0, _ := decls[0].(map[string]any)
	if d0["name"] != "read_files" {
		t.Errorf("declaration name = %v, want read_files", d0["name"])
	}
	contents, _ := decoded["contents"].([]any)
	if len(contents) != 1 {
		t.Fatalf("contents = %v", contents)
	}
	c0, _ := contents[0].(map[string]any)
	if c0["role"] != "user" {
		t.Errorf("prompt role = %v, want user", c0["role"])
	}
}

func TestRequestBody_ZeroBudgetAndThinkingOmitted(t *testing.T) {
	body, err := requestBody("hi", nil, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	if strings.Contains(string(body), "generationConfig") {
		t.Errorf("body carries a generationConfig with no budget/thinking set: %s", body)
	}
	if strings.Contains(string(body), "systemInstruction") {
		t.Errorf("body carries a systemInstruction with no persona: %s", body)
	}
	if strings.Contains(string(body), `"tools"`) {
		t.Errorf("body carries tools when none are declared: %s", body)
	}
}

func TestRequestBody_ThinkingBudgetOnly(t *testing.T) {
	body, err := requestBody("hi", nil, nil, 0, 32768, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		GenerationConfig struct {
			ThinkingConfig map[string]any `json:"thinkingConfig"`
		} `json:"generationConfig"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.GenerationConfig.ThinkingConfig["thinkingBudget"].(float64) != 32768 {
		t.Errorf("thinkingConfig = %v, want thinkingBudget 32768", decoded.GenerationConfig.ThinkingConfig)
	}
	if _, present := decoded.GenerationConfig.ThinkingConfig["thinkingLevel"]; present {
		t.Errorf("thinkingConfig must not carry thinkingLevel when only a budget is set: %v", decoded.GenerationConfig.ThinkingConfig)
	}
}

func TestRequestBody_ToolExchangeMapping(t *testing.T) {
	prior := []llm.Message{
		{Role: "user", Content: "read it"},
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"name":"notes.txt"}`, Signature: "sig-1"}}},
		{Role: "tool", Content: "ORANGE", ToolCallID: "call_1"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	if strings.Contains(string(body), `"role":"tool"`) {
		t.Fatalf("body carries the OpenAI-only role \"tool\" (Vertex rejects it): %s", body)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Contents) != 3 {
		t.Fatalf("contents len = %d, want 3: %s", len(decoded.Contents), body)
	}
	if decoded.Contents[1].Role != "model" || decoded.Contents[1].Parts[0]["functionCall"] == nil {
		t.Errorf("assistant tool call not mapped to model/functionCall: %+v", decoded.Contents[1])
	}
	if decoded.Contents[1].Parts[0]["thoughtSignature"] != "sig-1" {
		t.Errorf("the replayed functionCall part must echo the thoughtSignature: %+v", decoded.Contents[1].Parts[0])
	}
	if decoded.Contents[2].Role != "user" {
		t.Errorf("tool result role = %q, want user", decoded.Contents[2].Role)
	}
	fr, _ := decoded.Contents[2].Parts[0]["functionResponse"].(map[string]any)
	if fr == nil || fr["name"] != "read_files" {
		t.Errorf("tool result not mapped to functionResponse with its tool name: %+v", decoded.Contents[2])
	}
}

func TestParseResponse_TextToolCallUsage(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"4"},{"functionCall":{"name":"read_files","args":{"name":"notes.txt"}},"thoughtSignature":"sig-1"}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":2,"totalTokenCount":13}}`)
	resp, err := parseResponse(raw)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if resp.Text != "4" {
		t.Errorf("text = %q, want 4", resp.Text)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "read_files" || resp.ToolCalls[0].ID == "" {
		t.Errorf("tool calls = %+v, want one named read_files with an id", resp.ToolCalls)
	}
	if resp.ToolCalls[0].Signature != "sig-1" {
		t.Errorf("tool call signature = %q, want sig-1 (the thoughtSignature must be captured)", resp.ToolCalls[0].Signature)
	}
	if !resp.Usage.Reported || resp.Usage.PromptTokens != 11 || resp.Usage.TotalTokens != 13 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestParseResponse_NoUsableAnswer(t *testing.T) {
	if _, err := parseResponse([]byte(`{"candidates":[]}`)); err == nil {
		t.Error("empty candidates: expected an error")
	}
	if _, err := parseResponse([]byte(`not json`)); err == nil {
		t.Error("malformed body: expected an error")
	}
}

func TestClient_Complete_SendsServiceAccountTokenAndParsesAnswer(t *testing.T) {
	var tokenHits int32
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenHits, 1)
		_, _ = w.Write([]byte(`{"access_token":"tok-abc","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var genHits int32
	var authHeader string
	genSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&genHits, 1)
		authHeader = r.Header.Get("Authorization")
		if !strings.HasSuffix(r.URL.Path, ":generateContent") {
			t.Errorf("path = %q, want ...:generateContent", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"4"}]}}]}`))
	}))
	defer genSrv.Close()

	c, err := New(Config{
		ProviderName: "vertex-flash-3.8",
		BaseURL:      genSrv.URL + "/v1/projects/e2e/locations/global/publishers/google/models",
		APIKey:       writeTestKey(t, tokenSrv.URL),
		Model:        "gemini-3.8-flash",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for i := 0; i < 2; i++ {
		resp, err := c.Complete(context.Background(), llm.Request{Prompt: "what is 2+2?"})
		if err != nil {
			t.Fatalf("Complete: %v", err)
		}
		if resp.Text != "4" {
			t.Fatalf("answer = %q, want 4", resp.Text)
		}
	}
	if authHeader != "Bearer tok-abc" {
		t.Errorf("Authorization = %q, want Bearer tok-abc", authHeader)
	}
	if h := atomic.LoadInt32(&genHits); h != 2 {
		t.Errorf("generate endpoint hits = %d, want 2", h)
	}
	if h := atomic.LoadInt32(&tokenHits); h != 1 {
		t.Errorf("token endpoint hits = %d, want 1 (token cached and reused)", h)
	}
}

func TestNew_MissingCredentialIsProviderError(t *testing.T) {
	_, err := New(Config{ProviderName: "vertex-flash-3.8", BaseURL: "https://x", APIKey: "/no/such/key.json", Model: "m"})
	if err == nil {
		t.Fatal("expected a credential error")
	}
	var perr *llm.ProviderError
	if !asProviderError(err, &perr) {
		t.Fatalf("error = %T, want *llm.ProviderError", err)
	}
	if perr.Provider != "vertex-flash-3.8" {
		t.Errorf("ProviderError.Provider = %q", perr.Provider)
	}
}

// asProviderError is a tiny local errors.As shim (avoids importing errors twice).
func asProviderError(err error, target **llm.ProviderError) bool {
	for err != nil {
		if pe, ok := err.(*llm.ProviderError); ok {
			*target = pe
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

// TestParseResponse_DecodesCachedAndThinkingTokens pins round 072 (ADR 0044;
// closes #149): the Gemini/Vertex usageMetadata's `cachedContentTokenCount` and
// `thoughtsTokenCount` map into Usage.CachedTokens / Usage.ThinkingTokens, and the
// family rule is DISJOINT — CandidatesTokenCount is preserved verbatim (no
// subtraction of thoughts), unlike the OpenAI-compatible adapter.
func TestParseResponse_DecodesCachedAndThinkingTokens(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":390564,"cachedContentTokenCount":389538,"candidatesTokenCount":100,"thoughtsTokenCount":4096,"totalTokenCount":394760}}`)
	resp, err := parseResponse(raw)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	u := resp.Usage
	if !u.Reported {
		t.Fatal("usage must be reported")
	}
	if u.CachedTokens != 389538 {
		t.Errorf("CachedTokens = %d, want 389538 (cachedContentTokenCount)", u.CachedTokens)
	}
	if u.PromptTokens != 390564 {
		t.Errorf("PromptTokens = %d, want 390564", u.PromptTokens)
	}
	if u.ThinkingTokens != 4096 {
		t.Errorf("ThinkingTokens = %d, want 4096 (thoughtsTokenCount)", u.ThinkingTokens)
	}
	// DISJOINT: Gemini's candidatesTokenCount excludes thoughts, so it is verbatim.
	if u.CompletionTokens != 100 {
		t.Errorf("CompletionTokens = %d, want 100 (verbatim; Gemini is disjoint, no subtraction)", u.CompletionTokens)
	}
	// The cached count materially changes the miss (the whole point).
	if miss := u.PromptTokens - u.CachedTokens; miss != 1026 {
		t.Errorf("miss = %d, want 1026 (prompt - cached); a 0 cache would miss the whole prompt", miss)
	}
}

// TestParseResponse_AbsentCachedFieldsStayZero pins the degenerate path: a
// usageMetadata that omits the cached/thinking fields leaves them 0 (a legal cold
// cache), never negative.
func TestParseResponse_AbsentCachedFieldsStayZero(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":2,"totalTokenCount":13}}`)
	resp, err := parseResponse(raw)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if resp.Usage.CachedTokens != 0 || resp.Usage.ThinkingTokens != 0 {
		t.Errorf("absent fields must stay 0, got cached=%d thinking=%d", resp.Usage.CachedTokens, resp.Usage.ThinkingTokens)
	}
	if resp.Usage.PromptTokens != 11 || resp.Usage.TotalTokens != 13 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}
