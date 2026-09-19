# Technical Research: E2E suite throughput (round 055)

**Plan Package**: `specs/plans/055-e2e-suite-throughput`
**Truth Root**: `specs/truth`
**Owner**: `/axb-technical-research`
**Status**: complete — clarify round 1 CLOSED (Q1 → B, Q2 → A, Q3 → A). Decision-driven; each decision carries its **measured** evidence.

> Every number below was measured on **this host** (darwin/arm64, Go 1.26.6) on 2026-09-19, in the **same session**, and is recorded here as the *owning* artifact for the round's measured claims (the round-041 lesson: a measured claim has one owner).

---

## D1 — Parallelism: on by default at `4`, overridable (`TELL_ME_E2E_CONCURRENCY`)

**Decision** (spec Q1 → B): the godog suite runs with `Concurrency = 4` by default; the env seam `TELL_ME_E2E_CONCURRENCY` overrides it (values `< 1` are clamped to the default by the harness; godog itself clamps `< 1` to 1).

**Measured evidence** (full `tests/e2e` package, `go test -count=1`, same session):

| Concurrency | wall-clock | vs serial |
| --- | --- | --- |
| 1 (serial baseline) | **50.4–51.0 s** (2 runs) | 1.00× |
| 4 (the chosen default) | **16.66–17.07 s** (final 5-run evidence: 16.66/16.77/16.76/17.07/16.93) | **~3.0×** |
| 8 | 14.9 s | 3.4× |

**Why 4, not 8**: the marginal gain from 4 → 8 is ~1.2 s while the concurrent-child count doubles; 4 keeps headroom on a 2–4-core CI box and on a loaded laptop. The seam means a fast machine can raise it locally without a code change. **Why the gain is real and not CPU-bound**: it is **wait-overlap**, not CPU parallelism — the serial floor is dominated by the three never-answering MCP scenarios (~11 s each) plus the `sleep`-based timeout legs, so overlapping *waiting* is what yields ~3× (review §10). That is also why a small CI still benefits from 4.

**Why on by default (not opt-in)**: the complaint the round answers is the *gate's* wall-clock; hiding the speedup behind an env var would leave the default slow.

**Not a dependency change**: `Options.Concurrency` exists in the pinned `github.com/cucumber/godog v0.16.0` (`go.mod` unchanged).

---

## D2 — Timing-sensitive scenarios: **no re-margining required**; ADR-0010 is the standing remedy

**Decision** (spec FR-004): the timing-sensitive scenarios keep their current bounds and witness shapes. The measured evidence is the **N = 5 consecutive green parallel runs** at the default (D1's table) — every timing scenario passed in all five. The ADR-0010 doctrine (a generous test-local margin; non-vacuity by the failure's *shape*) is recorded as the **remedy to apply if a future flake appears**, not as a speculative change now.

**The timing observers that were exercised** (all green, 5/5):

| Scenario class | Mechanism | Bound | Why it is already sound |
| --- | --- | --- | --- |
| `[Tool Output]` idle-gap resume (round 040) | child `sleep 2`; forced threshold `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS=50` | the *idle gap* is the subject | the witness asserts the resume + the following synchronous clear, not a wall-clock budget |
| `execute_command` never-returns / descendant (round 024) | child `sleep 60`, tool `"timeout": 1`; `pollProcessGone(…, 2s)` | the *bound* is the subject | the bound is genuinely the test's subject (ADR-0010's own exception); the descendant leg additionally asserts the late sentinel was **not** written |
| harness stdin handshake | `markerDeadline = 10s` | ceiling only | 10 s is a generous ceiling, not a tight budget |

**Race witness (review R-A)**: `go test -race -count=1 ./tests/e2e/` is **green (19.7 s, no `DATA RACE`)** under the new default concurrency — reproduced 2026-09-19 on the reference host. This is the round's *executed* carrier for the "isolation-sound ⇒ parallel-safe" claim (the pinned godog wraps the formatter per scenario and guards its failure flag with a mutex; `testing.T.Run` concurrency is permitted). An optional `-race` E2E check in `verify` is a **forward item** (`verify` has no `-race` member today).

**Residual (recorded, not closed)**: if a *future* host/CI is slow enough that a scenario's margin is exceeded, the remedy is ADR-0010's — raise **that** test's local margin and keep the shape assertion; do **not** pin scenarios serial or introduce a hand-maintained timing-feature list (which would drift as features move).

---

## D3 — The fast subset: `godog.paths` via `make test-fast`; **never the gate**

**Decision** (spec Q2 → A): a `Makefile` target `test-fast` runs a **strict subset** selected with `godog.paths`; it prints a `SUBSET — NOT THE GATE` banner naming the selection, and **refuses to run** when the selection resolves to the whole contract.

**Selector rationale (measured)**: the contract is 240 Examples over 6 modules; the sizes are lopsided:

