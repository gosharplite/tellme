package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T017 — Then: tellme explains on stderr that "{reason}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme explains on stderr that "([^"]*)"$`, thenExplainsOnStderr)
	})
}

// thenExplainsOnStderr (必查 呈現結果): stderr carries a readable message
// corresponding to {reason}; the reason must not be only on stdout.
func thenExplainsOnStderr(ctx context.Context, reason string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, reason) {
		return fmt.Errorf("stderr %q does not explain %q", sc.stderr, reason)
	}
	return nil
}
