package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T004 — Given: the session history already holds a tool-using exchange with no provider token
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session history already holds a tool-using exchange with no provider token$`, givenUnsignedToolUsingExchange)
	})
}

// givenUnsignedToolUsingExchange: append a tool-using turn whose step carries no
// provider token (the provider-neutral shape), reusing the shared writer.
func givenUnsignedToolUsingExchange(ctx context.Context) error {
	return writeToolUsingExchange(ctx, "")
}
