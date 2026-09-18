package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 034: the tool-loop log is the DECOMPOSED shape — `[Tool Engine] Step i/M`,
// `[Tool Reason]`, `[Tool Action] <tool>(<sorted args>)`, `[Tool Result] <tool>:
// <snippet>` (ADR 0005). Round 046 (R4 of #92, ADR 0015): the loop renders those
// lines through the injected `Lines` port; the loop's own tests may not import
// internal/ui, so they pin the SCHEDULE via a deterministic fake renderer (the
// real bytes are pinned in internal/ui + E2E). The injected clock seam still keeps
// the loop deterministic.

var fixedLogClock = func() time.Time {
	return time.Date(2026, 9, 15, 12, 34, 56, 0, time.UTC)
}

// newLogLoop builds a loop wired with the fake renderer + fixed clock + a tool
// registry containing read_files, so the tool-line schedule is exercised.
func newLogLoop(buf *bytes.Buffer, calls ...llm.ToolCall) *AgentLoop {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: calls},
		{Text: "done"},
	}}
	return &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Stderr:   buf,
		Now:      fixedLogClock,
		Lines:    fakeRenderer{},
	}
}

func TestLogRendersDecomposedLinesWithReason(t *testing.T) {
	var buf bytes.Buffer
	a := newLogLoop(&buf, llm.ToolCall{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`})
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	for _, want := range []string{
		"ENGINE 1/1",
		"REASON because",
		"ACTION read_files(",
		"RESULT read_files: hi",
	} {
		if !strings.Contains(log, want) {
			t.Errorf("tool-loop log missing %q; log=%q", want, log)
		}
	}
}

func TestLogOmitsReasonLineWhenNoReason(t *testing.T) {
	var buf bytes.Buffer
	a := newLogLoop(&buf, llm.ToolCall{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`})
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	if strings.Contains(log, "REASON ") {
		t.Errorf("a reasonless call rendered a reason line (FR-006 defensive tolerance); log=%q", log)
	}
	if !strings.Contains(log, "ACTION read_files(") {
		t.Errorf("the action line was not rendered for a reasonless call; log=%q", log)
	}
}

// TestLogIsSilentWhenNoRendererInjected (round-046 review TD-3) pins the seam's
// documented default: a nil Lines renderer writes NO tool lines (and no blank, no
// yield). The failure mode is silent — a miswire would suppress every tool
// diagnostic and only the (slow, coarse) E2E would notice — so the promise gets a
// direct pin.
func TestLogIsSilentWhenNoRendererInjected(t *testing.T) {
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Stderr:   &buf,
		Now:      fixedLogClock,
		// Lines deliberately nil — the documented no-op default.
	}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("a nil Lines renderer wrote tool lines; out=%q", buf.String())
	}
}
