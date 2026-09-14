package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T015 — Then: the run reports the share of the prompt served from cache
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports the share of the prompt served from cache$`, thenCacheShare)
	})
}

// thenCacheShare (必查 呈現結果): the summary line carries a cache-hit percentage.
func thenCacheShare(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	_, _, _, _, pct, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line with a cache share; stderr=%q", sc.stderr)
	}
	if pct < 0 || pct > 100 {
		return fmt.Errorf("cache share %.1f%% is out of range", pct)
	}
	return nil
}
