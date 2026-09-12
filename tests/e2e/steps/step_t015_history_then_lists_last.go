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
// the arranged history, in order, each line as `role: content` (RF-2).
func thenListsLastMessages(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	var want []string
	for _, x := range sc.arrangedExchanges {
		want = append(want, "user: "+x.prompt, "assistant: "+x.answer)
	}
	if len(want) > count {
		want = want[len(want)-count:]
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
	return nil
}
