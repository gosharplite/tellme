package tools

import "io"

// OutputSink streams a shell command's complete output lines as `[Tool Output]`
// diagnostics (round 034 ADR 0005 D7 / FR-010; relocated to the domain by round
// 044 / ADR 0013). Begin/End bracket the block — the CLI stops/resumes the
// spinner and writes the header/separator lines there; the command tool tees the
// child's stdout/stderr bytes into Writer, which assembles complete lines
// (dropping a trailing partial). A zero sink (nil Writer) emits no block, and the
// block renders unconditionally (pause/resume is a no-op when the spinner is
// gated off).
//
// It is a neutral domain port (no infrastructure type) so the application layer
// (internal/cli) names no infrastructure package while still binding the sink
// from its own `ui` coordinator (round-044 composition-root extraction).
type OutputSink struct {
	// Begin opens the block (once per call, before the child starts).
	Begin func()
	// Writer receives the child's raw stdout/stderr bytes; the CLI's
	// ui.ToolOutputWriter assembles the complete `[Tool Output]` lines.
	Writer io.Writer
	// End closes the block (once, after the child is reaped).
	End func()
}

// Enabled reports whether the sink is bound (a writer is set). It is the
// zero-value-is-disabled contract the command tool relies on at four call sites
// (round-044 fix-3: kept as a METHOD so the zero value stays safe — an
// `Enabled func() bool` field would default to nil and panic).
func (s OutputSink) Enabled() bool { return s.Writer != nil }
