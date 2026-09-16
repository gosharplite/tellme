package di

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeFakeGh writes an executable `gh` shim into a temp dir and points PATH at
// it, so the production resolver can be exercised without a real `gh`. NOTE: the
// shim must be self-contained — a bare `sleep` would NOT be found under the
// shim-only PATH (exit 127) — so a hanging shim restores a real PATH before
// sleeping (round-032 implementation-review N1).
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
//
// N1 fold: the shim MUST be self-contained. PATH points only at the shim dir, so
// a bare `sleep` is not found (exit 127) and the shim would fail INSTANTLY —
// making the test pass vacuously (indistinguishable from "bounded"). Restoring a
// real PATH makes the shim genuinely sleep past the 200 ms bound, so the bound is
// the thing under test (fails if the bound is removed).
func TestNewGhTokenResolver_Bounded(t *testing.T) {
	writeFakeGh(t, "PATH=\"/usr/bin:/bin\"\nexec sleep 3")
	start := time.Now()
	_, err := NewGhTokenResolver(200 * time.Millisecond)(context.Background())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a hanging gh must return an error")
	}
	if elapsed > time.Second {
		t.Fatalf("the resolver stalled for %v (bound 200ms); the fast-fail bound is not applied", elapsed)
	}
}

// T032 [UNIT] — a missing `gh` returns an error (no run failure).
func TestNewGhTokenResolver_MissingGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no gh on PATH
	if _, err := NewGhTokenResolver(time.Second)(context.Background()); err == nil {
		t.Fatal("a missing gh must return an error")
	}
}
