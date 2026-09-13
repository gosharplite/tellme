package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T006 — Then: the tool activity is reported before the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tool activity is reported before the answer$`, thenToolActivityBeforeAnswer)
	})
}

// thenToolActivityBeforeAnswer (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the tool-loop log line appears BEFORE the answer bytes. The log line
// is located by the scripted tool name the scenario arranged (not by any log
// token — truth does not pin the log format). Runs against the merged
// cross-stream witness (round 010).
func thenToolActivityBeforeAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	if sc.scriptedTool == "" {
		return fmt.Errorf("scenario did not record a scripted tool (round-010 tool-ordering witness)")
	}
	m := mergedView(sc)
	tool := strings.Index(m, sc.scriptedTool)
	ans := strings.Index(m, sc.scriptedAnswer)
	if tool < 0 {
		return fmt.Errorf("merged capture carried no tool-loop log line naming %q; merged=%q", sc.scriptedTool, m)
	}
	if ans < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	if tool >= ans {
		return fmt.Errorf("tool activity (idx %d) did not precede the answer (idx %d); merged=%q", tool, ans, m)
	}
	return nil
}
