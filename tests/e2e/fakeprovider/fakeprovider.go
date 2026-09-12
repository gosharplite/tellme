// Package fakeprovider is a leaf E2E test-support package: an in-process,
// OpenAI-compatible provider used by the chat acceptance scenarios and the
// offline-path canary. It records every request it receives so a step can
// assert a provider was hit exactly once — or, on the offline paths, never.
package fakeprovider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// Provider is a scriptable OpenAI-compatible fake.
type Provider struct {
	srv *httptest.Server

	mu          sync.Mutex
	answer      string
	errorStatus int
	noAnswer    bool
	bodies      []string
	closeOnce   sync.Once
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

// RequestCount returns how many requests the provider has received.
func (p *Provider) RequestCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.bodies)
}

// LastBody returns the most recently received request body ("" when none).
func (p *Provider) LastBody() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.bodies) == 0 {
		return ""
	}
	return p.bodies[len(p.bodies)-1]
}

func (p *Provider) handle(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	p.bodies = append(p.bodies, string(raw))
	status, answer, noAnswer := p.errorStatus, p.answer, p.noAnswer
	p.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case status != 0:
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":{"message":"scripted error"}}`))
	case noAnswer:
		_, _ = w.Write([]byte(`{"choices":[]}`))
	default:
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":` + jsonString(answer) + `}}]}`))
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
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
