package history

import "io"

// TurnsLogStore persists and reads a session's turn log — the rendered turn
// chrome the diagnostic stream showed while the session ran (round 053, closing
// #103; ADR 0022). It is the per-mode plain-text file
// `$TELL_ME_HOME/output/<mode>/turns.log`: read by the `-t` offline command and
// archived by `--new` (alongside `history.jsonl` / `tokens.log`). It is
// best-effort — a write failure never fails a turn.
type TurnsLogStore interface {
	// Writer returns an append handle for the active turn log. The caller must
	// Close it when the run ends.
	Writer() (io.WriteCloser, error)
	// Read returns the active turn log's contents (an empty string when the file
	// is absent — not an error).
	Read() (string, error)
	// Archive moves the active turn log aside (the `--new` rotation).
	Archive() error
}
