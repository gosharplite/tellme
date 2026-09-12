package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T017 — Then: tellme sends no request to any provider
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme sends no request to any provider$`, thenNoProviderRequest)
	})
}

// thenNoProviderRequest (必查 呈現結果 / 權威狀態): every configured fake provider
// recorded zero requests.
func thenNoProviderRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for name, f := range sc.fakeByProvider {
		if n := f.RequestCount(); n != 0 {
			return fmt.Errorf("provider %q received %d requests, want none", name, n)
		}
	}
	for _, f := range sc.fakes {
		if n := f.RequestCount(); n != 0 {
			return fmt.Errorf("a provider received %d requests, want none", n)
		}
	}
	return nil
}
