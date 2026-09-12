package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: the rendered width is "{width}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the rendered width is "([^"]*)"$`, givenRenderedWidth)
	})
}

// givenRenderedWidth sets the environment override TELL_ME_WRAP_WIDTH so the
// effective rendered width resolves to {width} (怎麼做 / 權威狀態落地: the effective
// rendered width resolves to {width}; 回寫: none).
func givenRenderedWidth(ctx context.Context, width string) error {
	scenarioFrom(ctx).setEnv("TELL_ME_WRAP_WIDTH", width)
	return nil
}
