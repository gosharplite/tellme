package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// T006 — Then: the turn is headed "Turn {number}" for the active mode
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn is headed "Turn ([0-9]+)" for the active mode$`, thenTurnHeaded)
	})
}

// thenTurnHeaded (必查 呈現結果): the diagnostic stream carries a header line
// `╭─⠿ Turn {number} - <mode>` (the number equals the session's AI-endpoint-call
// count + 1 — see specs/truth/features/cli/chat/dsl.md, round 027; the mode is
// non-empty), and it is NOT written to standard output.
func thenTurnHeaded(ctx context.Context, number string) error {
	want, err := strconv.Atoi(number)
	if err != nil {
		return fmt.Errorf("invalid turn number %q: %w", number, err)
	}
	sc := scenarioFrom(ctx)
	n, mode, ok := turnHeaderParts(sc.stderr)
	if !ok {
		return fmt.Errorf("the diagnostic stream carried no `Turn …` header; stderr=%q", sc.stderr)
	}
	if n != want {
		return fmt.Errorf("the header carried Turn %d, want %d; stderr=%q", n, want, sc.stderr)
	}
	if mode == "" {
		return fmt.Errorf("the header carried no mode; stderr=%q", sc.stderr)
	}
	if hasTurnHeader(sc.stdout) {
		return fmt.Errorf("the turn header leaked to standard output; stdout=%q", sc.stdout)
	}
	return nil
}
