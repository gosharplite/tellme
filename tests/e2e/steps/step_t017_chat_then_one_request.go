package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T017 — Then: tellme sends exactly one request to the provider "{provider}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme sends exactly one request to the provider "([^"]*)"$`, thenSendsExactlyOneRequest)
	})
}

// thenSendsExactlyOneRequest (必查 呈現結果 / 權威狀態): the named provider's fake
// recorded exactly one request, carrying the prompt and the resolved model; no
// other provider's fake was contacted.
func thenSendsExactlyOneRequest(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.fakeByProvider[provider]
	if f == nil {
		return fmt.Errorf("no fake provider registered for %q", provider)
	}
	if n := f.RequestCount(); n != 1 {
		return fmt.Errorf("provider %q received %d requests, want exactly 1", provider, n)
	}
	for name, other := range sc.fakeByProvider {
		if name != provider && other.RequestCount() != 0 {
			return fmt.Errorf("provider %q also received %d requests", name, other.RequestCount())
		}
	}
	body := f.LastBody()
	if sc.lastPrompt != "" && !strings.Contains(body, sc.lastPrompt) {
		return fmt.Errorf("the request to %q did not carry the prompt: %q", provider, body)
	}
	if !strings.Contains(body, "deepseek-v4-flash") {
		return fmt.Errorf("the request to %q did not carry the resolved model: %q", provider, body)
	}
	return nil
}
