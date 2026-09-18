package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/ui"
)

// Round 036 (issue #74, operator Q2): callRenderer.OnCallEnd re-emits the round's
// reasons grouped as the per-call TAIL.
//
// Round 046 (R4 of #92, ADR 0015): the blank-reason predicate is single-owned
// UPSTREAM — the loop filters the round's reasons through
// ui.ToolLineRenderer.ReasonLine before they reach OnCallEnd — so the tail prints
// the reasons it is given, with no re-check. The former defensive guard here
// (round-036 review TD-1, tracked for consolidation on issue #69) is DELETED; the
// real-path suppression witness is the loop tier
// (internal/agent TestLogOmitsReasonLineForWhitespaceOnlyReason).

func newTestCallRenderer(stderr *bytes.Buffer) *callRenderer {
	return &callRenderer{
		env:     runtimeEnv{stderr: stderr, clock: func() time.Time { return r036Clock }},
		res:     resolution{Mode: "butler", Selected: "test", Provider: config.Provider{Model: "test-model"}},
		pricing: ui.Pricing{},
	}
}

// r036Clock is a local clock for this package's tail-format assertion: the ui
// package's clock vars are unexported and cannot cross packages, so the expected
// timestamp here is derived from this clock (round-036 review N-3).
var r036Clock = time.Date(2026, 9, 17, 8, 0, 0, 0, time.UTC)

func TestCallTailStillRendersANonBlankReasonLine(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCallRenderer(&buf)
	r.OnCallEnd(0, llm.Usage{Reported: false}, []string{"because"}, false)
	// Round 039: the trailing grouped reason block is preceded by exactly one
	// blank line.
	if got := buf.String(); got != "\n[08:00:00] [Tool Reason] because\n" {
		t.Errorf("tail reason rendering = %q; want a blank line then one [Tool Reason] line", got)
	}
}

// Round 039: the post-status group (measured payload + metrics + `Ready`) is
// preceded by exactly one blank line, after the trailing reason block (which is
// itself preceded by one blank).
func TestCallTailBlanksBeforeReasonsAndPostStatus(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCallRenderer(&buf)
	r.OnCallEnd(0, llm.Usage{Reported: true, PromptTokens: 10, CompletionTokens: 5}, []string{"checking"}, false)
	out := buf.String()

	// Exactly one blank line before the reason block, reason line right after it.
	if !strings.HasPrefix(out, "\n[08:00:00] [Tool Reason] checking\n") {
		t.Errorf("expected one blank line before the reason block; out=%q", out)
	}
	// Exactly one blank line before the measured payload line (post-status group).
	if !strings.Contains(out, "checking\n\n[08:00:00] Payload: 10/") {
		t.Errorf("expected one blank line before the post-status group; out=%q", out)
	}
	// No blank inside the reason block and no double blank before the payload.
	if strings.Contains(out, "checking\n\n\n") {
		t.Errorf("more than one blank line before the post-status group; out=%q", out)
	}
}

// Round 039 review B1: a TOOL-LESS turn (only a final call, no rendered tool
// round) must gain NO blank before its post-status group — FR-009 / the
// edge-case list / ADR 0008 D5's closing sentence all say a non-tool turn is
// unchanged.
func TestCallTailNoBlankBeforePostStatusWithoutToolRound(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCallRenderer(&buf)
	// The tool-less path: the loop fires exactly one final call, whose tail is
	// deferred; nothing sets renderedToolRound.
	r.OnCallEnd(0, llm.Usage{Reported: true, PromptTokens: 10, CompletionTokens: 5}, nil, true)
	r.EmitFinalTail()
	out := buf.String()
	if strings.HasPrefix(out, "\n") {
		t.Errorf("a tool-less turn gained a leading blank line before the post-status group; out=%q", out)
	}
	if !strings.HasPrefix(out, "[08:00:00] Payload: 10/") {
		t.Errorf("the post-status group should start flush for a tool-less turn; out=%q", out)
	}
}
