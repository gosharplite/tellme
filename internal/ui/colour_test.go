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

	t.Run("measured payload: the mode and the token number are green", func(t *testing.T) {
		t.Parallel()
		meas := formatPayloadMeasuredColour(ts, 438165, 1000000, "butler", "deepseek-flash", true)
		if !strings.Contains(meas, g+"438165"+r) {
			t.Errorf("measured payload token number not green: %q", meas)
		}
		if !strings.Contains(meas, "- "+g+"butler"+r+" - ") {
			t.Errorf("measured payload mode not green: %q", meas)
		}

		// Disabled ⇒ byte-identical to the plain formatter.
		if got, want := formatPayloadMeasuredColour(ts, 438165, 1000000, "butler", "deepseek-flash", false), FormatPayloadMeasured(ts, 438165, 1000000, "butler", "deepseek-flash"); got != want {
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
		line := formatPayloadMeasuredColour(ts, 1, 1000000, "", "m", true)
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

// Round 057 (ADR 0027): the grey `[Tool Output]` frame + the yellow `[Tool Action]`
// line, and the estimated-payload increment line.
func TestChromeColourRound057(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 9, 19, 6, 25, 11, 0, time.UTC)
	const gray = "\033[0;90m"
	const yellow = "\033[0;33m"
	const r = "\033[0m"

	t.Run("tool output header + separators are grey; block plain when off", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		w := &ToolOutputWriter{W: &buf, Now: func() time.Time { return ts }, Colour: true}
		w.Begin()
		_, _ = w.Write([]byte("hello\n"))
		w.End()
		out := buf.String()
		if !strings.Contains(out, gray+"[06:25:11] [Tool Output] Executing... (Output shown below)"+r) {
			t.Errorf("header not grey: %q", out)
		}
		if n := strings.Count(out, gray+ToolOutputSeparator+r); n != 2 {
			t.Errorf("expected both separators grey; got %d in %q", n, out)
		}

		var plain strings.Builder
		pw := &ToolOutputWriter{W: &plain, Now: func() time.Time { return ts }}
		pw.Begin()
		_, _ = pw.Write([]byte("hello\n"))
		pw.End()
		if got := plain.String(); got != "[06:25:11] [Tool Output] Executing... (Output shown below)\n"+ToolOutputSeparator+"\n[06:25:11] [Tool Output] hello\n"+ToolOutputReset+ToolOutputSeparator+"\n" {
			t.Errorf("plain block = %q", got)
		}
	})

	t.Run("action line is yellow; plain when off", func(t *testing.T) {
		t.Parallel()
		line := formatToolActionColour(ts, "read_files", `{"path":"a"}`, true)
		if !strings.HasPrefix(line, yellow) || !strings.HasSuffix(line, r) {
			t.Errorf("action line not wrapped yellow: %q", line)
		}
		if !strings.Contains(line, "[06:25:11] [Tool Action] read_files(path: a)") {
			t.Errorf("action text missing: %q", line)
		}
		if got, want := formatToolActionColour(ts, "read_files", `{"path":"a"}`, false), FormatToolAction(ts, "read_files", `{"path":"a"}`); got != want {
			t.Errorf("colour-off action = %q, want plain %q", got, want)
		}
	})

}

// TestChromeColourRound058 (ADR 0028): the streamed `[Tool Output]` CONTENT line
// is grey too — the whole block reads as one grey region — and the plain path is
// byte-identical (the content line included). Round-labelled so the behaviour is
// greppable by round (R-058-2).
func TestChromeColourRound058(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 9, 19, 6, 25, 11, 0, time.UTC)
	const gray = "\033[0;90m"
	const r = "\033[0m"

	var buf strings.Builder
	w := &ToolOutputWriter{W: &buf, Now: func() time.Time { return ts }, Colour: true}
	w.Begin()
	_, _ = w.Write([]byte("hello\n"))
	_, _ = w.Write([]byte("world\n"))
	w.End()
	out := buf.String()
	if !strings.Contains(out, gray+"[06:25:11] [Tool Output] hello"+r) ||
		!strings.Contains(out, gray+"[06:25:11] [Tool Output] world"+r) {
		t.Errorf("each content line must be grey; got %q", out)
	}
	// The content text is unchanged: the grey wraps the sanitized line only.
	if !strings.Contains(out, "[06:25:11] [Tool Output] hello") {
		t.Errorf("content text changed: %q", out)
	}

	// Colour OFF ⇒ the whole block is byte-identical (content line included).
	var plain strings.Builder
	pw := &ToolOutputWriter{W: &plain, Now: func() time.Time { return ts }}
	pw.Begin()
	_, _ = pw.Write([]byte("hello\n"))
	pw.End()
	if got := plain.String(); got != "[06:25:11] [Tool Output] Executing... (Output shown below)\n"+ToolOutputSeparator+"\n[06:25:11] [Tool Output] hello\n"+ToolOutputReset+ToolOutputSeparator+"\n" {
		t.Errorf("plain block = %q", got)
	}
}

// TestPayloadEstimateRound057 pins the round-057 (ADR 0027) estimated payload line:
// a signed increment, no budget, the green MODE accent, and the plain colour-off path.
func TestPayloadEstimateRound057(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 9, 19, 6, 25, 11, 0, time.UTC)
	t.Run("payload estimate: signed increment, no budget, green mode", func(t *testing.T) {
		t.Parallel()
		if got, want := FormatPayloadEstimate(ts, 203148, 100, "butler", "deepseek-flash"), "[06:25:11] Payload: +100 ~203148 tokens - butler - deepseek-flash"; got != want {
			t.Errorf("estimate = %q; want %q", got, want)
		}
		if got, want := FormatPayloadEstimate(ts, 5, 0, "butler", "m"), "[06:25:11] Payload: +0 ~5 tokens - butler - m"; got != want {
			t.Errorf("no-predecessor estimate = %q; want %q", got, want)
		}
		if got, want := FormatPayloadEstimate(ts, 203000, -1000, "butler", "m"), "[06:25:11] Payload: -1000 ~203000 tokens - butler - m"; got != want {
			t.Errorf("shrunken estimate = %q; want %q", got, want)
		}
		if strings.Contains(FormatPayloadEstimate(ts, 5, 0, "butler", "m"), "/") {
			t.Errorf("estimate must not show a budget")
		}
		col := formatPayloadEstimateColour(ts, 203148, 100, "butler", "deepseek-flash", true)
		if !strings.Contains(col, "- "+"\033[0;32m"+"butler"+"\033[0m"+" - ") {
			t.Errorf("estimate mode not green: %q", col)
		}
		if strings.Contains(col, "\033[0;32m"+"203148"+"\033[0m") {
			t.Errorf("estimated token number must stay plain: %q", col)
		}
		if got, want := formatPayloadEstimateColour(ts, 1, 0, "butler", "m", false), FormatPayloadEstimate(ts, 1, 0, "butler", "m"); got != want {
			t.Errorf("colour-off estimate = %q, want plain %q", got, want)
		}
	})
}
