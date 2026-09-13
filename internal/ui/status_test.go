package ui

import (
	"strings"
	"testing"
	"time"
)

// TestFormatPayloadStatus pins the reference-parity status-line shape and the
// `~`-vs-bare distinction, and that the line carries no `tellme:` prefix
// (round-009 research Decisions 3 & 4; FR-014).
func TestFormatPayloadStatus(t *testing.T) {
	ts := time.Date(2026, 9, 13, 13, 16, 25, 0, time.UTC)
	est := FormatPayloadStatus(ts, 1234, 1000000, "butler", "deepseek-v4-flash", true)
	if est != "[13:16:25] Payload: ~1234/1000000 tokens - butler - deepseek-v4-flash" {
		t.Errorf("estimated line = %q", est)
	}
	act := FormatPayloadStatus(ts, 1234, 1000000, "butler", "deepseek-v4-flash", false)
	if act != "[13:16:25] Payload: 1234/1000000 tokens - butler - deepseek-v4-flash" {
		t.Errorf("measured line = %q", act)
	}
	if strings.Contains(act, "~") {
		t.Errorf("measured line must not carry the ~ estimate marker: %q", act)
	}
	if !strings.Contains(est, "~") {
		t.Errorf("estimated line must carry the ~ estimate marker: %q", est)
	}
	if strings.Contains(est, "tellme:") {
		t.Errorf("status line must not carry the reserved tellme: prefix: %q", est)
	}
}
