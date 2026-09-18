package agent

import "time"

// ToolLineRenderer renders the agent loop's four diagnostic tool-line kinds and
// decides whether a model-authored `reason` renders a `[Tool Reason]` row.
//
// It is the loop's presentation-facing port, a peer to LoopObserver. The split of
// responsibilities is the round-046 rule (ADR 0015):
//
//   - the LOOP owns the SCHEDULE — which lines, when, in what order, written to
//     its own Stderr (including the round-039 per-call leading blank); and
//   - the RENDERER owns the BYTES — the pure formatting (the single-owned
//     `internal/ui` sanitize/fold/cap chain) and the blank-reason predicate.
//
// Why the port exists: `internal/agent` (tier 4) previously imported
// `internal/ui` (tier 5) — an upward RULE-A edge, and the last entry in the
// layer-discipline baseline (ADR 0011). Routing the loop's rendering through this
// domain-declared port removes that edge (agent -> domain/agent is downward)
// while the formatting stays single-owned in `internal/ui` (ADR 0015).
//
// The port carries NO presentation types; its implementation (`ui.ToolLineRenderer`)
// adapts the pure formatters. Every method takes the caller's clock reading so the
// line bytes are stamped from the loop's injected `Now` seam and stay identical.
type ToolLineRenderer interface {
	// EngineLine renders the per-executed-round `[Tool Engine] Step i/M` marker.
	EngineLine(t time.Time, step, total int) string
	// ActionLine renders the `[Tool Action] <tool>(<args>)` line.
	ActionLine(t time.Time, tool, arguments string) string
	// ResultLine renders the `[Tool Result] <tool>: <snippet>` line.
	ResultLine(t time.Time, tool, result string) string
	// ReasonLine renders the `[Tool Reason] <reason>` line and reports whether it
	// renders at all. ONE evaluation of the single reason transform decides both
	// the line and the decision, so the blank-reason predicate has one owner on
	// the real path. When `renders` is false — a blank (empty / whitespace-only)
	// or escape-only reason — the `line` return value is **UNSPECIFIED**: the
	// production adapter returns ("", false), and a caller MUST honour `renders`
	// (never the line's content) and MUST NOT print `line` when `renders` is
	// false. The `renders` result is exhaustive for a given reason — the loop uses
	// it for both the begin line and the tail filter (ADR 0015). This
	// documented-unspecified wording is deliberate: the loop's suppression witness
	// exercises a renderer whose suppression return is *non-empty* (round-046
	// fold-review N-3/F-1), proving the loop depends on `renders`, not on a
	// particular suppressed-line spelling.
	ReasonLine(t time.Time, reason string) (line string, renders bool)
}
