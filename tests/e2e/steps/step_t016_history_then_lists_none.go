package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T016 — Then: tellme lists no messages
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme lists no messages$`, thenListsNoMessages)
	})
}

// thenListsNoMessages (必查 呈現結果): stdout carries no messages.
func thenListsNoMessages(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.TrimSpace(sc.stdout) != "" {
		return fmt.Errorf("stdout = %q, want no messages", sc.stdout)
	}
	return nil
}
