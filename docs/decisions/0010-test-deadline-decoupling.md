# ADR 0010 — Test deadlines: a test must not hardcode a tight wall-clock budget that is not its subject

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Related:** issue [#87](https://github.com/gosharplite/tellme/issues/87) (the flake this ADR resolves);
  issue [#89](https://github.com/gosharplite/tellme/issues/89) (the R-4 bound-headroom measurement + wiring pin);
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

The defect was **in the test**, not the product: `TestNewGhTokenResolver_TrimsToken` — whose subject is **token trimming** (an instant `echo` shim) — **hardcoded its own `2 * time.Second`** wall-clock budget. That literal is **not** the caller's bound: production's fast-fail bound is **`mcpDiscoveryBound = 3 * time.Second`** (`internal/cli/mcp_discovery.go:19`; the sole call site is `di.NewGhTokenResolver(mcpDiscoveryBound)` at `:46`), and the 2 s figure is the *sibling* **`ghWaitDelay = 2 * time.Second`** (the post-SIGKILL wait delay in `mcp_factory.go:36`) — both introduced in the same commit (round-032 `0a3ad9c`), so the test literal almost certainly mirrored its sibling. Under whole-suite contention the shim's **fork+exec** was delayed past that 2 s budget (measured **≈14–17×** the quiet-host cost: 0.14 s standalone ⇒ ≈1.9–2.4 s loaded), so the deadline fired and the kill surfaced as a failure. The same real-time coupling sat on the bounded test's `elapsed > 1 s` **ceiling** (only a 5× margin over its 200 ms bound), a latent false-red on a correct implementation under load.

This reddened the standard gate, which is also the **session-closeout gate** (`SESSION-CLOSEOUT.md` Step 2 — `go test -count=1 ./...`; closeout **Rule 1** forbids closing out on a red gate) — at roughly a coin flip on a contended host. It is the same **determinism** value that ADR-036 parity protects for `time.Sleep`, on the adjacent axis of wall-clock **deadlines**.

**Provenance lesson (the most transferable part).** The round's first narrative called the 2 s literal *"the production fast-fail constant"* — an inference from a plausible-looking number, **not** read from the code (production's bound is 3 s and was never 2 s). The round's own method (*"measured, not hypothesised"*) is exactly what caught it: a constant is a **claim**, and a claim about production belongs verified against the call site, not inferred from a sibling literal.

## Decision

**A test must not hardcode a tight wall-clock budget that is not its subject.** Concretely:

**D1 — Decouple subjects that are not the bound.** A test whose subject is *not* a wall-clock bound (e.g. token trimming) MUST take a **generous, test-local** deadline (or none), never a tight hardcoded literal. Only a dedicated test whose subject **is** boundedness keeps a tight, explicit budget.

**D2 — Ceilings clear a host-speed margin.** A **wall-clock ceiling** assertion (e.g. "the bound fired within X") MUST clear a measured **host-speed margin** while **preserving falsifiability** against the unbounded case: the ceiling must remain strictly below the unbounded-case measurement (here: 2 s ceiling vs ≈3 s unbounded sleep). The ceiling is a **harness value**, independent of the production bound (here it is numerically equal to `ghWaitDelay` while being *below* `mcpDiscoveryBound` — do not "align" them).

**D3 — Non-vacuity for a bound under test.** A test asserting that a bound fires MUST also assert **non-vacuity** — that the subject **actually hit the bound** rather than failing for an unrelated, instant reason — so an instantly-failing shim (e.g. exit 127 from an unresolved child) cannot pass the test vacuously. The robust discriminator is the **failure's shape**, not elapsed-vs-bound arithmetic: assert the child was **killed by the deadline** (`exec.ExitError` whose `ExitCode() == -1`, i.e. signalled) rather than exited on its own with a real status. A bare `elapsed >= bound` assertion is unreliable when the bound is near the host's spawn cost (a vacuous shim can also take ≈ the bound to fork+exec — measured: the `elapsed >= bound` form **flaked to PASS** for an instant shim at a 200 ms bound). This supersedes a PATH-in-shim workaround (round-032 N1) with a direct invariant.

**D4 — Dominant PATH for test shims.** A test that shadows an executable via a temp-dir shim MUST give the shim a **dominant** `PATH` — the shim directory **first**, then the inherited `PATH` — so the target binary is **shadowed** while the shim's own children **resolve** normally.

## Alternatives considered

1. **Retry, or `t.Skip` on a `signal: killed`** — rejected: masks nondeterminism (the issue's acceptance forbids it; contradicts ADR-036's spirit).
2. **Absolute interpreter + absolute `sleep` (issue option A) as *the* fix** — rejected: the **deadline**, not PATH lookup, is what fired (an instant shim still exceeded 2 s under load). Retained only as the D4 fallback for a host lacking `sleep` on the inherited `PATH`.
3. **Mark the tests `t.Parallel()`** — rejected: the flake is **host load**, not intra-package concurrency; parallelising does not decouple the deadline.
4. **Tighten the bound (issue option C)** — rejected: backwards for the positive test (a tighter budget makes the flake *more* likely).
5. **Dominant `PATH` alone (issue option B) as *the* fix** — rejected: genuine hygiene (it retires the N1 in-shim `PATH` restoration), but not load-bearing — an instant shim can still blow a tight deadline.
6. **Keep the bounded test's `> 1 s` ceiling** — rejected: the same host-speed coupling would reintroduce a load-dependent false red on the **bounded** test; 2 s keeps a 10× margin over the 200 ms bound while staying strictly below the ≈3 s unbounded case.

## Consequences

- `TestNewGhTokenResolver_TrimsToken` passes a generous, **test-local** bound; `TestNewGhTokenResolver_Bounded` remains the **sole** falsifiability carrier of the resolver's boundedness (200 ms bound; the **exit-code** non-vacuity pin — the child was signal-killed by the deadline, `ee.ExitCode() == -1`; a 2 s ceiling). `TestNewGhTokenResolver_MissingGh` is a recorded **non-change** (it spawns no child).
- `writeFakeGh` sets a **dominant** `PATH`; the hanging shim simplifies to a bare `exec sleep 3` (the round-032 N1 contortion is retired).
- **No production code changes** — the resolver, `ghWaitDelay`, `mcpDiscoveryBound`, the MCP factory, and the CLI are untouched; `go.mod` / `go.sum` unchanged; `make verify-no-test-sleep` stays green (the shim's shell `sleep` is not a Go `time.Sleep`).
- **Portability**: the rule and the harness change are POSIX-only (Linux/macOS), matching tellme's scope; the `di` tests remain hermetic (no real `gh`, no network) and must pass both standalone and **under contention**.
- **Residual (recorded, R-3)**: the 2 s ceiling was validated in the `go test -count=20 ./internal/infrastructure/di/`-under-`tests/e2e` configuration; the **whole-suite package parallelism** (`go test ./...` runs package binaries concurrently) that produced the original red is **not** covered by the ceiling's margin. Under D3 the ceiling adds **no discriminating power** (an unbounded resolver ⇒ the `sleep 3` child exits 0 ⇒ `err == nil` ⇒ red; a `WaitDelay`-only return ⇒ `exec.ErrWaitDelay`, not `*exec.ExitError` ⇒ red), so it *is* the residual host coupling; dropping it is an operator-level option (round-041 Q3 was operator-locked to keep it).
- **Forward item (R-4)**: the production bound's headroom for a **real** `gh` is unmeasured — the observed ≈1.9–2.4 s contended spawn cost is a **trivial shim**, while `mcpDiscoveryBound = 3 s` must also cover a real binary plus a network token refresh (fail-open on timeout). Filed as **[#89](https://github.com/gosharplite/tellme/issues/89)** (measure under contention; a value change is a production change ⇒ its own ADR; plus a wiring pin). Production's 3 s is `Accepted` truth (FR-020) — **not** changed by this round.
- **Verification caveat**: this round has **no CI** on the PR; every witness is author-local (the round-031 SC-002 "manual closeout step, not part of `make verify`" precedent).
- **Scope**: the sibling wall-clock-assertion class elsewhere in the suite is **out of scope** (its own round).
