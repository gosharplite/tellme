package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T026 — Then: tellme sends no request to the provider "{provider}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme sends no request to the provider "([^"]*)"$`, thenSendsNoRequest)
	})
}

// thenSendsNoRequest (必查 呈現結果): the named provider's fake recorded ZERO
// requests — an aborted or non-interactive-empty submission must not reach any
// provider.
func thenSendsNoRequest(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.fakeByProvider[provider]
	if f == nil {
		return fmt.Errorf("no fake provider registered for %q", provider)
	}
	if n := f.RequestCount(); n != 0 {
		return fmt.Errorf("provider %q received %d requests, want 0", provider, n)
	}
	return nil
}
