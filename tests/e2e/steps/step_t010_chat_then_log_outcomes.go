package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: the tool usage shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tool usage shows the tool "([^"]*)" with (\d+) successes, (\d+) failures, and (\d+) timeouts$`, thenLogOutcomes)
	})
}

// thenLogOutcomes (必查 權威狀態): the global log holds exactly {ok} ok, {error}
// error, and {timeout} timeout records for {tool}.
func thenLogOutcomes(ctx context.Context, tool, okS, errS, toS string) error {
	sc := scenarioFrom(ctx)
	ok, _ := strconv.Atoi(okS)
	errN, _ := strconv.Atoi(errS)
	to, _ := strconv.Atoi(toS)
	recs, err := sc.toolUsageLogRecords()
	if err != nil {
		return err
	}
	gotOK, gotErr, gotTo := toolUsageCounts(recs, tool)
	if gotOK != ok || gotErr != errN || gotTo != to {
		return fmt.Errorf("tool usage for %q = ok %d, error %d, timeout %d; want ok %d, error %d, timeout %d",
			tool, gotOK, gotErr, gotTo, ok, errN, to)
	}
	return nil
}
