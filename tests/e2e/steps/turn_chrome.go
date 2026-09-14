package steps

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// Round-017 turn-chrome helpers (the non-TUI operator chrome). Kept in ONE file
// so the per-sentence step files stay independent (Zero Shared Edits), mirroring
// tui_chrome.go / payload_status.go / ordering_witness.go.

var (
	// reTurnAck matches the reference's input-capture acknowledgement line.
	reTurnAck = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Input captured\. Processing\.\.\.`)
	// reTurnHeader matches `╭─⠿ Turn <N>` with an optional ` - <mode>` suffix.
	reTurnHeader = regexp.MustCompile(`╭─⠿ Turn ([0-9]+)(?: - (\S+))?`)
)

// turnRuleWidth mirrors the product's fixed rule width (single-sourced so a
// product-side change cannot make the guard vacuous — the round-012 TD1 pattern).
const turnRuleWidth = 80

// turnRuleStr is the rule computed once (round-017 implementation-review finding 2).
var turnRuleStr = strings.Repeat("─", turnRuleWidth)

func turnRuleLiteral() string { return turnRuleStr }

// hasTurnAck reports whether s carries the input-capture acknowledgement.
func hasTurnAck(s string) bool { return reTurnAck.MatchString(harness.StripANSI(s)) }

// hasTurnRule reports whether s carries a line that is exactly the rule (the
// full-width `─` × 80) — an EXACT line, so a TUI editor border (`╭───╮`) does not
// match.
func hasTurnRule(s string) bool {
	rule := turnRuleLiteral()
	for _, ln := range strings.Split(harness.StripANSI(s), "\n") {
		if strings.TrimRight(ln, "\r") == rule {
			return true
		}
	}
	return false
}

// turnHeaderParts returns the (turn number, mode, found) of the first turn header.
func turnHeaderParts(s string) (int, string, bool) {
	m := reTurnHeader.FindStringSubmatch(harness.StripANSI(s))
	if m == nil {
		return 0, "", false
	}
	n, _ := strconv.Atoi(m[1])
	return n, m[2], true
}

// hasTurnHeader reports whether s carries a turn header.
func hasTurnHeader(s string) bool {
	_, _, ok := turnHeaderParts(s)
	return ok
}

// hasAnyTurnChrome reports whether s carries any part of the turn chrome (the
// acknowledgement, the rule, or the header) — the negative-boundary predicate.
func hasAnyTurnChrome(s string) bool {
	return hasTurnAck(s) || hasTurnRule(s) || hasTurnHeader(s)
}
