package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T023 — Then: tellme lists only the operator's messages
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme lists only the operator's messages$`, thenOperatorOnly)
	})
}

// thenOperatorOnly (必查 呈現結果): stdout carries the stored prompt and answer (as
// role-headed bodies, round 073) and NO tool-step text — the widened tool
// activity must not be surfaced by `-l` (clarify Q2 → A).
func thenOperatorOnly(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	blocks := listingBlocks(sc.stdout)
	if len(blocks) != 2*len(sc.arrangedExchanges) {
		return fmt.Errorf("listed %d messages, want %d; stdout=%q", len(blocks), 2*len(sc.arrangedExchanges), sc.stdout)
	}
	i := 0
	for _, x := range sc.arrangedExchanges {
		if blocks[i].role != "[USER]" || !strings.Contains(blocks[i].body, x.prompt) {
			return fmt.Errorf("message %d = %+v, want [USER] carrying %q", i, blocks[i], x.prompt)
		}
		i++
		if blocks[i].role != "[MODEL]" || !strings.Contains(blocks[i].body, strings.ReplaceAll(x.answer, "**", "")) {
			return fmt.Errorf("message %d = %+v, want [MODEL] carrying %q", i, blocks[i], x.answer)
		}
		i++
	}
	for _, toolText := range []string{"read_files", "arguments", "the launch code is ORANGE"} {
		if strings.Contains(stripANSI(sc.stdout), toolText) {
			return fmt.Errorf("the listing surfaced tool activity (%q): stdout=%q", toolText, sc.stdout)
		}
	}
	return nil
}
