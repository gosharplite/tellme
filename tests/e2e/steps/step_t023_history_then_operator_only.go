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
// `role: content` lines) and NO tool-step text — the widened tool activity must
// not be surfaced by `-l`.
func thenOperatorOnly(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var want []string
	for _, x := range sc.arrangedExchanges {
		want = append(want, "user: "+x.prompt, "assistant: "+x.answer)
	}
	out := strings.TrimRight(sc.stdout, "\n")
	var got []string
	if out != "" {
		got = strings.Split(out, "\n")
	}
	if len(got) != len(want) {
		return fmt.Errorf("listed %d lines, want %d; stdout=%q", len(got), len(want), sc.stdout)
	}
	for i := range want {
		if got[i] != want[i] {
			return fmt.Errorf("listed line %d = %q, want %q", i, got[i], want[i])
		}
	}
	for _, toolText := range []string{"read_files", "arguments", "the launch code is ORANGE"} {
		if strings.Contains(sc.stdout, toolText) {
			return fmt.Errorf("the listing surfaced tool activity (%q): stdout=%q", toolText, sc.stdout)
		}
	}
	return nil
}
