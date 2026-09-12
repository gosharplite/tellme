package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T016 — When: the operator runs tellme's diagnostic with the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme's diagnostic with the prompt "([^"]*)"$`, whenDiagnosticWithPrompt)
	})
}

// whenDiagnosticWithPrompt runs `tellme -d "{prompt}"` (怎麼做 / 權威狀態落地:
// -d dispatch precedes the prompt turn, so the diagnostic report is produced;
// 回寫: capture exit code, stdout, stderr, and the fake's recorded requests).
func whenDiagnosticWithPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	sc.args = []string{"-d", prompt}
	sc.run()
	return nil
}
