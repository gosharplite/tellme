package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round 039 (operator spacing request; ADR 0008) — the live turn output is grouped
// by blank lines: one blank before EACH tool call's begin block, one blank before
// the grouped post-call tail reason block, and one blank before the post-status
// group (measured payload + metrics + `Ready`).

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^each tool call's report begins after a blank line$`, thenEachCallReportBeginsAfterBlank)
		ctx.Then(`^the action of the call without a reason begins after a blank line$`, thenReasonlessActionBeginsAfterBlank)
		ctx.Then(`^the trailing reason summary follows a blank line$`, thenTrailingReasonSummaryFollowsBlank)
		ctx.Then(`^the turn's closing status follows a blank line$`, thenTurnClosingStatusFollowsBlank)
		ctx.Then(`^the closing status is preceded by exactly the frame gap$`, thenTurnClosingStatusNoBlank)
	})
}

// r039BlockStart returns the index of a tool call's begin block (the `[Tool
// Reason]` line when the line before the `[Tool Action]` line is a reason line,
// else the action line itself).
func r039BlockStart(lines []string, actionIdx int) int {
	if actionIdx > 0 && strings.Contains(lines[actionIdx-1], toolReasonMarker) {
		return actionIdx - 1
	}
	return actionIdx
}

func thenEachCallReportBeginsAfterBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	total, blank := 0, 0
	for i, l := range lines {
		if !strings.Contains(l, toolActionMarker) {
			continue
		}
		total++
		if begin := r039BlockStart(lines, i); begin > 0 && lines[begin-1] == "" {
			blank++
		}
	}
	if total == 0 {
		return fmt.Errorf("no tool call was reported; stderr=%q", sc.stderr)
	}
	if blank != total {
		return fmt.Errorf("only %d/%d tool-call begin blocks were preceded by a blank line; stderr=%q", blank, total, sc.stderr)
	}
	return nil
}

func thenReasonlessActionBeginsAfterBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	found := false
	for i, l := range lines {
		if !strings.Contains(l, toolActionMarker) {
			continue
		}
		if i > 0 && strings.Contains(lines[i-1], toolReasonMarker) {
			continue // this call stated a reason; its block starts at the reason line
		}
		found = true
		if i == 0 || lines[i-1] != "" {
			return fmt.Errorf("a reason-less call's action line was not preceded by a blank line; stderr=%q", sc.stderr)
		}
	}
	if !found {
		return fmt.Errorf("no reason-less tool call was reported; stderr=%q", sc.stderr)
	}
	return nil
}

func thenTrailingReasonSummaryFollowsBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	var idx []int
	for i, l := range lines {
		if strings.Contains(l, toolReasonMarker) {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Reason]` line; stderr=%q", sc.stderr)
	}
	// The trailing summary is the last maximal run of consecutive reason lines.
	start := len(idx) - 1
	for start > 0 && idx[start-1] == idx[start]-1 {
		start--
	}
	first := idx[start]
	if first == 0 || lines[first-1] != "" {
		return fmt.Errorf("the trailing reason summary was not preceded by a blank line; stderr=%q", sc.stderr)
	}
	return nil
}

// measuredPayloadIndexes returns the line indexes of every measured (`~`-less)
// `Payload:` line — the first line of a post-status group. Shared by the two
// closing-status assertions so their predicate cannot drift (round-039 fold
// review).
func measuredPayloadIndexes(lines []string) []int {
	var idx []int
	for i, l := range lines {
		if strings.Contains(l, "Payload: ") && !strings.Contains(l, "Payload: ~") {
			idx = append(idx, i)
		}
	}
	return idx
}

func thenTurnClosingStatusFollowsBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	idx := measuredPayloadIndexes(lines)
	if len(idx) == 0 {
		return fmt.Errorf("standard error carried no measured payload status line; stderr=%q", sc.stderr)
	}
	for _, i := range idx {
		if i == 0 || lines[i-1] != "" {
			return fmt.Errorf("the turn's closing status was not preceded by a blank line; stderr=%q", sc.stderr)
		}
	}
	return nil
}

// thenTurnClosingStatusNoBlank (round 039 review B1; hardening per the fold
// review): a TOOL-LESS turn gains no blank before its post-status group. The
// round-017 frame gap already puts exactly ONE blank between the pre-flight
// estimate and the measured payload, so the assertion is TARGETED — the line
// immediately before the measured `Payload:` line must be that single frame-gap
// blank whose own predecessor is the pre-flight (`Payload: ~…`) line. A global
// `\n\n\n` absence would go silent if the frame gap were ever retuned; this form
// fails visibly in both directions (it asserts the separation is exactly the frame
// gap, not an added blank). This step is valid only on a TOOL-LESS turn.
func thenTurnClosingStatusNoBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	idx := measuredPayloadIndexes(lines)
	if len(idx) == 0 {
		return fmt.Errorf("standard error carried no measured payload status line; stderr=%q", sc.stderr)
	}
	for _, i := range idx {
		if i >= 2 && lines[i-1] == "" && strings.Contains(lines[i-2], "Payload: ~") {
			continue // exactly the round-017 frame gap — the correct, unchanged shape
		}
		return fmt.Errorf("a tool-less turn's closing status was not preceded by exactly the frame gap (an extra blank?); stderr=%q", sc.stderr)
	}
	return nil
}
