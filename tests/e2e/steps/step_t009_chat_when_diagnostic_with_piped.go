package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T009 — When: the operator runs tellme's diagnostic with "{content}" piped in
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme's diagnostic with "([^"]*)" piped in$`, whenDiagnosticWithPiped)
	})
}

// whenDiagnosticWithPiped runs `tellme -d` with {content} on a piped standard
// input; the -d dispatch precedes any prompt handling, so the diagnostic report
// is produced (怎麼做 / 權威狀態落地 / 回寫).
func whenDiagnosticWithPiped(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin(content)
	sc.args = []string{"-d"}
	sc.run()
	return nil
}
