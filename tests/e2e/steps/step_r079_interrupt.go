package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 079 (ADR 0051; closes #159) — the operator-interruption persistence
// steps. The live interruption is produced by a scripted `execute_command` whose
// command signals tellme itself (`kill -INT $PPID`) — a REAL SIGINT to the turn
// process, deterministic and hermetic (no pty, no stall harness). The tool step
// the command completed is then persisted by the product's interrupted-turn path.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs the command "([^"]*)" and then answers with "([^"]*)"$`, givenProviderRunsCommand)
		ctx.Given(`^the session history already holds an interrupted exchange with (\d+) tool step$`, givenHistoryHoldsInterruptedExchange)
		ctx.Then(`^tellme stored an interrupted turn in the session history with (\d+) tool step$`, thenStoredInterruptedTurn)
		ctx.Then(`^the interrupted turn was recorded with (\d+) completed provider call$`, thenInterruptedTurnRecordedCalls)
		ctx.Then(`^the session history's last turn was closed with the operator-interruption answer$`, thenLastTurnClosedWithInterruption)
		ctx.Then(`^the request closes the earlier turn with the assistant answer "([^"]*)"$`, thenRequestClosesEarlierTurn)
		ctx.Then(`^tellme reports on stderr that the turn was interrupted by the operator$`, thenReportsInterruption)
		ctx.Then(`^tellme does not explain on stderr that "([^"]*)"$`, thenDoesNotExplainOnStderr)
	})
}

// givenProviderRunsCommand scripts the fake to return ONE `execute_command` tool
// call carrying the given command (and a reason), then a final answer — so the
// command can signal tellme mid-turn.
func givenProviderRunsCommand(ctx context.Context, provider, command, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	// Safety bound: if the signal somehow did not land, the loop stops after a few
	// rounds instead of running to the default 1000-round bound (a fast failure,
	// never a suite hang). `the tool-loop limit is "5"` in the feature sets it too.
	sc.setEnv("MAX_TOOL_LOOP", "5")
	args, err := json.Marshal(map[string]string{"command": command, "reason": "e2e interruption probe"})
	if err != nil {
		return err
	}
	f.Script(fakeprovider.Reply{ToolName: "execute_command", Arguments: string(args)}, fakeprovider.Reply{Answer: answer})
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// givenHistoryHoldsInterruptedExchange arranges a persisted partial turn (the
// exact shape the product writes for an interrupted turn) so a resumed run's
// replay is assertable without a second live interruption.
func givenHistoryHoldsInterruptedExchange(ctx context.Context, steps int) error {
	sc := scenarioFrom(ctx)
	stepList := make([]history.Step, 0, steps)
	for i := 0; i < steps; i++ {
		stepList = append(stepList, history.Step{
			Tool:      "execute_command",
			Arguments: `{"command":"kill -INT $PPID"}`,
			Result:    "Exit Code: -1\n",
		})
	}
	entry := history.Entry{
		Prompt: "Explore the repository.",
		Answer: history.InterruptedTurnAnswer,
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
	// Append (O_APPEND) — the row's verb (N-079-2; equivalent on a fresh home).
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

// thenStoredInterruptedTurn (必查 權威狀態): the interrupted turn was persisted as
// EXACTLY ONE history entry carrying the given number of completed tool steps.
func thenStoredInterruptedTurn(ctx context.Context, want int) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) != 1 {
		return fmt.Errorf("history holds %d entries, want exactly 1 (the interrupted turn)", len(entries))
	}
	if got := len(entries[0].Steps); got != want {
		return fmt.Errorf("the interrupted turn has %d tool step(s), want %d", got, want)
	}
	return nil
}

// thenInterruptedTurnRecordedCalls (必查 權威狀態; fold F-079-2): the persisted
// interrupted turn carries the given `calls` count — the completed inference
// rounds. The round's SC-001/D9 claim that the E2E asserts `calls` needs this
// carrier (without it, `Calls: 0` leaves the E2E green).
func thenInterruptedTurnRecordedCalls(ctx context.Context, want int) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("history is empty, want the interrupted turn")
	}
	if got := entries[len(entries)-1].Calls; got != want {
		return fmt.Errorf("the interrupted turn's persisted calls = %d, want %d (the completed inference rounds)", got, want)
	}
	return nil
}

// thenLastTurnClosedWithInterruption (必查 權威狀態): the persisted turn's answer is
// the synthetic `history.InterruptedTurnAnswer` — the close that makes the stored
// history replay as a valid `… assistant` sequence.
func thenLastTurnClosedWithInterruption(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("history is empty, want the interrupted turn")
	}
	got := entries[len(entries)-1].Answer
	if got != history.InterruptedTurnAnswer {
		return fmt.Errorf("last turn's answer = %q, want the synthetic %q", got, history.InterruptedTurnAnswer)
	}
	return nil
}

// thenReportsInterruption (必查 呈現結果): the informational interruption line
// reached the diagnostic stream.
func thenReportsInterruption(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, "interrupted by operator") {
		return fmt.Errorf("stderr does not report the operator interruption: %q", sc.stderr)
	}
	return nil
}

// thenDoesNotExplainOnStderr (必查 呈現結果; TD-079-1): the given class phrase does
// NOT appear on stderr — the interrupted-but-saved path is not a failure and must
// carry no `tellme: <phrase>` line.
func thenDoesNotExplainOnStderr(ctx context.Context, reason string) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(sc.stderr, "tellme: "+reason) {
		return fmt.Errorf("stderr carries the failure phrase %q though the interrupted turn was saved: %q", reason, sc.stderr)
	}
	return nil
}

// thenRequestClosesEarlierTurn (必查 呈現結果 / 權威狀態): the resumed request's
// replayed conversation closes the earlier (kept) turn with an ASSISTANT message
// carrying the synthetic answer, BEFORE the current prompt — i.e. the stored
// partial turn replays as a valid `user … assistant` sequence (I-1: role
// alternation holds on both families).
func thenRequestClosesEarlierTurn(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider was arranged")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode the recorded request: %w", err)
	}
	closeIdx := -1
	for i, m := range msgs {
		if m.Role == "assistant" && m.Content == answer {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 {
		return fmt.Errorf("no assistant message carries the earlier turn's closing answer %q; messages=%+v", answer, msgs)
	}
	for i := closeIdx + 1; i < len(msgs); i++ {
		if msgs[i].Role == "user" {
			return nil
		}
	}
	return fmt.Errorf("the closing assistant answer %q is not followed by the current prompt; messages=%+v", answer, msgs)
}
