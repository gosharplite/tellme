package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T023 — Then: the interactive prompt reports the session's token usage and turn count
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt reports the session's token usage and turn count$`, thenReportsTokensAndTurns)
	})
}

// thenReportsTokensAndTurns (必查 呈現結果): the captured output carries both a token
// usage figure and a turn count figure (the dashboard header). Presence is
// matched case-insensitively on the labels "token" and "turn".
func thenReportsTokensAndTurns(ctx context.Context) error {
	out := strings.ToLower(renderedOutput(scenarioFrom(ctx)))
	if !strings.Contains(out, "token") {
		return fmt.Errorf("the interactive prompt did not report the session's token usage; output=%q", out)
	}
	if !strings.Contains(out, "turn") {
		return fmt.Errorf("the interactive prompt did not report the session's turn count; output=%q", out)
	}
	return nil
}
