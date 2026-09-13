package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 — Then: the request to the provider "{provider}" allows at most {tokens} output tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request to the provider "([^"]*)" allows at most (\d+) output tokens$`, thenRequestOutputBudget)
	})
}

// thenRequestOutputBudget (必查 呈現結果): the fake recorded exactly one request to
// {provider} whose generation config caps the output at {tokens} tokens.
func thenRequestOutputBudget(ctx context.Context, provider string, tokens int) error {
	sc := scenarioFrom(ctx)
	f := sc.fakeByProvider[provider]
	if f == nil {
		return fmt.Errorf("no fake provider registered for %q", provider)
	}
	body := f.LastBody()
	var req struct {
		GenerationConfig struct {
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return fmt.Errorf("the request to %q was not valid JSON: %q", provider, body)
	}
	if req.GenerationConfig.MaxOutputTokens != tokens {
		return fmt.Errorf("the request to %q allows %d output tokens, want %d", provider, req.GenerationConfig.MaxOutputTokens, tokens)
	}
	return nil
}
