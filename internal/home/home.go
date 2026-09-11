// Package home resolves the runtime home (TELL_ME_HOME) and manages the
// per-mode session workspace (output/<mode>/).
package home

import (
	"errors"
	"os"
	"path/filepath"
)

// Workspace describes the resolved per-mode session workspace under the runtime
// home.
type Workspace struct {
	// Path is the resolved workspace directory path.
	Path string
}

// ErrNotDirectory is returned when the workspace path exists but is not a
// directory (a run must not overwrite it).
var ErrNotDirectory = errors.New("the workspace path is not a directory")

// EnsureWorkspace creates the per-mode session workspace under home on first run
// and reuses it on later runs. It is idempotent — an existing directory is
// returned untouched (NFR-002). When the path exists as a non-directory it
// returns ErrNotDirectory (with the path populated).
func EnsureWorkspace(home, mode string) (Workspace, error) {
	path := filepath.Join(home, "output", mode)

	info, err := os.Stat(path)
	switch {
	case err == nil && !info.IsDir():
		return Workspace{Path: path}, ErrNotDirectory
	case err == nil:
		return Workspace{Path: path}, nil // reuse: do not re-create
	case !errors.Is(err, os.ErrNotExist):
		return Workspace{Path: path}, err
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return Workspace{Path: path}, err
	}
	return Workspace{Path: path}, nil
}
