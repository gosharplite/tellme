package steps

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 039 (issue #80; ADR 0008) — the terminal-safe-line policy is generalized
// to every `[Tool …]` formatter: the `[Tool Reason]`, `[Tool Result]`, and
// `[Tool Action]` lines are control-free (same class as `[Tool Output]`).

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" with a reason that carries terminal control data and then answers with "([^"]*)"$`, givenProviderReadWithControlReason)
		ctx.Then(`^the run reported a reason line free of terminal control sequences$`, thenReasonControlFree)
		ctx.Then(`^the run reported the result for the tool call "([^"]*)" free of terminal control sequences$`, thenResultControlFree)
		ctx.Then(`^the run reported the action for the tool call "([^"]*)" free of terminal control sequences$`, thenActionControlFree)
	})
}

// r039ControlFree reports whether a line carries no terminal control data (ESC, a
// C0 control other than TAB, or DEL) AND is valid UTF-8.
func r039ControlFree(line string) bool {
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == 0x1b || (c < 0x20 && c != '\t') || c == 0x7f {
			return false
		}
	}
	return utf8.ValidString(line)
}

// r039LinesContaining returns the stderr lines carrying substr.
func r039LinesContaining(stderr, substr string) []string {
	var out []string
	for _, l := range stderrLines(stderr) {
		if strings.Contains(l, substr) {
			out = append(out, l)
		}
	}
	return out
}

// givenProviderReadWithControlReason scripts a `read_files` call whose top-level
// `reason` carries an ANSI escape (SGR set + reset) around visible text, then the
// answer. The loop must render the reason line control-free.
func givenProviderReadWithControlReason(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgsWithReason(path, "\x1b[31mchecking the launch code\x1b[0m")},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

func thenReasonControlFree(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := r039LinesContaining(sc.stderr, toolReasonMarker)
	if len(lines) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Reason]` line; stderr=%q", sc.stderr)
	}
	for _, l := range lines {
		if !r039ControlFree(l) {
			return fmt.Errorf("a `[Tool Reason]` line carried a terminal control sequence; line=%q stderr=%q", l, sc.stderr)
		}
	}
	return nil
}

func thenResultControlFree(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	lines := r039LinesContaining(sc.stderr, toolResultMarker+tool+": ")
	if len(lines) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Result] %s: …` line; stderr=%q", tool, sc.stderr)
	}
	for _, l := range lines {
		if !r039ControlFree(l) {
			return fmt.Errorf("a `[Tool Result]` line carried a terminal control sequence; line=%q stderr=%q", l, sc.stderr)
		}
	}
	return nil
}

func thenActionControlFree(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	lines := r039LinesContaining(sc.stderr, toolActionMarker+tool+"(")
	if len(lines) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Action] %s(…)` line; stderr=%q", tool, sc.stderr)
	}
	for _, l := range lines {
		if !r039ControlFree(l) {
			return fmt.Errorf("a `[Tool Action]` line carried a terminal control sequence; line=%q stderr=%q", l, sc.stderr)
		}
	}
	return nil
}
