// Package openai is the OpenAI-compatible provider adapter: the first concrete
// transport behind llm.Gateway (round-004 research Decisions 1–5). It speaks the
// Chat Completions wire shape over Go's stdlib net/http — no provider SDK.
package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// defaultTimeout bounds a provider request so a stalled endpoint cannot hang the
// CLI indefinitely (review finding #2).
const defaultTimeout = 300 * time.Second

// maxResponseBytes bounds the response body read so a rogue or misconfigured
// upstream cannot exhaust memory (review finding #4).
const maxResponseBytes = 32 << 20 // 32 MiB

// Config binds a resolved provider entry to the adapter.
type Config struct {
	ProviderName  string
	BaseURL       string
	APIKey        string
	Model         string
	MaxTokens     int
	Headers       map[string]string
	ThinkingLevel string
	Persona       string
}

// Client is the OpenAI-compatible adapter.
type Client struct {
	cfg  Config
	http *http.Client
}

var _ llm.Gateway = (*Client)(nil)

// New builds an adapter with a timeout-bounded default HTTP client.
func New(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: defaultTimeout}}
}

// NewWithHTTPClient builds an adapter with an explicit HTTP client.
func NewWithHTTPClient(cfg Config, c *http.Client) *Client {
	if c == nil {
		c = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{cfg: cfg, http: c}
}

// Complete sends exactly one non-streaming Chat Completions request and returns
// the normalized answer (round-004 research Decision 3 & 4). Every failure is
// wrapped in a *llm.ProviderError.
func (c *Client) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	body, err := requestBody(c.cfg.Model, req.Prompt, req.Messages, req.Tools, c.cfg.MaxTokens, c.cfg.ThinkingLevel, c.cfg.Persona)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL(c.cfg.BaseURL), bytes.NewReader(body))
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	for k, v := range requestHeaders(c.cfg.APIKey, c.cfg.Headers) {
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
		if msg := extractErrorMessage(raw); msg != "" {
			return llm.Response{}, c.wrap(fmt.Errorf("provider returned status %d: %s", resp.StatusCode, msg))
		}
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

// requestURL appends the Chat Completions path to the provider base URL.
func requestURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/chat/completions"
}

// requestBody builds the JSON request body (pure helper). The `messages` array
// is `prior` followed — when a non-empty prompt is given — by the current user
// prompt as the LAST message. A prior assistant message carrying tool calls
// emits `tool_calls`, and a prior tool result emits `tool_call_id`. An empty
// prompt is NOT appended: the agent tool loop folds the whole active turn
// (prompt + tool exchanges) into `prior` on its later rounds so the chronology
// stays `user → assistant(tool_calls) → tool(result)` (review PR #25 BLOCKER-1).
// When no tool definitions are given the body is byte-identical to rounds
// 004–007 (round-008 research Decision 2).
func requestBody(model, prompt string, prior []llm.Message, toolDefs []llm.ToolDef, maxTokens int, thinkingLevel, persona string) ([]byte, error) {
	messages := make([]map[string]any, 0, len(prior)+1)
	for _, m := range prior {
		msg := map[string]any{"role": m.Role, "content": messageContent(m)}
		if len(m.ToolCalls) > 0 {
			tcs := make([]map[string]any, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				tcs = append(tcs, map[string]any{
					"id":   tc.ID,
					"type": "function",
					"function": map[string]any{
						"name":      tc.Name,
						"arguments": tc.Arguments,
					},
				})
			}
			msg["tool_calls"] = tcs
		}
		if m.ToolCallID != "" {
			msg["tool_call_id"] = m.ToolCallID
		}
		messages = append(messages, msg)
	}
	if prompt != "" {
		messages = append(messages, map[string]any{"role": "user", "content": prompt})
	}
	// Round 011 — the configured persona is the leading `system` message of the
	// request (round-011 research Decision 1). An empty persona adds nothing.
	if persona != "" {
		messages = append([]map[string]any{{"role": "system", "content": persona}}, messages...)
	}

	payload := map[string]any{
		"model":    model,
		"messages": messages,
	}
	if len(toolDefs) > 0 {
		tools := make([]map[string]any, 0, len(toolDefs))
		for _, td := range toolDefs {
			params := td.Parameters
			if len(params) == 0 {
				params = json.RawMessage(`{"type":"object","properties":{}}`)
			}
			tools = append(tools, map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        td.Name,
					"description": td.Description,
					"parameters":  params,
				},
			})
		}
		payload["tools"] = tools
	}
	if maxTokens > 0 {
		payload["max_tokens"] = maxTokens
	}
	if thinkingLevel != "" {
		payload["reasoning_effort"] = thinkingLevel
	}
	return json.Marshal(payload)
}

