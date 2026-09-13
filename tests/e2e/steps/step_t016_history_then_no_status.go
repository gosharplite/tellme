package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T016 — Then: no payload status is reported
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^no payload status is reported$`, thenNoPayloadStatus)
	})
}

// thenNoPayloadStatus (必查 呈現結果): the captured standard error carries no
// payload status line at all (a non-prompt run).
func thenNoPayloadStatus(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if hasPayloadStatusLine(sc.stderr) {
		return fmt.Errorf("a non-prompt run reported a payload status; stderr=%q", sc.stderr)
	}
	return nil
}
