// Package prompt implements tellme's opt-in interactive TUI prompt (round 015,
// `-i`/`USE_TUI_PROMPT`). It is a Bubble Tea model: a multi-line editor, a live
// suggestion list with a selection cursor, and a session dashboard header.
//
// The package is deliberately stream-injected: the caller binds output to the
// diagnostic stream (env.stderr) so `stdout` stays byte-exact for piping
// (round-015 PR #38 review BLOCKER), and input comes from the terminal. That
// makes the model unit-testable with scripted keys and no real pty (research
// Decision 6).
package prompt

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Source yields the candidate suggestions for the current query. It is the
// injection seam: production wires the multi-source engine
// (internal/app/suggestions); unit tests inject a fake (round-015 T027/T029).
type Source interface {
	Suggest(query string) []string
}

// Dashboard carries the header values shown above the editor: the active
// provider/model, the payload token usage vs the budget, and the turn count
// (round-015 research Decision 5 — reused state, no new accounting).
type Dashboard struct {
	Provider string
	Tokens   int
	Budget   int
	Turns    int
}

// Model is the Bubble Tea model for the interactive prompt. Input and output are
// injected; the caller binds output to the diagnostic stream so stdout is
// untouched.
type Model struct {
	in   io.Reader
	out  io.Writer
	src  Source
	dash Dashboard
	ed   editor
	sug  suggester

	submitted bool
	aborted   bool
}

// New builds the prompt model over the injected streams and suggestion source.
// The caller binds out to env.stderr (round-015 PR #38 review BLOCKER). The
// initial suggestions are seeded for the empty query.
func New(in io.Reader, out io.Writer, src Source, dash Dashboard) *Model {
	m := &Model{in: in, out: out, src: src, dash: dash, ed: newEditor(), sug: newSuggester()}
	m.refresh()
	return m
}

// Streams returns the injected input and output — the caller binds them into the
// Bubble Tea program (Run).
func (m *Model) Streams() (io.Reader, io.Writer) { return m.in, m.out }

// baseStyle is the root style (terminal visual direction from ui/ui-plan.md).
var baseStyle = lipgloss.NewStyle()

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model: the terminal keybindings — Ctrl+S/Alt+Enter
// submit, Enter inserts a newline, Tab/Shift+Tab cycle the suggestions,
// Esc/Ctrl+C abort, and typing refreshes the suggestions.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.aborted = true
		return m, tea.Quit
	case tea.KeyCtrlS:
		m.submitted = strings.TrimSpace(m.ed.value()) != ""
		return m, tea.Quit
	case tea.KeyEnter:
		if key.Alt {
			m.submitted = strings.TrimSpace(m.ed.value()) != ""
			return m, tea.Quit
		}
		m.ed.insert("\n")
		return m, nil
	case tea.KeyTab:
		m.sug.cycle(1)
		return m, nil
	case tea.KeyShiftTab:
		m.sug.cycle(-1)
		return m, nil
	case tea.KeyRunes:
		m.ed.insert(string(key.Runes))
		m.refresh()
		return m, nil
	case tea.KeySpace:
		// Bubble Tea reports a space as its own key type, not a rune.
		m.ed.insert(" ")
		m.refresh()
		return m, nil
	}
	return m, nil
}

// View implements tea.Model: the dashboard header, the editor, and the
// suggestion list (the selection cursor is `>`), styled for the terminal.
func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(m.dashboardLine())
	b.WriteString("\n")
	b.WriteString(m.ed.value())
	b.WriteString("\n")
	for i, it := range m.sug.items {
		if i == m.sug.cursor {
			b.WriteString("> " + it + "\n")
		} else {
			b.WriteString("  " + it + "\n")
		}
	}
	return baseStyle.Render(b.String())
}

// dashboardLine renders the session dashboard header (provider + token usage +
// turn count).
func (m *Model) dashboardLine() string {
	return fmt.Sprintf("provider: %s | tokens: %d/%d | turns: %d",
		m.dash.Provider, m.dash.Tokens, m.dash.Budget, m.dash.Turns)
}

// refresh recomputes the suggestions for the current editor content. An empty
// query yields the seeds from the source.
func (m *Model) refresh() {
	if m.src == nil {
		m.sug.set(nil)
		return
	}
	m.sug.set(m.src.Suggest(m.ed.value()))
}

// Submitted is the composed (trimmed) prompt; valid when WasSubmitted.
func (m *Model) Submitted() string { return strings.TrimSpace(m.ed.value()) }

// WasSubmitted reports whether the operator submitted (Ctrl+S / Alt+Enter).
func (m *Model) WasSubmitted() bool { return m.submitted }

// WasAborted reports whether the operator aborted (Ctrl+C / Esc).
func (m *Model) WasAborted() bool { return m.aborted }
