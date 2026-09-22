package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// Round 078 (ADR 0050) — the transport-retry Thens.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn finishes with the provider's answer$`, thenTurnFinishesWithAnswer)
		ctx.Then(`^tellme sent the request exactly once$`, thenSentRequestOnce)
		ctx.Then(`^tellme sent the request exactly twice$`, thenSentRequestTwice)
		ctx.Then(`^tellme sent the request exactly three times$`, thenSentRequestThreeTimes)
	})
}

// thenTurnFinishesWithAnswer (必查 呈現結果): the run succeeded and the answer the
// fake scripted ("recovered") reached stdout — i.e. a retry absorbed the blip.
func thenTurnFinishesWithAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.Success {
		return fmt.Errorf("exit code = %d, want %d (success); stderr=%q", sc.exitCode, cli.Success, sc.stderr)
	}
	if !strings.Contains(sc.stdout, "recovered") {
		return fmt.Errorf("stdout does not carry the provider's answer; stdout=%q", sc.stdout)
	}
	return nil
}

// thenSentRequestOnce (必查 呈現結果): exactly one attempt — a non-retryable
// failure must not be retried.
func thenSentRequestOnce(ctx context.Context) error { return assertRequestCount(ctx, 1) }

// thenSentRequestTwice (必查 呈現結果): exactly two attempts — one retry absorbed
// a single drop.
func thenSentRequestTwice(ctx context.Context) error { return assertRequestCount(ctx, 2) }

// thenSentRequestThreeTimes (必查 呈現結果): exactly three attempts — the bound is
// two retries (three attempts), no more.
func thenSentRequestThreeTimes(ctx context.Context) error { return assertRequestCount(ctx, 3) }

// assertRequestCount asserts the single fake provider received exactly want
// requests (the retry count/order witness; the retry delay is pinned to 0 ms by
// the scenario default, so this measures the COUNT, never wall-clock).
func assertRequestCount(ctx context.Context, want int) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider was arranged")
	}
	if got := f.RequestCount(); got != want {
		return fmt.Errorf("provider received %d requests, want %d", got, want)
	}
	return nil
}
