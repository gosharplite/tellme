package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// T012 [BDD-RED] — Then: the review shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the review shows the tool "([^"]*)" with (\d+) successes, (\d+) failures, and (\d+) timeouts$`, thenReviewOutcomes)
	})
}

// thenReviewOutcomes (必查 呈現結果): the --tool-usage stdout lists {tool} with the
// expected per-outcome counts.
func thenReviewOutcomes(ctx context.Context, tool, okS, errS, toS string) error {
	sc := scenarioFrom(ctx)
	ok, _ := strconv.Atoi(okS)
	errN, _ := strconv.Atoi(errS)
	to, _ := strconv.Atoi(toS)
	counts, found := parseReportLine(sc.stdout, tool)
	if !found {
		return fmt.Errorf("the report does not list the tool %q; stdout=%q", tool, sc.stdout)
	}
	if counts.ok != ok || counts.errN != errN || counts.timeout != to {
		return fmt.Errorf("the report for %q = ok %d, error %d, timeout %d; want ok %d, error %d, timeout %d",
			tool, counts.ok, counts.errN, counts.timeout, ok, errN, to)
	}
	return nil
}
