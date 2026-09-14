package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T017 — Then: the post-turn status trails the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the post-turn status trails the answer$`, thenStatusTrailsAnswer)
	})
}

// thenStatusTrailsAnswer (必查 呈現結果): in the MERGED (stdout+stderr) capture,
// both the metrics line and the Ready summary appear AFTER the answer bytes.
func thenStatusTrailsAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	ans := sc.scriptedAnswer
	if !sc.scriptedAnswerSet {
		return fmt.Errorf("no scripted answer recorded")
	}
	// The answer is rendered (ANSI-coloured) in the merged capture, so strip the
	// SGR sequences before searching for the plain answer text.
	merged := stripPostTurnANSI(sc.merged)
	ai := strings.LastIndex(merged, ans)
	if ai < 0 {
		return fmt.Errorf("the merged capture carried no answer %q; merged=%q", ans, merged)
	}
	mi := reMetrics.FindStringIndex(merged)
	if mi == nil || mi[0] < ai {
		return fmt.Errorf("the metrics line did not trail the answer; merged=%q", merged)
	}
	ri := reReady.FindStringIndex(merged)
	if ri == nil || ri[0] < ai {
		return fmt.Errorf("the Ready summary did not trail the answer; merged=%q", merged)
	}
	return nil
}
