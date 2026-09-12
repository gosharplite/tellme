package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T009 — Then: the captured standard output is the rendered answer, not its raw Markdown
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output is the rendered answer, not its raw Markdown$`, thenCapturedIsRendered)
	})
}

// thenCapturedIsRendered (必查 呈現結果): the captured stdout carries the answer's
// rendered form — (i) the literal Markdown emphasis markers the scripted answer
// carries (`**`, `_`) are absent, AND (ii) the answer's underlying words are
// present (round-006 research Decision 1/4; M4 — the words-present half makes a
// renderer that emits nothing fail). The renderer's ANSI escapes are stripped
// first (terminal/profile-dependent).
func thenCapturedIsRendered(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !sc.scriptedAnswerSet {
		return fmt.Errorf("the rendered check requires a scripted answer (Given: a configured provider … whose endpoint answers with …)")
	}
	visible := harness.StripANSI(sc.stdout)
	for _, marker := range []string{"**", "_"} {
		if strings.Contains(sc.scriptedAnswer, marker) && strings.Contains(visible, marker) {
			return fmt.Errorf("captured output still carries the literal Markdown marker %q (not rendered): %q", marker, sc.stdout)
		}
	}
	words := strings.NewReplacer("**", "", "_", "").Replace(sc.scriptedAnswer)
	if !strings.Contains(visible, words) {
		return fmt.Errorf("captured output %q does not carry the rendered answer words %q", visible, words)
	}
	return nil
}
