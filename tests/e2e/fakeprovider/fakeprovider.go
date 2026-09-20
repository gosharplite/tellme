// Package fakeprovider is a leaf E2E test-support package: an in-process,
// OpenAI-compatible provider used by the chat acceptance scenarios and the
// offline-path canary. It records every request it receives so a step can
// assert a provider was hit exactly once — or, on the offline paths, never — and
// can serve scripted tool-call responses so the agent tool loop is exercisable
// (round-008 T006).
package fakeprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// Reply is one scripted provider response: a tool-call request (ToolName set) or
// a plain answer text (otherwise). Round 008. Round 019 adds Tools: a single
// response carrying several tool-call requests (the several-tool spinner case).
type Reply struct {
	ToolName  string
	Arguments string
	Answer    string
	// Tools, when non-empty, scripts ONE response carrying every listed tool-call
	// request (round 019). Mutually exclusive with ToolName.
	Tools []ToolRequest
	// FinishReason, when non-empty, scripts the response's provider finish reason
	// (round 030): the OpenAI-compatible `choices[0].finish_reason` or the Vertex
	// `candidates[0].finishReason`. "" omits it (a healthy response). A truncation
	// scripts "length" (OpenAI-compatible) / "MAX_TOKENS" (Vertex) to drive the guard.
	FinishReason string
}

// ToolRequest is one tool call in a multi-tool scripted reply (round 019).
type ToolRequest struct {
	Name      string
	Arguments string
}

// Provider is a scriptable OpenAI-compatible fake.
type Provider struct {
	srv *httptest.Server

	mu          sync.Mutex
	answer      string
	errorStatus int
	noAnswer    bool
	script      []Reply
	served      int
	bodies      []string
	auths       []string
	paths       []string
	usage       *usageCounts
	vertex      bool
	closeOnce   sync.Once
}

// usageCounts is a scripted `usage` block the fake reports on its answers
// (round-009 T006). Round 018 adds the cached/reasoning DETAILS and a `detailed`
// flag; when detailed the wire `completion_tokens` is the INCLUSIVE value
// (completion + thinking), matching the OpenAI-compatible transport.
type usageCounts struct {
	prompt     int
	cached     int
	completion int
	thinking   int
	total      int
	detailed   bool
}

// Start launches the fake provider on a loopback listener.
func Start() *Provider {
	p := &Provider{answer: "ok"}
	p.srv = httptest.NewServer(http.HandlerFunc(p.handle))
	return p
}

// URL returns the base endpoint a configuration should point at.
func (p *Provider) URL() string { return p.srv.URL }

// Close stops the server (idempotent).
func (p *Provider) Close() { p.closeOnce.Do(func() { p.srv.Close() }) }

// Answer scripts the provider to reply with a fixed answer text.
func (p *Provider) Answer(text string) { p.mu.Lock(); p.answer = text; p.mu.Unlock() }

// ErrorStatus scripts the provider to reply with a non-2xx status.
func (p *Provider) ErrorStatus(code int) { p.mu.Lock(); p.errorStatus = code; p.mu.Unlock() }

// NoAnswer scripts the provider to reply with a body carrying no usable answer.
func (p *Provider) NoAnswer() { p.mu.Lock(); p.noAnswer = true; p.mu.Unlock() }

// VertexMode switches the fake to the Vertex AI `:generateContent` response shape
// (round-013 T002). The default remains OpenAI-compatible.
func (p *Provider) VertexMode() { p.mu.Lock(); p.vertex = true; p.mu.Unlock() }

// LastAuthHeader returns the Authorization header of the most recent recorded
// completion request ("" when none). Round 013 asserts the service-account token.
func (p *Provider) LastAuthHeader() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.auths) == 0 {
		return ""
	}
	return p.auths[len(p.auths)-1]
}

// PathAt returns the request path of recorded completion i (i < 0 → the last).
// Round 013 uses it to assert the Vertex model (`…/<model>:generateContent`),
// which lives in the URL, not the body.
func (p *Provider) PathAt(i int) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.paths) == 0 {
		return ""
	}
	if i < 0 || i >= len(p.paths) {
		i = len(p.paths) - 1
	}
	return p.paths[i]
}

// ReportUsage scripts the provider to include a `usage` block on its answers
// (round-009 T006), so the CLI's post-turn measured status line can be asserted.
func (p *Provider) ReportUsage(prompt, completion int) {
	p.mu.Lock()
	p.usage = &usageCounts{prompt: prompt, completion: completion, total: prompt + completion}
	p.mu.Unlock()
}

