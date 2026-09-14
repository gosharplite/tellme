package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 — Then: the run reports the token metrics of the request that just completed
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports the token metrics of the request that just completed$`, thenMetricsReported)
	})
}

// thenMetricsReported (必查 呈現結果): the captured standard error carries a metrics
// line matching `[HH:MM:SS] [<provider>] M: <n> H: <n> C: <n> Th: <n>` (the
// timestamp matched by pattern); it must NOT be written to standard output.
func thenMetricsReported(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasMetricsLine(sc.stderr) {
		return fmt.Errorf("standard error carried no metrics line; stderr=%q", sc.stderr)
	}
	if hasMetricsLine(sc.stdout) {
		return fmt.Errorf("the metrics line must not be written to standard output; stdout=%q", sc.stdout)
	}
	return nil
}
