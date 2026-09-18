package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 036 (issue #74, operator Q2): a blank reason — empty OR whitespace-only
// after folding+trimming — must emit NO `[Tool Reason]` line. Round 046 (R4 of
// #92, ADR 0015): the blank-reason predicate is single-owned by the renderer
// port; the loop merely HONORS the renderer's decision, so this pins the loop's
// suppression on the real path (the transform semantics stay in internal/ui).

func TestLogOmitsReasonLineForWhitespaceOnlyReason(t *testing.T) {
	for _, reason := range []string{"   ", "\n", "\t"} {
		var buf bytes.Buffer
		a := newLogLoop(&buf, llm.ToolCall{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"` + reason + `"}`})
		if _, err := a.Run(context.Background(), "q", nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
		log := buf.String()
		if strings.Contains(log, "REASON ") {
			t.Errorf("a whitespace-only reason %q rendered a reason line; log=%q", reason, log)
		}
		if !strings.Contains(log, "ACTION read_files(") {
			t.Errorf("the action line was not rendered for reason %q; log=%q", reason, log)
		}
		// Round 039: a reason-less call's action line is still the start of its
		// begin block, so a blank line precedes it (the per-call blank-line group).
		if !lineBeforeContains(log, "ACTION read_files(", "") {
			t.Errorf("the reason-less call's action line was not preceded by a blank line; log=%q", log)
		}
	}
}

func TestLogStillRendersANonBlankReasonLine(t *testing.T) {
	// Control: a real reason still renders exactly one reason line, so the
	// suppression is not a blanket "never render".
	var buf bytes.Buffer
	a := newLogLoop(&buf, llm.ToolCall{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`})
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if log := buf.String(); !strings.Contains(log, "REASON because") {
		t.Errorf("a non-blank reason was dropped; log=%q", log)
	}
}
