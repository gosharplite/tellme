package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// maxStdinBytes bounds the standard-input read so an unbounded stream cannot
// exhaust memory (round-005 research Decision 2; matches the reference's cap).
const maxStdinBytes = 1 << 20 // 1 MiB

// MultiLineHint is the POSIX-terminal hint printed before the interactive
// multi-line read (round-012 research Decision 3). It is exported so the E2E step
// oracle can single-source the literal instead of duplicating it (round-012
// review TD1). It carries no `tellme: ` prefix, so the frozen class-phrase
// vocabulary is unchanged.
const MultiLineHint = "[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]"

// TUIHint is the POSIX-terminal announcement printed to the diagnostic stream
// (stderr) when the interactive TUI prompt (round 015, `-i`/`USE_TUI_PROMPT`)
// engages. It is exported so the E2E step oracle can single-source the literal
// (specs/truth/features/cli/chat/dsl.md — `the interactive prompt is shown`)
// instead of duplicating it, so a product-side text change cannot make the guard
// vacuous (mirroring MultiLineHint). It carries no `tellme: ` prefix, so the
// frozen class-phrase vocabulary is unchanged (still 11).
const TUIHint = "[Interactive prompt. Type a prompt; Tab accepts a suggestion, Ctrl+S sends, Esc cancels]"

// defaultIsTerminal reports whether v is connected to a terminal.
//
// It performs a real isatty query (golang.org/x/term.IsTerminal), NOT a bare
// os.ModeCharDevice test: the latter is true for *any* character device —
// /dev/null, /dev/zero, /dev/random — so promoting it to a behaviour gate made
// the interactive reader engage on a redirected null device and masked a
// missing-configuration failure as exit 0 (round-012 review BLOCKER B1). A real
// isatty is false for those devices and true only for an actual terminal.
//
// Only *os.File streams can be a terminal; any other reader/writer (e.g. an
// injected fake) is treated as non-terminal. Adopting x/term is a deliberate
// reversal of round-012 research Decision 1, recorded in
// docs/decisions/0003-terminal-detection-isatty.md; the module is already in the
// graph transitively (via glamour), so no new module is downloaded.
func defaultIsTerminal(v any) bool {
	f, ok := v.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// terminalColumns reports the terminal width in columns for an *os.File stream,
// or 0 when v is not a terminal (round 025). It mirrors defaultIsTerminal's
// *os.File requirement and backs the spinner's row-aware clear.
func terminalColumns(v any) int {
	f, ok := v.(*os.File)
	if !ok {
		return 0
	}
	cols, _, err := term.GetSize(int(f.Fd()))
	if err != nil || cols <= 0 {
		return 0
	}
	return cols
}

// combinePrompt assembles the prompt from the positional argument(s) and the
// piped standard-input content: the arguments joined by single spaces, then a
// newline and the piped content when it is non-empty, then trimmed (round-005
// research Decision 3; matches the reference's main chat path `args + "\n" +
// stdin`).
func combinePrompt(args []string, piped []byte) string {
	prompt := strings.Join(args, " ")
	if len(piped) > 0 {
		prompt = prompt + "\n" + string(piped)
	}
	return strings.TrimSpace(prompt)
}

// readPipedStdin reads standard input bounded by maxStdinBytes; content beyond
// the cap is not read (round-005 research Decision 2).
func readPipedStdin(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, maxStdinBytes))
}

// resolvePrompt decides the prompt for a prompt-turn dispatch: it reads piped
// standard input only when stdin is not a terminal, then combines it with the
// positional argument(s) (round-005 research Decisions 3 & 6). An empty result
// means no prompt was supplied, so the caller falls through to boot.
func resolvePrompt(args []string, stdin io.Reader, isTTY func(any) bool) (string, error) {
	var piped []byte
	if !isTTY(stdin) {
		data, err := readPipedStdin(stdin)
		if err != nil {
			return "", err
		}
		piped = data
	}
	return combinePrompt(args, piped), nil
}

// readInteractivePrompt prints the multi-line hint to the diagnostic stream and
// reads the prompt from a terminal to EOF (Ctrl+D), bounded by maxStdinBytes. It
// reports ok=false when the read is cancelled (ctx) — the caller then sends no
// request (round-012 research Decisions 2 & 4). POSIX-only; there is no Windows
// variant (round-012 research Decision 5).
//
// Lifetime bound (round-012 review TD2): the read runs in a goroutine that the
// owning process's exit bounds. On cancel (ctx) the goroutine stays blocked on
// Read until that exit — deliberate for tellme's one-shot CLI, and recorded here
// because any future long-lived reuse (a multi-turn REPL / TUI prompt mode) would
// turn it into an accumulating goroutine leak.
func readInteractivePrompt(ctx context.Context, stdin io.Reader, stderr io.Writer) (string, bool) {
	_, _ = fmt.Fprintln(stderr, MultiLineHint)
	done := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(io.LimitReader(stdin, maxStdinBytes))
		done <- data
	}()
	select {
	case data := <-done:
		return strings.TrimSpace(string(data)), true
	case <-ctx.Done():
		return "", false
	}
}
