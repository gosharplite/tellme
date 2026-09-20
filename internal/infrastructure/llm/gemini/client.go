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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
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

// extractErrorMessage pulls the provider's structured error message out of a
// Vertex error body so a non-2xx detail is actionable. Vertex returns
// {"error":{"code":...,"message":"...","status":"..."}}.
func extractErrorMessage(raw []byte) string {
	var decoded struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err == nil {
		if msg := strings.TrimSpace(decoded.Error.Message); msg != "" {
			return msg
		}
	}
	return ""
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
//
// Round 065 (ADR 0035; closes #132): a model round's tool results are BATCHED.
// When a `model` turn carries N `functionCall` parts, the next N `tool` results
// are emitted as ONE `user` turn carrying N `functionResponse` parts (in call
// order) — Vertex requires the responses to a function-call turn to share one
// turn (`len(functionResponse) == len(functionCall)` per turn). Round-scoped
// media: the round's media messages are buffered and emitted as their own
// standalone `user` turns AFTER the batched function-response turn (so the
// #1441 inlineData/functionResponse ordering hazard still cannot arise). The
// batching is invisible for N == 1 and for media-free rounds, so those paths are
// shape-identical to before (I-2/I-3); the OpenAI-compatible family is untouched.
// Round 066 adds the id axis (below); the round-065 FIFO name matching it named
// is now the fallback of the id-keyed pairing.
//
// A round boundary is a `model` turn, a plain-text message, or the prompt. A
// tool result is recognised by its `ToolCallID` or its `tool` role (every live
// producer sets BOTH: the loop's `tool` message and the replay path's
// synthesised `call_step_<n>`). A result with no id (or an id that matches no
// call of the round) pairs by the FIFO name fallback and emits no `id` key — so
// an older/id-less persisted step still serializes as a valid functionResponse
// rather than a stray text turn.
//
// Round 066 (ADR 0036; closes #134): the round's parts are ID-LINKED on the
// wire — every emitted `functionCall` carries the call's `id`, and every
// `functionResponse` carries its result's `id` (omitted when empty). Each result
// is bound to its call BY `ToolCallID` (order-independent), so the pairing
// survives a future out-of-order/concurrent result set; the positional FIFO
// name match is retained as the FALLBACK for a result with no id (or an id that
// matches no call of the round). The emitted batched turn still lists the round's
// results in CALL order, so the wire shape is unchanged (the added `id` key is
// the sole difference).
func buildContents(prompt string, prior []llm.Message) []map[string]any {
	return buildRound(prompt, prior)
}

// UnpairedCallIDs returns, in call order across the whole prior, the ids of the
// Gemini/Vertex tool calls that received no result by their round boundary — the
// M < N boundary drop made ACCOUNTABLE (round 067; ADR 0037). Round 068 (ADR
// 0038) makes the account single-owned and family-neutral: this accessor
// DELEGATES to llm.UnpairedToolCalls, and the account's live consumer is the CLI
// diagnostic decorator (internal/cli/unpaired_gateway.go) over that single owner,
// which renders a `[Tool Warning]` line for a short round (delivering ADR 0037
// §Forward RF-067-1). This wrapper is retained as the adapter-facing seam (its
// pins carry the delegation); the emitted batched turn still carries only the M
// results the round produced. A round whose calls are all paired contributes
// nothing.
func UnpairedCallIDs(prior []llm.Message) []string {
	// Round 068 (ADR 0038): the round-boundary account has ONE owner, the
	// family-neutral llm.UnpairedToolCalls; the adapter delegates to it (the
	// property pin TestUnpairedCallIDs_AgreesWithEmittedBody ties the account to
	// the emitted body, since the adapter no longer shares the account's walk).
	return llm.UnpairedToolCalls(prior)
}

// buildRound is the single builder: it walks the prior once and returns the
// Vertex `contents` (with the prompt appended when non-empty). The unpaired-call
// account is single-owned by the family-neutral llm.UnpairedToolCalls (round 068;
// ADR 0038) — the adapter's UnpairedCallIDs delegates there.
func buildRound(prompt string, prior []llm.Message) []map[string]any {
	b := newRoundBuilder()
	b.consume(prior)
	b.flush()
	if prompt != "" {
		b.contents = append(b.contents, map[string]any{"role": "user", "parts": []map[string]any{{"text": prompt}}})
	}
	return b.contents
}

// newRoundBuilder builds an empty round builder.
func newRoundBuilder() *roundBuilder {
	return &roundBuilder{resultParts: map[int]map[string]any{}}
}

// consume walks the prior conversation, buffering each round's calls/results and
// flushing a round at its boundary (a `model` turn, a plain-text turn, or the
// prompt). Shared by the single builder buildRound so the walk cannot drift.
func (b *roundBuilder) consume(prior []llm.Message) {
	for _, m := range prior {
		switch {
		case len(m.ToolCalls) > 0:
			b.modelTurn(m.ToolCalls)
		case m.ToolCallID != "" || (m.Role == "tool" && len(m.Media) == 0):
			b.result(m)
		case len(m.Media) > 0:
			// Round 063 (ADR 0033 D2/D3): a media-bearing message that is NOT a
			// tool result is its OWN `user` turn — media lead the turn. Round 065
			// buffers it so the round's media turns are emitted AFTER the batched
			// function-response turn; it is never merged INTO a functionResponse
			// turn, so the Vertex parser's ordering hazard (#1441: a user turn
			// whose functionResponse precedes an inlineData part) cannot arise by
			// construction. (The tool-result case above is restricted to a
			// media-free `tool` message, so a media-bearing one is still carried —
			// I-4: never silently lose an image.)
			b.mediaTurns = append(b.mediaTurns, inlineDataParts(m.Content, m.Media))
		default:
			b.textTurn(m)
		}
	}
}

// callEntry is one function call of the current round awaiting its result.
type callEntry struct {
	id   string // the call's wire id (may be empty)
	name string
	used bool
}

// roundBuilder accumulates a round's `contents` and buffers the round's tool
// results + media until the round boundary (see [buildContents]).
type roundBuilder struct {
	contents     []map[string]any
	pending      []callEntry            // this round's function calls awaiting their result
	resultParts  map[int]map[string]any // one functionResponse PART per matched call index
	extraResults []map[string]any       // results with no call to bind (kept, name "")
	mediaTurns   [][]map[string]any     // this round's standalone media turns
}

// flush emits the buffered round: the batched function-response turn (when any
// results were buffered) — its parts in CALL order, with any unbound results
// appended — followed by the round's standalone media turns. It also drops EVERY
// call still pending at the boundary: a round boundary means no further result
// can arrive for the flushed round, so a call left unpaired (a round that yielded
// M < N results) contributes no part — the batched turn carries the M parts the
// round actually produced (the doc-recorded shape — a provider that rejects it
// surfaces the same 400, never a silent mispair).
func (b *roundBuilder) flush() {
	if len(b.resultParts) > 0 || len(b.extraResults) > 0 {
		parts := make([]map[string]any, 0, len(b.pending))
		for i := range b.pending {
			if p, ok := b.resultParts[i]; ok {
				parts = append(parts, p)
			}
		}
		parts = append(parts, b.extraResults...)
		b.contents = append(b.contents, map[string]any{"role": "user", "parts": parts})
	}
	b.resultParts = map[int]map[string]any{}
	b.extraResults = nil
	b.pending = nil
	for _, parts := range b.mediaTurns {
		b.contents = append(b.contents, map[string]any{"role": "user", "parts": parts})
	}
	b.mediaTurns = make([][]map[string]any, 0)
}

// modelTurn flushes the previous round, emits this round's `model` turn, and
// records its calls for pairing.
func (b *roundBuilder) modelTurn(calls []llm.ToolCall) {
	b.flush() // a new model turn starts a new round
	parts := make([]map[string]any, 0, len(calls))
	for _, tc := range calls {
		b.pending = append(b.pending, callEntry{id: tc.ID, name: tc.Name})
		parts = append(parts, functionCallPart(tc))
	}
	b.contents = append(b.contents, map[string]any{"role": "model", "parts": parts})
}

// functionCallPart builds a `model` turn's functionCall part (with the call's id
// when non-empty + the replayed `thoughtSignature` when present).
func functionCallPart(tc llm.ToolCall) map[string]any {
	var args any = map[string]any{}
	if strings.TrimSpace(tc.Arguments) != "" {
		_ = json.Unmarshal([]byte(tc.Arguments), &args)
	}
	fc := map[string]any{"name": tc.Name, "args": args}
	if tc.ID != "" {
		fc["id"] = tc.ID
	}
	part := map[string]any{"functionCall": fc}
	// Gemini 3 requires the model's `thoughtSignature` to be echoed back on the
	// replayed functionCall part (else HTTP 400).
	if tc.Signature != "" {
		part["thoughtSignature"] = tc.Signature
	}
	return part
}

// result buffers a tool result, bound to its call by `ToolCallID` (else the FIFO
// fallback), for the current round (not emitted yet). The response `id` is
// emitted ONLY when it equals the id of the call it bound to (the reference's
// `response.id == call.id` invariant) — a foreign id (an unmatched result whose
// id is on no call of the round) is omitted, so the wire never carries a
// `functionResponse.id` that is absent from the round's `functionCall` parts.
func (b *roundBuilder) result(m llm.Message) {
	idx := b.bind(m.ToolCallID)
	fr := map[string]any{"name": "", "response": map[string]any{"content": m.Content}}
	if idx < 0 {
		b.extraResults = append(b.extraResults, map[string]any{"functionResponse": fr})
		return
	}
	b.pending[idx].used = true
	fr["name"] = b.pending[idx].name
	if m.ToolCallID != "" && m.ToolCallID == b.pending[idx].id {
		fr["id"] = m.ToolCallID
	}
	b.resultParts[idx] = map[string]any{"functionResponse": fr}
}

// bind returns the index of the call a result pairs with — by `ToolCallID`
// (identity, order-independent), else the FIFO fallback by arrival — or -1.
func (b *roundBuilder) bind(id string) int {
	if id != "" {
		for i := range b.pending { // primary: pair by identity
			if !b.pending[i].used && b.pending[i].id == id {
				return i
			}
		}
	}
	for i := range b.pending { // fallback: FIFO by arrival (id-less / unmatched)
		if !b.pending[i].used {
			return i
		}
	}
	return -1
}

// textTurn flushes the previous round and emits a plain-text turn.
func (b *roundBuilder) textTurn(m llm.Message) {
	b.flush()
	b.contents = append(b.contents, map[string]any{
		"role":  vertexRole(m.Role),
		"parts": []map[string]any{{"text": m.Content}},
	})
}

// inlineDataParts builds a `user` turn's parts for a media-bearing message
// (round 063; ADR 0033 D3): an optional leading text part, then one `inlineData`
// blob per media part. The keys are camelCase proto-JSON (`inlineData`/
// `mimeType`) — consistent with the adapter's other emitted keys — and the data
// is base64 StdEncoding (the Vertex REST proto-JSON form). The MIME kind is
// resolved upstream by the shared `read_image` sniff, so no second sniffer exists.
func inlineDataParts(text string, media []domaintools.MediaPart) []map[string]any {
	parts := make([]map[string]any, 0, len(media)+1)
	if text != "" {
		parts = append(parts, map[string]any{"text": text})
	}
	for _, mp := range media {
		parts = append(parts, map[string]any{
			"inlineData": map[string]any{
				"mimeType": mp.MIMEType,
				"data":     base64.StdEncoding.EncodeToString(mp.Data),
			},
		})
	}
	return parts
}

// buildGenerationConfig builds the Vertex `generationConfig` (output cap +
// thinking config), returning nil when nothing is set. Vertex rejects
// `thinkingBudget` and `thinkingLevel` together ("thinking_budget and
// thinking_level are not supported together"), so exactly one is sent: the level
// when configured (the Gemini 3 knob), else the budget.
func buildGenerationConfig(maxTokens, thinkingBudget int, thinkingLevel string) map[string]any {
	gen := map[string]any{}
	if maxTokens > 0 {
		gen["maxOutputTokens"] = maxTokens
	}
	thinking := map[string]any{}
	switch {
	case thinkingLevel != "":
		thinking["thinkingLevel"] = thinkingLevel
	case thinkingBudget > 0:
		thinking["thinkingBudget"] = thinkingBudget
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
// declarations, returning nil when none are offered. Each declaration's
// parameters are projected onto the provider's supported schema surface
// (round 061 / ADR 0031) — Vertex parses `parameters` as a CLOSED proto, so an
// unprojected keyword (an MCP server's annotation, `anyOf`, …) fails the whole
// request.
func buildToolDeclarations(toolDefs []llm.ToolDef) []map[string]any {
	if len(toolDefs) == 0 {
		return nil
	}
	decls := make([]map[string]any, 0, len(toolDefs))
	for _, td := range toolDefs {
		params := td.Parameters
		if len(params) == 0 {
			params = json.RawMessage(freeformParameters)
		}
		decls = append(decls, map[string]any{"name": td.Name, "description": td.Description, "parameters": ProjectSchema(params)})
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

// geminiFinishReasonMaxTokens is the Vertex/Gemini finish reason for an
// output-cap truncation (round 030): the output budget was exhausted.
const geminiFinishReasonMaxTokens = "MAX_TOKENS"

// geminiTruncationError is the generic output-cap-truncation error (round 030).
func geminiTruncationError() error {
	return fmt.Errorf("response truncated at %s: the output budget was exhausted before the model finished; increase MAX_TOKENS or shorten the prompt", geminiFinishReasonMaxTokens)
}

// geminiFunctionCallTruncationError is the function-call-aware output-cap
// truncation error (round 030 / FR-003): it names the tool whose arguments were
// cut off and cannot be safely dispatched.
func geminiFunctionCallTruncationError(tool string) error {
	return fmt.Errorf("response truncated at %s during function call (tool=%q): the tool arguments are incomplete and cannot be safely dispatched; increase MAX_TOKENS or break the call up", geminiFinishReasonMaxTokens, tool)
}

// callID resolves a function call's wire id (round 067; ADR 0037; RF-066-2): the
// provider's own `functionCall.id` when it is present (after trimming surrounding
// whitespace) and non-empty — reference parity (`fromSDKFunctionCall` reads
// `f.ID` first) — else the deterministic positional `call_<n>` fallback
// (unchanged from pre-067). A blank/whitespace-only provider id is treated as
// ABSENT (never an empty `id` on the wire — ADR 0036 D2/D4); a present one is
// returned TRIMMED, so the two whitespace cases share one normalisation (F-067-4)
// and the echoed id never carries stray padding. The fallback spelling stays
// `call_<n>`; the reference's `gemini-call-<index>-<name>` is deliberately NOT
// adopted.
func callID(callIdx int, providerID string) string {
	if id := strings.TrimSpace(providerID); id != "" {
		return id
	}
	return fmt.Sprintf("call_%d", callIdx)
}

// parseResponse extracts candidates[0].content.parts (text + functionCall) and
// usageMetadata (pure helper). A response must carry either answer text or at
// least one function call; an empty response is an error.
func parseResponse(raw []byte) (llm.Response, error) {
	var decoded struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text             string `json:"text"`
					ThoughtSignature string `json:"thoughtSignature"`
					FunctionCall     *struct {
						ID   string          `json:"id"`
						Name string          `json:"name"`
						Args json.RawMessage `json:"args"`
					} `json:"functionCall"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata *struct {
			PromptTokenCount        int `json:"promptTokenCount"`
			CandidatesTokenCount    int `json:"candidatesTokenCount"`
			TotalTokenCount         int `json:"totalTokenCount"`
			CachedContentTokenCount int `json:"cachedContentTokenCount"`
			ThoughtsTokenCount      int `json:"thoughtsTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return llm.Response{}, fmt.Errorf("unreadable provider response: %w", err)
	}
	if len(decoded.Candidates) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	// Round 030 — an output-cap truncation is a loud failure, function-call-aware.
	if decoded.Candidates[0].FinishReason == geminiFinishReasonMaxTokens {
		for _, p := range decoded.Candidates[0].Content.Parts {
			if p.FunctionCall != nil {
				return llm.Response{}, geminiFunctionCallTruncationError(p.FunctionCall.Name)
			}
		}
		return llm.Response{}, geminiTruncationError()
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
				ID:        callID(callIdx, p.FunctionCall.ID),
				Name:      p.FunctionCall.Name,
				Arguments: args,
				Signature: p.ThoughtSignature,
			})
			continue
		}
		sb.WriteString(p.Text)
	}
	resp.Text = sb.String()
	if decoded.UsageMetadata != nil {
		um := decoded.UsageMetadata
		// Round 072 (ADR 0044; closes #149): decode the CACHED and THINKING counts.
		// Gemini reports the reused-conversation portion under
		// `cachedContentTokenCount`; omitting it made `CachedTokens` 0, so the miss
		// (`prompt − cached`) billed the whole prompt at the MISS rate (~10×). The
		// DISJOINTNESS is family-specific: Gemini's `candidatesTokenCount` EXCLUDES
		// `thoughtsTokenCount` (total = prompt + candidates + thoughts), so — unlike
		// the OpenAI-compatible adapter, whose `completion_tokens` INCLUDES reasoning
		// — there is NO subtraction here. Every count is floored at 0 defensively.
		resp.Usage = llm.Usage{
			Reported:         true,
			PromptTokens:     max(0, um.PromptTokenCount),
			CachedTokens:     max(0, um.CachedContentTokenCount),
			CompletionTokens: max(0, um.CandidatesTokenCount),
			ThinkingTokens:   max(0, um.ThoughtsTokenCount),
			TotalTokens:      max(0, um.TotalTokenCount),
		}
	}
	if strings.TrimSpace(resp.Text) == "" && len(resp.ToolCalls) == 0 {
		return llm.Response{}, fmt.Errorf("provider response carried no usable answer")
	}
	return resp, nil
}
