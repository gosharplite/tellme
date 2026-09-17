package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// Round 040 (issue #83) — WS-B dual elapsed timer (ADR 0009 D1/D2). The dual-timer
// Then (T005).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner shows the time since the prompt and the current model call's time$`, thenSpinnerDualTimer)
	})
}

// thenSpinnerDualTimer (必查 / 呈現結果): the captured standard error carries a
// spinner line whose elapsed segment is TWO whole-second figures inside one
// parentheses — `(<total>s <call>s)` — the first the total since prompt capture,
// the second the current model call's elapsed, both unlabelled (ADR 0009 D1).
// 不該發生: the spinner must not show a single figure, and must not write the
// spinner to standard output. (The per-call reset arithmetic cannot be observed
// E2E on a fast scripted turn — it is the T006 unit pin, SC-003.)
func thenSpinnerDualTimer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reSpinnerElapsed.MatchString(sc.stderr) {
		return fmt.Errorf("the spinner carries no dual `(<total>s <call>s)` elapsed segment: stderr=%q", sc.stderr)
	}
	if hasSpinnerFrame(sc.stdout) {
		return fmt.Errorf("the progress spinner was written to standard output: stdout=%q", sc.stdout)
	}
	return nil
}
