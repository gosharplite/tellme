package steps

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/godog"
)

// [BDD-RED] — Then: the run reported the action value for the tool call "{tool}" shortened to at most {runes} runes
//
// Round 034 rune-cap row (chat/dsl.md line 381; the tasks.md Phase-3 list listed
// 15 rows but the Feature carries 17 — this and T007b close the gap). On the
// `[Tool Action] {tool}(…)` line every rendered argument value is ≤ {runes}
// runes, any over-cap value ending in exactly one U+2026 (FR-003 / G3).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the action value for the tool call "([^"]*)" shortened to at most ([0-9]+) runes$`, thenActionShortened)
	})
}

// thenActionShortened (必查 呈現結果): the action's argument values are rune-capped.
func thenActionShortened(ctx context.Context, tool, cap string) error {
	sc := scenarioFrom(ctx)
	n, err := strconv.Atoi(cap)
	if err != nil {
		return fmt.Errorf("invalid rune cap %q: %w", cap, err)
	}
	bodies := toolActionArgsFor(sc.stderr, tool)
	if len(bodies) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Action] %s(…)` line; stderr=%q", tool, sc.stderr)
	}
	for _, body := range bodies {
		for _, v := range splitArgValues(body) {
			if got := utf8.RuneCountInString(v); got > n {
				return fmt.Errorf("an argument value is %d runes, over the %d-rune cap; value=%q", got, n, v)
			}
		}
	}
	if !strings.Contains(strings.Join(bodies, ", "), "…") {
		return fmt.Errorf("no over-cap argument value was shortened with a U+2026; stderr=%q", sc.stderr)
	}
	return nil
}

// splitArgValues splits a rendered `k: v, k: v` argument body into its value
// strings (the text after the first `: ` of each comma-separated pair).
func splitArgValues(body string) []string {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	parts := strings.Split(body, ", ")
	vals := make([]string, 0, len(parts))
	for _, p := range parts {
		if i := strings.Index(p, ": "); i >= 0 {
			vals = append(vals, p[i+2:])
		} else {
			vals = append(vals, p)
		}
	}
	return vals
}
