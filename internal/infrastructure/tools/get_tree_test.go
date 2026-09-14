package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round 021 T032: get_tree emits a connector tree, defaults max_depth to 2, and
// lists `.git` without recursing into it (RED until Phase 4D). The default-depth
// assertion pins the reference's depth cut (entries down to max_depth levels).

func TestGetTreeShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src", "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "src", "pkg", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "pkg", "deep", "leaf.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := getTree{}.Execute(context.Background(), `{"path":"`+dir+`","reason":"r"}`)
	if err != nil {
		t.Fatalf("get_tree: %v", err)
	}
	for _, want := range []string{"src", "main.go", "pkg"} {
		if !strings.Contains(got, want) {
			t.Errorf("get_tree missing %q; got %q", want, got)
		}
	}
	if strings.Contains(got, "leaf.go") {
		t.Errorf("get_tree reached beyond the default depth to leaf.go; got %q", got)
	}
}

func TestGetTreeSkipsGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("gitconfig"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := getTree{}.Execute(context.Background(), `{"path":"`+dir+`","reason":"r"}`)
	if err != nil {
		t.Fatalf("get_tree: %v", err)
	}
	if !strings.Contains(got, ".git") {
		t.Errorf("get_tree must list .git; got %q", got)
	}
	if strings.Contains(got, "config") {
		t.Errorf("get_tree must not descend into .git; got %q", got)
	}
}
