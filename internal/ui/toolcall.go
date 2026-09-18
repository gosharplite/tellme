package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Round-034 decomposed tool-call rendering
// (specs/truth/features/cli/chat/watching-the-tool-loop.feature; ADR 0005). The
// pure formatters below supersede the round-022 single-line `FormatToolLog`:
//
//	[HH:MM:SS] [Tool Engine] Step <i>/<m>
//	[HH:MM:SS] [Tool Reason] <reason>
//	[HH:MM:SS] [Tool Action] <tool>(<sorted k: v>)
//	[HH:MM:SS] [Tool Result] <tool>: <snippet>
//
// Truncation is RUNE-safe (ADR 0005 D5): a maximum total rendered rune length,
// one U+2026 counted inside the cap, cut on a rune boundary, evaluated on the
// folded value. Argument keys are sorted ascending and the `reason` key is
// excluded (ADR 0005 D6); ordering/value rendering NEVER mutates the call's
// arguments (round-014 replay fidelity).

// argValueCap is the maximum total rendered rune length of one argument value
// (FR-003): fold first, truncate iff > 189 runes → first 188 runes + one U+2026.
const argValueCap = 189

// resultValueCap is the maximum total rendered rune length of a result snippet
// (FR-004): fold first, truncate iff > 200 runes → first 199 runes + one U+2026.
const resultValueCap = 200

// reasonValueCap is the maximum total rendered rune length of the model-authored
// reason (round 036, issue #74, ADR 0006): fold + trim first, truncate iff > 200
// runes → first 199 runes + one U+2026. It mirrors resultValueCap, so the reason
// joins its sibling sanitize+cap family (cap set {189, 200, 200}).
const reasonValueCap = 200

// FormatToolEngine renders the per-executed-round step marker (FR-001):
// `[HH:MM:SS] [Tool Engine] Step <step>/<total>`.
func FormatToolEngine(t time.Time, step, total int) string {
	return fmt.Sprintf("[%s] [Tool Engine] Step %d/%d", formatClock(t), step, total)
}

// toolReasonText is the ONE reason transform (round 039 review RF-1): fold + trim
// → sanitize → cap. Both the rendered row (`FormatToolReason`) and the "would this
// render a row?" predicate (`ToolReasonRenders`) derive from it, so the chain has a
// single definition and cannot drift (it has already changed twice in three rounds
// — 036 fold/trim, 039 sanitize).
func toolReasonText(reason string) string {
	return capRunes(sanitizeControl(oneLine(strings.TrimSpace(reason))), reasonValueCap)
}

// FormatToolReason renders one reason line (FR-002/FR-005, round 036; round 039):
// `[HH:MM:SS] [Tool Reason] <reason>`. The reason is the ONLY model-authored
// free-text field in the tool log, so it is sanitized + capped like its siblings
// (the `toolReasonText` transform): folded (`\n`/`\r` → space), trimmed to a
// single line, **sanitized** (every terminal control sequence removed — round 039,
// issue #80, ADR 0008), then rune-capped at reasonValueCap (one U+2026 inside the
// cap), so the cap bounds the VISIBLE output. A blank (empty / whitespace-only)
// reason is suppressed by the CALLERS (not here), so this stays a pure formatter
// and never returns an empty-string sentinel.
func FormatToolReason(t time.Time, reason string) string {
	return formatToolReasonLine(t, toolReasonText(reason))
}

// formatToolReasonLine builds the `[HH:MM:SS] [Tool Reason] <text>` row from an
// ALREADY-TRANSFORMED reason text. It is the ONE place the row's format literal
// lives, shared by FormatToolReason (the pure formatter) and
// ToolLineRenderer.ReasonLine (the port adapter), so the two can never drift
// (round-046 review TD-6).
func formatToolReasonLine(t time.Time, text string) string {
	return fmt.Sprintf("[%s] [Tool Reason] %s", formatClock(t), text)
}

