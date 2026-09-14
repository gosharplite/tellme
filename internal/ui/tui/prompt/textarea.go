package prompt

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// editorPlaceholder is the reference placeholder — it carries the keybinding
// hints, so the `-i` surface needs no separate status line (round-016 strict
// parity with tell-me-go).
const editorPlaceholder = "Type your message here... (Alt+Enter or Ctrl+S to submit, Esc to abort)"

// editorHeight is the reference fixed editor height (rows).
const editorHeight = 10

// defaultEditorWidth is the editor width before a WindowSizeMsg arrives (the
// reference default).
const defaultEditorWidth = 80

// textareaStyle frames the editor with the reference border (NormalBorder, fg 240).
var textareaStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), true).
	BorderForeground(lipgloss.Color("240"))

// editor wraps the bubbles multi-line textarea with the reference chrome. It
// delegates every editing key to the underlying textarea (round-016 architect D2:
// backspace, navigation, word movement all work) and exposes only the value +
// width seams the model needs.
type editor struct {
	ta textarea.Model
}

// newEditor builds the reference editor: bordered, fixed height, placeholder, no
// line numbers.
func newEditor() editor {
	ta := textarea.New()
	ta.Placeholder = editorPlaceholder
	ta.Focus()
	ta.SetHeight(editorHeight)
	ta.SetWidth(defaultEditorWidth)
	ta.ShowLineNumbers = false
	return editor{ta: ta}
}

// value is the editor's current text.
func (e editor) value() string { return e.ta.Value() }

// setValue replaces the editor's text.
func (e *editor) setValue(v string) { e.ta.SetValue(v) }

// setWidth sets the editor width (clamped to at least one column).
func (e *editor) setWidth(w int) {
	if w < 1 {
		w = 1
	}
	e.ta.SetWidth(w)
}

// update delegates a message to the bubbles textarea (round-016 architect D2).
func (e *editor) update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	e.ta, cmd = e.ta.Update(msg)
	return cmd
}

// view renders the bordered editor.
func (e editor) view() string { return textareaStyle.Render(e.ta.View()) }
