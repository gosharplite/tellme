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

func thenTurnClosingStatusFollowsBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	for i, l := range lines {
		if !strings.Contains(l, "Payload: ") || strings.Contains(l, "Payload: ~") {
			continue // only the measured (~-less) payload line starts the post-status group
		}
		if i == 0 || lines[i-1] != "" {
			return fmt.Errorf("the turn's closing status was not preceded by a blank line; stderr=%q", sc.stderr)
		}
		return nil
	}
	return fmt.Errorf("standard error carried no measured payload status line; stderr=%q", sc.stderr)
}
