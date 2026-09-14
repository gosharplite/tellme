package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T004 — Then: the input capture is announced before the turn frame
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the input capture is announced before the turn frame$`, thenInputCaptureBeforeFrame)
	})
}

// thenInputCaptureBeforeFrame (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the input-capture acknowledgement appears BEFORE the horizontal rule.
func thenInputCaptureBeforeFrame(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	m := mergedView(sc)
	ack := firstMatchIndex(reTurnAck, m)
	rule := strings.Index(m, turnRuleLiteral())
	if ack < 0 {
		return fmt.Errorf("merged capture carried no input-capture acknowledgement; merged=%q", m)
	}
	if rule < 0 {
		return fmt.Errorf("merged capture carried no horizontal rule; merged=%q", m)
	}
	if ack >= rule {
		return fmt.Errorf("the acknowledgement (idx %d) did not precede the rule (idx %d); merged=%q", ack, rule, m)
	}
	return nil
}
