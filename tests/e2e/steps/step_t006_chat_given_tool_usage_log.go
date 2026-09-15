package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Given: the tool usage already records that the tool "{tool}" was used {count} times with the outcome "{outcome}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the tool usage already records that the tool "([^"]*)" was used (\d+) times with the outcome "([^"]*)"$`, givenToolUsageRecords)
	})
}

// givenToolUsageRecords (怎麼做 / 權威狀態落地 / 回寫): create $HOME/.tellme/ if
// absent, then append {count} JSON lines recording {tool} with the mapped outcome
// to tools-count.jsonl.
func givenToolUsageRecords(ctx context.Context, tool, count, outcome string) error {
	sc := scenarioFrom(ctx)
	n, err := strconv.Atoi(count)
	if err != nil {
		return fmt.Errorf("count %q is not a number", count)
	}
	mapped, ok := mapToolOutcomeWord(outcome)
	if !ok {
		return fmt.Errorf("unknown outcome %q", outcome)
	}
	for i := 0; i < n; i++ {
		if err := sc.appendToolUsage(tool, mapped); err != nil {
			return err
		}
	}
	return nil
}
