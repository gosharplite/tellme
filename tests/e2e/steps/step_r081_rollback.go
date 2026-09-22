package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Round 081 (ADR 0053) — roll back the last N turns of the session history.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		// Givens
		ctx.Given(`^the session history then holds a plain exchange$`, givenHistoryThenPlainExchange)
		// Whens
		ctx.When(`^the operator rolls back the last turn$`, whenRollbackLastTurn)
		ctx.When(`^the operator rolls back the last (\d+) turns$`, whenRollbackLastNTurns)
		ctx.When(`^the operator rolls back a count of (\d+) turns$`, whenRollbackNTurns)
		ctx.When(`^the operator rolls back the last turn and asks "([^"]*)"$`, whenRollbackThenAsk)
		ctx.When(`^the operator rolls back the last turn and starts a fresh session$`, whenRollbackAndFresh)
		// Thens
		ctx.Then(`^the active history holds exactly (\d+) exchanges?$`, thenActiveHistoryHoldsExactly)
		ctx.Then(`^tellme reports that 1 turn was rolled back$`, thenReportsOneTurnRolledBack)
		ctx.Then(`^tellme reports that (\d+) turns were rolled back$`, thenReportsNTurnsRolledBack)
		ctx.Then(`^the remaining exchange carries its recorded tool step$`, thenRemainingExchangeCarriesStep)
		ctx.Then(`^the session archive holds no exchanges$`, thenArchiveHoldsNoExchanges)
		ctx.Then(`^the last persisted exchange asks "([^"]*)" and answers "([^"]*)"$`, thenLastExchangeAsksAndAnswers)
		ctx.Then(`^the provider was asked against the trimmed history$`, thenProviderAskedAgainstTrimmedHistory)
	})
}

// givenHistoryThenPlainExchange appends one plain (step-free) exchange so it
// becomes the LAST exchange (used with the tool-using Given).
func givenHistoryThenPlainExchange(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(historyEntry{Prompt: "plain Q", Answer: "plain A"})
	if err != nil {
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
	if err := f.Close(); err != nil {
		return err
	}
	sc.recordExchange("plain Q", "plain A")
	return nil
}

func whenRollbackLastTurn(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-b"}
	sc.run()
	return nil
}

func whenRollbackLastNTurns(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-b", strconv.Itoa(count)}
	sc.run()
	return nil
}

func whenRollbackNTurns(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-b", strconv.Itoa(count)}
	sc.run()
	return nil
}

func whenRollbackThenAsk(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-b", prompt}
	sc.run()
	return nil
}

func whenRollbackAndFresh(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-b", "--new"}
	sc.run()
	return nil
}

func thenActiveHistoryHoldsExactly(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read session history: %w", err)
	}
	if len(entries) != count {
		return fmt.Errorf("the active history holds %d exchanges, want %d; %+v", len(entries), count, entries)
	}
	return nil
}

func thenReportsOneTurnRolledBack(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "Rolled back 1 turn") {
		return fmt.Errorf("stdout = %q, want a rollback confirmation naming 1 turn", sc.stdout)
	}
	if strings.Contains(sc.stdout, "tellme: ") {
		return fmt.Errorf("stdout = %q, a rollback confirmation must carry no `tellme: ` class phrase", sc.stdout)
	}
	if strings.Contains(sc.stderr, "tellme: ") {
		return fmt.Errorf("stderr = %q, a standalone rollback must not emit a class phrase", sc.stderr)
	}
	return nil
}

func thenReportsNTurnsRolledBack(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	want := fmt.Sprintf("Rolled back %d turn", count)
	if !strings.Contains(sc.stdout, want) {
		return fmt.Errorf("stdout = %q, want a rollback confirmation naming %d turns", sc.stdout, count)
	}
	if strings.Contains(sc.stdout, "tellme: ") {
		return fmt.Errorf("stdout = %q, a rollback confirmation must carry no `tellme: ` class phrase", sc.stdout)
	}
	return nil
}

// thenRemainingExchangeCarriesStep asserts the surviving history.jsonl line still
// carries its recorded steps array (a rollback of a LATER turn must not mutate an
// earlier turn's tool activity).
func thenRemainingExchangeCarriesStep(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read history: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return fmt.Errorf("the active history is empty, want 1 tool-using exchange")
	}
	var e widenedEntry
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		return fmt.Errorf("decode history line: %w", err)
	}
	if len(e.Steps) != 1 || e.Steps[0].Tool == "" {
		return fmt.Errorf("surviving exchange steps = %+v, want its recorded tool step", e.Steps)
	}
	return nil
}

func thenArchiveHoldsNoExchanges(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(sc.historyArchivePath())
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("the archive holds %d exchanges, want none (rollback must not archive)", len(entries))
	}
	return nil
}

func thenLastExchangeAsksAndAnswers(ctx context.Context, prompt, answer string) error {
	sc := scenarioFrom(ctx)
	entries, err := readSessionEntries(sc)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("the active history is empty, want a final exchange %q/%q", prompt, answer)
	}
	last := entries[len(entries)-1]
	if last.Prompt != prompt || last.Answer != answer {
		return fmt.Errorf("last exchange = %q/%q, want %q/%q", last.Prompt, last.Answer, prompt, answer)
	}
	return nil
}

// thenProviderAskedAgainstTrimmedHistory (F-081-4) asserts the prompt-bearing
// rollback built the request against the TRIMMED conversation: the surviving
// exchange is present and the rolled-back one is absent from the wire body.
func thenProviderAskedAgainstTrimmedHistory(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if len(sc.fakes) == 0 {
		return fmt.Errorf("no fake provider was configured")
	}
	body := sc.fakes[0].LastBody()
	if !strings.Contains(body, "first Q") {
		return fmt.Errorf("the request body does not carry the surviving exchange; body=%q", body)
	}
	if strings.Contains(body, "second Q") {
		return fmt.Errorf("the request body still carries the rolled-back exchange (build-then-rollback regression); body=%q", body)
	}
	return nil
}
