package prompt

import (
	"context"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Run drives the interactive prompt through the Bubble Tea runtime and returns
// the composed prompt text plus whether it was submitted (ok). It is bound to the
// injected streams; the caller passes the diagnostic stream as out so stdout
// stays byte-exact (round-015 PR #38 review BLOCKER). Driving the real Bubble Tea
// runtime keeps the surface faithful; the model is also unit-testable directly.
// debounce is the suggestion-refresh debounce (<=0 refreshes synchronously — the
// TELL_ME_TUI_DEBOUNCE=0 hermetic seam).
func Run(ctx context.Context, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error) {
	m := New(ctx, in, out, src)
	defer m.Destroy() // cancel any in-flight fetch on exit (round-016 architect TD1)
	m.SetDebounce(debounce)
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
