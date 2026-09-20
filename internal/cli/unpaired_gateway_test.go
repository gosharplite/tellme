package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// recordingGateway records the last request it received and returns a stub answer.
type recordingGateway struct{ got llm.Request }

func (g *recordingGateway) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	g.got = req
	return llm.Response{Text: "ok"}, nil
}

// TestUnpairedGateway_EmitsOnShortRound pins round 068 (ADR 0038; surfaces
// RF-067-1): the decorator surfaces the unpaired ids of a short round (M < N)
// BEFORE the request is sent, and the request still reaches the inner gateway
// unchanged (Q2 → A: informational).
func TestUnpairedGateway_EmitsOnShortRound(t *testing.T) {
	inner := &recordingGateway{}
	var got []string
	gw := withUnpairedDiagnostic(inner, func(ids []string) { got = ids })
	req := llm.Request{Messages: []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "a", Name: "t"}, {ID: "b", Name: "t"}}},
		{Role: "tool", Content: "x", ToolCallID: "a"},
	}}
	if _, err := gw.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("emit got %v, want [b] (the unpaired call)", got)
	}
	if len(inner.got.Messages) != 2 {
		t.Fatalf("inner received %d messages, want the request unchanged (2)", len(inner.got.Messages))
	}
}

// TestUnpairedGateway_SilentOnHappyPath pins I-7: the shipped M == N path emits
// nothing.
func TestUnpairedGateway_SilentOnHappyPath(t *testing.T) {
	inner := &recordingGateway{}
	emitted := false
	gw := withUnpairedDiagnostic(inner, func([]string) { emitted = true })
	req := llm.Request{Messages: []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "a", Name: "t"}}},
		{Role: "tool", Content: "x", ToolCallID: "a"},
	}}
	if _, err := gw.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if emitted {
		t.Fatal("a fully paired round must emit no diagnostic (I-7)")
	}
}

// TestUnpairedEmitter_WritesToStderr pins Q1 → A: the diagnostic is a `[Tool …]`
// line on stderr (never stdout), and no-op when empty.
func TestUnpairedEmitter_WritesToStderr(t *testing.T) {
	var buf bytes.Buffer
	env := runtimeEnv{stderr: &buf, stdout: &bytes.Buffer{}, clock: func() time.Time { return time.Unix(0, 0) }}
	emit := unpairedEmitter(env, fakeLines{})
	emit(nil) // empty ⇒ no write
	if buf.Len() != 0 {
		t.Fatalf("no ids must write nothing, got %q", buf.String())
	}
	emit([]string{"b", "c"})
	if !strings.Contains(buf.String(), "<unpaired:b,c>") {
		t.Fatalf("stderr = %q, want the unpaired line", buf.String())
	}
}

// TestWithUnpairedDiagnostic_NilEmitterPassthrough pins the nil-emitter default.
func TestWithUnpairedDiagnostic_NilEmitterPassthrough(t *testing.T) {
	inner := &recordingGateway{}
	if _, ok := withUnpairedDiagnostic(inner, nil).(*recordingGateway); !ok {
		t.Fatal("a nil emitter must return the gateway unwrapped")
	}
}
