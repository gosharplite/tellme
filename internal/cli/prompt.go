package cli

import (
	"io"
	"os"
	"strings"
)

// maxStdinBytes bounds the standard-input read so an unbounded stream cannot
// exhaust memory (round-005 research Decision 2; matches the reference's cap).
const maxStdinBytes = 1 << 20 // 1 MiB

// defaultIsTerminal reports whether v is connected to a terminal, using the
// dependency-free standard-library character-device check (round-005 research
// Decision 1 — no golang.org/x/term). Only *os.File streams can be a terminal;
// any other reader/writer (e.g. an injected fake) is treated as non-terminal.
func defaultIsTerminal(v any) bool {
	f, ok := v.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
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