// ReportUsageDetails scripts the provider to include a `usage` block carrying
// the cached and reasoning token DETAILS on EVERY response it serves — the
// answer AND any scripted tool-call response — so a tool turn accumulates TWO
// usage-bearing calls (round-018: `$#1 < $#2`). The wire `completion_tokens` is
// the INCLUSIVE value (completion + thinking), from which the CLI derives the
// exclusive completion (round-018 FR-002 / research D1).
func (p *Provider) ReportUsageDetails(prompt, cached, completion, thinking int) {
	p.mu.Lock()
	p.usage = &usageCounts{
		prompt:     prompt,
		cached:     cached,
		completion: completion,
		thinking:   thinking,
		total:      prompt + completion + thinking,
		detailed:   true,
	}
	p.mu.Unlock()
}

// usageBlock renders the JSON `usage` fragment (including the leading `,`) for a
// scripted usage, or "" when none. When detailed, `completion_tokens` = completion
// + thinking (inclusive) with the reasoning count in the details.
func usageBlock(u *usageCounts) string {
	if u == nil {
		return ""
	}
	if !u.detailed {
		return fmt.Sprintf(`,"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d}`,
			u.prompt, u.completion, u.total)
	}
	return fmt.Sprintf(`,"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d,"prompt_tokens_details":{"cached_tokens":%d},"completion_tokens_details":{"reasoning_tokens":%d}}`,
		u.prompt, u.completion+u.thinking, u.total, u.cached, u.thinking)
}

// NoUsage scripts the provider to omit the `usage` block (the default), so the
// CLI omits the post-turn measured status line.
func (p *Provider) NoUsage() {
	p.mu.Lock()
	p.usage = nil
	p.mu.Unlock()
}

// Script sets the ordered replies the provider serves. Once the sequence is
// exhausted the final reply repeats — so a single-element script models an
// "always" behaviour (round-008 T006).
func (p *Provider) Script(replies ...Reply) {
	p.mu.Lock()
	p.script = replies
	p.served = 0
	p.mu.Unlock()
}

// Reset rewinds the scripted reply cursor to the start so the SAME scripted
// exchange can be replayed (round-010 ordering witness re-run). Recorded request
// bodies are left intact.
func (p *Provider) Reset() {
	p.mu.Lock()
	p.served = 0
	p.mu.Unlock()
}

// Snapshot returns the fake's restorable state — the script-reply cursor and the
// recorded-request count — so a caller can run a side-effect-free replay and then
// Restore() to leave the request count and script sequence unchanged.
func (p *Provider) Snapshot() (served, requests int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.served, len(p.bodies)
}

// Restore rewinds the fake to a Snapshot(): the script cursor and the number of
// recorded requests. Bodies recorded beyond the snapshot are dropped, so a
// replayed run leaves no trace on RequestCount.
func (p *Provider) Restore(served, requests int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.served = served
	if requests >= 0 && requests < len(p.bodies) {
		p.bodies = p.bodies[:requests]
	}
}

// RequestCount returns how many requests the provider has received.
func (p *Provider) RequestCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.bodies)
}

// LastBody returns the most recently received request body ("" when none).
func (p *Provider) LastBody() string { return p.BodyAt(-1) }

// BodyAt returns the raw request body of request i (i < 0 → the last).
func (p *Provider) BodyAt(i int) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.bodies) == 0 {
		return ""
	}
	if i < 0 || i >= len(p.bodies) {
		i = len(p.bodies) - 1
	}
	return p.bodies[i]
}

// ToolNamesAt returns the wire tool-definition names sent on request i (i < 0 →
// the last). Round-008 T006 for the OpenAI-compatible wire
// (`tools[].function.name`); round 063 also reads the Vertex/Gemini wire
// (`tools[].functionDeclarations[].name`), so the offered-set assertions work on
// either family.
func (p *Provider) ToolNamesAt(i int) []string {
	var req struct {
		Tools []struct {
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
			FunctionDeclarations []struct {
				Name string `json:"name"`
			} `json:"functionDeclarations"`
		} `json:"tools"`
	}
	_ = json.Unmarshal([]byte(p.BodyAt(i)), &req)
	names := make([]string, 0, len(req.Tools))
	for _, t := range req.Tools {
		// A given entry carries one wire shape or the other; concatenating both
		// (rather than switching) is deliberate, so a malformed body carrying
		// both emits both names rather than silently picking one.
		if t.Function.Name != "" {
			names = append(names, t.Function.Name)
		}
		for _, d := range t.FunctionDeclarations {
			names = append(names, d.Name)
		}
	}
	return names
}

