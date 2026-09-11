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

// thenExplainsOnStderr (必查 呈現結果): stderr carries the frozen message for the
// failure class — a line pinned to start with the `tellme: {reason}` prefix
// (round 002 / FR-004); the reason must not be only on stdout.
func thenExplainsOnStderr(ctx context.Context, reason string) error {
	sc := scenarioFrom(ctx)
	want := "tellme: " + reason
	for _, line := range strings.Split(sc.stderr, "\n") {
		if strings.HasPrefix(line, want) {
			return nil
		}
	}
	return fmt.Errorf("no stderr line starts with the frozen message %q; got %q", want, sc.stderr)
}
