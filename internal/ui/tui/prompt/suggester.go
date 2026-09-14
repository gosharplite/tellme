package prompt

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// suggesterHeader is the reference suggestion-list header.
const suggesterHeader = "Suggestions:"

var (
	suggesterStyle  = lipgloss.NewStyle().Padding(0, 1)
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Background(lipgloss.Color("235"))
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// suggester holds the suggestion list and the selection cursor. The cursor
// defaults to 0 so the first item is the current choice when the list is
// populated (round-016 architect D4: aligns with the Gherkin "marks one
// suggestion as the current choice" and ui/screens/entry.txt).
type suggester struct {
	items  []string
	cursor int
}

// newSuggester builds an empty suggestion list.
func newSuggester() suggester { return suggester{} }

// set replaces the items and keeps the cursor in range (default 0).
func (s *suggester) set(items []string) {
	s.items = items
	if len(items) == 0 {
		s.cursor = 0
		return
	}
	if s.cursor < 0 || s.cursor >= len(items) {
		s.cursor = 0
	}
}

// cycle moves the selection cursor by delta (wrapping).
func (s *suggester) cycle(delta int) {
	if len(s.items) == 0 {
		return
	}
	s.cursor = (s.cursor + delta + len(s.items)) % len(s.items)
}

// selected returns the current choice, or "" when the list is empty.
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
