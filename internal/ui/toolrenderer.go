package ui

import (
	"strings"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
)

// ToolLineRenderer is the production implementation of
// agentport.ToolLineRenderer (round 046 / ADR 0015): a thin adapter over the pure
// tool-line formatters. It is the SINGLE production owner of the blank-reason
// predicate — ReasonLine evaluates the one reason transform (`toolReasonText`)
// ONCE and derives both the rendered line and the `renders` decision.
//
// The loop (internal/agent) holds this via the port, so `internal/agent` no
// longer imports `internal/ui` (the RULE-A edge that the layer-discipline
// baseline recorded). The loop keeps the write schedule; this adapter owns the
// bytes.
type ToolLineRenderer struct{ colour bool }

// Compile-time conformance to the port the loop consumes.
var _ agentport.ToolLineRenderer = ToolLineRenderer{}

// EngineLine renders the per-executed-round `[Tool Engine] Step i/M` marker.
func (ToolLineRenderer) EngineLine(t time.Time, step, total int) string {
	return FormatToolEngine(t, step, total)
}

// ActionLine renders the `[Tool Action] <tool>(<args>)` line. Round 057
// (ADR 0027): the whole line is yellow when the adapter's colour flag is set.
func (r ToolLineRenderer) ActionLine(t time.Time, tool, arguments string) string {
	return formatToolActionColour(t, tool, arguments, r.colour)
}

// ResultLine renders the `[Tool Result] <tool>: <snippet>` line.
func (ToolLineRenderer) ResultLine(t time.Time, tool, result string) string {
	return FormatToolResult(t, tool, result)
}

// ReasonLine renders the `[Tool Reason] <reason>` line and reports whether it
// renders at all. It evaluates `toolReasonText` ONCE: a blank (empty /
// whitespace-only) or escape-only reason renders nothing (("", false)); otherwise
// the line is built by the SHARED formatToolReasonLine helper (round-046 review
// TD-6), so it is byte-identical to FormatToolReason(t, reason). Round 054
// (ADR 0023): the whole line is green when the adapter's colour flag is set.
func (r ToolLineRenderer) ReasonLine(t time.Time, reason string) (string, bool) {
	text := toolReasonText(reason)
	if strings.TrimSpace(text) == "" {
		return "", false
	}
	return green(formatToolReasonLine(t, text), r.colour), true
}
