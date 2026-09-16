package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T016 [BDD-RED] — Then: the progress spinner does not appear while the command's output streams
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner does not appear while the command's output streams$`, thenNoSpinnerWhileStreaming)
	})
}

// thenNoSpinnerWhileStreaming (必查 呈現結果): within the `[Tool Output]` block
// (header through last streamed line), stderr carries no live spinner status
// line — the spinner is yielded once per call and is not visible while a shell
// command streams (round 034 FR-012 / ADR 0005 D7 / G8).
func thenNoSpinnerWhileStreaming(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	head, last := toolOutputBlockIndexes(sc.stderr)
	if head < 0 {
		return fmt.Errorf("standard error carried no `[Tool Output]` block to inspect; stderr=%q", sc.stderr)
	}
	if hasSpinnerStatusBetween(lines, head, last) {
		return fmt.Errorf("a live spinner status line appeared inside the `[Tool Output]` block; stderr=%q", sc.stderr)
	}
	return nil
}
