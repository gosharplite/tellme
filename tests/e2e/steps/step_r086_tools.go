package steps

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Round 086 (ADR 0057) — the `-l` listing's per-turn tool-activity line
// `[TOOLS] - M (N calls)` (between the turn's [USER] block and its [MODEL]
// header, yellow on a terminal stdout).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing reports each turn's tool activity$`, thenListingReportsEachTurnsToolActivity)
		ctx.Then(`^the listing accents the tool-activity line in yellow$`, thenListingAccentsToolLineYellow)
	})
}

// parseToolHeader parses the round-086 tool-activity line:
// `[TOOLS] - M (N calls)` (with M) or `[TOOLS] (N calls)` (no index — the
// non-positive-index fallback). It returns the index (0 when absent), the count,
// and whether the line is a tool line at all.
func parseToolHeader(line string) (index int, calls int, ok bool) {
	rest, found := strings.CutPrefix(line, "[TOOLS]")
	if !found {
		return 0, 0, false
	}
	rest = strings.TrimSpace(rest)
	if v, f := strings.CutPrefix(rest, "- "); f {
		numPart, tail, f2 := strings.Cut(v, " (")
		if !f2 {
			return 0, 0, false
		}
		n, err := strconv.Atoi(strings.TrimSpace(numPart))
		if err != nil {
			return 0, 0, false
		}
		index = n
		rest = "(" + tail
	}
	if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, " calls)") {
		return 0, 0, false
	}
	numStr := strings.TrimSuffix(strings.TrimPrefix(rest, "("), " calls)")
	c, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil {
		return 0, 0, false
	}
	return index, c, true
}

// listedTurn is one parsed turn block: the [USER]/[MODEL] backward indices, the
// tool line's index/count (hasTool when present), and the two visible bodies.
type listedTurn struct {
	turnIndex  int
	modelIndex int
	toolIndex  int
	toolCalls  int
	hasTool    bool
	userBody   string
	modelBody  string
}

// listedTurns splits a listing capture into its turn blocks, recognising the
// round-082 role headers and the round-086 tool line. ANSI is stripped first
// (the rendered model body's glamour style output is not the listing contract).
func listedTurns(out string) []listedTurn {
	lines := strings.Split(stripANSI(out), "\n")
	var turns []listedTurn
	var cur *listedTurn
	var pending *struct {
		index, calls int
	}
	section := ""
	flush := func() {
		if cur != nil {
			turns = append(turns, *cur)
			cur = nil
		}
	}
	applyPending := func(t *listedTurn) {
		if pending != nil {
			t.toolIndex, t.toolCalls, t.hasTool = pending.index, pending.calls, true
			pending = nil
		}
	}
	for _, raw := range lines {
		t := strings.TrimRight(raw, " \t")
		if role, idx, ok := parseListingHeader(t); ok {
			if role == "[USER]" {
				flush()
				cur = &listedTurn{turnIndex: idx}
				applyPending(cur)
				section = "user"
				continue
			}
			if cur == nil {
				cur = &listedTurn{turnIndex: idx}
			}
			cur.modelIndex = idx
			applyPending(cur)
			section = "model"
			continue
		}
		if ti, tc, ok := parseToolHeader(strings.TrimSpace(t)); ok {
			pending = &struct{ index, calls int }{ti, tc}
			section = ""
			continue
		}
		if strings.TrimSpace(t) == "" {
			continue
		}
		if cur == nil {
			continue
		}
		switch section {
		case "user":
			cur.userBody += t + "\n"
		case "model":
			cur.modelBody += t + "\n"
		}
	}
	flush()
	return turns
}

// thenListingReportsEachTurnsToolActivity (必查 呈現結果; round 086, ADR 0057): for
// EVERY listed turn, stdout carries exactly one `[TOOLS] - M (N calls)` line
// between the turn's [USER] block and its [MODEL] header, with M the turn's
// backward index (equal to the role headers' index) and N the number of tool
// steps the ARRANGED turn recorded. The expectation is derived from the arranged
// exchanges AND the `-l N` request (the last N MESSAGES) — never from the
// observed output — so a missing line, a wrong index, a `Calls`-derived count, or
// an extra line reddens.
func thenListingReportsEachTurnsToolActivity(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	turns := listedTurns(sc.stdout)

	total := len(sc.arrangedExchanges)
	// The listed window = the last N messages (two per turn); a turn is listed
	// when either of its messages is in the window.
	type msg struct{ turn int }
	var msgs []msg
	for i := range sc.arrangedExchanges {
		turnIndex := total - i
		msgs = append(msgs, msg{turnIndex}, msg{turnIndex})
	}
	if n := arrangedListCount(sc); len(msgs) > n {
		msgs = msgs[len(msgs)-n:]
	}
	var wantIdx []int
	seen := map[int]bool{}
	for _, m := range msgs {
		if !seen[m.turn] {
			seen[m.turn] = true
			wantIdx = append(wantIdx, m.turn)
		}
	}
	wantCalls := map[int]int{}
	for i, x := range sc.arrangedExchanges {
		wantCalls[total-i] = x.toolCalls
	}

	if len(turns) != len(wantIdx) {
		return fmt.Errorf("the listing showed %d turns, want %d; stdout=%q", len(turns), len(wantIdx), sc.stdout)
	}
	for i, w := range wantIdx {
		t := turns[i]
		if t.turnIndex != w || t.modelIndex != w {
			return fmt.Errorf("turn block %d indices = user %d / model %d, want %d; stdout=%q", i, t.turnIndex, t.modelIndex, w, sc.stdout)
		}
		if !t.hasTool {
			return fmt.Errorf("turn %d carries no `[TOOLS] - M (N calls)` line; stdout=%q", w, sc.stdout)
		}
		if t.toolIndex != w {
			return fmt.Errorf("turn %d tool line index = %d, want %d; stdout=%q", w, t.toolIndex, w, sc.stdout)
		}
		if t.toolCalls != wantCalls[w] {
			return fmt.Errorf("turn %d tool count = %d, want %d (len(Steps), not Calls); stdout=%q", w, t.toolCalls, wantCalls[w], sc.stdout)
		}
	}
	return nil
}

// thenListingAccentsToolLineYellow (必查 呈現結果; round 086, ADR 0057): on a
// terminal stdout the WHOLE `[TOOLS] - M (N calls)` label is wrapped in the
// reference's yellow SGR (`\033[0;33m … \033[0m`).
func thenListingAccentsToolLineYellow(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "\x1b[0;33m[TOOLS] - ") {
		return fmt.Errorf("the whole [TOOLS] label must be accented yellow; stdout=%q", sc.stdout)
	}
	if !strings.Contains(sc.stdout, "\x1b[0m") {
		return fmt.Errorf("the accent must be reset with \\x1b[0m; stdout=%q", sc.stdout)
	}
	return nil
}
