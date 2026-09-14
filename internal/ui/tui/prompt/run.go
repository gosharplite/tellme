package prompt

import (
	"context"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// Run drives the interactive prompt through the Bubble Tea runtime and returns
// the composed prompt text plus whether it was submitted (ok). It is bound to the
// injected streams; the caller passes the diagnostic stream as out so stdout
// stays byte-exact (round-015 PR #38 review BLOCKER). Driving the real Bubble Tea
// runtime keeps the surface faithful (research Decision 1); the model is also
// unit-testable directly (research Decision 6).
func Run(ctx context.Context, in io.Reader, out io.Writer, src Source, dash Dashboard) (string, bool, error) {
	m := New(in, out, src, dash)
	rin, rout := m.Streams()
	p := tea.NewProgram(m, tea.WithInput(rin), tea.WithOutput(rout), tea.WithContext(ctx))
	final, err := p.Run()
	if err != nil {
		return "", false, err
	}
	fm, ok := final.(*Model)
	if !ok || !fm.WasSubmitted() {
		return "", false, nil
	}
	return fm.Submitted(), true, nil
}
