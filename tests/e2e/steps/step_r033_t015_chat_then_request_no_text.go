package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T015 [BDD-RED] — Then: the request carried none of the text "{text}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried none of the text "([^"]*)"$`, thenRequestCarriedNoneText)
	})
}

// thenRequestCarriedNoneText (必查 呈現結果 / 權威狀態): none of the fake's recorded
// request message contents contains {text} — skill content must NOT be injected
// into the request (the skills system is on-demand only; FR-006).
func thenRequestCarriedNoneText(ctx context.Context, text string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	for i := 0; i < f.RequestCount(); i++ {
		if body := f.BodyAt(i); strings.Contains(body, text) {
			return fmt.Errorf("request %d carried %q; body=%q", i, text, body)
		}
	}
	return nil
}