// ToolReasonRenders reports whether a reason value renders a non-empty
// `[Tool Reason]` line: it applies the SAME `toolReasonText` transform as
// FormatToolReason and tests whether the rendered text is non-blank. It lets a
// caller suppress a reason that would render as a dangling prefix row — an empty /
// whitespace-only reason (round 036) or an escape-only reason (round 039, issue
// #80, ADR 0008) — without the pure formatter returning an empty-string sentinel.
//
// Round 046 (R4 of #92, ADR 0015): the PRODUCTION owner of this predicate is now
// ToolLineRenderer.ReasonLine (it evaluates `toolReasonText` once and returns both
// the line and the decision), so this helper has no production caller — it is
// retained as a TEST-FACING public helper (its own unit pin) and as the documented
// statement of the predicate. A future round may fold it onto ReasonLine.
func ToolReasonRenders(reason string) bool {
	return strings.TrimSpace(toolReasonText(reason)) != ""
}

// FormatToolAction renders the action line — the tool name plus its sorted,
// rune-capped argument list with `reason` excluded (FR-003):
// `[HH:MM:SS] [Tool Action] <tool>(<sorted k: v>)`. Unparseable or non-object
// arguments render the empty argument list (`<tool>()`).
func FormatToolAction(t time.Time, tool, arguments string) string {
	return fmt.Sprintf("[%s] [Tool Action] %s(%s)", formatClock(t), tool, formatToolArgs(arguments))
}

// FormatToolResult renders the result line — the tool name plus the folded,
// sanitized, rune-capped result snippet (FR-004; round 039): the result text is
// sanitized (control sequences removed — ADR 0008) before it is rune-capped, so a
// `read_files` of an escape-bearing file cannot tint the terminal via its snippet:
// `[HH:MM:SS] [Tool Result] <tool>: <snippet>`.
func FormatToolResult(t time.Time, tool, text string) string {
	return fmt.Sprintf("[%s] [Tool Result] %s: %s", formatClock(t), tool, capRunes(sanitizeControl(oneLine(text)), resultValueCap))
}

// formatToolArgs renders the sorted, reason-excluded, rune-capped `k: v, k: v`
// argument list (FR-003 / ADR 0005 D6). Round 039 (issue #80, ADR 0008): both the
// argument **keys** and the rendered **values** are sanitized (control sequences
// removed) — each is folded → sanitized, and a value is additionally rune-capped
// so the cap bounds the visible output. The key list is sorted AFTER sanitizing
// (round-039 review RF-2), so the rendered order is ascending in what is actually
// printed — a raw key carrying control data cannot render out of ascending order.
// Unparseable / non-object / empty (after removing `reason`) arguments render "".
func formatToolArgs(arguments string) string {
	dec := json.NewDecoder(strings.NewReader(arguments))
	dec.UseNumber()
	var raw map[string]any
	if err := dec.Decode(&raw); err != nil || raw == nil {
		return ""
	}
	delete(raw, "reason")
	if len(raw) == 0 {
		return ""
	}
	type argKV struct{ key, val string }
	kvs := make([]argKV, 0, len(raw))
	for k, v := range raw {
		kvs = append(kvs, argKV{
			key: sanitizeControl(oneLine(k)),
			val: capRunes(sanitizeControl(oneLine(renderArgValue(v))), argValueCap),
		})
	}
	// Sort AFTER sanitizing (round-039 review RF-2), so the printed order is
	// ascending in what is actually rendered. Tie-break on the value keeps the
	// order deterministic if two distinct raw keys sanitize to the same text.
	sort.Slice(kvs, func(i, j int) bool {
		if kvs[i].key != kvs[j].key {
			return kvs[i].key < kvs[j].key
		}
		return kvs[i].val < kvs[j].val
	})
	parts := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		parts = append(parts, kv.key+": "+kv.val)
	}
	return strings.Join(parts, ", ")
}

// renderArgValue renders one decoded JSON value (FR-003 / ADR 0005 D6): a
// json.Number's raw literal (`1000000`, never `1e+06`), strings unquoted,
// booleans/`null` literal, arrays/objects compact JSON.
func renderArgValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case json.Number:
		return x.String()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

// oneLine folds newlines to spaces so a multi-line value cannot break the
// single-line log contract. It is shared by the round-034 decomposed formatters
// (`FormatToolAction` / `FormatToolResult`) and, since round 036, by
// `FormatToolReason` (which also trims). Relocated here round 036 (issue #74)
// from the now-retired `toollog.go` — its sole consumer family is in this file.
func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
}

// capRunes truncates s to at most max RUNES, keeping the first max-1 runes plus
// one U+2026 counted INSIDE the cap (FR-003/FR-004 / ADR 0005 D5). The cut is
// on a rune boundary — never a byte slice.
func capRunes(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(rs[:max-1]) + "…"
}
