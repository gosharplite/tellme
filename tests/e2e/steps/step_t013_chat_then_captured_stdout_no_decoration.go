package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 — Then: the captured standard output carries no terminal decoration
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output carries no terminal decoration$`, thenCapturedStdoutNoDecoration)
	})
}

// thenCapturedStdoutNoDecoration (必查 呈現結果): narrowed to FR-007's actual
// contract (grill Q4) — when stdout is not a terminal, tellme introduces NO
// presentation control codes of its own; the redirected stream is exactly the
// answer bytes verbatim plus the single appended terminating newline. An answer
// that itself carries control bytes is passed through unchanged and must NOT be
// flagged (the previous content-blind `strings.Contains(stdout, "\x1b[")` did
// flag it, contradicting the requirement).
func thenCapturedStdoutNoDecoration(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !sc.scriptedAnswerSet {
		return fmt.Errorf("the no-decoration check requires a scripted answer (Given: a configured provider … whose endpoint answers with …)")
	}
	want := sc.scriptedAnswer + "\n"
	if sc.stdout != want {
		return fmt.Errorf("stdout = %q, want the answer bytes verbatim + one appended newline (%q): system-added decoration detected", sc.stdout, want)
	}
	return nil
}
