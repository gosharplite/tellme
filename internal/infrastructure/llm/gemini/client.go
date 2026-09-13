// Package gemini is the Vertex AI Gemini provider adapter: the second concrete
// transport behind llm.Gateway (round-013 research Decisions 1–2). It speaks the
// Vertex AI `:generateContent` wire shape over Go's stdlib net/http — no provider
// SDK — and authenticates with a service-account key file through a stdlib
// OAuth2 flow (auth.go; round-013 research Decision 3).
//
// The endpoint is built directly from the configured URL with NO hostname
// validation, so a loopback Vertex-shaped fake passes transparently in E2E
// (research Decision 2 / review D2). Prior conversation messages are re-mapped to
// Vertex `contents`: assistant tool calls under `role: "model"` (`functionCall`),
// tool results under `role: "user"` (`functionResponse`) — Vertex REST rejects
// `role: "tool"` (review D1).
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

const (
	defaultTimeout   = 300 * time.Second
	maxResponseBytes = 32 << 20 // 32 MiB
)

// Config binds a resolved gemini provider entry to the adapter.
type Config struct {
	ProviderName   string
	BaseURL        string
	APIKey         string // path to a service-account JSON credential (ends in .json)
	Model          string
	MaxTokens      int
	Headers        map[string]string
	ThinkingBudget int
	ThinkingLevel  string
	Persona        string
}

// Client is the Vertex/Gemini adapter.
type Client struct {
	cfg   Config
	http  *http.Client
	token *tokenSource
}

var _ llm.Gateway = (*Client)(nil)

// New builds the adapter. It reads and validates the service-account credential
// NOW (round-013 research Decision 4) — a missing / unreadable / non-`.json` /
// invalid key file is a credential failure surfaced as a *llm.ProviderError (the
// frozen class phrase `the provider request failed`, exit 6).
func New(cfg Config) (*Client, error) {
	c := &Client{cfg: cfg, http: &http.Client{Timeout: defaultTimeout}}
	ts, err := newTokenSource(cfg.APIKey, c.http)
	if err != nil {
		return nil, &llm.ProviderError{Provider: cfg.ProviderName, Err: err}
	}
	c.token = ts
	return c, nil
}

// Complete sends exactly one non-streaming Vertex `:generateContent` request and
// returns the normalized answer. Every failure is wrapped in a *llm.ProviderError.
func (c *Client) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	body, err := requestBody(req.Prompt, req.Messages, req.Tools, c.cfg.MaxTokens, c.cfg.ThinkingBudget, c.cfg.ThinkingLevel, c.cfg.Persona)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL(c.cfg.BaseURL, c.cfg.Model), bytes.NewReader(body))
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	token, err := c.token.Token(ctx)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	for k, v := range requestHeaders(token, c.cfg.Headers) {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return llm.Response{}, c.wrap(fmt.Errorf("provider returned status %d", resp.StatusCode))
	}
	answer, err := parseResponse(raw)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	return answer, nil
}

func (c *Client) wrap(err error) error {
	return &llm.ProviderError{Provider: c.cfg.ProviderName, Err: err}
}

// requestURL appends the model + `:generateContent` to the configured base URL.
// No hostname validation (research Decision 2 / review D2): the endpoint is built
// directly from the configured URL so a loopback Vertex-shaped fake passes.
func requestURL(baseURL, model string) string {
	return strings.TrimRight(baseURL, "/") + "/" + model + ":generateContent"
}

// requestBody builds the Vertex JSON request body (pure helper). `contents` is
// the mapped conversation (prior messages + the current prompt when non-empty);
// the persona becomes `systemInstruction`; `generationConfig` carries the output
// cap and the thinking config only when set. When no tools are offered the body
// carries no `tools` key.
func requestBody(prompt string, prior []llm.Message, toolDefs []llm.ToolDef, maxTokens, thinkingBudget int, thinkingLevel, persona string) ([]byte, error) {
	payload := map[string]any{"contents": buildContents(prompt, prior)}
	if persona != "" {
		payload["systemInstruction"] = map[string]any{"parts": []map[string]any{{"text": persona}}}
	}
	if gen := buildGenerationConfig(maxTokens, thinkingBudget, thinkingLevel); gen != nil {
		payload["generationConfig"] = gen
	}
	if decls := buildToolDeclarations(toolDefs); decls != nil {
		payload["tools"] = []map[string]any{{"functionDeclarations": decls}}
	}
	return json.Marshal(payload)
}

