package home

import (
	"os"
	"path/filepath"
	"testing"
)

// T014 — unit tests for EnsureWorkspace idempotency (research Decision 1 & 3).
func TestEnsureWorkspaceCreatesThenReuses(t *testing.T) {
	root := t.TempDir()

	ws, err := EnsureWorkspace(root, "butler")
	if err != nil {
		t.Fatalf("EnsureWorkspace (create) error = %v", err)
	}
	want := filepath.Join(root, "output", "butler")
	if ws.Path != want {
		t.Fatalf("Path = %q, want %q", ws.Path, want)
	}
	info, err := os.Stat(want)
	if err != nil || !info.IsDir() {
		t.Fatalf("workspace not created as a directory: err=%v", err)
	}

	// Idempotent reuse: a sentinel placed inside must survive a second call.
	sentinel := filepath.Join(want, "state.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	ws2, err := EnsureWorkspace(root, "butler")
	if err != nil {
		t.Fatalf("EnsureWorkspace (reuse) error = %v", err)
	}
	if ws2.Path != want {
		t.Fatalf("reuse Path = %q, want %q", ws2.Path, want)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel not preserved on reuse: %v", err)
	}
}

func TestEnsureWorkspaceNotADirectory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "output", "butler")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("setup dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("file"), 0o644); err != nil {
		t.Fatalf("setup file: %v", err)
	}

	ws, err := EnsureWorkspace(root, "butler")
	if err != ErrNotDirectory {
		t.Fatalf("EnsureWorkspace error = %v, want ErrNotDirectory", err)
	}
	if ws.Path != path {
		t.Fatalf("Path = %q, want %q", ws.Path, path)
	}
}
