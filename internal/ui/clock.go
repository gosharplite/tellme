package ui

import "time"

// formatClock renders the shared `HH:MM:SS` clock token used by every
// `internal/ui` timestamp formatter — the payload status line
// (FormatPayloadStatus), the input-capture acknowledgement (FormatInputCaptured),
// the post-turn metrics line (FormatMetrics), and the tool-loop log line
// (FormatToolLog) — so the four `stderr` surfaces stay in lockstep (round-022
// research Decision 2 / review R-1). The caller supplies the time via the
// injected clock seam, so assertions stay deterministic.
func formatClock(t time.Time) string { return t.Format("15:04:05") }
