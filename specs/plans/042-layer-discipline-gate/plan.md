# System Analysis Plan — round 042 (`042-layer-discipline-gate`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/042-layer-discipline-gate/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks (NOT this branch)

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling) ✓ done

docs/decisions/
├── 0011-layer-discipline-gate.md  # governance — ADD (the rule + baseline policy) ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (the gate is a dev surface, not the `tellme`
CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
Makefile                            # CHANGED — add the `verify-architecture` target; wire it into `verify`; `.PHONY` + `help`
tools/arch/                         # NEW — the build-tagged (`//go:build arch`) Go guard + its self-test + an untagged `doc.go`
tools/arch/baseline.txt             # NEW — the committed, sorted baseline (8 known violations)
docs/decisions/0011-layer-discipline-gate.md   # NEW — the rule + baseline policy; index row in docs/decisions/README.md
specs/truth/techstack.md            # MODIFY — Build & Tooling: the layer-discipline gate row + the `verify` aggregate list ✓ done
go.mod / go.sum                     # unchanged — no new dependency
internal/** , cmd/** , tests/**     # unchanged — no product code / no CLI behaviour change
```

**Structure Decision**: Round 042 is a **build/quality-pipeline change**, not a runtime-interface change.
It adds a **layer-discipline gate** (`Makefile` `verify-architecture` → a `-tags=arch` Go guard over
`tools/arch/`) that enforces the pinned **two-part predicate** over the module's own import graph (the
`CROSS_TARGETS` union, anchored to the module root) and diffs it against a committed **baseline**
(`tools/arch/baseline.txt`), as a member of `make verify`. There is **no** new endpoint, **no** persisted
state, **no** CLI behaviour change, and **no** new dependency, consistent with `research.md` D1–D12. The
truth changes are `specs/truth/techstack.md` (Build & Tooling) + the governance **ADR 0011**.

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
ADR 0011) are RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/042-layer-discipline-gate`; truth root
`specs/truth`; truth-delta `specs/plans/042-layer-discipline-gate/truth-delta.md`; interfaces: **none**;
the round's delivery is the `Makefile` gate + the `tools/arch/` guard + the baseline + the ADR + the
`techstack.md` truth row.

---

### Pinned layer ranking + the rule the gate enforces

| Tier | Packages | Notes |
| --- | --- | --- |
| 0 Pure | `internal/domain/**` | imports only `internal/domain/**` + stdlib (RULE-C) |
| 1 Shared utilities | `internal/config`, `internal/home` | `infrastructure → config/home` is **legal** |
| 2 Application | `internal/app/**` | RULE-B: MUST NOT import `internal/infrastructure/**` |
| 3 Infrastructure | `internal/infrastructure/**` | |
| 4 Agent | `internal/agent` | |
| 5 Presentation | `internal/ui`, `internal/ui/tui/**` | above `agent` ⇒ `agent→ui` is a violation (RULE-A) |
| 6 CLI | `internal/cli` | RULE-B: MUST NOT import `internal/infrastructure/**` |
| — | `cmd/tellme`, `tests/**`, `tools/**` | **exempt** |

**The predicate (two parts — review B-1).** **(A)** no **upward** import (imported tier > importer tier);
**(B)** the application tiers (`internal/app/**`, `internal/cli`) MUST NOT import `internal/infrastructure/**`;
**(C)** `internal/domain/**` purity; **(D)** default-deny for an unranked `internal/**` package.
*(A one-part "no upward import" rule cannot reproduce the 7 `cli → infrastructure` entries — they are **downward** imports RULE-B forbids. Hence two parts.)*

**Baseline = 8** (re-measured from the gate at implementation) — 7 `cli → infrastructure` (RULE-B) +
`agent → ui` (RULE-A); **0** unranked, **0** cycles. The 7 are **R2**'s to remove and the 8th is
**R3/R4**'s (so the count reaches 0 across R2–R4). Normative rule + worked examples: **ADR 0011**.

**Enumeration anchor (review B-2)**: the guard runs `go list` with `cmd.Dir = <module root>` (a Go
test's CWD is its package dir, so `./...` must be anchored), evaluates the **union over `CROSS_TARGETS`**
(review TD-1), and its self-test asserts the enumeration itself (package count + known governed
packages) before any ranking assertion.

---

### Gating blockers

*(none — the operator locked the theme (R1 of [#92](https://github.com/gosharplite/tellme/issues/92)) and
answered clarify Q1 (broad rule → now a two-part predicate) + Q3 (fail-on-stale); Q2 was resolved by
measurement. No open decision gates the round. The **plan half is complete**; `/axb-tasks` and
`/axb-implement` are intentionally **out of scope for this branch** per the operator's instruction.)*
