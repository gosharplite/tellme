// Package prompt implements tellme's opt-in interactive TUI prompt (round 015,
// `-i`/`USE_TUI_PROMPT`). Round 016 aligns it to tell-me-go's surface (strict
// parity): a bordered multi-line editor above a styled suggestion list, with the
// keybinding hints in the placeholder and NO session metrics header.
//
// The package is deliberately stream-injected: the caller binds output to the
// diagnostic stream (env.stderr) so `stdout` stays byte-exact for piping
// (round-015 PR #38 review BLOCKER), and input comes from the terminal. That
// makes the model unit-testable with scripted keys and no real pty.
package prompt

import (
	"context"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DefaultDebounceDuration is the suggestion-refresh debounce, matching the
// reference (round-016 research Decision 2).
const DefaultDebounceDuration = 100 * time.Millisecond

// maxSuggestionLines is the drop threshold: an entry with more than three lines
// is not offered (round-016 FR-006).
const maxSuggestionLines = 3

// modelStyle pads the whole prompt block (reference root style).
var modelStyle = lipgloss.NewStyle().Padding(1, 1)

// Source yields the candidate suggestions for the current query. It carries ctx
// so a superseded fetch can be cancelled (round-016 architect D1). It is the
// injection seam: production wires the multi-source engine
// (internal/app/suggestions); unit tests inject a fake.
type Source interface {
	Suggest(ctx context.Context, query string) []string
}

// Model is the Bubble Tea model for the interactive prompt. Input and output are
// injected; the caller binds output to the diagnostic stream so stdout is
// untouched.
type Model struct {
	in  io.Reader
	out io.Writer
	src Source
	ed  editor
	sug suggester

	debounce time.Duration
	parent   context.Context
	ctx      context.Context
	cancel   context.CancelFunc

	submitted bool
	aborted   bool
}

// debounceMsg is delivered after the debounce delay to (re)compute suggestions.
type debounceMsg struct{ value string }

// New builds the prompt model over the injected streams and suggestion source.
// parent is the run context: the model derives its cancelable fetch contexts
// from it, so cancelling the run (or Destroy) aborts an in-flight fetch
// (round-016 architect TD1). The caller binds out to env.stderr (round-015 PR #38
// review BLOCKER). The initial suggestions are seeded for the empty query.
func New(parent context.Context, in io.Reader, out io.Writer, src Source) *Model {
	if parent == nil {
		parent = context.Background()
	}
	m := &Model{in: in, out: out, src: src, ed: newEditor(), sug: newSuggester(), debounce: DefaultDebounceDuration}
	m.parent = parent
	m.ctx, m.cancel = context.WithCancel(m.parent)
	m.computeSuggestions()
	return m
}

// Streams returns the injected input and output — the caller binds them into the
// Bubble Tea program (Run).
func (m *Model) Streams() (io.Reader, io.Writer) { return m.in, m.out }

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Destroy cancels any in-flight suggestion fetch (round-016 architect D1).
func (m *Model) Destroy() {
	if m.cancel != nil {
		m.cancel()
	}
}

// Update implements tea.Model. Command keys are intercepted here; every editing
// key is delegated to the textarea (round-016 architect D2), and a value change
// schedules a debounced, cancelable refresh.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.ed.setWidth(msg.Width - 4)
		return m, nil
	case debounceMsg:
		if msg.value == m.ed.value() {
			m.computeSuggestions()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Type == tea.KeyCtrlC || key.Type == tea.KeyEsc:
		m.aborted = true
		return m, tea.Quit
	case key.Type == tea.KeyCtrlS:
		return m, m.trySubmit()
	case key.Type == tea.KeyEnter && key.Alt:
		return m, m.trySubmit()
	case key.Type == tea.KeyTab:
		m.accept(1)
		return m, nil
	case key.Type == tea.KeyShiftTab:
		m.accept(-1)
		return m, nil
	}
	// Delegate every editing key to the bubbles textarea (D2). Enter inserts a
	// newline; backspace/arrows/word movement all work.
	old := m.ed.value()
	cmd := m.ed.update(key)
	if m.ed.value() != old {
		m.cancel() // abort the in-flight fetch (D1) before scheduling the next
		m.ctx, m.cancel = context.WithCancel(m.parent)
		if m.debounce <= 0 {
			m.computeSuggestions() // synchronous refresh (the hermetic E2E seam)
			return m, cmd
		}
		return m, tea.Batch(cmd, m.scheduleDebounce())
	}
	return m, cmd
}

