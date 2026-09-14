package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Then: the progress spinner shows how long it has waited.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner shows how long it has waited$`, thenSpinnerShowsElapsed)
	})
}

// thenSpinnerShowsElapsed (必查 / 呈現結果): the spinner line carries an `(<n>s)`
// elapsed segment (whole seconds). 不該發生: the elapsed segment must be missing.
func thenSpinnerShowsElapsed(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reSpinnerElapsed.MatchString(sc.stderr) {
		return fmt.Errorf("the spinner carries no elapsed segment: stderr=%q", sc.stderr)
	}
	return nil
}
