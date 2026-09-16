package steps

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/godog"
)

// [BDD-RED] — Then: the run reported the result for the tool call "{tool}" shortened to at most {runes} runes
//
// Round 034 rune-cap row (chat/dsl.md line 380; the tasks.md Phase-3 list listed
// 15 rows but the Feature carries 17 — this and T006b close the gap). The
// `[Tool Result] {tool}: <snippet>` line's snippet is ≤ {runes} runes and ends
// with exactly one U+2026 (FR-004 / G3).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the result for the tool call "([^"]*)" shortened to at most ([0-9]+) runes$`, thenResultShortened)
	})
}

// thenResultShortened (必查 呈現結果): the result snippet is rune-capped.
func thenResultShortened(ctx context.Context, tool, cap string) error {
	sc := scenarioFrom(ctx)
	n, err := strconv.Atoi(cap)
	if err != nil {
		return fmt.Errorf("invalid rune cap %q: %w", cap, err)
	}
	snippet, ok := resultSnippetFor(sc.stderr, tool)
	if !ok {
		return fmt.Errorf("standard error carried no `[Tool Result] %s: …` line; stderr=%q", tool, sc.stderr)
	}
	if got := utf8.RuneCountInString(snippet); got > n {
		return fmt.Errorf("the result snippet is %d runes, over the %d-rune cap; snippet=%q", got, n, snippet)
	}
	if !strings.HasSuffix(snippet, "…") {
		return fmt.Errorf("the shortened result snippet does not end with one U+2026; snippet=%q", snippet)
	}
	return nil
}

// resultSnippetFor returns the snippet after `[Tool Result] <tool>: `.
func resultSnippetFor(stderr, tool string) (string, bool) {
	prefix := toolResultMarker + tool + ": "
	for _, l := range stderrLines(stderr) {
		if i := strings.Index(l, prefix); i >= 0 {
			return strings.TrimRight(l[i+len(prefix):], "\r"), true
		}
	}
	return "", false
}
