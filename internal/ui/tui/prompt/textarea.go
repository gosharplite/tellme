package prompt

import "github.com/charmbracelet/bubbles/textarea"

// editor wraps the bubbles multi-line textarea — the multi-line input surface
// (Enter inserts a newline; submit is a separate keybinding handled by the
// model).
type editor struct {
	ta textarea.Model
}

// newEditor builds the multi-line editor.
func newEditor() editor { return editor{ta: textarea.New()} }

// value is the editor's current text.
func (e *editor) value() string { return e.ta.Value() }

// insert appends s at the cursor.
func (e *editor) insert(s string) { e.ta.InsertString(s) }
