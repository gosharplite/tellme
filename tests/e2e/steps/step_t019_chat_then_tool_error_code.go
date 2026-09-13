package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T019 — Then: tellme exits with the tool error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the tool error code$`, thenToolErrorCode)
	})
}

// toolErrorCode is the pinned round-008 tool error code
// (specs/truth/features/cli/chat/dsl.md): distinct from success (0) and every
// other class (usage 2, configuration 3, environment 4, diagnostic 5, provider 6).
const toolErrorCode = 7

// thenToolErrorCode (必查 呈現結果): the exit code equals the pinned tool error
// code 7.
func thenToolErrorCode(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != toolErrorCode {
		return fmt.Errorf("exit code = %d, want %d (the tool error code)", sc.exitCode, toolErrorCode)
	}
	return nil
}
