package ui

import (
	"strings"
	"testing"
	"time"
)

// Round-034 T021 unit pins for the [Tool Output] writer: the pinned header /
// separator literals, the per-line format, the line-assembly across writes, and
// the trailing-partial-line drop (FR-010 / ADR 0005 D7).

func TestToolOutputHeaderAndSeparator(t *testing.T) {
	if got, want := FormatToolOutputHeader(r034Clock), "[20:29:51] [Tool Output] Executing... (Output shown below)"; got != want {
		t.Errorf("FormatToolOutputHeader = %q; want %q", got, want)
	}
	if len(ToolOutputSeparator) != 60 || strings.Trim(ToolOutputSeparator, "-") != "" {
		t.Errorf("ToolOutputSeparator = %q; want 60 hyphens", ToolOutputSeparator)
	}
}

func TestFormatToolOutputLine(t *testing.T) {
	if got, want := FormatToolOutputLine(r034Clock, "hi"), "[20:29:51] [Tool Output] hi"; got != want {
		t.Errorf("FormatToolOutputLine = %q; want %q", got, want)
	}
}

func TestToolOutputWriter_AssemblesLinesAndDropsTrailingPartial(t *testing.T) {
	var sb strings.Builder
	w := &ToolOutputWriter{W: &sb, Now: func() time.Time { return r034Clock }}
	w.Begin()
	// "alpha\n" completes a line; "beta" is a partial carried across calls;
	// "\ngamma" completes beta, leaving "gamma" partial at End.
	if _, err := w.Write([]byte("alpha\nbeta")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("\ngamma")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	w.End()

	got := sb.String()
	if !strings.Contains(got, FormatToolOutputLine(r034Clock, "alpha")) {
		t.Errorf("missing alpha line; got %q", got)
	}
	if !strings.Contains(got, FormatToolOutputLine(r034Clock, "beta")) {
		t.Errorf("missing beta line (partial not completed); got %q", got)
	}
	if strings.Contains(got, "gamma") {
		t.Errorf("the trailing partial line was flushed, not dropped; got %q", got)
	}
	if n := strings.Count(got, ToolOutputSeparator); n != 2 {
		t.Errorf("want an opening and a closing separator (2), got %d; got %q", n, got)
	}
}

func TestToolOutputWriter_NilWriterIsNoOp(t *testing.T) {
	w := &ToolOutputWriter{}
	w.Begin()
	if _, err := w.Write([]byte("x\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	w.End() // must not panic
}
