package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round 038 (issue #78 + #76) — tool-output sanitization and the empty-submit
// no-op. Registered from this one cohesive file (the Zero-Shared-Edits
// one-file-per-sentence split was not needed for this session). The sanitizer
// /prompt helpers live here too.

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a colouring command and then answers with "([^"]*)"$`, givenRunsColouringCommand)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a colouring command that is stopped at its time limit and then answers with "([^"]*)"$`, givenRunsColouringCommandStopped)
		ctx.When(`^the operator submits an empty prompt then submits the prompt "([^"]*)" at the interactive prompt$`, whenSubmitEmptyThenSubmit)
		ctx.Then(`^the run streamed the command's output free of terminal control sequences$`, thenStreamedOutputControlFree)
		ctx.Then(`^the terminal is left in its default state$`, thenTerminalNeutral)
	})
}

// --- helpers ---

// r038ContentLines returns the emitted `[Tool Output] <text>` content lines
// (excluding the header), so a check can be scoped off the block's own close
// restore (a deliberate ESC).
func r038ContentLines(stream string) []string {
	var out []string
	for _, ln := range strings.Split(stream, "\n") {
		i := strings.Index(ln, toolOutputMarker)
		if i < 0 {
			continue
		}
		if strings.HasPrefix(ln[i+len(toolOutputMarker):], "Executing... (Output shown below)") {
			continue
		}
		out = append(out, ln)
	}
	return out
}

// r038ControlFree reports whether every `[Tool Output]` content line is free of
// terminal control data (ESC, a C0 control other than TAB, or DEL).
func r038ControlFree(stream string) bool {
	for _, ln := range r038ContentLines(stream) {
		for i := 0; i < len(ln); i++ {
			c := ln[i]
			if c == 0x1b || (c < 0x20 && c != '\t') || c == 0x7f {
				return false
			}
		}
	}
	return true
}

// r038RestoredAfterBlock reports whether a default-state restore appears after
// the last `[Tool Output]` line (the block closes neutral).
func r038RestoredAfterBlock(stream string) bool {
	last := strings.LastIndex(stream, toolOutputMarker)
	if last < 0 {
		return false
	}
	return strings.Contains(stream[last:], "\x1b[0m")
}

// --- Givens ---

func givenRunsColouringCommand(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command",
		commandArgs(map[string]any{"command": "printf '\\033[31mERROR: failed\\n'", "reason": "r"}), answer)
}

func givenRunsColouringCommandStopped(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptToolThenAnswer(sc, provider, "execute_command",
		commandArgs(map[string]any{"command": "printf '\\033[31mworking\\n'; sleep 60", "timeout": 1, "reason": "r"}), answer)
}

// --- When ---

// whenSubmitEmptyThenSubmit scripts `-i` with an EMPTY submit first, then the
// real prompt submit (round 038 FR-003). The last byte of the key sequence is
// the terminal key delivered after the frame paints; the leading empty submit is
// therefore delivered in the compose batch and must be a no-op for the run to
// continue to the real submit.
func whenSubmitEmptyThenSubmit(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	launchTUI(sc, tuiKeySubmit+unescapeText(prompt)+tuiKeySubmit)
	return nil
}

// --- Thens ---

func thenStreamedOutputControlFree(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := r038ContentLines(sc.stderr)
	if len(lines) == 0 {
		return fmt.Errorf("no `[Tool Output]` content line to inspect; stderr=%q", sc.stderr)
	}
	if !r038ControlFree(sc.stderr) {
		return fmt.Errorf("a streamed `[Tool Output]` content line carried a terminal control sequence; stderr=%q", sc.stderr)
	}
	return nil
}

func thenTerminalNeutral(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !r038RestoredAfterBlock(sc.stderr) {
		return fmt.Errorf("no default-state restore after the `[Tool Output]` block; stderr=%q", sc.stderr)
	}
	return nil
}
