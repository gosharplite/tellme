package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T004 — Then: the estimated payload status is reported before the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the estimated payload status is reported before the answer$`, thenEstimatedStatusBeforeAnswer)
	})
}

// thenEstimatedStatusBeforeAnswer (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the pre-flight estimated (`~`) payload status line appears BEFORE the
// answer bytes. Runs against the merged cross-stream witness (round 010).
func thenEstimatedStatusBeforeAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	m := mergedView(sc)
	est := firstMatchIndex(reEstimatedPayload, m)
	ans := strings.Index(m, sc.scriptedAnswer)
	if est < 0 {
		return fmt.Errorf("merged capture carried no estimated payload status line; merged=%q", m)
	}
	if ans < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	if est >= ans {
		return fmt.Errorf("estimated payload status (idx %d) did not precede the answer (idx %d); merged=%q", est, ans, m)
	}
	return nil
}
