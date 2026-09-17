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

// Round 036 (issue #74, operator Q2): the per-call TAIL is the second reason
// emission site (callRenderer.OnCallEnd re-emits the round's reasons grouped).
// A blank reason must emit NO `[Tool Reason]` line there either (and no bare
// newline — the caller-side trim check is what keeps the pure formatter from
// being asked to return an empty-string sentinel, which `Fprintln` would still
// turn into a blank line).

func newTestCallRenderer(stderr *bytes.Buffer) *callRenderer {
	return &callRenderer{
		env:     runtimeEnv{stderr: stderr, clock: func() time.Time { return r036Clock }},
		res:     resolution{Mode: "butler", Selected: "test", Provider: config.Provider{Model: "test-model"}},
		pricing: ui.Pricing{},
	}
}

var r036Clock = time.Date(2026, 9, 17, 8, 0, 0, 0, time.UTC)

func TestCallTailOmitsBlankReasonLine(t *testing.T) {
	for _, reason := range []string{"   ", "\n"} {
		var buf bytes.Buffer
		r := newTestCallRenderer(&buf)
		r.OnCallEnd(0, llm.Usage{Reported: false}, []string{reason}, false)
		if strings.Contains(buf.String(), "[Tool Reason]") {
			t.Errorf("a blank tail reason %q rendered a [Tool Reason] line; out=%q", reason, buf.String())
		}
		if strings.Contains(buf.String(), "\n") {
			t.Errorf("a blank tail reason %q wrote a bare newline; out=%q", reason, buf.String())
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
