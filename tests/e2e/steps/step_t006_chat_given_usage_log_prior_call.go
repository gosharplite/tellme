package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T006 — Given: the session's usage log already records a prior call
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session's usage log already records a prior call$`, givenUsageLogPriorCall)
	})
}

// givenUsageLogPriorCall (怎麼做 / 權威狀態落地 / 回寫): append one prior call record
// to the per-mode usage log (`$TELL_ME_HOME/output/<mode>/tokens.log`).
func givenUsageLogPriorCall(ctx context.Context) error {
	return writePriorUsage(scenarioFrom(ctx))
}
