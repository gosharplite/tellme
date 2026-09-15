package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] — Then: the prompt "{prompt}" is not echoed on the diagnostic output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the prompt "([^"]*)" is not echoed on the diagnostic output$`, thenPromptNotEchoed)
	})
}

// thenPromptNotEchoed (必查 呈現結果): the captured stderr does NOT repeat the
// prompt `{prompt}` as its own line — the positional / Ctrl+D surfaces already
// show the typed text (FR-008), so only the `-i` submit echoes.
func thenPromptNotEchoed(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	p := unescapeText(prompt)
	for _, ln := range strings.Split(sc.stderr, "\n") {
		if strings.TrimSpace(strings.TrimRight(ln, "\r")) == p {
			return fmt.Errorf("the prompt was echoed as its own line; stderr=%q", sc.stderr)
		}
	}
	return nil
}
