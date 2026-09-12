// Package openai is the OpenAI-compatible provider adapter: the first concrete
// transport behind llm.Gateway (round-004 research Decisions 1–5). It speaks the
// Chat Completions wire shape over Go's stdlib net/http — no provider SDK.
package openai

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
	body, err := requestBody(c.cfg.Model, req.Prompt, c.cfg.MaxTokens, c.cfg.ThinkingLevel)
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
	text, err := parseAnswer(raw)
	if err != nil {
		return llm.Response{}, c.wrap(err)
	}
	return llm.Response{Text: text}, nil
}

func (c *Client) wrap(err error) error {
	return &llm.ProviderError{Provider: c.cfg.ProviderName, Err: err}
}

// requestURL appends the Chat Completions path to the provider base URL.
func requestURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/chat/completions"
}

// requestBody builds the JSON request body (pure helper).
func requestBody(model, prompt string, maxTokens int, thinkingLevel string) ([]byte, error) {
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	if maxTokens > 0 {
		payload["max_tokens"] = maxTokens
	}
	if thinkingLevel != "" {
		payload["reasoning_effort"] = thinkingLevel
	}
	return json.Marshal(payload)
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

// parseAnswer extracts choices[0].message.content (pure helper).
func parseAnswer(raw []byte) (string, error) {
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", fmt.Errorf("unreadable provider response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("provider response carried no usable answer")
	}
	text := decoded.Choices[0].Message.Content
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("provider response carried no usable answer")
	}
	return text, nil
}
