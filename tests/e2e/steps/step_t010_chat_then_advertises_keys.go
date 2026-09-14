package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T010 — Then: the interactive prompt advertises the submit and abort keys
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt advertises the submit and abort keys$`, thenAdvertisesKeys)
	})
}

// thenAdvertisesKeys (必查 呈現結果): the captured output carries the editor
// placeholder naming the submit key (`Ctrl+S`) and the abort key (`Esc`).
func thenAdvertisesKeys(ctx context.Context) error {
	out := renderedOutput(scenarioFrom(ctx))
	if !strings.Contains(out, "Ctrl+S") || !strings.Contains(out, "Esc") {
		return fmt.Errorf("the interactive prompt did not advertise the submit/abort keys; output=%q", out)
	}
	return nil
}
