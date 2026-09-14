package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T008 — Then: the turn frame is separated from the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn frame is separated from the answer$`, thenTurnFrameSeparated)
	})
}

// thenTurnFrameSeparated (必查 呈現結果): in the MERGED capture, a blank line
// separates the frame's last line (the pre-flight payload line) from the answer.
func thenTurnFrameSeparated(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	m := mergedView(sc)
	ans := strings.Index(m, sc.scriptedAnswer)
	if ans < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	p := strings.Index(m, "Payload: ")
	if p < 0 {
		return fmt.Errorf("merged capture carried no pre-flight payload line; merged=%q", m)
	}
	nl := strings.Index(m[p:], "\n")
	if nl < 0 {
		return fmt.Errorf("the payload line was not newline-terminated; merged=%q", m)
	}
	if p+nl+1 > ans {
		return fmt.Errorf("the payload line (end %d) did not precede the answer (idx %d); merged=%q", p+nl+1, ans, m)
	}
	// The bytes between the payload line's newline and the answer must open with
	// a newline — i.e. a blank line separates the frame from the answer.
	seg := m[p+nl+1 : ans]
	if !strings.HasPrefix(seg, "\n") {
		return fmt.Errorf("no blank line separated the frame from the answer; segment=%q", seg)
	}
	return nil
}
