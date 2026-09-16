package tools

import "io"

// ToolOutputSink streams a shell command's complete output lines as `[Tool
// Output]` diagnostics (round 034 ADR 0005 D7 / FR-010). Begin/End bracket the
// block — the CLI stops/resumes the spinner and writes the header/separator
// lines there; the command tool tees the child's stdout/stderr bytes into
// Writer, which assembles complete lines (dropping a trailing partial). A zero
// sink (nil Writer) emits no block, and the block renders unconditionally
// (pause/resume is a no-op when the spinner is gated off).
type ToolOutputSink struct {
	// Begin opens the block (once per call, before the child starts).
	Begin func()
	// Writer receives the child's raw stdout/stderr bytes; the CLI's
	// ui.ToolOutputWriter assembles the complete `[Tool Output]` lines.
	Writer io.Writer
	// End closes the block (once, after the child is reaped).
	End func()
}

// Enabled reports whether the sink is bound (a writer is set).
func (s ToolOutputSink) Enabled() bool { return s.Writer != nil }
