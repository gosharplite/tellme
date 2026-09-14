package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Then: the progress spinner names the model it is waiting for.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner names the model it is waiting for$`, thenSpinnerNamesModel)
	})
}

// thenSpinnerNamesModel (必查 / 呈現結果): the captured stderr carries a spinner line
// whose status is the `Thinking [<model>]...` form carrying the active provider's
// configured MODEL attribute (reference parity — not the registry key). 不該發生:
// the model-phase spinner must not omit the model label.
func thenSpinnerNamesModel(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	model, err := spinnerSelectedModel(sc)
	if err != nil {
		return err
	}
	if model == "" {
		return fmt.Errorf("the scenario's active provider names no MODEL")
	}
	want := "Thinking [" + model + "]..."
	if !strings.Contains(sc.stderr, want) {
		return fmt.Errorf("the spinner did not name the model %q (want %q): stderr=%q", model, want, sc.stderr)
	}
	return nil
}
