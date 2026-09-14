package ui

import (
	"strings"
	"testing"
	"time"
)

func TestFormatMetrics(t *testing.T) {
	ts := time.Date(2026, 9, 14, 11, 46, 52, 0, time.UTC)
	got := FormatMetrics(ts, "deepseek-flash", UsageCounts{Miss: 236, Hit: 56576, Completion: 43, Thinking: 27})
	want := "[11:46:52] [deepseek-flash] M: 236 H: 56576 C: 43 Th: 27"
	if got != want {
		t.Errorf("FormatMetrics = %q, want %q", got, want)
	}
}

func TestFormatMetricsThAlwaysShown(t *testing.T) {
	ts := time.Date(2026, 9, 14, 11, 47, 16, 0, time.UTC)
	got := FormatMetrics(ts, "butler", UsageCounts{Miss: 1, Hit: 2, Completion: 3, Thinking: 0})
	if !strings.Contains(got, "Th: 0") {
		t.Errorf("the Th segment must be shown even at zero, got %q", got)
	}
	want := "[11:47:16] [butler] M: 1 H: 2 C: 3 Th: 0"
	if got != want {
		t.Errorf("FormatMetrics = %q, want %q", got, want)
	}
}

func TestFormatReady(t *testing.T) {
	got := FormatReady(0.0005, 0.0005, 0.0005, 236, 56576, 70, 99.6)
	want := "╰─⠿ Ready ($0.0005 $0.0005 $0.0005 - M: 236 H: 56576 O: 70 - 99.6%)"
	if got != want {
		t.Errorf("FormatReady = %q, want %q", got, want)
	}
}