// SetDebounce overrides the refresh debounce. A non-positive value refreshes
// synchronously on each change — the hermetic E2E seam (TELL_ME_TUI_DEBOUNCE=0),
// so the scripted keys observe suggestions without a timing pause (round-016).
func (m *Model) SetDebounce(d time.Duration) { m.debounce = d }

// trySubmit submits the composed prompt when the editor holds non-whitespace
// text. Round 038 (issue #76): an EMPTY (or whitespace-only) submit is a no-op —
// it returns no command so the model keeps running (the prompt stays open),
// mirroring the reference's submit() (which returns false and keeps the model
// running). A real submit marks the model submitted and returns tea.Quit.
func (m *Model) trySubmit() tea.Cmd {
	if strings.TrimSpace(m.ed.value()) == "" {
		return nil // empty submit: stay in the prompt
	}
	m.submitted = true
	return tea.Quit
}

// accept inserts the current choice into the editor: it replaces only the last
// token when the line is multi-word and the suggestion is a single token,
// otherwise the whole line (round-016 FR-007 / last-token heuristic).
func (m *Model) accept(delta int) {
	if len(m.sug.items) == 0 {
		return
	}
	m.sug.cycle(delta)
	sel := m.sug.selected()
	if sel == "" {
		return
	}
	cur := m.ed.value()
	if i := strings.LastIndex(cur, " "); i != -1 && !strings.Contains(sel, " ") {
		m.ed.setValue(cur[:i+1] + sel)
	} else {
		m.ed.setValue(sel)
	}
}

// scheduleDebounce returns a command that fires debounceMsg after the delay.
func (m *Model) scheduleDebounce() tea.Cmd {
	return tea.Tick(m.debounce, func(time.Time) tea.Msg { return debounceMsg{value: m.ed.value()} })
}

// computeSuggestions refreshes the list for the current editor value using the
// cancelable fetch context; entries spanning more than three lines are dropped
// (FR-006).
func (m *Model) computeSuggestions() {
	if m.src == nil {
		m.sug.set(nil)
		return
	}
	raw := m.src.Suggest(m.ctx, m.ed.value())
	filtered := make([]string, 0, len(raw))
	for _, s := range raw {
		if strings.Count(strings.TrimSpace(s), "\n") < maxSuggestionLines {
			filtered = append(filtered, s)
		}
	}
	m.sug.set(filtered)
}

// View implements tea.Model: the bordered editor above the styled suggestion
// list. There is NO dashboard header and NO status line (round-016 strict
// parity with tell-me-go). Round 023: once the operator submits
// (Ctrl+S / Alt+Enter) or aborts (Esc / Ctrl+C) the frame is CLEARED (empty),
// matching tell-me-go's View() — the editor must not linger on screen while the
// turn surface takes over. Earlier frames remain in the raw captured stream, so
// the round-016 chrome assertions still hold; the E2E teardown witness reads the
// cleared final frame via terminal reduction (the model-level clear is the
// authoritative pin — round-023 T007).
func (m *Model) View() string {
	if m.submitted || m.aborted {
		return ""
	}
	return modelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.ed.view(), "\n", m.sug.view()))
}

// Submitted is the composed (trimmed) prompt; valid when WasSubmitted.
func (m *Model) Submitted() string { return strings.TrimSpace(m.ed.value()) }

// WasSubmitted reports whether the operator submitted (Ctrl+S / Alt+Enter).
func (m *Model) WasSubmitted() bool { return m.submitted }

// WasAborted reports whether the operator aborted (Esc / Ctrl+C).
func (m *Model) WasAborted() bool { return m.aborted }