// messageContent renders one conversation message's `content` field (round 062;
// ADR 0032). A message with NO media returns the plain string content — so a
// text-only turn is byte-identical to before. A message WITH media returns a
// content ARRAY: a leading text part when the message states text, then one
// inline base64 `image_url` block per media part (media-first).
func messageContent(m llm.Message) any {
	if len(m.Media) == 0 {
		return m.Content
	}
	parts := make([]any, 0, len(m.Media)+1)
	if m.Content != "" {
		parts = append(parts, map[string]any{"type": "text", "text": m.Content})
	}
	for _, mp := range m.Media {
		parts = append(parts, map[string]any{
			"type":      "image_url",
			"image_url": map[string]any{"url": dataURI(mp)},
		})
	}
	return parts
}

// dataURI renders an inline base64 data URI for one media part.
func dataURI(mp llm.MediaPart) string {
	return "data:" + mp.MIMEType + ";base64," + base64.StdEncoding.EncodeToString(mp.Data)
}

// requestHeaders builds the outbound headers (pure helper).
func requestHeaders(apiKey string, extra map[string]string) map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}
	for k, v := range extra {
		headers[k] = v
	}
	return headers
}

// extractErrorMessage pulls the provider's structured error message out of an
// error body (review finding #3), so a non-2xx detail is actionable.
func extractErrorMessage(raw []byte) string {
	var decoded struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err == nil {
		return strings.TrimSpace(decoded.Error.Message)
	}
	return ""
}

// toolCall is the OpenAI wire shape for a model's tool-call request.
type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// openAIFinishReasonLength is the OpenAI-compatible finish reason for an
// output-cap truncation (round 030): the output budget was exhausted.
const openAIFinishReasonLength = "length"

// checkTruncation surfaces an output-cap truncation as a loud error, or returns
// nil for a healthy finish reason (round 030). The call site sits after the
// `len(choices)==0` early-return (a choice-less body carries no finish reason)
// and BEFORE the generic content-empty "no usable answer" check, so an empty +
// truncated response is reported as a truncation — a deliberate inversion of the
// reference's empty-content-first order (research D6).
func checkTruncation(finishReason string) error {
	if finishReason != openAIFinishReasonLength {
		return nil
	}
	return fmt.Errorf("response truncated at max_tokens (finish_reason=%q): the output budget was exhausted before the model finished; increase MAX_TOKENS or shorten the prompt", finishReason)
}

// parseResponse extracts choices[0].message.content and
// choices[0].message.tool_calls (pure helper). A response must carry either
// answer text or at least one tool call; an empty response is an error.
func parseResponse(raw []byte) (llm.Response, error) {
	var decoded struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []toolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			PromptDetails    *struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
			CompletionDetails *struct {
				ReasoningTokens int `json:"reasoning_tokens"`
			} `json:"completion_tokens_details"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return llm.Response{}, fmt.Errorf("unreadable provider response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	// Round 030 — an output-cap truncation is a loud failure (see checkTruncation).
	if err := checkTruncation(decoded.Choices[0].FinishReason); err != nil {
		return llm.Response{}, err
	}
	resp := llm.Response{Text: decoded.Choices[0].Message.Content}
	if decoded.Usage != nil {
		reasoning := 0
		if decoded.Usage.CompletionDetails != nil {
			reasoning = decoded.Usage.CompletionDetails.ReasoningTokens
		}
		cached := 0
		if decoded.Usage.PromptDetails != nil {
			cached = decoded.Usage.PromptDetails.CachedTokens
		}
		resp.Usage = llm.Usage{
			Reported:     true,
			PromptTokens: decoded.Usage.PromptTokens,
			CachedTokens: cached,
			// Round 018 FR-002: the wire `completion_tokens` INCLUDES reasoning,
			// so store the EXCLUSIVE completion (disjoint from ThinkingTokens).
			// Floor at zero: a rogue/exclusive proxy could report a malformed
			// count (round-018 review #2).
			CompletionTokens: max(0, decoded.Usage.CompletionTokens-reasoning),
			ThinkingTokens:   reasoning,
			TotalTokens:      decoded.Usage.TotalTokens,
		}
	}
	for _, tc := range decoded.Choices[0].Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, llm.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}
	if strings.TrimSpace(resp.Text) == "" && len(resp.ToolCalls) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	return resp, nil
}
