package ui

import (
	"strings"
	"testing"
	"time"
)

// Round 046 (R4 of #92, ADR 0015): ToolLineRenderer is the production adapter the
// agent loop renders through. ReasonLine is the single owning definition of the
// blank-reason predicate on the real path — it evaluates the reason transform
// ONCE and returns BOTH the rendered line and the render decision.

var r046Clock = time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)

func TestToolLineRendererReasonLineBlankReturnsFalse(t *testing.T) {
	r := ToolLineRenderer{}
	for _, reason := range []string{"", "   ", "\n", "\t", "\u001b[31m"} {
		line, renders := r.ReasonLine(r046Clock, reason)
		if renders {
			t.Errorf("ReasonLine(%q) renders = true; want false (blank / whitespace-only / escape-only)", reason)
		}
		if line != "" {
			t.Errorf("ReasonLine(%q) line = %q; want the empty string", reason, line)
		}
	}
}

func TestToolLineRendererReasonLineMatchesFormatToolReason(t *testing.T) {
	r := ToolLineRenderer{}
	for _, reason := range []string{"because", "  spaced  ", "a\nb", strings.Repeat("x", 300)} {
		line, renders := r.ReasonLine(r046Clock, reason)
		if !renders {
			t.Errorf("ReasonLine(%q) renders = false; want true", reason)
			continue
		}
		if want := FormatToolReason(r046Clock, reason); line != want {
			t.Errorf("ReasonLine(%q) = %q; want byte-equal to FormatToolReason %q", reason, line, want)
		}
		if !ToolReasonRenders(reason) {
			t.Errorf("ReasonLine(%q) says renders but ToolReasonRenders says false — the owner drifted", reason)
		}
	}
}

func TestToolLineRendererDelegatesToTheFormatters(t *testing.T) {
	r := ToolLineRenderer{}
	if got, want := r.EngineLine(r046Clock, 1, 3), FormatToolEngine(r046Clock, 1, 3); got != want {
		t.Errorf("EngineLine = %q; want %q", got, want)
	}
	if got, want := r.ActionLine(r046Clock, "read_files", `{"a":"b"}`), FormatToolAction(r046Clock, "read_files", `{"a":"b"}`); got != want {
		t.Errorf("ActionLine = %q; want %q", got, want)
	}
	if got, want := r.ResultLine(r046Clock, "read_files", "hi"), FormatToolResult(r046Clock, "read_files", "hi"); got != want {
		t.Errorf("ResultLine = %q; want %q", got, want)
	}
}
