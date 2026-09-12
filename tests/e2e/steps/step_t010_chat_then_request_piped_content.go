package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T010 — Then: the request carried the piped content "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried the piped content "([^"]*)"$`, thenRequestCarriedPipedContent)
	})
}

// thenRequestCarriedPipedContent (必查 呈現結果): the single recorded request's
// `messages[0].content` equals {content} (trimmed).
func thenRequestCarriedPipedContent(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	got, err := singleRequestPrompt(sc)
	if err != nil {
		return err
	}
	if strings.TrimSpace(got) != content {
		return fmt.Errorf("the request carried prompt %q, want %q", got, content)
	}
	return nil
}

// singleRequestPrompt returns `messages[0].content` of the one request recorded
// across the scenario's registered fakes.
func singleRequestPrompt(sc *scenarioContext) (string, error) {
	var found *fakeprovider.Provider
	for _, f := range sc.fakeByProvider {
		if f.RequestCount() > 0 {
			if found != nil {
				return "", fmt.Errorf("more than one provider received a request")
			}
			found = f
		}
	}
	if found == nil {
		return "", fmt.Errorf("no provider received a request")
	}
	body := found.LastBody()
	var decoded struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		return "", fmt.Errorf("the recorded request was not valid JSON: %q", body)
	}
	if len(decoded.Messages) == 0 {
		return "", fmt.Errorf("the recorded request carried no message: %q", body)
	}
	return decoded.Messages[0].Content, nil
}
