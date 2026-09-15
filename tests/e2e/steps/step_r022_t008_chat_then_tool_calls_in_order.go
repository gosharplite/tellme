package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 [BDD-RED] — Then: the run reported the tool calls in order "{tool_a}" and "{tool_b}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the tool calls in order "([^"]*)" and "([^"]*)"$`, thenToolCallsInOrder)
	})
}

// thenToolCallsInOrder (必查 呈現結果): the captured stderr carries a
// `[HH:MM:SS] [Tool] {tool_a}` line BEFORE a `[HH:MM:SS] [Tool] {tool_b}` line
// (round-022 FR-005 call order, review TD3).
func thenToolCallsInOrder(ctx context.Context, toolA, toolB string) error {
	sc := scenarioFrom(ctx)
	logs := toolLogs(sc.stderr)
	idxA, idxB := -1, -1
	for i, tl := range logs {
		if idxA < 0 && tl.name == toolA {
			idxA = i
		}
		if idxB < 0 && tl.name == toolB {
			idxB = i
		}
	}
	if idxA < 0 {
		return fmt.Errorf("no tool-loop log line named %q; stderr=%q", toolA, sc.stderr)
	}
	if idxB < 0 {
		return fmt.Errorf("no tool-loop log line named %q; stderr=%q", toolB, sc.stderr)
	}
	if idxA >= idxB {
		return fmt.Errorf("tool %q (position %d) was not reported before %q (position %d); stderr=%q", toolA, idxA, toolB, idxB, sc.stderr)
	}
	return nil
}
