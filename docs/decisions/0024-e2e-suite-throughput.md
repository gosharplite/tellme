# ADR 0024 — E2E suite throughput: parallel scenarios by default + a subset that is never the gate

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0010](0010-test-deadline-decoupling.md) (the standing test-deadline doctrine this round reuses instead of pinning scenarios serial), [ADR 0012](0012-hermetic-make-go-env.md) (the `Makefile` boundary the new target sits behind), [ADR 0011](0011-layer-discipline-gate.md) (the `-count=1` load-bearing invocation); issue [#87](https://github.com/gosharplite/tellme/issues/87) (the `di` flake — the resource-pressure precedent); round 040 (**TD-1** — `Strict: true`; the report-and-ignore failure mode the subset invariant guards against); round 055 (`specs/plans/055-e2e-suite-throughput` — this ADR's round)

## Context

The round's trigger is an operator observation: *"I see you do full test and it takes more than 60 sec."* Measured on the reference host (darwin/arm64, Go 1.26.6, 2026-09-19, one session): `go test -count=1 ./...` ≈ **55 s**, of which `tests/e2e` ≈ **50 s** — a single godog suite (`TestFeatures`) executing the executable CLI contract **serially**: **46 feature files / 240 Examples / 1770 steps** (`research.md` owns the round's measured numbers), ≈205 ms per scenario.

The suite is already **isolation-sound**: scenario state is per-scenario (`context.WithValue`), each scenario gets a fresh temp `TELL_ME_HOME` **and** a fresh temp `HOME` (round 026), the `steps` package holds only immutable package-level values, and the binary is built once per process (`sync.Once`, unique temp dir). The suite therefore *can* be parallelised — the risk is not correctness of state but **wall-clock observers** (the `#87`-class flake), and the repo has a standing doctrine for those: **ADR 0010**.

## Decision

**D1 — parallelism is on by default at `4`, overridable.** The suite's `godog.Options` gains `Concurrency: e2eConcurrency()`, where `e2eConcurrency()` returns the `TELL_ME_E2E_CONCURRENCY` value when it parses to ≥ 1, else the constant **`e2eDefaultConcurrency = 4`**. Measured: **~50.7 s → ~16.9 s (3.0×)**; at 8 it is 14.9 s (a ~1.2 s gain for 2× the concurrent children — hence 4). The gain is **wait-overlap, not CPU parallelism** — the serial floor is dominated by the three never-answering MCP scenarios (~11 s each) plus the `sleep`-based timeout legs — which is why 4 still helps a small CI. No dependency change (`Concurrency` exists in the pinned `godog v0.16.0`).

**D2 — the timing-sensitive scenarios are protected by doctrine, not by pinning.** No scenario is pinned serial and no static "timing features" list is introduced (it would drift as features move). The measured evidence is **N = 5 consecutive green full parallel runs**; if a future host/CI is slow enough to exceed a timing scenario's margin, the remedy is **ADR 0010's** — raise *that test's* local margin and keep the shape-based non-vacuity assertion.

