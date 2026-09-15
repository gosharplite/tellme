package ui

import (
	"strings"
	"testing"
	"time"
)

// Round 022 T011: the pure FormatToolLog formatter and the shared clock token.
// The injected fixed time keeps the assertions deterministic.

var toolLogClock = time.Date(2026, 9, 15, 12, 34, 56, 0, time.UTC)

func TestFormatToolLogWithReason(t *testing.T) {
	got := FormatToolLog(toolLogClock, "get_tree", "show the repo tree")
	want := "[12:34:56] [Tool] get_tree - show the repo tree"
	if got != want {
		t.Errorf("FormatToolLog = %q; want %q", got, want)
	}
}

func TestFormatToolLogWithoutReason(t *testing.T) {
	got := FormatToolLog(toolLogClock, "read_files", "")
	want := "[12:34:56] [Tool] read_files"
	if got != want {
		t.Errorf("FormatToolLog = %q; want %q", got, want)
	}
}

// TestFormatClockSharedAcrossFormatters pins review-2 R-1: the four internal/ui
// timestamp formatters share ONE clock token (`15:04:05`).
func TestFormatClockSharedAcrossFormatters(t *testing.T) {
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	formatters := map[string]string{
		"tool log":      FormatToolLog(ts, "t", "r"),
		"payload":       FormatPayloadStatus(ts, 1, 2, "mode", "model", true),
		"input capture": FormatInputCaptured(ts),
		"metrics":       FormatMetrics(ts, "provider", UsageCounts{Miss: 1}),
	}
	for name, got := range formatters {
		if !strings.Contains(got, "[03:04:05]") {
			t.Errorf("%s = %q; want the shared [03:04:05] clock token", name, got)
		}
	}
}
