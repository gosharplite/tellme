package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 [BDD-RED] — Then: the run streamed no command output block for the writing command
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run streamed no command output block for the writing command$`, thenNoOutputBlock)
	})
}

// thenNoOutputBlock (必查 呈現結果): a command that carries `output_file` binds
// its streams to a file, so stderr carries NO `[Tool Output]` block (round 034
// FR-010).
func thenNoOutputBlock(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if hasToolOutputBlock(sc.stderr) {
		return fmt.Errorf("standard error carried a `[Tool Output]` block for an output_file command; stderr=%q", sc.stderr)
	}
	return nil
}
