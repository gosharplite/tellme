package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 024 — the tool resource contract, execute_command, and the reader
// retrofit. Registered from this one cohesive file (the Zero-Shared-Edits
// one-file-per-sentence split exists for PARALLEL dispatch, which this session
// did not use; the sentences are independent regardless).

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command and then answers with "([^"]*)"$`, givenRunsCommand)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that exits non-zero and then answers with "([^"]*)"$`, givenRunsCommandNonZero)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that never returns and then answers with "([^"]*)"$`, givenRunsCommandNeverReturns)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that spawns a long-lived descendant and then answers with "([^"]*)"$`, givenRunsCommandDescendant)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command producing a great deal of output and then answers with "([^"]*)"$`, givenRunsCommandGreatOutput)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that writes its output to a file and then reads it and then answers with "([^"]*)"$`, givenRunsCommandToFile)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint lists the current directory with a small result budget and then answers with "([^"]*)"$`, givenListsSmallBudget)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint shows the folder tree with a small result budget and then answers with "([^"]*)"$`, givenShowsTreeSmallBudget)
		ctx.Given(`^the working directory contains a file "([^"]*)" whose text is longer than the old per-file cap$`, givenOldCapFile)

		ctx.Then(`^tellme ran the command using its execute_command tool$`, thenRanCommand)
		ctx.Then(`^the command result carried the exit status "([^"]*)"$`, thenCommandExit)
		ctx.Then(`^the command result recorded that the command was stopped$`, thenCommandStopped)
		ctx.Then(`^the command result reported that its output went to "([^"]*)"$`, thenCommandOutputTarget)
		ctx.Then(`^the command result was trimmed to what the run can hold$`, thenCommandTrimmed)
		ctx.Then(`^the command's descendant process is no longer running$`, thenDescendantGone)
		ctx.Then(`^the listing was trimmed to what the run can hold$`, thenListingTrimmed)
		ctx.Then(`^the tree was trimmed to what the run can hold$`, thenTreeTrimmed)
		ctx.Then(`^the part of "([^"]*)" that tellme read does not end with a truncation marker$`, thenReadNotTruncated)
		ctx.Then(`^the read result reports that "([^"]*)" was not read$`, thenReadNotRead)
	})
}

// commandArgs marshals execute_command args.
func commandArgs(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// scriptToolThenAnswer scripts the fake to request one tool call, then answer.
func scriptToolThenAnswer(sc *scenarioContext, provider, tool, arguments, answer string) error {
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: tool, Arguments: arguments},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// --- Givens ---

func givenRunsCommand(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command", commandArgs(map[string]any{"command": "echo hi", "reason": "r"}), answer)
}

func givenRunsCommandNonZero(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command", commandArgs(map[string]any{"command": "exit 3", "reason": "r"}), answer)
}

func givenRunsCommandNeverReturns(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command", commandArgs(map[string]any{"command": "sleep 60", "timeout": 1, "reason": "r"}), answer)
}

func givenRunsCommandDescendant(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	// The deterministic fixture: a descendant that holds the pipe, records its pid
	// to child.pid, and schedules a sentinel well past the timeout.
	command := "(sleep 60; echo late > child.sentinel) & echo $! > child.pid; wait"
	return scriptToolThenAnswer(sc, provider, "execute_command", commandArgs(map[string]any{"command": command, "timeout": 1, "reason": "r"}), answer)
}

func givenRunsCommandGreatOutput(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command", commandArgs(map[string]any{"command": "yes", "reason": "r"}), answer)
}

func givenRunsCommandToFile(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "execute_command", Arguments: commandArgs(map[string]any{"command": "echo hello", "output_file": "out.txt", "reason": "r"})},
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgsWithReason("out.txt", "read it back")},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

func givenListsSmallBudget(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "list_files", commandArgs(map[string]any{"path": ".", "max_output_tokens": 100, "reason": "r"}), answer)
}

func givenShowsTreeSmallBudget(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "get_tree", commandArgs(map[string]any{"max_depth": 2, "max_output_tokens": 100, "reason": "r"}), answer)
}

func givenOldCapFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	// Larger than the retired 100000-byte per-file cap but well within the default
	// byte budget (1000000 B), so the reader returns it whole.
	content := strings.Repeat("m", 120000)
	return os.WriteFile(filepath.Join(sc.workDir, filepath.FromSlash(name)), []byte(content), 0o644)
}

// --- Thens ---

func thenRanCommand(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "execute_command", "") {
		return fmt.Errorf("no execute_command tool call was recorded")
	}
	return nil
}

func thenCommandExit(ctx context.Context, code string) error {
	if _, ok := toolResultContaining(ctx, "Exit Code: "+code); !ok {
		return fmt.Errorf("no command result carries exit status %q", code)
	}
	return nil
}

func thenCommandStopped(ctx context.Context) error {
	if _, ok := toolResultContaining(ctx, "stopped at the time limit"); !ok {
		return fmt.Errorf("no command result records a stop")
	}
	return nil
}

func thenCommandOutputTarget(ctx context.Context, path string) error {
	if _, ok := toolResultContaining(ctx, "Output written to "+path); !ok {
		return fmt.Errorf("no command result names %q as the capture target", path)
	}
	return nil
}

func thenCommandTrimmed(ctx context.Context) error {
	if _, ok := toolResultContaining(ctx, "(truncated)"); !ok {
		return fmt.Errorf("no command result was trimmed")
	}
	return nil
}

func thenDescendantGone(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	pidPath := filepath.Join(sc.workDir, "child.pid")
	raw, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("the descendant fixture did not record its pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return fmt.Errorf("child.pid is not a pid: %v", err)
	}
	if !pollProcessGone(pid, 2*time.Second) {
		return fmt.Errorf("the descendant process %d is still alive after the stop", pid)
	}
	if _, err := os.Stat(filepath.Join(sc.workDir, "child.sentinel")); err == nil {
		return fmt.Errorf("the descendant wrote its late sentinel — it was not stopped with the group")
	}
	return nil
}

func thenListingTrimmed(ctx context.Context) error {
	return thenTrimmed(ctx, "listing")
}

func thenTreeTrimmed(ctx context.Context) error {
	return thenTrimmed(ctx, "tree")
}

func thenTrimmed(ctx context.Context, what string) error {
	res := lastToolResultOf(ctx)
	if !strings.Contains(res, "(truncated)") {
		return fmt.Errorf("the %s result was not trimmed; got %q", what, tailStr(res, 120))
	}
	return nil
}

func thenReadNotTruncated(ctx context.Context, name string) error {
	res := lastToolResultOf(ctx)
	if strings.Contains(res, "(truncated)") {
		return fmt.Errorf("the read of %q was truncated but must be returned whole; got %q", name, tailStr(res, 120))
	}
	return nil
}

func thenReadNotRead(ctx context.Context, name string) error {
	res := lastToolResultOf(ctx)
	if !strings.Contains(res, "not read") || !strings.Contains(res, name) {
		return fmt.Errorf("the read result does not name %q as not read; got %q", name, tailStr(res, 160))
	}
	return nil
}

// --- helpers ---

// lastToolResultOf returns the last recorded tool result for the scenario's only
// fake.
func lastToolResultOf(ctx context.Context) string {
	f := scenarioFrom(ctx).onlyFake()
	if f == nil {
		return ""
	}
	return lastToolResult(f)
}

// toolResultContaining scans every recorded tool result for one containing sub.
func toolResultContaining(ctx context.Context, sub string) (string, bool) {
	f := scenarioFrom(ctx).onlyFake()
	if f == nil {
		return "", false
	}
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role == "tool" && strings.Contains(m.Content, sub) {
				return m.Content, true
			}
		}
	}
	return "", false
}

// pollProcessGone reports whether the pid is gone within the deadline (the group
// is reaped asynchronously).
func pollProcessGone(pid int, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for {
		if err := syscall.Kill(pid, 0); err != nil {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// tailStr returns the last n characters of s (for diagnostics).
func tailStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
