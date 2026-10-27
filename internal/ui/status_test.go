package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/render"
)

// TestFormatPayloadMeasured pins the round-009 measured status-line shape
// (round-057 TD-057-1: the estimated `~`-prefixed shape no longer lives on this
// formatter — it is FormatPayloadEstimate), and that the line carries no
// `tellme:` prefix (round-009 research Decisions 3 & 4; FR-014).
func TestFormatPayloadMeasured(t *testing.T) {
	ts := time.Date(2026, 9, 13, 13, 16, 25, 0, time.UTC)
	got := FormatPayloadMeasured(render.PayloadStatus{At: ts, Tokens: 1234, Budget: 1000000, Mode: "butler", Model: "deepseek-v4-flash"})
	want := "[13:16:25] Payload: 1234/1000000 tokens - butler - deepseek-v4-flash"
	if got != want {
		t.Errorf("measured line = %q; want %q", got, want)
	}
	if strings.Contains(got, "~") {
		t.Errorf("the measured line must not carry the ~ estimate marker: %q", got)
	}
	if strings.Contains(got, "tellme:") {
		t.Errorf("the status line must not carry the reserved tellme: prefix: %q", got)
	}
}
