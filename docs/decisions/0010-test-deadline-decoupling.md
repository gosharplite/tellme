# ADR 0010 — Test deadlines: a test must not pace itself on a production fast-fail constant

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Related:** issue [#87](https://github.com/gosharplite/tellme/issues/87) (the flake this ADR resolves);
  round 032 (the bounded `gh` resolver + its implementation-review **N1** — PATH-only-to-shim non-vacuity hack);
  round 040 (where the flake surfaced: `PR #86` review, attributed as pre-existing);
  **ADR-036 parity** (`make verify-no-test-sleep` — "no `time.Sleep` for synchronization": the same *determinism* value, a different axis — wall-clock **deadlines** rather than sleeps);
  round 041 (`specs/plans/041-di-resolver-test-load-tolerance` — this ADR's round).

## Context

`internal/infrastructure/di`'s bounded `gh`-token-resolver unit tests failed non-deterministically during a whole-suite run (`go test ./...` / `make test`):

```
--- FAIL: TestNewGhTokenResolver_TrimsToken (2.00s)
    mcp_factory_test.go:30: unexpected error: signal: killed
```

The resolver (`NewGhTokenResolver(bound)`) is deterministic **given its deadline**: it spawns `gh auth token` in its own process group, bounded by `context.WithTimeout(ctx, bound)`, and on deadline SIGKILLs the group (`cmd.Cancel`), so `cmd.Output()` returns an `*exec.ExitError` with the exact text `signal: killed`.

The defect was **in the test**, not the product: `TestNewGhTokenResolver_TrimsToken` — whose subject is **token trimming** (an instant `echo` shim) — passed the **production fast-fail constant `2 * time.Second`** as *its own* budget. Under whole-suite contention the shim's **fork+exec** was delayed past 2 s (measured **≈14–17×** the quiet-host cost: 0.14 s standalone ⇒ ≈1.9–2.4 s loaded), so the deadline fired and the kill surfaced as a failure. The same real-time coupling sat on the bounded test's `elapsed > 1 s` **ceiling** (only a 5× margin over its 200 ms bound), a latent false-red on a correct implementation under load.

This reddened the standard gate, which is also the **session-closeout gate** (`SESSION-CLOSEOUT.md` Step 2 — `go test -count=1 ./...`; closeout **Rule 1** forbids closing out on a red gate) — at roughly a coin flip on a contended host. It is the same **determinism** value that ADR-036 parity protects for `time.Sleep`, on the adjacent axis of wall-clock **deadlines**.

## Decision

**A test must not pace itself on a production fast-fail constant.** Concretely:

**D1 — Decouple subjects that are not the bound.** A test whose subject is *not* the bound (e.g. token trimming) MUST pass a **generous, test-local** resolver bound (or none), never a production fast-fail constant. Only a dedicated test whose subject **is** boundedness keeps a tight bound.

**D2 — Ceilings clear a host-speed margin.** A **wall-clock ceiling** assertion (e.g. "the bound fired within X") MUST clear a measured **host-speed margin** while **preserving falsifiability** against the unbounded case: the ceiling must remain strictly below the unbounded-case measurement (here: 2 s ceiling vs ≈3 s unbounded sleep).

**D3 — Non-vacuity for a bound under test.** A test asserting that a bound fires MUST also assert **non-vacuity** — that the resolver actually **waited** (e.g. `elapsed >= bound`) — so a shim that fails instantly (e.g. exit 127 from an unresolved child) cannot pass the test vacuously. This supersedes a PATH-in-shim workaround (round-032 N1) with a direct invariant.

**D4 — Dominant PATH for test shims.** A test that shadows an executable via a temp-dir shim MUST give the shim a **dominant** `PATH` — the shim directory **first**, then the inherited `PATH` — so the target binary is **shadowed** while the shim's own children **resolve** normally.

## Alternatives considered

1. **Retry, or `t.Skip` on a `signal: killed`** — rejected: masks nondeterminism (the issue's acceptance forbids it; contradicts ADR-036's spirit).
2. **Absolute interpreter + absolute `sleep` (issue option A) as *the* fix** — rejected: the **deadline**, not PATH lookup, is what fired (an instant shim still exceeded 2 s under load). Retained only as the D4 fallback for a host lacking `sleep` on the inherited `PATH`.
3. **Mark the tests `t.Parallel()`** — rejected: the flake is **host load**, not intra-package concurrency; parallelising does not decouple the deadline.
4. **Tighten the bound (issue option C)** — rejected: backwards for the positive test (a tighter budget makes the flake *more* likely).
5. **Dominant `PATH` alone (issue option B) as *the* fix** — rejected: genuine hygiene (it retires the N1 in-shim `PATH` restoration), but not load-bearing — an instant shim can still blow a tight deadline.
6. **Keep the bounded test's `> 1 s` ceiling** — rejected: the same host-speed coupling would reintroduce a load-dependent false red on the **bounded** test; 2 s keeps a 10× margin over the 200 ms bound while staying strictly below the ≈3 s unbounded case.

## Consequences

- `TestNewGhTokenResolver_TrimsToken` passes a generous, **test-local** bound; `TestNewGhTokenResolver_Bounded` remains the **sole** falsifiability carrier of the resolver's boundedness (200 ms bound; the `elapsed >= bound` vacuity pin; a 2 s ceiling). `TestNewGhTokenResolver_MissingGh` is a recorded **non-change** (it spawns no child).
- `writeFakeGh` sets a **dominant** `PATH`; the hanging shim simplifies to a bare `exec sleep 3` (the round-032 N1 contortion is retired).
- **No production code changes** — the resolver, `ghWaitDelay`, the MCP factory, and the CLI are untouched; `go.mod` / `go.sum` unchanged; `make verify-no-test-sleep` stays green (the shim's shell `sleep` is not a Go `time.Sleep`).
- **Portability**: the rule and the harness change are POSIX-only (Linux/macOS), matching tellme's scope; the `di` tests remain hermetic (no real `gh`, no network) and must pass both standalone and **under contention**.
- **Scope**: the sibling wall-clock-assertion class elsewhere in the suite is **out of scope** (its own round).
