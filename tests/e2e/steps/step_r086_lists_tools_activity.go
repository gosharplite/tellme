package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round 007 T023 / round 086 (ADR 0057) — Then: tellme lists the tools' activity
// but not their contents.
//
// The round-007 step asserted the listing surfaces the operator's messages and
// NOT the widened tool activity. Round 086 amended that reading: the listing now
// surfaces each turn's tool COUNT (`[TOOLS] - M (N calls)`), while the round-073
// clarify Q2 -> A rule holds for the tool CONTENT (name, arguments, result). The
// phrase is renamed so the DSL row and this step state the new contract.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme lists the tools' activity but not their contents$`, thenListsToolActivityNotContents)
	})
}

// thenListsToolActivityNotContents (必查 呈現結果): every listed turn is the [USER]
// prompt + [MODEL] answer pair PLUS the turn's `[TOOLS] - M (N calls)` count line
// (round 086), while the tool CONTENT — the tool name, its arguments, and its
// result — is NOT surfaced.
func thenListsToolActivityNotContents(ctx context.Context) error {
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
	// Round 086: the count line IS surfaced …
	if !strings.Contains(stripANSI(sc.stdout), "[TOOLS] - ") {
		return fmt.Errorf("the listing must surface each turn's tool activity; stdout=%q", sc.stdout)
	}
	// … but the tool CONTENT (name, arguments, result) must NOT be.
	for _, toolText := range []string{"read_files", "arguments", "the launch code is ORANGE"} {
		if strings.Contains(stripANSI(sc.stdout), toolText) {
			return fmt.Errorf("the listing surfaced tool content (%q): stdout=%q", toolText, sc.stdout)
		}
	}
	return nil
}