// MessagesAt returns the decoded `messages` array of request i (i < 0 → the
// last). Round-008 T006.
func (p *Provider) MessagesAt(i int) []map[string]any {
	var req struct {
		Messages []map[string]any `json:"messages"`
	}
	_ = json.Unmarshal([]byte(p.BodyAt(i)), &req)
	return req.Messages
}

func (p *Provider) handle(w http.ResponseWriter, r *http.Request) {
	// Round-013 D4: the OAuth2 token endpoint is served WITHOUT being recorded
	// as a provider request (recorded requests count completions only).
	if strings.HasSuffix(r.URL.Path, "/token") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600,"token_type":"Bearer"}`))
		return
	}

	raw, _ := io.ReadAll(r.Body)
	auth := r.Header.Get("Authorization")
	p.mu.Lock()
	p.bodies = append(p.bodies, string(raw))
	p.auths = append(p.auths, auth)
	p.paths = append(p.paths, r.URL.Path)
	status, answer, noAnswer, usage, vertex := p.errorStatus, p.answer, p.noAnswer, p.usage, p.vertex
	hasScript := len(p.script) > 0
	var reply Reply
	if hasScript {
		idx := p.served
		if idx >= len(p.script) {
			idx = len(p.script) - 1
		}
		reply = p.script[idx]
		p.served++
	}
	p.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case status != 0:
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":{"message":"scripted error"}}`))
	case noAnswer:
		_, _ = w.Write([]byte(noAnswerBody(vertex)))
	case hasScript:
		_, _ = w.Write([]byte(scriptedBody(reply, vertex, usage)))
	default:
		_, _ = w.Write([]byte(plainAnswerBody(answer, vertex, usage)))
	}
}

// noAnswerBody is the body for a provider answer with no usable content.
func noAnswerBody(vertex bool) string {
	if vertex {
		return `{"candidates":[]}`
	}
	return `{"choices":[]}`
}

// plainAnswerBody is the body for the default (unscripted) answer.
func plainAnswerBody(answer string, vertex bool, u *usageCounts) string {
	if vertex {
		return vertexAnswerBody(answer, u, "")
	}
	return answerBody(answer, u, "")
}

// scriptedBody picks the scripted reply's wire body, family-aware. Extracted so
// the handler's switch stays below the cyclop gate (round 019).
func scriptedBody(reply Reply, vertex bool, u *usageCounts) string {
	switch {
	case len(reply.Tools) > 0:
		if vertex {
			return vertexMultiToolCallBody(reply.Tools, reply.FinishReason)
		}
		return multiToolCallBody(reply.Tools, u, reply.FinishReason)
	case reply.ToolName != "":
		if vertex {
			return vertexToolCallBody(reply.ToolName, reply.Arguments, reply.FinishReason)
		}
		return toolCallBody(reply.ToolName, reply.Arguments, u, reply.FinishReason)
	default:
		if vertex {
			return vertexAnswerBody(reply.Answer, u, reply.FinishReason)
		}
		return answerBody(reply.Answer, u, reply.FinishReason)
	}
}

// openAIFinishReason renders the `finish_reason` fragment for an OpenAI-compatible
// choice (round 030), or "" when none is scripted (a healthy response omits it).
func openAIFinishReason(finishReason string) string {
	if finishReason == "" {
		return ""
	}
	return `,"finish_reason":` + jsonString(finishReason)
}

// vertexFinishReason renders the `finishReason` fragment for a Vertex candidate
// (round 030), or "" when none is scripted (a healthy response omits it).
func vertexFinishReason(finishReason string) string {
	if finishReason == "" {
		return ""
	}
	return `,"finishReason":` + jsonString(finishReason)
}

func answerBody(text string, u *usageCounts, finishReason string) string {
	return `{"choices":[{"message":{"role":"assistant","content":` + jsonString(text) + `}` + openAIFinishReason(finishReason) + `}]` + usageBlock(u) + `}`
}

