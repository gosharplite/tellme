# System Analysis Plan — round 047 (`047-application-import-ceiling-gate`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/047-application-import-ceiling-gate/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks (NOT this branch)

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling / Layer-discipline gate) ✓ done

docs/decisions/
├── 0016-application-import-ceiling.md   # governance — ADD (RULE-E) ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (the gate is a dev surface, not the `tellme`
CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
tools/arch/arch_test.go             # CHANGED — add RULE-E (sanctioned-set section in the tier table + the rule) + the coverage self-test
tools/arch/baseline.txt             # CHANGED — gain the 3 RULE-E residual edges (header comment updated); generated, never transcribed
docs/decisions/0016-application-import-ceiling.md  # NEW — RULE-E; index row in docs/decisions/README.md ✓ done
specs/truth/techstack.md            # MODIFY — Build & Tooling: the Layer-discipline gate row names RULE-E ✓ done
Makefile                            # unchanged — RULE-E rides the existing `verify-architecture` target (no new member)
go.mod / go.sum                     # unchanged — no new dependency
internal/** , cmd/** , tests/**     # unchanged — no product code / no CLI behaviour change
```

**Structure Decision**: Round 047 is a **build/quality-pipeline extension**, not a runtime-interface
change. It adds **RULE-E** — an application-tier **import ceiling** — **inside** the round-042 guard
(`tools/arch`, ADR 0011), reusing the whole mechanism (module-root-anchored `go list`, the
`CROSS_TARGETS` union, the merged production+test graph, the SCC pass, the `baseline.txt` ratchet) and
the committed baseline, per `research.md` D1–D11. There is **no** new endpoint, **no** persisted state,
**no** CLI behaviour change, **no** new Makefile target, and **no** new dependency. The truth changes
are `specs/truth/techstack.md` (Build & Tooling) + the governance **ADR 0016**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes the **build / quality pipeline**,
which is **not** a system boundary (backend / frontend / CLI): the `tellme` CLI end's observable
behaviour is unchanged, so there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state; the baseline is a **repo artifact**, not runtime state).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the gate is a dev surface (`make verify`), not the `tellme` binary's CLI contract).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — a build gate is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes (`techstack.md` +
ADR 0016) are RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/047-application-import-ceiling-gate`;
truth root `specs/truth`; truth-delta `specs/plans/047-application-import-ceiling-gate/truth-delta.md`;
interfaces: **none**; the round's delivery is **RULE-E in the `tools/arch` guard** + the **3 baseline
edges** + **ADR 0016** + the `techstack.md` truth row.

---

### The rule RULE-E adds (normative: ADR 0016 + the guard's tier table)

The existing gate enforces **four** rules over the pinned layer ranking (ADR 0011 **D1/D2**):
**(A)** no **upward** import; **(B)** application tiers MUST NOT import `internal/infrastructure/**`;
**(C)** `internal/domain/**` purity; **(D)** default-deny for an unranked `internal/**` package.
**RULE-E** is a **fifth, independent** rule — a **positive ceiling** on the application tiers' downward imports.

| Tier | Packages | RULE-E? |
| --- | --- | --- |
| 0 Pure | `internal/domain/**` | **sanctioned** (allowed) |
| 1 Shared utilities | `internal/config`, `internal/home` | **sanctioned** (allowed) |
| 2 Application | `internal/app/**` | **governed** — may import `domain` + `config`/`home` + `app/**` only |
| 3 Infrastructure | `internal/infrastructure/**` | not sanctioned for the app tiers (RULE-B ∩ RULE-E) |
| 4 Agent | `internal/agent` | **not sanctioned** for the app tiers |
| 5 Presentation | `internal/ui`, `internal/ui/tui/**` | **not sanctioned** for the app tiers |
| 6 CLI | `internal/cli` | **governed** — may import `domain` + `config`/`home` + `app/**` only |
| — | `cmd/tellme`, `tests/**`, `tools/**` | **exempt** |

**RULE-E (ADR 0016 D1–D6).** For a **governed application tier** (`internal/app/**`, `internal/cli`) an
import of an **`internal/**` package is a violation unless it is in the sanctioned set**
(`internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`); **stdlib** is allowed by
construction; **third-party** module imports are **out of RULE-E's scope** (a recorded residual); the set
is **normative in the guard's tier table** (ADR 0011 **D7**); **default-deny**; **fail-on-stale allow-list**
(an unused sanctioned entry fails); violations are **deduped by edge** with RULE-A/B/C.

**Baseline = 3** (re-measured from the gate at implementation) — `internal/cli -> internal/agent`,
`internal/cli -> internal/ui`, `internal/cli -> internal/ui/tui/prompt`; `internal/app/**` has **0**
residual (it imports only `domain` + the sanctioned `config`); RULE-A/B/C stay **0**; **0** cycles; **0**
unranked. The 3 are the **R5 de-coupling** workstream ([#101](https://github.com/gosharplite/tellme/issues/101)),
removed one-for-one until the baseline is header-only again.

**Enumeration / mechanism (inherited, not re-derived).** The guard runs `go list` with
`cmd.Dir = <module root>` (a Go test's CWD is its package dir, so `./...` must be anchored), evaluates the
**union over `CROSS_TARGETS`**, merges the production **+** test import graph for the rule and evaluates
the **production-only** graph for the SCC pass, and its self-test asserts the enumeration itself before
any ranking assertion (ADR 0011 **D4/D5/D8**). **No** new Makefile target — RULE-E rides
`verify-architecture` (already a member of `make verify`); the `-count=1` invocation stays load-bearing.

---

### Gating blockers

*(none — the operator locked the theme (**R5.1** of [#101](https://github.com/gosharplite/tellme/issues/101))
and answered clarify Q1–Q5 one at a time (Q1-A gate-first slice · Q2-A normative allow-list,
default-deny, fail-on-stale allow-list · Q3-A both application tiers · Q4-A the sanctioned set as
measured · Q5-A ADR 0016 + slug). No open decision gates the round.)*
