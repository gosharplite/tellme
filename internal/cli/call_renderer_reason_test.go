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
// reasons grouped as the per-call TAIL. This pin drives OnCallEnd DIRECTLY, so it
// exercises the tail's DEFENSIVE guard — production routes the reasons through
// agent.reasonsOf, which filters blanks upstream (round-036 review TD-1). The
// real-path witness is the loop-tier TestLogOmitsReasonLineForWhitespaceOnlyReason;
// the single-ownership consolidation of this redundant predicate is tracked on
// issue #69.
//
// A blank reason must emit NO `[Tool Reason]` line here (and no bare newline — the
// caller-side trim check is what keeps the pure formatter from being asked to
// return an empty-string sentinel, which `Fprintln` would still turn into a blank
// line).

func newTestCallRenderer(stderr *bytes.Buffer) *callRenderer {
	return &callRenderer{
		env:     runtimeEnv{stderr: stderr, clock: func() time.Time { return r036Clock }},
		res:     resolution{Mode: "butler", Selected: "test", Provider: config.Provider{Model: "test-model"}},
		pricing: ui.Pricing{},
	}
}

// r036Clock is a local clock for this package's tail-format assertion: the ui
// package's r034Clock is unexported and cannot cross packages, so the expected
// timestamp here is derived from this clock (round-036 review N-3).
var r036Clock = time.Date(2026, 9, 17, 8, 0, 0, 0, time.UTC)

func TestCallTailOmitsBlankReasonLine(t *testing.T) {
	for _, reason := range []string{"   ", "\n"} {
		var buf bytes.Buffer
		r := newTestCallRenderer(&buf)
		r.OnCallEnd(0, llm.Usage{Reported: false}, []string{reason}, false)
		if strings.Contains(buf.String(), "[Tool Reason]") {
			t.Errorf("a blank tail reason %q rendered a [Tool Reason] line; out=%q", reason, buf.String())
		}
		// The tail must write NOTHING for a blank reason — assert emptiness
		// directly rather than "no newline" (round-036 review N-1).
		if buf.Len() != 0 {
			t.Errorf("a blank tail reason %q wrote output; out=%q", reason, buf.String())
		}
	}
}

func TestCallTailStillRendersANonBlankReasonLine(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCallRenderer(&buf)
	r.OnCallEnd(0, llm.Usage{Reported: false}, []string{"because"}, false)
	if got := buf.String(); got != "[08:00:00] [Tool Reason] because\n" {
		t.Errorf("tail reason rendering = %q; want one [Tool Reason] line", got)
	}
}
