package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Round 053 (closes #103; ADR 0022) — the offline session commands (`-l`, `-t`)
// select the session named by the `-c` configuration's MODE.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds the configuration "([^"]*)" in mode "([^"]*)" holding the answer "([^"]*)"$`, givenConfigSessionHoldsAnswer)
		ctx.Given(`^the runtime home holds the configuration "([^"]*)" in mode "([^"]*)" whose turn log holds "([^"]*)"$`, givenConfigSessionTurnsLog)
		ctx.When(`^the operator asks tellme to list the last (\d+) messages of the configuration "([^"]*)"$`, whenListLastOfConfig)
		ctx.When(`^the operator reviews the turn log of the configuration "([^"]*)"$`, whenReviewTurnsLogOfConfig)
		ctx.When(`^the operator starts a fresh session of the configuration "([^"]*)"$`, whenNewOfConfig)
		ctx.Then(`^tellme lists the assistant message "([^"]*)"$`, thenListsAssistantMessage)
		ctx.Then(`^tellme prints exactly the turn log line "([^"]*)"$`, thenPrintsExactlyTurnLog)
		ctx.Then(`^tellme prints nothing$`, thenPrintsNothing)
		ctx.Then(`^the session's turn log holds the turn progress$`, thenSessionTurnsLogHoldsProgress)
		ctx.Then(`^the "([^"]*)" session holds no active exchanges$`, thenNamedSessionNoExchanges)
		ctx.Then(`^the "([^"]*)" session archived the exchange "([^"]*)" and "([^"]*)"$`, thenNamedSessionArchivedExchange)
	})
}

// writeNamedConfig writes a resolvable configuration named {config} (under
// configs/) whose MODE is {mode}.
func writeNamedConfig(sc *scenarioContext, config, mode string) error {
	return sc.writeFile(filepath.Join("configs", config), []byte(wellFormedConfig(mode, "deepseek-flash")))
}

// givenConfigSessionHoldsAnswer arranges the named configuration and seeds its
// MODE's session with one exchange (prompt "hello", answer {answer}).
func givenConfigSessionHoldsAnswer(ctx context.Context, config, mode, answer string) error {
	sc := scenarioFrom(ctx)
	if err := writeNamedConfig(sc, config, mode); err != nil {
		return err
	}
	dir := filepath.Join(sc.home, "output", mode)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(historyEntry{Prompt: "hello", Answer: answer})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "history.jsonl"), append(line, '\n'), 0o644)
}

// givenConfigSessionTurnsLog arranges the named configuration and seeds its
// MODE's session with a turn log holding {content}.
func givenConfigSessionTurnsLog(ctx context.Context, config, mode, content string) error {
	sc := scenarioFrom(ctx)
	if err := writeNamedConfig(sc, config, mode); err != nil {
		return err
	}
	dir := filepath.Join(sc.home, "output", mode)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "turns.log"), []byte(content), 0o644)
}

// whenListLastOfConfig runs `tellme -l {count} -c <config>` (no positional prompt).
func whenListLastOfConfig(ctx context.Context, count int, config string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-l", strconv.Itoa(count), "-c", sc.homePath(filepath.Join("configs", config))}
	sc.run()
	return nil
}

// whenReviewTurnsLogOfConfig runs `tellme -t -c <config>` (no positional prompt).
func whenReviewTurnsLogOfConfig(ctx context.Context, config string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-t", "-c", sc.homePath(filepath.Join("configs", config))}
	sc.run()
	return nil
}

// whenNewOfConfig runs `tellme --new -c <config>` (no positional prompt).
func whenNewOfConfig(ctx context.Context, config string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"--new", "-c", sc.homePath(filepath.Join("configs", config))}
	sc.run()
	return nil
}

// thenListsAssistantMessage (必查 呈現結果): the last line on stdout is
// `assistant: {answer}` — the named configuration's session, not any other's.
func thenListsAssistantMessage(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	want := "assistant: " + answer
	lines := strings.Split(strings.TrimRight(sc.stdout, "\n"), "\n")
	if len(lines) == 0 || lines[len(lines)-1] != want {
		return fmt.Errorf("the last listed line must be %q; stdout=%q", want, sc.stdout)
	}
	return nil
}

// thenPrintsExactlyTurnLog (必查 呈現結果): stdout is exactly {content}.
func thenPrintsExactlyTurnLog(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	if strings.TrimRight(sc.stdout, "\n") != strings.TrimRight(content, "\n") {
		return fmt.Errorf("stdout must be exactly %q; got %q", content, sc.stdout)
	}
	return nil
}

// thenPrintsNothing (必查 呈現結果): stdout carries no characters.
func thenPrintsNothing(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.stdout != "" {
		return fmt.Errorf("stdout must be empty; got %q", sc.stdout)
	}
	return nil
}

// thenSessionTurnsLogHoldsProgress (必查 權威狀態 / the write side of the turn
// log): the resolved session's turns.log holds the rendered turn chrome — the
// `╭─⠿ Turn …` header and the payload status line. Round 057 (F-057-2): the
// payload check pins the CURRENT shape (`Payload: +<delta> ~<n> …`), not the bare
// substring `Payload:` that the pre-057 shape would also satisfy.
func thenSessionTurnsLogHoldsProgress(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.historyDir(), "turns.log"))
	if err != nil {
		return fmt.Errorf("the session turn log must be written: %w", err)
	}
	got := string(data)
	if !strings.Contains(got, "╭─⠿ Turn") || !reTurnLogEstimate.MatchString(got) {
		return fmt.Errorf("the turn log must carry the turn chrome (header + the round-057 payload estimate); got %q", got)
	}
	return nil
}

// thenNamedSessionNoExchanges (必查 權威狀態): the named mode's active
// history.jsonl holds no exchange lines (a fresh-session start archived them).
func thenNamedSessionNoExchanges(ctx context.Context, mode string) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(filepath.Join(sc.home, "output", mode, "history.jsonl"))
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("the %q session holds %d exchanges, want none", mode, len(entries))
	}
	return nil
}

// thenNamedSessionArchivedExchange (必查 權威狀態): the named mode's archive holds
// the exchange whose prompt/answer are {prompt}/{answer}.
func thenNamedSessionArchivedExchange(ctx context.Context, mode, prompt, answer string) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(filepath.Join(sc.home, "output", mode, "history.archive.jsonl"))
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Prompt == prompt && e.Answer == answer {
			return nil
		}
	}
	return fmt.Errorf("the %q archive does not hold %q / %q; got %+v", mode, prompt, answer, entries)
}
