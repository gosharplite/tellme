package ui

import "strings"

// oneLine folds newlines to spaces so a multi-line value cannot break the
// single-line log contract. It is shared by the round-034 decomposed formatters
// (`FormatToolAction` / `FormatToolResult`).
//
// (Round 034 T031: the round-022 `FormatToolLog` single-line formatter was
// removed with the rest of the round-022 rendering — only this helper survives.)
func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
}
