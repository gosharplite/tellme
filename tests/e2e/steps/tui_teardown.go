package steps

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Round-023 teardown helpers. Kept in ONE file (Zero Shared Edits), mirroring
// tui_chrome.go / turn_chrome.go. The inline TUI emits cursor moves and line
// erases; a raw substring check over the accumulated buffer would still see the
// editor frame written on earlier draw cycles, so the teardown witness reduces
// the stream to its FINAL screen first.

// reduceTerminal replays a captured byte stream through a minimal line-oriented
// terminal model and returns the resulting screen rows: it handles `\n`, `\r`,
// CSI cursor moves (`A`/`B`/`C`/`D`), line erase (`K`/`2K`), erase-below
// (`J`/`2J`), and cursor home (`H`/`f`); SGR (`m`) and unknown sequences are
// ignored. The result is what the operator would finally see.
func reduceTerminal(out string) []string {
	s := &screen{}
	for i := 0; i < len(out); {
		switch c := out[i]; {
		case c == '\n':
			s.row++
			s.col = 0
			s.ensure(s.row)
			i++
		case c == '\r':
			s.col = 0
			i++
		case c == 0x1b && i+1 < len(out) && out[i+1] == '[':
			j := i + 2
			for j < len(out) && (out[j] < '@' || out[j] > '~') {
				j++
			}
			if j >= len(out) {
				i = len(out)
				break
			}
			s.csi(out[i+2:j], out[j])
			i = j + 1
		default:
			r, size := utf8.DecodeRuneInString(out[i:])
			s.put(r)
			i += size
		}
	}
	rows := make([]string, 0, len(s.lines))
	for _, ln := range s.lines {
		rows = append(rows, string(ln))
	}
	return rows
}

// screen is the minimal terminal state: rune rows plus a cursor.
type screen struct {
	lines [][]rune
	row   int
	col   int
}

func (s *screen) ensure(row int) {
	for len(s.lines) <= row {
		s.lines = append(s.lines, []rune{})
	}
}

func (s *screen) put(r rune) {
	s.ensure(s.row)
	ln := s.lines[s.row]
	for len(ln) <= s.col {
		ln = append(ln, ' ')
	}
	ln[s.col] = r
	s.lines[s.row] = ln
	s.col++
}

// csi dispatches a CSI sequence (params is the raw parameter bytes, final the
// final byte). Relative moves clamp at 0; unhandled finals are ignored.
func (s *screen) csi(params string, final byte) {
	n := csiParam(params)
	switch final {
	case 'A':
		s.row = clamp0(s.row - n)
	case 'B':
		s.row += n
		s.ensure(s.row)
	case 'C':
		s.col += n
	case 'D':
		s.col = clamp0(s.col - n)
	case 'K':
		s.eraseLine(params)
	case 'J':
		s.eraseBelow(params)
	case 'H', 'f':
		s.row, s.col = 0, 0
	}
}

// eraseLine blanks part of the cursor's line: `2K` the whole line, `1K` up to the
// cursor, otherwise from the cursor to the end.
func (s *screen) eraseLine(params string) {
	s.ensure(s.row)
	ln := s.lines[s.row]
	switch params {
	case "1":
		for i := 0; i <= s.col && i < len(ln); i++ {
			ln[i] = ' '
		}
	case "2":
		for i := range ln {
			ln[i] = ' '
		}
	default:
		for i := s.col; i < len(ln); i++ {
			ln[i] = ' '
		}
	}
	s.lines[s.row] = ln
}

// eraseBelow blanks the cursor's line (from the cursor) and every line below;
// `2J` blanks the whole screen.
func (s *screen) eraseBelow(params string) {
	s.ensure(s.row)
	if params == "2" {
		for i := range s.lines {
			s.lines[i] = []rune{}
		}
		return
	}
	ln := s.lines[s.row]
	for i := s.col; i < len(ln); i++ {
		ln[i] = ' '
	}
	s.lines[s.row] = ln
	for i := s.row + 1; i < len(s.lines); i++ {
		s.lines[i] = []rune{}
	}
}

// csiParam parses a CSI count (an absent/invalid/zero count is 1; a `?`-prefixed
// private parameter — e.g. cursor show/hide — is treated as 1 and ignored by the
// caller).
func csiParam(params string) int {
	if params == "" || params[0] == '?' {
		return 1
	}
	v, err := strconv.Atoi(params)
	if err != nil || v <= 0 {
		return 1
	}
	return v
}

func clamp0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// terminalHasBorder reports whether the reduced screen still shows the editor
// frame (a `┌`/`└` border rune).
func terminalHasBorder(rows []string) bool {
	for _, ln := range rows {
		if strings.ContainsAny(ln, "┌└") {
			return true
		}
	}
	return false
}
