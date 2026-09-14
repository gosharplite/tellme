package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T011 [BDD-RED] — Then: the progress spinner no longer appears once the answer is written.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner no longer appears once the answer is written$`, thenSpinnerClearedAfterAnswer)
	})
}

// thenSpinnerClearedAfterAnswer (必查 / 呈現結果): in the MERGED (stdout+stderr)
// capture, no spinner frame (a braille frame followed by a phase status) appears
// after the answer bytes. 不該發生: the spinner must not survive into or after the
// answer. Uses the round-010 merged-stream witness.
func thenSpinnerClearedAfterAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.captureMerged()
	mv := mergedView(sc)
	answer := sc.scriptedAnswer
	if answer == "" {
		return fmt.Errorf("the scenario recorded no scripted answer to locate")
	}
	idx := strings.Index(mv, answer)
	if idx < 0 {
		return fmt.Errorf("the answer %q is missing from the merged capture: %q", answer, mv)
	}
	after := mv[idx+len(answer):]
	if hasSpinnerFrame(after) {
		return fmt.Errorf("a spinner frame survives after the answer: %q", after)
	}
	return nil
}
