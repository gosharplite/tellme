package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 — Then: the request to the provider "{provider}" carried the service-account access token
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request to the provider "([^"]*)" carried the service-account access token$`, thenServiceAccountToken)
	})
}

// thenServiceAccountToken (必查 呈現結果): the fake recorded exactly one request to
// {provider} carrying an `Authorization: Bearer <token>` header whose token was
// minted at the fake token endpoint from the configured service-account
// credential (the fake's canned token). It must NOT carry the raw key path.
func thenServiceAccountToken(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.fakeByProvider[provider]
	if f == nil {
		return fmt.Errorf("no fake provider registered for %q", provider)
	}
	if n := f.RequestCount(); n != 1 {
		return fmt.Errorf("provider %q received %d requests, want exactly 1", provider, n)
	}
	got := f.LastAuthHeader()
	if got != "Bearer fake-token" {
		return fmt.Errorf("the request to %q carried Authorization %q, want \"Bearer fake-token\"", provider, got)
	}
	return nil
}
