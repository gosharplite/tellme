package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 080 (ADR 0052; closes #161) — persist a FAILED turn's completed tool
// steps. The failure is produced hermetically by the in-process fake provider:
// request 1 serves a `read_files` tool call (so a step completes), then request
// 2+ drops the connection forever (retry exhaustion) or answers 400 (non-
// retryable). The retry delay seam is collapsed to 0 ms.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and then always drops the connection$`, givenProviderReadsThenDrops)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and then rejects the request outright$`, givenProviderReadsThenRejects)
		ctx.Given(`^the session history already holds a failed exchange with (\d+) tool step$`, givenHistoryHoldsFailedExchange)
		ctx.Then(`^tellme stored the failed turn in the session history with (\d+) tool step$`, thenStoredFailedTurn)
		ctx.Then(`^the failed turn was recorded with (\d+) completed provider call$`, thenFailedTurnRecordedCalls)
		ctx.Then(`^the failed turn was closed with the answer "([^"]*)"$`, thenFailedTurnClosedWithAnswer)
		ctx.Then(`^tellme reports on stderr that it kept the completed tool step$`, thenReportsKeptStep)
		ctx.Then(`^the session history is empty$`, thenSessionHistoryEmpty)
	})
}

// readFilesCall is the `read_files` tool-call arguments JSON for one path.
func readFilesCall(path string) string {
	args, _ := json.Marshal(map[string]any{
		"filepaths": []string{path},
		"reason":    "inspect the notes",
	})
	return string(args)
}

// givenProviderReadsThenDrops scripts the fake to serve a `read_files` tool call
// on request 1, then DROP the connection on every later request (a transport
// failure the bounded retry cannot absorb) — so a step completes and the turn
// then FAILS.
func givenProviderReadsThenDrops(ctx context.Context, provider, path string) error {
	sc := scenarioFrom(ctx)
	// Collapse the round-078 retry delays so the exhaustion is instant (count,
	// never wall-clock).
	sc.setEnv("TELL_ME_FORCE_RETRY_DELAY_MS", "0")
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readFilesCall(path)},
		fakeprovider.Reply{Drop: true}, // the last reply repeats → permanent drop
	)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// givenProviderReadsThenRejects scripts the fake to serve a `read_files` tool call
// on request 1, then answer 400 on every later request (a NON-retryable failure)
// — so a step completes and the turn then FAILS at once.
func givenProviderReadsThenRejects(ctx context.Context, provider, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readFilesCall(path)},
		fakeprovider.Reply{ErrorStatus: 400},
	)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// givenHistoryHoldsFailedExchange arranges a persisted failed turn (the exact
// shape the product writes) so a resumed run's replay is assertable without a
// second live failure.
func givenHistoryHoldsFailedExchange(ctx context.Context, steps int) error {
	sc := scenarioFrom(ctx)
	stepList := make([]history.Step, 0, steps)
	for i := 0; i < steps; i++ {
		stepList = append(stepList, history.Step{
			Tool:      "read_files",
			Arguments: `{"filepaths":["notes.txt"],"reason":"inspect the notes"}`,
			Result:    "the launch code is ORANGE\n",
		})
	}
	entry := history.Entry{
		Prompt: "Read the notes, then keep going.",
		Answer: history.ProviderFailedTurnAnswer,
		Calls:  1,
		Steps:  stepList,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(sc.historyFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// thenStoredFailedTurn (必查 權威狀態): the failed turn was persisted as EXACTLY
// ONE history entry carrying the given number of completed tool steps.
func thenStoredFailedTurn(ctx context.Context, want int) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) != 1 {
		return fmt.Errorf("history holds %d entries, want exactly 1 (the failed turn)", len(entries))
	}
	if got := len(entries[0].Steps); got != want {
		return fmt.Errorf("the failed turn has %d tool step(s), want %d", got, want)
	}
	return nil
}

// thenFailedTurnRecordedCalls (必查 權威狀態; fold F-080-2): the persisted failed
// turn carries the given `calls` count — the completed inference rounds. Without
// this carrier a `Calls: 0` mutant leaves the E2E green.
func thenFailedTurnRecordedCalls(ctx context.Context, want int) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("history is empty, want the failed turn")
	}
	if got := entries[len(entries)-1].Calls; got != want {
		return fmt.Errorf("the failed turn's persisted calls = %d, want %d (the completed inference rounds)", got, want)
	}
	return nil
}

// thenFailedTurnClosedWithAnswer (必查 權威狀態; fold TD-080-1): the
// persisted failed turn's answer equals the EXPECTED class-specific failure
// answer — so a class swap (e.g. the tool answer on a provider failure) reddens.
func thenFailedTurnClosedWithAnswer(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("history is empty, want the failed turn")
	}
	if got := entries[len(entries)-1].Answer; got != answer {
		return fmt.Errorf("failed turn's answer = %q, want %q", got, answer)
	}
	return nil
}

// thenReportsKeptStep (必查 呈現結果): the informational keep line reached stderr.
func thenReportsKeptStep(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, "kept ") || !strings.Contains(sc.stderr, "completed tool step(s) in the session history") {
		return fmt.Errorf("stderr does not report the kept steps: %q", sc.stderr)
	}
	return nil
}

// thenSessionHistoryEmpty (必查 權威狀態): nothing was persisted (a zero-step
// failure, or a clean abort).
func thenSessionHistoryEmpty(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // a missing file is an empty history
		}
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("history holds %d entries, want none", len(entries))
	}
	return nil
}
