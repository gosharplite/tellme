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
		ctx.Then(`^tellme lists the assistant message "([^"]*)"$`, thenListsAssistantMessage)
		ctx.Then(`^tellme prints exactly the turn log line "([^"]*)"$`, thenPrintsExactlyTurnLog)
		ctx.Then(`^tellme prints nothing$`, thenPrintsNothing)
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
	if strings.TrimSpace(sc.stdout) != "" {
		return fmt.Errorf("stdout must be empty; got %q", sc.stdout)
	}
	return nil
}