| Module | Examples | Files |
| --- | --- | --- |
| `chat` | **197** | 39 |
| `configuration` | 16 | 1 |
| `history` | 15 | 3 |
| `workspace` | 5 | 1 |
| `usage` | 4 | 1 |
| `diagnostics` | 3 | 1 |
| **total** | **240** | **46** |

⇒ a `chat`-only subset would be **197/240 ≈ 82 %** of the contract and is therefore **not** a fast subset. The default selection is the **non-`chat` modules** (`configuration history workspace usage diagnostics` = **43 Examples**, ≈10 s serial, less with D1's concurrency).

**Why `godog.paths` and not tags**: godog **v0.16.0** exposes only `Paths` and `Tags` as selectors (there is **no name filter**), and the tree carries **zero tags**. A tag-based subset would require editing `specs/truth/features/**` — a **truth-owner** change (`/axb-dsl-refine`), i.e. a different round. Recorded as a forward item (RF-055-1).

**The guard (review B-055-1)** is enforced in the **harness** (`guardSelection`), on **resolved paths** — it resolves each selection to the set of `.feature` files it contains and refuses a set equal to the whole contract — so a traversal (`…/cli/../cli`) or an all-modules list is caught, and **both** entry points (`make test-fast` and a hand-typed `go test -args -godog.paths`) obey it. The `Makefile` only assembles the selection (and fails loud on an empty list or a missing module). `test-fast` prints the banner; the harness also prints it (visible under `-v`).

**The gate invariant** (spec FR-002/FR-006, proposed and *not* open to interpretation): the subset **selects for convenience** and **never excludes from the gate** — `make test` / `go test -count=1 ./...` always executes **all 240** Examples. `test-fast` is deliberately **not** a member of `verify` and is not referenced by `test`. The guard exists to prevent the round-040 TD-1 failure mode (scenarios silently dropping while the suite still exits 0).

---

## D4 — What is deliberately **unchanged**

| Element | State | Why |
| --- | --- | --- |
| `Strict: true` | kept | undefined/ambiguous steps must still FAIL (round-040 TD-1) |
| `-count=1` at the gate | kept | disables the test cache (the arch gate's whole-module `go list` cannot be cached) |
| `StopOnFailure` / `Randomize` | off | failure reporting / reproducible order unchanged |
| One binary build (`sync.Once`, temp dir) | kept | process-wide and read-only ⇒ already concurrency-safe |
| Per-scenario `TELL_ME_HOME` + `HOME` | kept | the isolation that makes D1 sound |
| `verify-no-test-sleep` | kept | the harness's only waits are child-process `sleep`s in the *fixture*, not Go test sleeps |
| `go.mod` / `go.sum` | unchanged | no dependency change |

---

## D5 — The measurable bar (spec Q3 → A)

- **Primary (gate-able)**: the full gate is **≤ 60 %** of the **paired serial baseline** measured on the **same host in the same session** (i.e. ≥ 1.6× faster). Measured **across sessions (review R-B)**: the **`tests/e2e` package** ratio is stable at **~33 %** (3.0×), while the **paired full gate** ratio is **load-sensitive — observed 38 % (55.0 s → 20.7 s) here and 47 % (54.0 s → 25.6 s) under the reviewer's concurrent load** (the full-gate figure sits near the ≈25 s backstop — which is exactly why the backstop is not a gate assertion). Either way the ≤ 60 % bar is met; record the **range (33–47 %)**, not a point.
- **Secondary (recorded, not asserted)**: a loose absolute **ceiling** — ≈ 25 s on the reference host — recorded in **ADR 0024** as a sanity backstop only (an absolute bar goes born-stale on a slower machine; ADR-0010).
- **Stability**: **N = 5** consecutive full parallel runs green (measured: 5/5, the D1 table).

---

## D6 — Truth impact (recorded via `truth-delta.md`)

`specs/truth/techstack.md` **MODIFY** ×4: *E2E runner / step definitions* (parallel default + the subset/gate invariant), *Host test harness* (the ADR-0010 standing remedy under parallelism), *Test strategy* (the gate's scope is all Examples, independent of `test-fast`), *Task runner* (the new `test-fast` target, explicitly not part of `verify`). `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine` are **NOOP** (no API/data/acceptance change). Governance: **ADR 0024** + the index row.

---

## Residual risks (recorded)

- **RF-055-1**: a *chat*-scope inner loop needs tags ⇒ a future `/axb-dsl-refine` round (writes `specs/truth/**`).
- **RF-055-2**: the absolute ceiling in ADR 0024 is a recorded backstop, not a gate assertion.
- **RF-055-3**: concurrency raises the peak concurrent child count (4 `tellme` processes + temp homes) — benign here, but a resource-constrained CI may want to lower `TELL_ME_E2E_CONCURRENCY` (the seam exists for exactly that).
- **RF-055-4**: the subset is module-granular; per-file selection is a possible refinement if `godog.paths` accepts file paths (unverified — not needed for the default).
