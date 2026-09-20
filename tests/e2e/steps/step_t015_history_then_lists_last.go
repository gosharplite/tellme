package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T015 — Then: tellme lists the last {count} messages
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme lists the last (\d+) messages$`, thenListsLastMessages)
	})
}

// thenListsLastMessages (必查 呈現結果): stdout lists the last {count} messages of
// the arranged history, in order, each as a role header line ([USER] / [MODEL])
// followed by its body (round 073; ADR 0045), with one blank line between
// consecutive messages.
func thenListsLastMessages(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	wantRole := make([]string, 0, 2*len(sc.arrangedExchanges))
	wantBody := make([]string, 0, 2*len(sc.arrangedExchanges))
	for _, x := range sc.arrangedExchanges {
		wantRole = append(wantRole, "[USER]", "[MODEL]")
		wantBody = append(wantBody, x.prompt, strings.ReplaceAll(x.answer, "**", ""))
	}
	if len(wantRole) > count {
		wantRole = wantRole[len(wantRole)-count:]
		wantBody = wantBody[len(wantBody)-count:]
	}
	blocks := listingBlocks(sc.stdout)
	if len(blocks) != len(wantRole) {
		return fmt.Errorf("listed %d messages, want %d; stdout=%q", len(blocks), len(wantRole), sc.stdout)
	}
	for i := range wantRole {
		if blocks[i].role != wantRole[i] {
			return fmt.Errorf("message %d header = %q, want %q; stdout=%q", i, blocks[i].role, wantRole[i], sc.stdout)
		}
		if !strings.Contains(blocks[i].body, wantBody[i]) {
			return fmt.Errorf("message %d body = %q, want it to carry %q; stdout=%q", i, blocks[i].body, wantBody[i], sc.stdout)
		}
	}
	if err := thenMessagesSeparatedByBlankLine(ctx); err != nil {
		return err
	}
	return nil
}
