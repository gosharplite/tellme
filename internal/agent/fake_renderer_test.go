package agent

import (
	"fmt"
	"strings"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
)

// fakeRenderer is a deterministic agentport.ToolLineRenderer used by the loop's
// unit tests. The loop's own _test.go files may NOT import internal/ui — the
// layer-discipline gate governs test imports too (ADR 0015) — so the tests pin
// the loop's SCHEDULE (which lines, in what order, with which blanks) via this
// fake's markers, while the real formatting is pinned at the ui tier + E2E.
//
// ReasonLine mirrors the production predicate CONTRACT: a reason that renders
// returns a marker line + true; a reason that does not render returns ("", false).
// The default decision is "non-blank after trimming"; a test may override it via
// `renders` to pin that the loop merely HONORS the renderer's decision (e.g. an
// escape-only reason the real sanitizer blanks).
type fakeRenderer struct {
	// renders, when non-nil, overrides the default ReasonLine decision.
	renders func(reason string) bool
}

func (fakeRenderer) EngineLine(_ time.Time, step, total int) string {
	return fmt.Sprintf("ENGINE %d/%d", step, total)
}

func (fakeRenderer) ActionLine(_ time.Time, tool, arguments string) string {
	return fmt.Sprintf("ACTION %s(%s)", tool, arguments)
}

func (fakeRenderer) ResultLine(_ time.Time, tool, result string) string {
	return fmt.Sprintf("RESULT %s: %s", tool, result)
}

func (f fakeRenderer) ReasonLine(_ time.Time, reason string) (string, bool) {
	renders := strings.TrimSpace(reason) != ""
	if f.renders != nil {
		renders = f.renders(reason)
	}
	if !renders {
		// Round-046 fold-review F-1: the suppressed case returns a DISTINGUISHABLE
		// sentinel (not ""), so a loop that ignored `renders` and printed the line
		// anyway would emit a line containing "REASON " — which the suppression
		// assertions (`!Contains(log, "REASON ")`) catch. A `""` return made the
		// mutant's output (a bare newline) invisible, so the pin lost its power.
		return "REASON suppressed " + reason, false
	}
	return "REASON " + reason, true
}

var _ agentport.ToolLineRenderer = fakeRenderer{}
