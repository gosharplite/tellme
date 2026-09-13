package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T005 — Then: the measured payload status is reported after the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the measured payload status is reported after the answer$`, thenMeasuredStatusAfterAnswer)
	})
}

// thenMeasuredStatusAfterAnswer (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the post-turn measured payload status line appears AFTER the answer
// bytes. Runs against the merged cross-stream witness (round 010).
func thenMeasuredStatusAfterAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	m := mergedView(sc)
	meas := firstMatchIndex(reMeasuredPayload, m)
	ans := strings.Index(m, sc.scriptedAnswer)
	if meas < 0 {
		return fmt.Errorf("merged capture carried no measured payload status line; merged=%q", m)
	}
	if ans < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	if meas <= ans {
		return fmt.Errorf("measured payload status (idx %d) did not follow the answer (idx %d); merged=%q", meas, ans, m)
	}
	return nil
}
