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
	"io"

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
}

// New builds the prompt model over the injected streams and suggestion source.
// The caller binds out to env.stderr (round-015 PR #38 review BLOCKER).
func New(in io.Reader, out io.Writer, src Source, dash Dashboard) *Model {
	return &Model{in: in, out: out, src: src, dash: dash, ed: newEditor(), sug: newSuggester()}
}

// baseStyle is the skeleton root style (the terminal visual direction lands with
// the Feature phase, round-015 T032).
var baseStyle = lipgloss.NewStyle()

// Init implements tea.Model. Skeleton: no initial command yet.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model. Skeleton: no key handling yet (lands with T032).
func (m *Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// View implements tea.Model. Skeleton render; the composed frame lands with T032.
func (m *Model) View() string { return baseStyle.Render(m.skeleton()) }

// skeleton references the injected state so the landing package is complete and
// free of unused symbols; it intentionally renders nothing yet — the composed
// frame (dashboard header, editor, suggestion list) lands with T032.
func (m *Model) skeleton() string {
	_ = m.in
	_ = m.out
	_ = m.src
	_ = m.dash
	return m.ed.value() + m.sug.top()
}
