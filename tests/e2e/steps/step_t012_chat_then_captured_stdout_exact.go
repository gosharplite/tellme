package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 — Then: the captured standard output is exactly "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output is exactly "([^"]*)"$`, thenCapturedStdoutExact)
	})
}

// thenCapturedStdoutExact (必查 呈現結果): the captured standard output equals the
// answer bytes verbatim, followed by exactly one newline appended by the CLI —
// i.e. `{answer}` (decoded through the escape convention) + a single "\n". This
// pins the corrected predicate (grill Q5): the answer's own trailing/embedded
// bytes are passed through; only the terminating newline is added.
func thenCapturedStdoutExact(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	want := unescapeText(answer) + "\n"
	if sc.stdout != want {
		return fmt.Errorf("stdout = %q, want %q", sc.stdout, want)
	}
	return nil
}
