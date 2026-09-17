package di

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Resolver test-deadline constants (ADR 0010 — a test must not pace itself on a
// production fast-fail constant).
//
// The production fast-fail constant (2s) is the *caller's* bound, not a test
// budget: under whole-suite contention an instant shim's fork+exec can exceed it
// and the resolver's own SIGKILL fires ("signal: killed" at exactly 2.00s, issue
// #87). So the trimming test — whose subject is trimming, not the bound — takes a
// generous, test-local deadline; only the bounded test keeps a tight bound and
// asserts the deadline actually fired.
const (
	// generousResolverBound decouples a non-bound test from host speed
	// (≈14–17× margin over the measured loaded spawn; ADR 0010 D1).
	generousResolverBound = 30 * time.Second
	// boundedResolverBound is the tight bound the bounded test exercises.
	boundedResolverBound = 200 * time.Millisecond
	// boundedCeiling catches an unbounded resolver (≈3s child) while clearing a
	// host-speed margin over the bound (ADR 0010 D2); it stays strictly below the
	// unbounded-case measurement so falsifiability survives.
	boundedCeiling = 2 * time.Second
)

// writeFakeGh writes an executable `gh` shim into a temp dir and points PATH at
// it so the production resolver can be exercised without a real `gh`.
//
// The shim dir is placed FIRST on PATH and the inherited PATH is appended
// (dominant PATH, ADR 0010 D4): `gh` is shadowed, while the shim's own children
// (e.g. `sleep`) resolve normally — so a hanging shim needs no in-shim PATH
// restoration (retiring the round-032 implementation-review N1 contortion). The
// inherited value is captured BEFORE t.Setenv so the append does not alias the
// value being replaced.
func writeFakeGh(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	inherited := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+inherited)
}

// T032 [UNIT] — the production bounded `gh` resolver trims the token.
//
// The subject is trimming, so the resolver deadline is a generous test-local
// constant (ADR 0010 D1): an instant shim must never be the variable under test,
// regardless of host load (issue #87).
func TestNewGhTokenResolver_TrimsToken(t *testing.T) {
	writeFakeGh(t, `echo "tok-123"`)
	tok, err := NewGhTokenResolver(generousResolverBound)(context.Background())
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
// This is the SOLE carrier of the resolver's boundedness. The shim genuinely
// hangs (a bare `exec sleep 3`, resolvable via the dominant PATH), so the 200ms
// bound must fire; the non-vacuity pin rejects an instantly-failing shim, which
// would otherwise pass the err!=nil check vacuously (the round-032 N1 hazard).
// The pin asserts the child was KILLED by the deadline (ExitCode() == -1, i.e.
// signalled), not that it exited on its own with a real status — a robust
// discriminator independent of host spawn cost (ADR 0010 D3). The ceiling clears
// a host-speed margin while staying strictly below the ≈3s unbounded case (D2).
func TestNewGhTokenResolver_Bounded(t *testing.T) {
	writeFakeGh(t, "exec sleep 3")
	start := time.Now()
	_, err := NewGhTokenResolver(boundedResolverBound)(context.Background())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a hanging gh must return an error")
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("the resolver returned %T (%v); want the deadline's process failure", err, err)
	}
	if ee.ExitCode() != -1 {
		t.Fatalf("the shim exited %d on its own (%v); the fast-fail deadline did not kill it — the bound was never exercised (vacuous)", ee.ExitCode(), err)
	}
	if elapsed > boundedCeiling {
		t.Fatalf("the resolver stalled for %v (bound %v); the fast-fail bound is not applied", elapsed, boundedResolverBound)
	}
}

// T032 [UNIT] — a missing `gh` returns an error (no run failure).
//
// Recorded non-change (ADR 0010 D4): PATH holds no `gh`, so exec.LookPath fails
// before any child is spawned — the test is not load-coupled to its bound.
func TestNewGhTokenResolver_MissingGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no gh on PATH
	if _, err := NewGhTokenResolver(time.Second)(context.Background()); err == nil {
		t.Fatal("a missing gh must return an error")
	}
}
