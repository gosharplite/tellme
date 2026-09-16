package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T015 [BDD-RED] — Then: the last post-turn status trails the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the last post-turn status trails the answer$`, thenLastStatusTrailsAnswer)
	})
}

// thenLastStatusTrailsAnswer (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the LAST `╰─⠿ Ready` tail appears AFTER the answer bytes — the final
// call's tail is deferred past the answer (round 034 FR-008 / G5).
func thenLastStatusTrailsAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	if sc.scriptedAnswer == "" {
		return fmt.Errorf("scenario did not record a scripted answer (last-tail witness)")
	}
	m := mergedView(sc)
	lastReady := strings.LastIndex(m, readyMarker)
	if lastReady < 0 {
		return fmt.Errorf("merged capture carried no `╰─⠿ Ready` tail; merged=%q", m)
	}
	answerAt := strings.Index(m, sc.scriptedAnswer)
	if answerAt < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	if lastReady < answerAt {
		return fmt.Errorf("the last post-turn status did not trail the answer; merged=%q", m)
	}
	return nil
}
