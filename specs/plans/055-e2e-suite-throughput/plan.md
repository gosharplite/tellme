# System Analysis Plan — round 055 (`055-e2e-suite-throughput`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/055-e2e-suite-throughput/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md                    # ✓ done (/axb-technical-research)
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY ×4 (Testing & Verification + Build & Tooling) ✓ done

docs/decisions/
├── 0024-e2e-suite-throughput.md   # governance — ADD ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (the round changes *how* the contract is
executed, not *what* it asserts). No `contracts/**` — `/axb-api-plan` **NOOP**. No `data/**` —
`/axb-data-plan` **NOOP**. No `ui/**` artifact (no UX surface).)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
tests/e2e/suite_test.go             # CHANGED — Concurrency (e2eDefaultConcurrency = 4) + the TELL_ME_E2E_CONCURRENCY seam
Makefile                            # CHANGED — a new `test-fast` target (subset via godog.paths; banner + never-the-gate guard); NOT a member of `verify`; `test` unchanged
docs/decisions/0024-e2e-suite-throughput.md   # NEW — ADR 0024; index row ✓ done
specs/truth/techstack.md            # MODIFY ×4 ✓ done
go.mod / go.sum                     # unchanged — no dependency change (Concurrency exists in the pinned godog v0.16.0)
internal/** , cmd/**                # unchanged — NO product code
specs/truth/features/**             # unchanged — the contract is not edited (the subset uses godog.paths)
```

**Structure Decision**: Round 055 is a **test-tooling / developer-throughput** change. It adds
(`tests/e2e/suite_test.go`) a `Concurrency` option — **on by default at 4**, overridable by
`TELL_ME_E2E_CONCURRENCY` — and (`Makefile`) a `test-fast` convenience target that runs a **subset**
of the executable contract via `godog.paths`. The suite's **isolation** (per-scenario
`TELL_ME_HOME` + `HOME`), its single binary build, `Strict: true`, and the gate's scope (**all 240
Examples**) are unchanged. There is **no** system boundary touched: the `tellme` CLI's observable
behaviour, the acceptance/interface Gherkin, the API, and the persisted state are all untouched, so
there is **no** interface to delegate or carry forward. The truth changes are
`specs/truth/techstack.md` (four rows) + the governance **ADR 0024**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes the **test harness and the
build tooling**, which are not a system boundary (backend / frontend / CLI): the `tellme` binary's
observable behaviour is unchanged, so no interface exists to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state; the *measured baseline* and the
>   *subset selection* are development-time facts, not system state).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the executable contract's
>   sentences, Examples and steps are byte-identical; only the runner's concurrency/selection changes).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — a test runner is not
>   a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes
(`techstack.md` ×4 + ADR 0024) are RD-side and owned by `/axb-technical-research` (already applied in
the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/055-e2e-suite-throughput`; truth
root `specs/truth`; truth-delta `specs/plans/055-e2e-suite-throughput/truth-delta.md`; interfaces:
**none**; the round's delivery is the **suite's parallel default** + the **`TELL_ME_E2E_CONCURRENCY`
seam** + the **`make test-fast`** subset target (with its banner + never-the-gate guard) + **ADR 0024**
+ the four `techstack.md` truth rows.

---

### The invariants this round introduces (normative: ADR 0024 + the harness)

| # | Invariant | Carrier |
| --- | --- | --- |
| **I-1** | **Parallel by default, overridable** — `Concurrency = e2eDefaultConcurrency (4)`, or the `TELL_ME_E2E_CONCURRENCY` value when it parses to ≥ 1 | `tests/e2e/suite_test.go` |
| **I-2** | **The gate's scope is all Examples** — `make test` / `go test -count=1 ./...` always executes all 240 Examples under `specs/truth/features/cli`; `Strict: true` and `-count=1` are unchanged | `Makefile` (`test` untouched) |
| **I-3** | **The subset never becomes the gate** — `test-fast` selects with `godog.paths`, prints a `SUBSET — NOT THE GATE` banner, and **refuses** a selection that resolves to the whole contract; `test-fast` is **not** a member of `verify` | `Makefile` (`test-fast`) |
| **I-4** | **Timing scenarios stay non-vacuous** — no scenario pinned serial, no static timing-feature list; the standing remedy is ADR-0010 (a generous test-local margin + shape-based non-vacuity) | `research.md` D2; ADR 0024 D2 |
| **I-5** | **The measured bar** — the gate ≤ 60 % of the paired serial baseline (same host/session); stability = N = 5 green runs; a loose ≈25 s ceiling is recorded as a backstop only | ADR 0024 D5; `research.md` D5 |

### Gating blockers

*(none — the operator locked the theme (**"Start 055 with both"**: US1 + US2) and answered clarify
Q1–Q3 one at a time (Q1 → B on-by-default concurrency · Q2 → A `godog.paths` subset, never the gate ·
Q3 → A the same-session ratio bar + N = 5). No open decision gates the round.)*
