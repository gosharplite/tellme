package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 — Then: the request carried the instruction "{instruction}" followed by the piped content "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried the instruction "([^"]*)" followed by the piped content "([^"]*)"$`, thenRequestCarriedInstructionThenContent)
	})
}

// thenRequestCarriedInstructionThenContent (必查 呈現結果): the recorded request's
// prompt equals {instruction} + a newline + {content}.
func thenRequestCarriedInstructionThenContent(ctx context.Context, instruction, content string) error {
	sc := scenarioFrom(ctx)
	got, err := singleRequestPrompt(sc)
	if err != nil {
		return err
	}
	want := instruction + "\n" + content
	if got != want {
		return fmt.Errorf("the request carried prompt %q, want %q", got, want)
	}
	return nil
}
