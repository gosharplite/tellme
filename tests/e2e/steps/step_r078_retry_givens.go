package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round 078 (ADR 0050) — the transport-retry Givens. Each scripts a fake whose
// endpoint fails in a specific way so the retry predicate's two sides are
// exercised: a transport DROP (retryable) and an outright rejection (not).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint drops the connection once and then answers$`, givenProviderDropOnceThenAnswer)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint drops the connection twice and then answers$`, givenProviderDropTwiceThenAnswer)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint always drops the connection$`, givenProviderAlwaysDrops)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint rejects the request outright$`, givenProviderRejectsOutright)
	})
}

// givenProviderDropOnceThenAnswer scripts the fake to drop the FIRST request's
// connection (a transport failure) and answer normally afterwards — so the turn
// succeeds on the second attempt, and exactly two requests were sent.
func givenProviderDropOnceThenAnswer(ctx context.Context, provider string) error {
	return arrangeFlakyProvider(ctx, provider, 1, 0)
}

// givenProviderDropTwiceThenAnswer drops the first TWO requests, so the turn
// succeeds on the third attempt (exactly three requests).
func givenProviderDropTwiceThenAnswer(ctx context.Context, provider string) error {
	return arrangeFlakyProvider(ctx, provider, 2, 0)
}

// givenProviderAlwaysDrops drops every request (a large count), so the bounded
// retry exhausts and the run fails with the frozen phrase + exit 6.
func givenProviderAlwaysDrops(ctx context.Context, provider string) error {
	return arrangeFlakyProvider(ctx, provider, 1000, 0)
}

// givenProviderRejectsOutright answers every request with a non-retryable 400
// (a client error), so the run fails at once — exactly one request.
func givenProviderRejectsOutright(ctx context.Context, provider string) error {
	return arrangeFlakyProvider(ctx, provider, 0, 400)
}

// arrangeFlakyProvider builds the fake with a drop count and/or an error status,
// registers it, and writes the default configuration selecting {provider}.
func arrangeFlakyProvider(ctx context.Context, provider string, dropFirst, errorStatus int) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Answer("recovered")
	if dropFirst > 0 {
		f.DropFirst(dropFirst)
	}
	if errorStatus != 0 {
		f.ErrorStatus(errorStatus)
	}
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}
