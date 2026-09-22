package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
// requests (the retry count witness; the retry delay is pinned to 0 ms by
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

// recordedTurn is the persisted history line's shape the accounting Then reads
// (the round-078 fold F-2 witness for the "one AI-endpoint call" invariant).
type recordedTurn struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
	Calls  int    `json:"calls"`
}

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme announces on stderr that it is retrying the provider request$`, thenAnnouncesRetry)
		ctx.Then(`^the turn was recorded as a single provider call$`, thenRecordedSingleCall)
		ctx.Then(`^the retry re-sent the same request$`, thenRetryResentSameRequest)
		ctx.Then(`^the run wrote nothing to standard output$`, thenWroteNothingToStdout)
	})
}

// thenAnnouncesRetry (必查 呈現結果; round-078 fold F-4): the retry line reached
// the diagnostic stream — the watched server drop is announced before the retry.
func thenAnnouncesRetry(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, "retrying the provider request in") {
		return fmt.Errorf("stderr does not announce the retry: %q", sc.stderr)
	}
	return nil
}

// thenRecordedSingleCall (必查 權威狀態; round-078 fold F-2): the retried turn was
// recorded as ONE provider call — the persisted history line carries `calls == 1`
// and exactly one usage record was written. This witnesses the round-078
// accounting invariant (I-4/D7) the count probes do not cover.
func thenRecordedSingleCall(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read history: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		return fmt.Errorf("history has %d lines, want 1: %q", len(lines), string(data))
	}
	var turn recordedTurn
	if err := json.Unmarshal([]byte(lines[0]), &turn); err != nil {
		return fmt.Errorf("decode history line: %w", err)
	}
	if turn.Calls != 1 {
		return fmt.Errorf("persisted calls = %d, want 1 (a retried call is one AI-endpoint call)", turn.Calls)
	}
	usagePath := filepath.Join(sc.historyDir(), "tokens.log")
	usage, err := os.ReadFile(usagePath)
	if err != nil {
		return fmt.Errorf("read usage log: %w", err)
	}
	if n := len(strings.Split(strings.TrimRight(string(usage), "\n"), "\n")); n != 1 {
		return fmt.Errorf("usage log has %d records, want 1: %q", n, string(usage))
	}
	return nil
}

// thenRetryResentSameRequest (必查 權威狀態; round-078 nit N-3): the retry re-sent
// the SAME request — attempt 0's body equals attempt 1's body (byte-exact).
func thenRetryResentSameRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider was arranged")
	}
	if f.RequestCount() < 2 {
		return fmt.Errorf("only %d request(s) recorded; need >= 2 to compare", f.RequestCount())
	}
	if a, b := f.BodyAt(0), f.BodyAt(1); a == "" || a != b {
		return fmt.Errorf("retry body differs from the original request:\nfirst=%q\nretry=%q", a, b)
	}
	return nil
}

// thenWroteNothingToStdout (必查 呈現結果; round-078 nit N-2): the failed turn's
// stdout is empty (the frozen phrase goes to stderr only).
func thenWroteNothingToStdout(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.stdout != "" {
		return fmt.Errorf("stdout = %q, want empty", sc.stdout)
	}
	return nil
}
