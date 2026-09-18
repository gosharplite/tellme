package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// TestChromeColour pins the round-054 (ADR 0023) green accents: with colour
// enabled the four elements carry the reference's `\033[0;32m…\033[0m`; with it
// disabled the lines are byte-identical to the plain form.
func TestChromeColour(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 9, 19, 6, 25, 11, 0, time.UTC)
	const g = "\033[0;32m"
	const r = "\033[0m"

	t.Run("payload: mode green in both; measured tokens green only", func(t *testing.T) {
		t.Parallel()
		est := formatPayloadStatusColour(ts, 282827, 1000000, "butler", "deepseek-flash", true, true)
		if !strings.Contains(est, "- "+g+"butler"+r+" - ") {
			t.Errorf("estimated payload mode not green: %q", est)
		}
		if strings.Contains(est, g+"282827"+r) {
			t.Errorf("estimated payload token number must stay plain: %q", est)
		}
		meas := formatPayloadStatusColour(ts, 438165, 1000000, "butler", "deepseek-flash", false, true)
		if !strings.Contains(meas, g+"438165"+r) {
			t.Errorf("measured payload token number not green: %q", meas)
		}
		if !strings.Contains(meas, "- "+g+"butler"+r+" - ") {
			t.Errorf("measured payload mode not green: %q", meas)
		}

		// Disabled ⇒ byte-identical to the plain formatters.
		if got, want := formatPayloadStatusColour(ts, 438165, 1000000, "butler", "deepseek-flash", false, false), FormatPayloadStatus(ts, 438165, 1000000, "butler", "deepseek-flash", false); got != want {
			t.Errorf("colour-off payload = %q, want plain %q", got, want)
		}
	})

	t.Run("ready: only the session cost is green", func(t *testing.T) {
		t.Parallel()
		line := formatReadyColour(0.0047, 0.0047, 1.0607, 742574, 95659136, 219965, 99.2, true)
		if !strings.Contains(line, g+"$1.0607"+r) {
			t.Errorf("session cost not green: %q", line)
		}
		if strings.Contains(line, g+"$0.0047"+r) {
			t.Errorf("call/turn costs must stay plain: %q", line)
		}
		if got, want := formatReadyColour(0.0047, 0.0047, 1.0607, 742574, 95659136, 219965, 99.2, false), FormatReady(0.0047, 0.0047, 1.0607, 742574, 95659136, 219965, 99.2); got != want {
			t.Errorf("colour-off ready = %q, want plain %q", got, want)
		}
	})

	t.Run("tool reason: whole line green", func(t *testing.T) {
		t.Parallel()
		line := formatToolReasonColour(ts, "read the config", true)
		if !strings.HasPrefix(line, g) || !strings.HasSuffix(line, r) {
			t.Errorf("reason line not wrapped green: %q", line)
		}
		if !strings.Contains(line, "[Tool Reason] read the config") {
			t.Errorf("reason text missing: %q", line)
		}
		if got, want := formatToolReasonColour(ts, "read the config", false), FormatToolReason(ts, "read the config"); got != want {
			t.Errorf("colour-off reason = %q, want plain %q", got, want)
		}
	})

	t.Run("empty mode is not wrapped", func(t *testing.T) {
		t.Parallel()
		line := formatPayloadStatusColour(ts, 1, 1000000, "", "m", false, true)
		if strings.Contains(line, g+r) {
			t.Errorf("an empty mode must not emit an empty green pair: %q", line)
		}
	})

	t.Run("metrics line is uncoloured", func(t *testing.T) {
		t.Parallel()
		line := Lines{colour: true}.Metrics(ts, "p", metrics.UsageCounts{Miss: 1})
		if strings.Contains(line, g) {
			t.Errorf("the metrics line must stay plain: %q", line)
		}
	})
}