// toolCallBody builds a tool-call response; the wire call id is deterministic
// ("call_1") so the loop pairs it with the tool result deterministically. Round
// 018 also attaches the scripted `usage` block (so a tool turn's first call
// reports usage too).
func toolCallBody(name, arguments string, u *usageCounts, finishReason string) string {
	call := `{"id":"call_1","type":"function","function":{"name":` + jsonString(name) + `,"arguments":` + jsonString(arguments) + `}}`
	return `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[` + call + `]}` + openAIFinishReason(finishReason) + `}]` + usageBlock(u) + `}`
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// multiToolCallBody builds a single OpenAI-compatible response carrying every
// listed tool call (deterministic ids call_1..call_N). Round 019.
func multiToolCallBody(tools []ToolRequest, u *usageCounts, finishReason string) string {
	calls := make([]string, 0, len(tools))
	for i, t := range tools {
		calls = append(calls, fmt.Sprintf(
			`{"id":"call_%d","type":"function","function":{"name":%s,"arguments":%s}}`,
			i+1, jsonString(t.Name), jsonString(t.Arguments)))
	}
	return `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[` +
		strings.Join(calls, ",") + `]}` + openAIFinishReason(finishReason) + `}]` + usageBlock(u) + `}`
}

// vertexMultiToolCallBody builds a single Vertex response carrying every listed
// functionCall part. Round 019.
func vertexMultiToolCallBody(tools []ToolRequest, finishReason string) string {
	parts := make([]string, 0, len(tools))
	for _, t := range tools {
		args := strings.TrimSpace(t.Arguments)
		if args == "" {
			args = "{}"
		}
		parts = append(parts, `{"functionCall":{"name":`+jsonString(t.Name)+`,"args":`+args+`},"thoughtSignature":"sig-1"}`)
	}
	return `{"candidates":[{"content":{"role":"model","parts":[` + strings.Join(parts, ",") + `]}` + vertexFinishReason(finishReason) + `}]}`
}

// vertexAnswerBody builds a Vertex `:generateContent` answer response (round-013).
// Round 072 (ADR 0044): a `detailed` usage block also emits
// `cachedContentTokenCount` and `thoughtsTokenCount`; the Vertex family is
// DISJOINT, so `candidatesTokenCount` is the plain completion (no `+thinking`).
func vertexAnswerBody(text string, u *usageCounts, finishReason string) string {
	usageJSON := ""
	if u != nil {
		if u.detailed {
			usageJSON = fmt.Sprintf(`,"usageMetadata":{"promptTokenCount":%d,"candidatesTokenCount":%d,"totalTokenCount":%d,"cachedContentTokenCount":%d,"thoughtsTokenCount":%d}`,
				u.prompt, u.completion, u.total, u.cached, u.thinking)
		} else {
			usageJSON = fmt.Sprintf(`,"usageMetadata":{"promptTokenCount":%d,"candidatesTokenCount":%d,"totalTokenCount":%d}`, u.prompt, u.completion, u.total)
		}
	}
	return `{"candidates":[{"content":{"role":"model","parts":[{"text":` + jsonString(text) + `}]}` + vertexFinishReason(finishReason) + `}]` + usageJSON + `}`
}

// vertexToolCallBody builds a Vertex functionCall response; `arguments` is a raw
// JSON object (Vertex `args` is an object, unlike OpenAI's string form).
func vertexToolCallBody(name, arguments string, finishReason string) string {
	args := strings.TrimSpace(arguments)
	if args == "" {
		args = "{}"
	}
	return `{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":` + jsonString(name) + `,"args":` + args + `},"thoughtSignature":"sig-1"}]}` + vertexFinishReason(finishReason) + `}]}`
}

// ConfigYAML builds a resolvable default configuration (MODE: butler) that
// selects `selected` and registers each provider name->url.
func ConfigYAML(selected string, providerURLs map[string]string) string {
	var sb strings.Builder
	sb.WriteString("MODE: butler\n")
	sb.WriteString("PERSON: \"e2e persona\"\n")
	sb.WriteString("SELECTED_PROVIDER: " + selected + "\n")
	sb.WriteString("PROVIDERS:\n")
	for name, url := range providerURLs {
		sb.WriteString("  " + name + ":\n")
		sb.WriteString("    TYPE: deepseek\n")
		sb.WriteString("    MODEL: deepseek-v4-flash\n")
		sb.WriteString("    URL: " + url + "\n")
		sb.WriteString("    API_KEY: test-key\n")
	}
	return sb.String()
}
