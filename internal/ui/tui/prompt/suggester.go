package prompt

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// suggesterHeader is the reference suggestion-list header.
const suggesterHeader = "Suggestions:"

// noChoice is the no-selection cursor sentinel: no suggestion row is the current
// choice. It mirrors the reference's `suggester{Index: -1}` (round 037) — the list
// opens unselected and resets to unselected on every refresh, so an empty editor
// never shows a "chosen" hint. `cycle` uses plain modulo arithmetic, so from
// noChoice the first Tab lands on index 0 (the first suggestion) exactly as the
// reference does.
const noChoice = -1

var (
	suggesterStyle  = lipgloss.NewStyle().Padding(0, 1)
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Background(lipgloss.Color("235"))
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// suggester holds the suggestion list and the selection cursor. The cursor
// defaults to noChoice (no suggestion selected) — round 037, aligning to the
// reference's `suggester{Index: -1}` (which supersedes round 016's "the first item
// is the current choice"); a fresh list, and every refreshed list, carry no
// highlight until the operator cycles.
type suggester struct {
	items  []string
	cursor int
}

// newSuggester builds an empty suggestion list with no selection.
func newSuggester() suggester { return suggester{cursor: noChoice} }

// set replaces the items and resets the selection to noChoice (mirroring the
// reference's `Update(suggestions, -1)`), so a refresh clears any prior highlight.
func (s *suggester) set(items []string) {
	s.items = items
	s.cursor = noChoice
}

// cycle moves the selection cursor by delta (wrapping). With plain modulo
// arithmetic, from noChoice the first cycle(+1) lands on index 0 (the first
// suggestion) — the reference's own arithmetic (round 037).
func (s *suggester) cycle(delta int) {
	if len(s.items) == 0 {
		return
	}
	s.cursor = (s.cursor + delta + len(s.items)) % len(s.items)
}

// selected returns the current choice, or "" when nothing is selected (an empty
// list or the noChoice sentinel).
func (s suggester) selected() string {
	if len(s.items) == 0 || s.cursor < 0 || s.cursor >= len(s.items) {
		return ""
	}
	return s.items[s.cursor]
}

// view renders the header + the styled rows (the `>` cursor marks the choice).
func (s suggester) view() string {
	if len(s.items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(suggesterHeader + "\n")
	for i, it := range s.items {
		prefix, style := "  ", unselectedStyle
		if i == s.cursor {
			prefix, style = "> ", selectedStyle
		}
		b.WriteString(prefix + style.Render(it) + "\n")
	}
	return suggesterStyle.Render(b.String())
}
