# Technical Research: resolver test load tolerance — decouple the unit tests from production fast-fail constants (round 041)

**Plan Package**: `specs/plans/041-di-resolver-test-load-tolerance`

**Status**: Decisions D1–D8. Ratifies the operator-locked Q1–Q6 (delegated recommendations) and the root-cause refinement in issue [#87](https://github.com/gosharplite/tellme/issues/87)’s comments. **Test-only + documentation** — no production Go file changes.

## Context

Anchor: issue [#87](https://github.com/gosharplite/tellme/issues/87) — `internal/infrastructure/di`’s bounded `gh`-token-resolver unit tests fail non-deterministically during a whole-suite run with `signal: killed` at exactly **2.00 s**.

### Root cause (settled — measured, not hypothesised)

The resolver is deterministic **given its deadline**. `NewGhTokenResolver(bound)` runs `exec.CommandContext(cctx, "gh", "auth", "token")` in its own process group with `cmd.Cancel = kill(-pgid, SIGKILL)` and `ghWaitDelay = 2 s`; when `bound` elapses, `cmd.Output()` returns an `*exec.ExitError` whose text is exactly `signal: killed`. The **positive** trimming test passes `2 * time.Second` — the **production fast-fail constant** — as *its own* budget, and under whole-suite contention the (instant) shim’s **fork+exec** is delayed past 2 s, so the deadline fires and the kill surfaces as the failure. Evidence measured on this host (darwin/arm64, 8 CPU, Go 1.26.6):

| Observation | Value |
| --- | --- |
| `TestNewGhTokenResolver_TrimsToken`, **standalone** (instant `echo` shim) | **0.14 s** |
| `TestNewGhTokenResolver_Bounded`, standalone (bound 200 ms + kill + reap) | 0.20 s |
| `TestNewGhTokenResolver_MissingGh`, standalone (no spawn — `LookPath` fails) | 0.00 s |
| `di` package, standalone | 0.57–1.08 s |
| `TestNewGhTokenResolver_TrimsToken`, **whole-suite** (contended) | **≈1.9–2.4 s** (package 2.69–5.14 s) |
| Whole-suite attempt **#1** (unfixed) | **FAIL** — `TrimsToken (2.00s)` `signal: killed`; package 5.136 s |
| Whole-suite attempt **#3** (unfixed) | PASS — package 2.693 s (the positive test squeaked in just under 2 s) |
| Whole-suite attempt **#2** (positive bound **widened to 30 s**) | **PASS** — package 2.951 s (the positive test ran ≈2.4 s, well inside 30 s) |

⇒ a **≈14–17× host-speed factor** between the quiet and the contended host, i.e. the same margin the issue observed (3 of 5 runs red; ~50 % here). The failure is **the bounded resolver working as designed** on a child that had not finished within the production budget — **not** a PATH-resolution defect and **not** a product defect. The defect is that a **test** paces itself on a **production fast-fail constant**.

Consequence for the issue’s suggested fixes: option **A** (absolute interpreter + absolute `sleep`) does **not** address the cause (the deadline — not PATH lookup — is what fired); option **C** (“bound them tighter”) is backwards for the positive test; option **B** (dominant PATH) is genuine **hygiene** but not load-bearing. The **load-bearing** fix is to **decouple the positive test from host speed**.

## Decisions

### D1 — The positive test takes a generous, test-local bound (the load-bearing fix)

`TestNewGhTokenResolver_TrimsToken` (subject = **token trimming**) takes a **generous** resolver bound from a **test-local named constant** (`generousResolverBound = 30 * time.Second`), **not** the production 2 s. 30 s is ~12× the worst observed loaded spawn (~2.4 s), so an instant shim can never be the variable under test. The bound is a **test fixture value**, not a production knob (the production bound is the caller’s; the `di` resolver is invoked by the CLI with the FR-020 discovery bound).

**Rationale**: the test’s subject is the **trim** (`strings.TrimSpace`), so its deadline must be invisible; only `TestNewGhTokenResolver_Bounded` exercises boundedness. This removes the coupling that reddens the standard gate — and the round-040 closeout gate is `go test -count=1 ./...` (`SESSION-CLOSEOUT.md` Step 2, **Rule 1 forbids closing out on a red gate**).

**Alternatives considered**:
- **Keep 2 s and retry / `t.Skip` on a kill** — rejected: masks nondeterminism (issue acceptance forbids it; ADR-036 spirit).
- **`t.Parallel()` on the `di` tests** — rejected: the flake is **host load**, not intra-package concurrency; parallelising does not decouple the deadline (and would widen the Q6 class).
- **A tighter bound** — rejected: strictly worse (issue option C).
- **A smaller generous value (e.g. 10 s, ~4× headroom)** — viable; recorded as a research-level number, subject to the falsifiability constraint (D3). 30 s is chosen for ≥10× headroom.

### D2 — `writeFakeGh` gives the shim a **dominant** PATH (hygiene; removes the N1 contortion)

`writeFakeGh` sets `PATH = <shim-dir> + os.PathListSeparator + <inherited PATH>` (capturing the inherited value **before** `t.Setenv`), so the shim dir is **first** — `gh` is still **shadowed** — while the shim’s **children** resolve normally. The hanging shim therefore simplifies from the round-032 N1 contortion (`PATH="/usr/bin:/bin"; exec sleep 3`) to a bare `exec sleep 3`.

**Rationale**: it removes a readability wart the round-032 review itself flagged as a hack. It is **not** the load-bearing fix (D1 is): the deadline fired on an *instant* shim, so PATH lookup was never the cause.

**Alternatives considered**:
- **Option A (absolute interpreter + absolute `sleep`)** — rejected as the fix (does not address the cause) and brittle (`sleep`’s path differs across POSIX hosts); it does **not** preserve the shadow-`gh` mechanism as cleanly.
- **Leave the shim-only PATH + restore PATH inside the shim** — rejected: keeps the N1 contortion for no benefit once the shim is dominant.
- **A hard-coded `#!/bin/sh` + `/bin/sleep`** — recorded as the fallback if a future host lacks `sleep` on the inherited PATH (the D3 vacuity pin makes a missing `sleep` fail loudly rather than pass vacuously).

### D3 — The bounded test keeps the tight bound + a raised ceiling, and gains a **non-vacuity pin**

`TestNewGhTokenResolver_Bounded` remains the **sole falsifiability carrier**: bound stays **200 ms**; the shim genuinely hangs (3 s); assertions become:

1. `err != nil` (a hanging `gh` is bounded to an error);
2. **`elapsed >= bound`** (the **non-vacuity pin** — a shim that exits instantly, e.g. exit 127 from an unresolved `sleep`, can no longer pass the test vacuously; the resolver must have actually waited out the deadline);
3. `elapsed <= boundedCeiling` where **`boundedCeiling = 2 * time.Second`** (raised from the round-032 `> 1 s`).

**Rationale**: the ceiling must clear the same ≈14–17× host-speed factor that delayed the positive test’s spawn, otherwise a **correct** resolver can false-red under load (the knife-edge the issue comment flagged: 0.2 s actual vs a 1 s ceiling, only 5× margin). At 2 s the margin over the 200 ms bound is 10×, while **falsifiability survives**: an **unbounded** resolver measures ≈3 s (the child’s sleep) `> 2 s` ⇒ red. Pin (2) closes the round-032 N1 hazard permanently and independently of PATH.

**Alternatives considered**:
- **Keep the 1 s ceiling** — rejected: the same class of host-speed coupling would reintroduce a load-dependent false red on the *bounded* test.
- **A larger ceiling (≈3 s)** — rejected: the unbounded case measures ≈3 s, so a 3 s ceiling would straddle the falsifiability boundary (margin ≈1×); 2 s is the strongest ceiling that keeps a real gap to 3 s.
- **No non-vacuity pin (rely on the shim restoring PATH)** — rejected: PATH-dependent vacuity is exactly what D2 removes; an explicit `elapsed >= bound` assertion is the robust invariant.

### D4 — the missing-`gh` test is a recorded **non-change**

`TestNewGhTokenResolver_MissingGh` sets `PATH` to an empty temp dir, so `exec.LookPath("gh")` fails **before any spawn** (measured 0.00 s) — it is not load-coupled and needs no change. Its `1 s` bound is irrelevant because no child runs. **Recorded non-change** (checked, no edit).

### D5 — no masking, no `time.Sleep`, no new dependency

No `t.Skip`, no retry loop, no relaxed/vacuous assertion (**FR-005**); the shim’s shell `sleep 3` is **not** a Go `time.Sleep`, so `make verify-no-test-sleep` stays green; stdlib-only (`os`, `time`, `context`, `path/filepath`, `testing` are already imported) — `go.mod` / `go.sum` unchanged.

### D6 — Witnesses (falsifiability)

| Layer | Witness |
| --- | --- |
| Whole-suite, **under contention** | `go test -count=20 ./...` run **concurrently with a full `./tests/e2e` run** stays green (SC-001) — the discriminating criterion; a quiet-host run cannot tell “fixed” from “quiet”. |
| Falsifiability (a) | revert the positive test to `2 * time.Second` ⇒ **under load** the positive test reddens (`TrimsToken (2.00s)` `signal: killed`) — **already reproduced** on the unfixed head (attempt #1; issue §observed). |
| Falsifiability (b) | make the resolver unbounded (`bound` ignored) ⇒ the bounded test measures ≈3 s `> boundedCeiling (2 s)` ⇒ red. |
| Falsifiability (c) | make the bounded shim vacuous (an instant exit-127 shim) ⇒ `elapsed >= bound` fails ⇒ red (the N1 hazard is now pinned). |

Witnesses (a)–(c) are reproduced then reverted in `/axb-implement`, per house style.

### D7 — Governance: ADR 0010 records the rule; `techstack.md` records the harness

- **ADD `docs/decisions/0010-test-deadline-decoupling.md`** (+ the `docs/decisions/README.md` index row): *“A test must not use a production fast-fail constant as its own deadline”* — with three sub-rules: **(i)** a test whose subject is not the bound uses a generous, test-local constant; **(ii)** a real-time **ceiling** assertion must clear a host-speed margin while preserving falsifiability against the unbounded case; **(iii)** a bound-under-test must assert **non-vacuity** (`elapsed >= bound`); **(iv)** a test’s `PATH` shim is **dominant** (shadow the target, resolve children).
- **MODIFY `specs/truth/techstack.md` (Testing & Verification)**: extend the **Host test harness** row with the test-deadline rule (ADR 0010) and the **Pure-helper unit tests** row with the `di` resolver harness specifics (dominant-PATH `gh` shim; the generous positive bound; `_Bounded` as the sole boundedness carrier with the `elapsed >= bound` vacuity pin and the 2 s ceiling).

**Rationale**: the round-035 session lesson — a forward item must live on a **durable, citable surface** (an ADR / truth row), not a frozen plan package.

**Alternatives considered**: a `research.md`-only note (rejected — freezes with the package); a code comment only (rejected — not citable as a project rule).

### D8 — Truth & round scope

- `/axb-technical-research` — **MODIFY** `techstack.md` (D7) + **ADD** ADR 0010.
- `/axb-api-plan` — **NOOP** (no HTTP surface). `/axb-data-plan` — checked **NOOP** (no persisted state; a unit-test fixture only). `/axb-dsl-refine` — **NOOP** (no new/changed CLI interface truth; the carrier is the unit test + the recorded rule — the round-020/031 non-BDD-tooling precedent).
- `/axb-spec-by-example` — **NOOP** (no user-facing CLI journey). `/axb-ui-plan` — skipped (plain CLI).
- **No production Go file changes** (FR-009); the round touches `internal/infrastructure/di/mcp_factory_test.go`, `specs/truth/techstack.md`, and `docs/decisions/**`.

## Residual risks (forward)

- **A future host factor larger than ~14–17×** could still delay the positive shim beyond 30 s (or the bounded kill+reap beyond 2 s). The 30 s / 2 s choices carry ≥10× headroom over the measured worst case; a pathological host is out of scope (the issue’s acceptance is “20× under contention”, not “any host any load”).
- **The non-vacuity pin assumes `bound` is the *only* reason the resolver waits** — true today (`context.WithTimeout` is the sole timer); a future resolver that waits for another reason would need the pin re-examined.
- **The dominant PATH relies on the inherited PATH containing `sleep`**; if absent the shim exits 127 and the `elapsed >= bound` pin fails **loudly** (never vacuously) — the intended failure mode.
- **The sibling wall-clock-assertion class elsewhere in the suite** (Q6) is **out of scope**, recorded as a forward item on `techstack.md`’s “Not Introduced Yet” / the round diff (its own round).

## Truth impact (ratified by the owners)

| Owner | Artifact | Action |
| --- | --- | --- |
| `/axb-technical-research` | `specs/truth/techstack.md` (Testing & Verification: Host test harness + Pure-helper unit tests rows) | **MODIFY** |
| `/axb-technical-research` | `docs/decisions/0010-test-deadline-decoupling.md` + `docs/decisions/README.md` | **ADD** |
| `/axb-api-plan` | `specs/truth/contracts/**` | **NOOP** (checked) |
| `/axb-data-plan` | `specs/truth/data/**` | **NOOP** (checked) |
| `/axb-dsl-refine` | `specs/truth/features/cli/**` | **NOOP** (checked) |
