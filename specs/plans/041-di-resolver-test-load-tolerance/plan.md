# System Analysis Plan — round 041 (`041-di-resolver-test-load-tolerance`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/041-di-resolver-test-load-tolerance/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Testing & Verification) ✓ done
docs/decisions/
├── 0010-test-deadline-decoupling.md   # /axb-technical-research — ADD ✓ done
└── README.md                          # CHANGED — index row 0010 ✓ done
```

*(No `features/acceptance/**` (no user-facing CLI behaviour — `/axb-spec-by-example` **NOOP**), no
`features/cli/**` change (`/axb-dsl-refine` **NOOP**), no `contracts/**` change (`/axb-api-plan` **NOOP**),
no `data/**` change (`/axb-data-plan` **NOOP**), and no `ui/**` artifact.)*

### Repository structure (root)

```text
internal/infrastructure/di/mcp_factory_test.go   # CHANGED — the resolver harness (bound + dominant PATH + vacuity pin)
specs/truth/techstack.md                         # MODIFY — Testing & Verification: Host test harness + Pure-helper unit tests rows ✓ done
docs/decisions/0010-*.md + README.md             # ADD + index row ✓ done
internal/infrastructure/di/mcp_factory.go        # UNCHANGED — no production code change (FR-009)
go.mod / go.sum                                  # unchanged — stdlib-only (NFR-004)
Makefile                                         # UNCHANGED — no gate semantics change (FR-008)
```

**Structure Decision**: Round 041 is a **test-harness determinism** change, not a runtime-interface change.
It retunes the resolver's **unit-test fixture** (`writeFakeGh` → dominant PATH), the positive test's
**bound** (a generous test-local constant, decoupled from the production 2 s fast-fail constant), and the
bounded test's **ceiling + non-vacuity pin** — so the `di` package is load-tolerant without weakening the
resolver's boundedness falsifiability. There is **no** new endpoint, **no** persisted state, **no** CLI
behaviour change, **no** production code change, and **no** new dependency, consistent with `research.md`
D1–D8. The round's only truth changes are `specs/truth/techstack.md` (Testing & Verification) and **ADR
0010**, both owned by `/axb-technical-research` (already applied in the plan/truth half).

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes a **Go unit-test harness** plus a
recorded **test rule**; neither is a system boundary (backend / frontend / CLI). The CLI end's observable
behaviour (flags, exit codes, `stdout`/`stderr`, persisted records) is **unchanged** — the resolver under
test is an internal seam whose production callers are untouched (`spec.md` FR-009) — so there is no
interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the round authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (no persisted or in-memory state; the round touches a test fixture only).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — no new/changed Gherkin or `DSLRow`; the carrier is a Go unit test + the recorded rule — the round-020/031 non-BDD-tooling precedent).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **NOOP** (no user-facing CLI journey; a test-harness fix is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's only truth changes (`techstack.md`
+ ADR 0010) are RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/041-di-resolver-test-load-tolerance`;
truth root `specs/truth`; truth-delta `specs/plans/041-di-resolver-test-load-tolerance/truth-delta.md`;
interfaces: **none**; the round's delivery is the retuned resolver harness (`mcp_factory_test.go`) + the
`techstack.md` Testing rows + ADR 0010.

---

### Gating blockers

*(none — the theme is issue [#87](https://github.com/gosharplite/tellme/issues/87), the operator delegated
the six scope decisions (Q1–Q6 → the recommendations, carried in `spec.md`), and `research.md` D1–D8
settle the mechanism. No open decision gates the round.)*


---

## Analysis focus (handoff to `/axb-tasks`)

The round is **test-only + documentation**; `/axb-tasks` breaks the work down directly (no interface
truth to reconcile). The focus set:

- **D1 (load-bearing)** — `TestNewGhTokenResolver_TrimsToken` passes a **generous test-local bound**
  (`generousResolverBound = 30 * time.Second`) instead of the production `2 * time.Second`.
- **D2 (hygiene)** — `writeFakeGh` sets a **dominant** `PATH` (`<shim-dir> + string(os.PathListSeparator) + <inherited PATH>`, the inherited value captured **before** `t.Setenv`); the hanging shim simplifies to a bare `exec sleep 3` (the round-032 N1 in-shim PATH restoration is retired; its explanatory comment is updated).
- **D3 (falsifiability)** — `TestNewGhTokenResolver_Bounded` keeps the `200 ms` bound and gains the
  **non-vacuity pin** (`elapsed >= bound`), and its ceiling moves `> 1s → > 2s` (`boundedCeiling`).
- **D4** — `TestNewGhTokenResolver_MissingGh` is a recorded **non-change** (spawns no child).
- **D5** — no `t.Skip` / no retry / no vacuous assertion; no Go `time.Sleep` (the shim's shell `sleep` is
  not a Go call, so `make verify-no-test-sleep` stays green); stdlib-only.
- **Witnesses (`research.md` D6)** — `go test -count=20 ./...` **under contention** (concurrent with a
  full `./tests/e2e` run) green; falsifiability (a) revert the positive bound ⇒ red under load,
  (b) unbounded resolver ⇒ ≈3 s > 2 s ceiling ⇒ red, (c) vacuous shim ⇒ `elapsed >= bound` fails — each
  reproduced then reverted.
- **Truth/ADR** — `specs/truth/techstack.md` (Testing & Verification) + **ADR 0010** are owned by
  `/axb-technical-research` and already folded; `truth-delta.md` carries the owner rows.

**No production Go file is modified** (`internal/infrastructure/di/mcp_factory.go` untouched —
`spec.md` FR-009); the changed-file set is `internal/infrastructure/di/mcp_factory_test.go`,
`specs/truth/techstack.md`, `docs/decisions/0010-test-deadline-decoupling.md`, and
`docs/decisions/README.md` (+ this plan package).
