package tools

import "io"

// OutputSink streams a shell command's complete output lines as `[Tool Output]`
// diagnostics (round 034 ADR 0005 D7 / FR-010; relocated to the domain by round
// 044 / ADR 0013). Begin/End bracket the block — the CLI stops/resumes the
// spinner and writes the header/separator lines there; the command tool tees the
// child's stdout/stderr bytes into Writer, which assembles complete lines
// (dropping a trailing partial). A nil sink emits no block, and the block renders
// unconditionally (pause/resume is a no-op when the spinner is gated off).
//
// It is a neutral domain port (no infrastructure type) so the application layer
// (internal/cli) names no infrastructure package. Round 051 (R5.5 of #92;
// ADR 0020) made it an INTERFACE the `internal/ui` coordinator satisfies
// DIRECTLY (review-deferral **F-8**), removing the former struct-of-funcs bridge
// and its partial-binding hole. The zero value is a nil interface; callers MUST
// guard with `sink != nil` before use.
type OutputSink interface {
	// Begin opens the block (once per call, before the child starts).
	Begin()
	// Writer receives the child's raw stdout/stderr bytes; the CLI assembles the
	// complete `[Tool Output]` lines from them.
	Writer() io.Writer
	// End closes the block (once, after the child is reaped).
	End()
	// Enabled reports whether the sink is live (a bound writer is set).
	Enabled() bool
}