**D3 — the fast subset is a `Makefile` convenience, selected with `godog.paths`, and can never be the gate.** A new target **`test-fast`** runs a **strict subset**:
- **selector**: `godog.paths` only. godog v0.16.0's only selectors are `Paths` and `Tags`, and the truth tree has **zero tags** — a tag-based subset would edit `specs/truth/features/**` (a truth-owner change), recorded as a forward item (RF-055-1) rather than smuggled into this round.
- **plumbing (recorded)**: godog's `TestSuite`/`Options` path never binds its command-line flags (`TestSuite.Run` skips `getDefaultOptions` when `Options != nil`), so godog's own `-godog.paths` cannot reach the suite through `go test -args`. The harness therefore registers the **same flag name and semantics** on the stdlib flag set in `init()` (`flag.StringVar(&e2ePathsOverride, "godog.paths", …)`); the default (flag absent) is the full contract. This keeps a genuine **path-based** selection without adding a dependency or writing truth.
- **default selection**: the **non-`chat` modules** (`configuration history workspace usage diagnostics` = **43 Examples**, ≈2 s). `chat` alone is **197/240 Examples**, so it is not a fast subset; `E2E_FAST_MODULES` overrides the selection.
- **banner**: the target prints `SUBSET — NOT THE GATE` naming the selected paths and the gate command; the **harness** prints the same banner whenever a selection is supplied.
- **guard (on RESOLVED paths, in the HARNESS)**: `tests/e2e/suite_test.go` `guardSelection` resolves a supplied selection to the **set of `.feature` files it contains** and **refuses** a set equal to the whole contract — so a traversal spelling (`…/cli/../cli`) and an all-modules list are both caught, and **both** entry points (`make test-fast` *and* a hand-typed `go test -args -godog.paths=…`) obey the invariant (round-055 review **B-055-1**). The `Makefile` only assembles the selection (failing loud on an empty list or a missing module); the previous string-equality clauses were removed as dead/mislabelled (review **TD-055-1**).
- **caveats (recorded, review R-055-x)**: only `paths` is registered on the stdlib flag set, so the other `godog.*` flags are **undefined** (the concurrency seam is `TELL_ME_E2E_CONCURRENCY`, **not** `-godog.concurrency`; passing another `godog.*` flag is a loud usage error); and `-args` is application-wide, so a repo-wide `go test ./... -args -godog.paths=…` fails every other package — a subset is a `tests/e2e`-only invocation.

**D4 — the gate's scope is untouched.** `make test` / `go test -count=1 ./...` still runs **all 240 Examples**. `test-fast` is deliberately **not** a member of `verify` and is not referenced by `test`. `Strict: true`, `-count=1`, `StopOnFailure=false`, `Randomize=off`, and the single-build harness are all unchanged.

**D5 — the measurable bar (operator-locked).** The full gate MUST be **≤ 60 %** of the paired serial baseline on the same host/session (measured: **33 %**, ≈3.0×); a loose absolute **ceiling of ≈25 s** is recorded here as a **sanity backstop only** (an absolute bar goes born-stale on a slower machine — ADR 0010); stability = **N = 5** green runs.

## Consequences

- The default gate's E2E wall-clock drops ≈3×; nothing about *what* is asserted changes.
- A green `test-fast` is never evidence about the whole contract; the banner and the guard make that visible and enforced.
- The gate remains cache-defeating (`-count=1`) and strict — the two properties the round-042 architecture gate depends on.
- No product code; `go.mod`/`go.sum` unchanged; POSIX-only (the `Makefile` block's hermetic boundary applies to the new target like every other).

## Forward items

- **RF-055-1** — a *chat*-scope inner loop needs **tags** (`@fast`), i.e. a `/axb-dsl-refine` round writing `specs/truth/features/**`.
- **RF-055-2** — the ≈25 s absolute ceiling is a recorded backstop, never a gate assertion.
- **RF-055-3** — concurrency raises the peak concurrent child count (4 `tellme` processes + temp homes); a resource-constrained CI lowers `TELL_ME_E2E_CONCURRENCY` (the seam exists for this).
- **RF-055-4** — the subset is module-granular; per-file selection via `godog.paths` is a possible refinement (unverified, not needed).
- **RF-055-5** — an optional `-race` E2E check in `verify` (~20 s; `verify` has no `-race` member today). The round records the executed witness (`go test -race -count=1 ./tests/e2e/` green, 19.7 s) but does not gate on it.
- **RF-055-6** — the `godog.*` flag **namespace mimicry**: only `paths` exists; moving the seam to a project namespace (`-tellme.e2e.paths` + a `TELL_ME_E2E_PATHS` env twin, keeping `godog.paths` as a documented alias) would remove the partial-mimicry expectation. Behaviour is safe today (review R-055-x).
