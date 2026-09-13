package prompt

import "github.com/charmbracelet/bubbles/textarea"

// editor wraps the bubbles multi-line textarea — the multi-line input surface
// (Enter inserts a newline; submit is a separate keybinding handled by the
// model). Behaviour lands with the Feature phase (round-015 T032).
type editor struct {
	ta textarea.Model
}

// newEditor builds the multi-line editor skeleton.
func newEditor() editor { return editor{ta: textarea.New()} }

// value is the editor's current text. Skeleton accessor used by the model.
func (e *editor) value() string { return e.ta.Value() }
