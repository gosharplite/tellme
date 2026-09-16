package di

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeFakeGh writes an executable `gh` shim into a temp dir and points PATH at
// it, so the production resolver can be exercised without a real `gh`.
func writeFakeGh(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

// T032 [UNIT] — the production bounded `gh` resolver trims the token.
func TestNewGhTokenResolver_TrimsToken(t *testing.T) {
	writeFakeGh(t, `echo "tok-123"`)
	tok, err := NewGhTokenResolver(2 * time.Second)(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "tok-123" {
		t.Fatalf("token = %q, want tok-123", tok)
	}
}

// T032 [UNIT] — an unresponsive `gh` is bounded by the fast-fail deadline and
// returns an error (the caller then warns and falls back to anonymous).
func TestNewGhTokenResolver_Bounded(t *testing.T) {
	writeFakeGh(t, `sleep 30`)
	start := time.Now()
	_, err := NewGhTokenResolver(200 * time.Millisecond)(context.Background())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a hanging gh must return an error")
	}
	if elapsed > 3*time.Second {
		t.Fatalf("the resolver stalled for %v (bound 200ms)", elapsed)
	}
}

// T032 [UNIT] — a missing `gh` returns an error (no run failure).
func TestNewGhTokenResolver_MissingGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no gh on PATH
	if _, err := NewGhTokenResolver(time.Second)(context.Background()); err == nil {
		t.Fatal("a missing gh must return an error")
	}
}
