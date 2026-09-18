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
	// Reader returns a streaming reader over the active turn log (symmetric with
	// Writer; a missing file reads as empty — not an error), so the `-t` reader
	// never materialises the whole log (RF-53-2).
	Reader() (io.ReadCloser, error)
	// Archive moves the active turn log aside (the `--new` rotation).
	Archive() error
}
