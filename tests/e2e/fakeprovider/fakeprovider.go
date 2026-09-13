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
// a plain answer text (otherwise). Round 008.
type Reply struct {
	ToolName  string
	Arguments string
	Answer    string
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
// (round-009 T006).
type usageCounts struct {
	prompt     int
	completion int
	total      int
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

// ToolNamesAt returns the wire tool-definition names (`tools[].function.name`)
// sent on request i (i < 0 → the last). Round-008 T006.
func (p *Provider) ToolNamesAt(i int) []string {
	var req struct {
		Tools []struct {
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
		} `json:"tools"`
	}
	_ = json.Unmarshal([]byte(p.BodyAt(i)), &req)
	names := make([]string, 0, len(req.Tools))
	for _, t := range req.Tools {
		names = append(names, t.Function.Name)
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
		if vertex {
			_, _ = w.Write([]byte(`{"candidates":[]}`))
		} else {
			_, _ = w.Write([]byte(`{"choices":[]}`))
		}
	case hasScript && reply.ToolName != "":
		if vertex {
			_, _ = w.Write([]byte(vertexToolCallBody(reply.ToolName, reply.Arguments)))
		} else {
			_, _ = w.Write([]byte(toolCallBody(reply.ToolName, reply.Arguments)))
		}
	case hasScript:
		if vertex {
			_, _ = w.Write([]byte(vertexAnswerBody(reply.Answer, usage)))
		} else {
			_, _ = w.Write([]byte(answerBody(reply.Answer, usage)))
		}
	default:
		if vertex {
			_, _ = w.Write([]byte(vertexAnswerBody(answer, usage)))
		} else {
			_, _ = w.Write([]byte(answerBody(answer, usage)))
		}
	}
}

func answerBody(text string, u *usageCounts) string {
	usageJSON := ""
	if u != nil {
		usageJSON = fmt.Sprintf(`,"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d}`, u.prompt, u.completion, u.total)
	}
	return `{"choices":[{"message":{"role":"assistant","content":` + jsonString(text) + `}}]` + usageJSON + `}`
}

// toolCallBody builds a tool-call response; the wire call id is deterministic
// ("call_1") so the loop pairs it with the tool result deterministically.
func toolCallBody(name, arguments string) string {
	call := `{"id":"call_1","type":"function","function":{"name":` + jsonString(name) + `,"arguments":` + jsonString(arguments) + `}}`
	return `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[` + call + `]}}]}`
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// vertexAnswerBody builds a Vertex `:generateContent` answer response (round-013).
func vertexAnswerBody(text string, u *usageCounts) string {
	usageJSON := ""
	if u != nil {
		usageJSON = fmt.Sprintf(`,"usageMetadata":{"promptTokenCount":%d,"candidatesTokenCount":%d,"totalTokenCount":%d}`, u.prompt, u.completion, u.total)
	}
	return `{"candidates":[{"content":{"role":"model","parts":[{"text":` + jsonString(text) + `}]}}]` + usageJSON + `}`
}

// vertexToolCallBody builds a Vertex functionCall response; `arguments` is a raw
// JSON object (Vertex `args` is an object, unlike OpenAI's string form).
func vertexToolCallBody(name, arguments string) string {
	args := strings.TrimSpace(arguments)
	if args == "" {
		args = "{}"
	}
	return `{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":` + jsonString(name) + `,"args":` + args + `},"thoughtSignature":"sig-1"}]}}]}`
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
