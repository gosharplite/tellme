package history

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// The per-mode turn-log file names (round 053; ADR 0022). The active file is
// plain text (the rendered turn chrome); `--new` appends it to the archive and
// removes the active file, mirroring the history/usage `--new` rotation.
const (
	turnsLogActiveFileName  = "turns.log"
	turnsLogArchiveFileName = "turns.archive.log"
)

// turnsLogStore is the file-backed TurnsLogStore rooted at the session
// workspace. It is best-effort at the call sites (a write failure is ignored).
type turnsLogStore struct {
	workspace string
}

var _ domainhistory.TurnsLogStore = (*turnsLogStore)(nil)

// NewTurnsLogStore returns a file-backed TurnsLogStore rooted at the workspace.
func NewTurnsLogStore(workspace string) *turnsLogStore {
	return &turnsLogStore{workspace: workspace}
}

func (s *turnsLogStore) activePath() string {
	return filepath.Join(s.workspace, turnsLogActiveFileName)
}
func (s *turnsLogStore) archivePath() string {
	return filepath.Join(s.workspace, turnsLogArchiveFileName)
}

// Writer opens the active turn log for appending (creating it if absent).
func (s *turnsLogStore) Writer() (io.WriteCloser, error) {
	return os.OpenFile(s.activePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

// Reader opens the active turn log for streaming; a missing file reads as empty
// (not an error), so the `-t` reader tolerates a session with no turn log.
func (s *turnsLogStore) Reader() (io.ReadCloser, error) {
	f, err := os.Open(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		return nil, err
	}
	return f, nil
}

// Archive appends the active turn log onto the archive and removes the active
// file (a missing active file is a no-op).
func (s *turnsLogStore) Archive() error {
	data, err := os.ReadFile(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	f, err := os.OpenFile(s.archivePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.activePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
