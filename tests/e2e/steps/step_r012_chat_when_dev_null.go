package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round-012 review RF2 — When: the operator starts tellme with the null device on
// standard input.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme with the null device on standard input$`, whenDevNullStdin)
	})
}

// whenDevNullStdin runs `tellme` (no positional prompt) with the null device
// (a character device that is NOT a terminal) on standard input — the round-012
// B1 pin (怎麼做 / 權威狀態落地: the process has run to completion; 回寫: captured
// exit code, stdout, stderr).
func whenDevNullStdin(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.devNullStdin()
	sc.args = nil
	sc.run()
	return nil
}
