package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T009 — Then: the reported metrics line shows {miss} missed, {cached} cached, {completion} completed, and {thinking} reasoning tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the reported metrics line shows ([0-9]+) missed, ([0-9]+) cached, ([0-9]+) completed, and ([0-9]+) reasoning tokens$`, thenMetricsValues)
	})
}

// thenMetricsValues (必查 呈現結果): the reported metrics line's M/H/C/Th equal the
// cited counts (with M = prompt − cached); Th is present even when zero.
func thenMetricsValues(ctx context.Context, miss, hit, completion, thinking int) error {
	sc := scenarioFrom(ctx)
	_, m, h, c, th, ok := metricsValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no metrics line; stderr=%q", sc.stderr)
	}
	if m != miss || h != hit || c != completion || th != thinking {
		return fmt.Errorf("metrics line = M:%d H:%d C:%d Th:%d, want M:%d H:%d C:%d Th:%d",
			m, h, c, th, miss, hit, completion, thinking)
	}
	return nil
}
