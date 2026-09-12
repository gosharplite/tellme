package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T009 — Then: the captured standard output is the rendered answer, not its raw Markdown
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output is the rendered answer, not its raw Markdown$`, thenCapturedIsRendered)
	})
}

// thenCapturedIsRendered (必查 呈現結果): the captured stdout carries the answer's
// rendered form — the literal Markdown emphasis markers the scripted answer
// carries (`**`, `_`) are absent (round-006 research Decisions 1 & 4).
func thenCapturedIsRendered(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !sc.scriptedAnswerSet {
		return fmt.Errorf("the rendered check requires a scripted answer (Given: a configured provider … whose endpoint answers with …)")
	}
	for _, marker := range []string{"**", "_"} {
		if strings.Contains(sc.scriptedAnswer, marker) && strings.Contains(sc.stdout, marker) {
			return fmt.Errorf("captured output still carries the literal Markdown marker %q (not rendered): %q", marker, sc.stdout)
		}
	}
	return nil
}
