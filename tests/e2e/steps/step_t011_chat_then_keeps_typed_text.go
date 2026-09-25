package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T011 — Then: the interactive prompt keeps the typed text "{text}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt keeps the typed text "([^"]*)"$`, thenKeepsTypedText)
	})
}

// thenKeepsTypedText (必查 呈現結果): the captured output carries each typed line in
// an EDITOR row (not merely in the suggestion list) — and, when the text spans
// several lines, each line sits on its OWN row. The multi-line clause is the
// round-093 strengthening: the pre-093 assertion only required each line to be a
// substring of SOME editor row, so a joined row (`line oneline two`) passed
// vacuously even though the product had dropped the line break (issue #191).
func thenKeepsTypedText(ctx context.Context, text string) error {
	out := renderedOutput(scenarioFrom(ctx))
	parts := typedLines(unescapeText(text))
	if len(parts) == 0 {
		return nil
	}
	rows := tuiEditorRows(out)
	for _, part := range parts {
		if !rowContains(rows, part) {
			return fmt.Errorf("the interactive prompt did not keep the typed text %q (missing %q in the editor); output=%q", text, part, out)
		}
	}
	// Multi-line: the typed lines must be kept APART — no single editor row may
	// carry two of them (a joined row is the defect this clause exists to catch).
	if len(parts) > 1 {
		if row := joinedRow(rows, parts); row != "" {
			return fmt.Errorf("the interactive prompt joined the typed lines onto one editor row %q (each line must be on its own row); output=%q", row, out)
		}
	}
	return nil
}

// joinedRow returns the first editor row that carries two distinct typed lines,
// or "" when every line sits on its own row. It is the multi-line half of the
// keep-the-typed-text assertion, extracted so a unit pin can redden on the joined
// state (the pre-093 `line oneline two` row).
func joinedRow(rows, parts []string) string {
	for _, row := range rows {
		carried := 0
		for _, part := range parts {
			if part != "" && strings.Contains(row, part) {
				carried++
			}
		}
		if carried > 1 {
			return row
		}
	}
	return ""
}

// typedLines returns the non-empty lines of a typed value.
func typedLines(text string) []string {
	parts := make([]string, 0, 2)
	for _, part := range strings.Split(text, "\n") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

// tuiEditorRows returns the ANSI-stripped editor rows (a bordered line carrying
// the `│` border rune) of the captured output.
func tuiEditorRows(out string) []string {
	rows := make([]string, 0, 8)
	for _, ln := range tuiVisibleLines(out) {
		if strings.Contains(ln, "│") {
			rows = append(rows, ln)
		}
	}
	return rows
}

// rowContains reports whether any row carries the substring.
func rowContains(rows []string, sub string) bool {
	for _, ln := range rows {
		if strings.Contains(ln, sub) {
			return true
		}
	}
	return false
}
