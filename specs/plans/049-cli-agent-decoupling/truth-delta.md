# Truth Delta: 049-cli-agent-decoupling

**Plan Package**: `specs/plans/049-cli-agent-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: `/axb-technical-research` (D1–D10 + **ADR 0018** + the `techstack.md` MODIFY rows below) has **landed**. The api/data/dsl-refine rows are the owners' records (each inspected its surface). **Clarify round 1 CLOSED**: **Q1 → (B) re-cut** · **Q2 → (i)** (all three crossing contracts → `internal/domain/agent`) · **Q3 → (a)** (reference + delete; no alias).
>
> **Round shape (Q1 → B)**: this is the **re-cut sub-slice 1** of the `cli → agent` de-coupling. **The RULE-E baseline does NOT move this round** (still **2**); the baseline-moving round is **sub-slice 2** (the `AgentLoop` construction inversion, **2 → 1**).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Adds the **round-049** note (R5.3; **ADR 0018**): the **re-cut sub-slice 1** of the `internal/cli → internal/agent` edge — the loop's crossing contracts (`Result`/`ErrIncomplete`/`ToolDefs`) were extracted to `internal/domain/agent` (referenced directly; no alias), reducing the CLI's `→ agent` references to the single `AgentLoop` construction. The RULE-E baseline figure stays **2** (unchanged) — a preparatory slice; the `→ agent` edge is removed by **sub-slice 2** (**2 → 1**). The rule + policy (ADR 0011/0016) are otherwise unchanged, and **`tools/arch/baseline.txt` is byte-identical**. | `truth-current`; round 049 FR-008; `research.md` D1/D6/D9. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | Notes that the loop's crossing **contracts** — the turn-result type (`Result`), the incomplete-turn error (`ErrIncomplete`), and the wire-def projection (`ToolDefs`) — now live in **`internal/domain/agent`** (peer of the observer/renderer ports) and are **referenced directly** by `internal/agent` (no alias); `internal/cli` reads them from `internal/domain/**`, so its only remaining `internal/agent` reference is the `AgentLoop` construction. Behaviour unchanged. | round 049 FR-001/FR-008; `research.md` D2/D3/D5. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the `verify` aggregate already names `verify-architecture`. The round adds no Makefile target — the gate rides that target — so the row is unchanged. | round 049 FR-008; `research.md` D9. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface (`specs/truth/` = `data/`, `features/`, `techstack.md`). This round is an internal contract relocation; it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5; `research.md` D8 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes: **no** persisted/runtime-state change — the refactor moves a contract (a struct + an error + a function), not data. | `spec.md` A5; `research.md` D8 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Inspected: a `cli`-reader refactor is **not** a `tellme` CLI-contract change — no feature Rule, Example, step, or `DSLRow` changes (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the gate + the identifier-count check + the adapted unit seams. | `spec.md` A2/A6; `research.md` D7/D8 — the rounds 020/031/036/041–048 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0018-cli-agent-contracts-extraction.md` (+ the `docs/decisions/README.md` index row) | Records the re-cut: the **re-cut sub-slice 1** decision (Q1-B), the three extracted **`internal/domain/agent`** contracts (Q2-i), the **reference + delete** choice (Q3-a, no alias), the **deferred** construction inversion (**sub-slice 2**, baseline **2 → 1**), the **baseline unchanged (2)** statement, **what the change is *not***, and its relation to ADR 0011/0016/0017/0015. | round 049 FR-007; `research.md` D6 — a structural change sub-slice 2 must cite. |
