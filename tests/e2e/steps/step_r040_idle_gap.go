package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 040 (issue #82) — WS-A streaming liveness (ADR 0009 D3/D4). The idle-gap
// state Given, the quiet-command provider Given, and the liveness Then (T004).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the command stays quiet for longer than the spinner's idle gap$`, givenIdleGapExceeded)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that prints a line and then stays quiet and then answers with "([^"]*)"$`, givenQuietCommandProvider)
		ctx.Then(`^the run shows the progress spinner again while the command stays quiet$`, thenSpinnerAgainWhileQuiet)
	})
}

// round040IdleGapMS is the small nonzero forced threshold the E2E uses so a REAL
// idle gap is observed without a wall-clock wait (dsl.md Given (round 040)).
const round040IdleGapMS = "50"

// givenIdleGapExceeded (怎麼做 / 權威狀態落地): force the round-040 idle-gap seam
// to a small nonzero value so the WS-A resume fires quickly; the diagnostics
// terminal Given sets TELL_ME_FORCE_STDERR_TTY alongside it. The scripted
// command's quiet stretch exceeds N + 2·P with margin (see the provider Given),
// so a real idle gap is observed.
func givenIdleGapExceeded(ctx context.Context) error {
	scenarioFrom(ctx).setEnv("TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS", round040IdleGapMS)
	return nil
}

// givenQuietCommandProvider scripts the fake to request ONE execute_command call
// whose command prints a line and then stays quiet (a child `sleep`, never a Go
// time.Sleep, so verify-no-test-sleep is untouched), then answers {answer}. The
// child's quiet stretch (2 s) exceeds N + 2·P with margin at N=50 ms / P≈200 ms.
func givenQuietCommandProvider(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "execute_command", Arguments: commandArgs(map[string]any{
			"command": "printf 'line one\\n'; sleep 2",
			"reason":  "run it",
		})},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "execute_command"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// thenSpinnerAgainWhileQuiet (必查 / 呈現結果): within the `[Tool Output]` block's
// span — bounded by the CLOSING SEPARATOR, not the last `[Tool Output]` marker
// line (R-1: the resumed frame shares the reset+separator `\n`-line and carries
// no `[Tool Output]` marker) — the captured standard error carries a live spinner
// frame followed by a phase status. 不該發生: the indicator must not stay hidden
// for the whole block (the retired whole-block pause would leave the span empty of
// any spinner frame).
func thenSpinnerAgainWhileQuiet(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	head, _ := toolOutputBlockIndexes(sc.stderr)
	if head < 0 {
		return fmt.Errorf("standard error carried no `[Tool Output]` block to inspect; stderr=%q", sc.stderr)
	}
	closing := closingSeparatorIndex(sc.stderr)
	if closing < 0 {
		return fmt.Errorf("the `[Tool Output]` block carried no closing separator; stderr=%q", sc.stderr)
	}
	if !hasSpinnerStatusBetween(lines, head, closing+1) {
		return fmt.Errorf("no spinner frame appeared inside the `[Tool Output]` block's quiet stretch; stderr=%q", sc.stderr)
	}
	return nil
}
