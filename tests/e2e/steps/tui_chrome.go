package steps

import (
	"strings"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// Round-016 chrome helpers for the strict-parity `-i` prompt assertions. Kept in
// ONE file so the per-sentence step files stay independent (Zero Shared Edits),
// mirroring tui_keys.go / shared_prompt_log.go.

// tuiVisibleLines splits captured output into ANSI-stripped visible lines.
func tuiVisibleLines(out string) []string {
	return strings.Split(harness.StripANSI(out), "\n")
}

// tuiHasBorder reports whether the output draws the editor border (NormalBorder
// runes) — the "framed around a multi-line editor" predicate.
func tuiHasBorder(out string) bool {
	for _, ln := range tuiVisibleLines(out) {
		if strings.Contains(ln, "┌") || strings.Contains(ln, "└") {
			return true
		}
	}
	return false
}

// tuiSuggestionsBeneath reports whether the `Suggestions:` header appears (the
// list is rendered beneath the editor frame).
func tuiSuggestionsBeneath(out string) bool {
	for _, ln := range tuiVisibleLines(out) {
		if strings.Contains(ln, suggesterHeaderLiteral) {
			return true
		}
	}
	return false
}

// suggesterHeaderLiteral mirrors the product's suggester header (single-sourced
// by the step's assertion, so a product-side change cannot make the guard
// vacuous — the round-012 TD1 pattern).
const suggesterHeaderLiteral = "Suggestions:"

// tuiCursorRows counts suggestion rows carrying the `>` selection cursor. The
// rendered cursor row is exactly the padded `"  > "` row (modelStyle Padding(1,1)
// + suggesterStyle Padding(0,1) = 2 leading spaces, then the `> ` cursor prefix).
// An UNselected row renders its `  ` row prefix (4 leading spaces), so a
// suggestion whose TEXT itself begins with `> ` (e.g. `> quoted reply`) renders
// as `"    > quoted reply"` and is correctly NOT counted — an exact-prefix match,
// not a `TrimLeft` "any `> `-leading line" proxy (round-037 review G-2; the proxy
// produced a false failure when the flipped exclusion ran on such a suggestion).
func tuiCursorRows(out string) int {
	n := 0
	for _, ln := range tuiVisibleLines(out) {
		if strings.HasPrefix(ln, "  > ") {
			n++
		}
	}
	return n
}

// tuiHasMetricsHeader reports whether any line matches the round-015 dashboard
// pattern `provider … | tokens … | turns …` (round-016 F3: a SPECIFIC pattern, not
// a bare `tokens`/`turns` substring).
func tuiHasMetricsHeader(out string) bool {
	for _, ln := range tuiVisibleLines(out) {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "provider ") && strings.Contains(t, "tokens ") && strings.Contains(t, "turns ") {
			return true
		}
	}
	return false
}

// tuiMaxLineWidth returns the longest visible line width in columns.
func tuiMaxLineWidth(out string) int {
	max := 0
	for _, ln := range tuiVisibleLines(out) {
		if n := len([]rune(ln)); n > max {
			max = n
		}
	}
	return max
}

// tuiOverlongSuggestionRow reports whether a rendered suggestion list row spans
// more than one line — a wrapped continuation carrying no `>`/`  ` row marker
// (round-016 FR-006: an entry over three lines is dropped, so a stray
// continuation means the drop failed).
func tuiOverlongSuggestionRow(out string) bool {
	inList := false
	for _, ln := range tuiVisibleLines(out) {
		if strings.Contains(ln, suggesterHeaderLiteral) {
			inList = true
			continue
		}
		if !inList {
			continue
		}
		t := strings.TrimRight(ln, "\r")
		if strings.TrimSpace(t) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(t, " "), "> ") {
			continue
		}
		if strings.HasPrefix(t, "  ") {
			continue
		}
		return true
	}
	return false
}

// tuiHasLineNumberGutter reports whether an editor row renders a line-number
// gutter (a bare leading number), which strict parity forbids.
func tuiHasLineNumberGutter(out string) bool {
	for _, ln := range tuiVisibleLines(out) {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		i := 0
		for i < len(t) && t[i] >= '0' && t[i] <= '9' {
			i++
		}
		if i > 0 && i < len(t) && t[i] == ' ' {
			return true
		}
	}
	return false
}

// tuiEditorRowContains reports whether an EDITOR row (a bordered line carrying
// the `│` border rune) contains the substring — distinguishing the editor
// content from a suggestion row, which also renders the text (so an
// insert-on-Tab assertion cannot pass vacuously off the suggestion list).
func tuiEditorRowContains(out, sub string) bool {
	for _, ln := range tuiVisibleLines(out) {
		if strings.Contains(ln, "│") && strings.Contains(ln, sub) {
			return true
		}
	}
	return false
}
