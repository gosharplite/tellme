package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T018 — Then: the run reports no post-turn status (interface root)
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports no post-turn status$`, thenNoPostTurnStatus)
	})
}

// thenNoPostTurnStatus (必查 呈現結果): neither the captured standard output nor
// standard error carries a metrics line or a `╰─⠿ Ready` summary line — used by
// the no-usage path and every non-prompt path (`--version`, `-l`, prompt-less
// `--new`).
func thenNoPostTurnStatus(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if hasMetricsLine(sc.stdout) || hasReadyLine(sc.stdout) {
		return fmt.Errorf("standard output carried the post-turn status; stdout=%q", sc.stdout)
	}
	if hasMetricsLine(sc.stderr) || hasReadyLine(sc.stderr) {
		return fmt.Errorf("standard error carried the post-turn status; stderr=%q", sc.stderr)
	}
	return nil
}
