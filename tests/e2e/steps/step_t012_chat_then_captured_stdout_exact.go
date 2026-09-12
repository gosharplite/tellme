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

// thenCapturedStdoutExact (必查 呈現結果): the captured standard output equals
// {answer} followed by exactly one trailing newline.
func thenCapturedStdoutExact(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	want := answer + "\n"
	if sc.stdout != want {
		return fmt.Errorf("stdout = %q, want %q", sc.stdout, want)
	}
	return nil
}
