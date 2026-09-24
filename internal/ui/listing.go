package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/gosharplite/tellme/internal/domain/render"
)

// listing is the render.Listing adapter (round 073; ADR 0045). It owns the
// listing's bytes: the role header lines, the model-body rendering, the blank
// separator, and the terminal-gated header accents. The renderer it uses is the
// SAME one the answer path uses (NewRenderer), so there is one Markdown
// rendering policy, not two.
type listing struct {
	r *Renderer
}

// NewListing builds the production history-listing renderer.
func NewListing() render.Listing { return listing{r: NewRenderer()} }

// Render writes the listing for msgs. Per message: the header line, the body
// normalised to end in exactly one newline, then one additional newline (the
// separator) — so consecutive messages are separated by exactly one blank line
// and the listing ends with a blank line. The model body is rendered as Markdown
// (the operator body is verbatim, a recorded divergence from the reference); Raw
// prints the model body verbatim and disables the header accents.
//
// Round 086 (ADR 0057): a MODEL message is prefixed by the turn's tool-activity
// line — `[TOOLS] - M (N calls)`, emitted as its own block (a blank line before
// and after) so it reads between the turn's `[USER]` block and the `[MODEL]`
// header. It rides the MODEL message (so a partially-listed turn still shows it);
// the whole label is yellow only under the colour gate.
func (l listing) Render(w io.Writer, msgs []render.ListingMessage, spec render.ListingSpec) {
	colour := spec.Colour && !spec.Raw
	for _, m := range msgs {
		if m.Role == render.ListingModel {
			_, _ = fmt.Fprintln(w, l.toolLine(m, colour))
			_, _ = fmt.Fprintln(w)
		}
		_, _ = fmt.Fprintln(w, l.header(m, colour))
		_, _ = fmt.Fprintln(w, l.body(m, spec))
		// One blank line after every message (the blank separator), so
		// consecutive messages cannot run together and the listing ends with a
		// blank line — the reference's Fprintln-after-every-block shape.
		_, _ = fmt.Fprintln(w)
	}
}

// toolLine renders the round-086 per-turn tool-activity line: `[TOOLS] - M (N
// calls)` (the turn's backward index M and its tool-call count N), or the bare
// `[TOOLS] (N calls)` when M is non-positive (a non-history caller — the
// round-082 bare-label fallback; never a ` - 0` suffix). The WHOLE label is the
// colour unit and is wrapped yellow only under the colour gate (the round-073
// stdout-terminal gate + `-r` off).
func (l listing) toolLine(m render.ListingMessage, colour bool) string {
	label := "[TOOLS]"
	if m.TurnIndex > 0 {
		label = fmt.Sprintf("%s - %d", label, m.TurnIndex)
	}
	label = fmt.Sprintf("%s (%d calls)", label, m.ToolCount)
	return yellow(label, colour)
}

// header renders the role header line: `[USER] - N` for the operator, `[MODEL] -
// N` for the model, accented (blue / magenta) only when colour is on.
//
// Round 082 (ADR 0054): the label carries the message's backward turn index N
// (1 = the most recent turn), so the listing maps 1:1 to `tellme -b [N]`. The
// WHOLE label (role word + index) is the colour unit. A non-positive index (a
// non-history caller) falls back to the bare `[USER]` / `[MODEL]` label.
func (l listing) header(m render.ListingMessage, colour bool) string {
	role := render.ListingOperator
	label := "[USER]"
	if m.Role == render.ListingModel {
		role = render.ListingModel
		label = "[MODEL]"
	}
	if m.TurnIndex > 0 {
		label = fmt.Sprintf("%s - %d", label, m.TurnIndex)
	}
	if role == render.ListingOperator {
		return blue(label, colour)
	}
	return magenta(label, colour)
}

// body returns the message body ready to be written as one line-terminated
// block: the model body rendered as Markdown (or verbatim under Raw), the
// operator body verbatim. It ensures the returned string ends in exactly one
// newline, so the caller's Fprintln contributes exactly the blank separator.
func (l listing) body(m render.ListingMessage, spec render.ListingSpec) string {
	text := m.Body
	if m.Role == render.ListingModel && !spec.Raw {
		rendered, degraded := l.r.Render(text, spec.Width)
		if degraded && spec.Warn != nil {
			l.r.WarnDegraded(spec.Warn)
		}
		// On degrade, Render returns the SANITIZED raw text (round-006 D5; the
		// answer path consumes that same value) — keep it, never the unsanitized
		// original.
		text = strings.Trim(rendered, "\n")
	}
	return strings.TrimRight(text, "\n")
}