// buildContents maps the conversation (prior messages + the current prompt) to
// Vertex `contents`: assistant tool calls under `role:"model"` functionCall, tool
// results under `role:"user"` functionResponse (Vertex rejects `role:"tool"`),
// and text turns under their mapped role.
func buildContents(prompt string, prior []llm.Message) []map[string]any {
	contents := make([]map[string]any, 0, len(prior)+1)
	pending := make([]string, 0) // names of tool calls awaiting their result
	for _, m := range prior {
		switch {
		case len(m.ToolCalls) > 0:
			parts := make([]map[string]any, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				pending = append(pending, tc.Name)
				var args any = map[string]any{}
				if strings.TrimSpace(tc.Arguments) != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &args)
				}
				parts = append(parts, map[string]any{"functionCall": map[string]any{"name": tc.Name, "args": args}})
			}
			contents = append(contents, map[string]any{"role": "model", "parts": parts})
		case m.ToolCallID != "":
			name := ""
			if len(pending) > 0 {
				name = pending[0]
				pending = pending[1:]
			}
			contents = append(contents, map[string]any{
				"role":  "user",
				"parts": []map[string]any{{"functionResponse": map[string]any{"name": name, "response": map[string]any{"content": m.Content}}}},
			})
		default:
			contents = append(contents, map[string]any{
				"role":  vertexRole(m.Role),
				"parts": []map[string]any{{"text": m.Content}},
			})
		}
	}
	if prompt != "" {
		contents = append(contents, map[string]any{"role": "user", "parts": []map[string]any{{"text": prompt}}})
	}
	return contents
}

// buildGenerationConfig builds the Vertex `generationConfig` (output cap +
// thinking config), returning nil when nothing is set.
func buildGenerationConfig(maxTokens, thinkingBudget int, thinkingLevel string) map[string]any {
	gen := map[string]any{}
	if maxTokens > 0 {
		gen["maxOutputTokens"] = maxTokens
	}
	thinking := map[string]any{}
	if thinkingBudget > 0 {
		thinking["thinkingBudget"] = thinkingBudget
	}
	if thinkingLevel != "" {
		thinking["thinkingLevel"] = thinkingLevel
	}
	if len(thinking) > 0 {
		gen["thinkingConfig"] = thinking
	}
	if len(gen) == 0 {
		return nil
	}
	return gen
}

// buildToolDeclarations maps the tool definitions to Vertex function
// declarations, returning nil when none are offered.
func buildToolDeclarations(toolDefs []llm.ToolDef) []map[string]any {
	if len(toolDefs) == 0 {
		return nil
	}
	decls := make([]map[string]any, 0, len(toolDefs))
	for _, td := range toolDefs {
		params := td.Parameters
		if len(params) == 0 {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		decls = append(decls, map[string]any{"name": td.Name, "description": td.Description, "parameters": params})
	}
	return decls
}

// vertexRole maps an llm.Message role to a Vertex `contents` role.
func vertexRole(role string) string {
	if role == "assistant" || role == "model" {
		return "model"
	}
	return "user"
}

// requestHeaders builds the outbound headers (pure helper). The service-account
// access token is attached as `Authorization: Bearer`.
func requestHeaders(token string, extra map[string]string) map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	for k, v := range extra {
		headers[k] = v
	}
	return headers
}

// parseResponse extracts candidates[0].content.parts (text + functionCall) and
// usageMetadata (pure helper). A response must carry either answer text or at
// least one function call; an empty response is an error.
func parseResponse(raw []byte) (llm.Response, error) {
	var decoded struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string `json:"text"`
					FunctionCall *struct {
						Name string          `json:"name"`
						Args json.RawMessage `json:"args"`
					} `json:"functionCall"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return llm.Response{}, fmt.Errorf("unreadable provider response: %w", err)
	}
	if len(decoded.Candidates) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	var resp llm.Response
	var sb strings.Builder
	callIdx := 0
	for _, p := range decoded.Candidates[0].Content.Parts {
		if p.FunctionCall != nil {
			callIdx++
			args := strings.TrimSpace(string(p.FunctionCall.Args))
			if args == "" {
				args = "{}"
			}
			resp.ToolCalls = append(resp.ToolCalls, llm.ToolCall{
				ID:        fmt.Sprintf("call_%d", callIdx),
				Name:      p.FunctionCall.Name,
				Arguments: args,
			})
			continue
		}
		sb.WriteString(p.Text)
	}
	resp.Text = sb.String()
	if decoded.UsageMetadata != nil {
		resp.Usage = llm.Usage{
			Reported:         true,
			PromptTokens:     decoded.UsageMetadata.PromptTokenCount,
			CompletionTokens: decoded.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      decoded.UsageMetadata.TotalTokenCount,
		}
	}
	if strings.TrimSpace(resp.Text) == "" && len(resp.ToolCalls) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	return resp, nil
}
