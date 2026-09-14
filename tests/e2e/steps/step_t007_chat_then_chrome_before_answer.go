package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T007 — Then: the turn chrome is shown before the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn chrome is shown before the answer$`, thenTurnChromeBeforeAnswer)
	})
}

// thenTurnChromeBeforeAnswer (必查 呈現結果): in the MERGED (stdout+stderr)
// capture, the WHOLE turn chrome — the acknowledgement, the rule, the header, and
// the pre-flight payload line — appears BEFORE the answer bytes.
func thenTurnChromeBeforeAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	m := mergedView(sc)
	ans := strings.Index(m, sc.scriptedAnswer)
	if ans < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	parts := []struct {
		name string
		idx  int
	}{
		{"input-capture acknowledgement", firstMatchIndex(reTurnAck, m)},
		{"horizontal rule", strings.Index(m, turnRuleLiteral())},
		{"turn header", firstMatchIndex(reTurnHeader, m)},
		{"pre-flight payload line", strings.Index(m, "Payload: ")},
	}
	for _, p := range parts {
		if p.idx < 0 {
			return fmt.Errorf("merged capture carried no %s; merged=%q", p.name, m)
		}
		if p.idx >= ans {
			return fmt.Errorf("the %s (idx %d) did not precede the answer (idx %d); merged=%q", p.name, p.idx, ans, m)
		}
	}
	return nil
}
