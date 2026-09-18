# Truth Delta: 046-blank-reason-owner-and-presentation-decoupling

**Plan Package**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | The row (round-034/036/039/045 text) describes the `[Tool Reason]` blank-reason guard as living at `agentloop.logAction` + `agentloop.reasonsOf` **plus a defensive guard in the `callRenderer.OnCallEnd` emit closure**, and records that `internal/agent` imports `internal/ui`. R4 adds: the loop renders its four tool lines through the **injected `agentport.ToolLineRenderer`** port (declared in `internal/domain/agent`, implemented by `internal/ui`), so **`internal/agent` no longer imports `internal/ui`** (the RULE-A edge is gone → the layer baseline is header-only); and the blank-reason predicate is **single-owned** by `ui.ToolLineRenderer.ReasonLine` (ONE evaluation of `ui.toolReasonText` returns both the line and the render decision), consumed at the two live sites (begin line, tail filter) with the **dead** `callRenderer.OnCallEnd` re-check removed. | `truth-current`; round 046 FR-010; `research.md` D1/D2/D3/D6/D8. |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | The row states *"At delivery the baseline holds the 8 known violations … so the count reaches 0 across R2–R4."* R4 delivers that **terminal state**: R2 removed the 7 `cli → infrastructure` edges (**8 → 1**), R3 left the 8th (baseline stayed **1**), and R4 removes `internal/agent → internal/ui` (**1 → 0**) — the committed baseline is **header-only (0)**; a re-introduced `internal/agent → internal/ui` import **fails** the gate (the anti-bypass rule holds at 0). | `truth-current`; round 046 FR-004/FR-010; `research.md` D8. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**` directory exists**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; `specs/truth/` contains only `data/`, `features/`, and `techstack.md`. This round replaces an internal presentation import with an injected port and consolidates a predicate — it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A4 / `research.md` D9. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes | No persisted/runtime state change: the round is a structural consolidation (an in-memory presentation ownership move); no record field, file location, or lifecycle changes. | `spec.md` A4 / `research.md` D9. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (note) | `specs/truth/features/cli/chat/dsl.md:55` — the **round-036 note** | The note (a historical record, not a `DSLRow`/step) asserted the blank-reason guard lives at `agentloop.logAction` + `agentloop.reasonsOf` **plus a defensive guard in the `callRenderer.OnCallEnd` `emit` closure**. R4 **deletes** that third site (C-R4-2), so the note's clause was reconciled to the single-owner reality: the owner is `ui.ToolLineRenderer.ReasonLine`; the two loop-tier sites stand and **the third no longer exists**. **No `DSLRow` and no step changes** — the class-phrase vocabulary and the topology audit are unchanged. | `truth-current` — the note described a site that no longer exists; round 046 FR-003/FR-010; `research.md` D9. |
| NOOP (checked) | `specs/truth/features/cli/**` — every other feature file and `DSLRow` | No user-facing CLI interface behaviour change: no feature Rule, Example, step, or `DSLRow` is added or changed. **Stale-row guard (measured — two greps, per the review's R-2 refinement):** (1) the SYMBOL grep `grep -rn 'ToolReasonRenders\|FormatTool' specs/truth/` returns the `chat/dsl.md` note (now reconciled, above) — the named formatters stay exported and `agentloop.logAction`/`agentloop.reasonsOf` still exist; (2) the BEHAVIOUR grep `grep -rn 'OnCallEnd' specs/truth/` returns the same note — confirming the symbol-existence guard could NOT have caught the stale clause, which is why the clause is reconciled and this limit is recorded. | `spec.md` A1/A5; `research.md` D9 — the non-BDD-refactor precedent (rounds 020/031/041/042/043/044/045). |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0015-loop-presentation-port.md` (+ the `docs/decisions/README.md` index row) | Records the rule: the agent loop renders its tool-line diagnostics through an injected **`agentport.ToolLineRenderer`** port; the **`internal/ui`** tier owns the formatting **and** the blank-reason predicate; the predicate is evaluated once per site on the real path. States its relation to ADR **0005 D1** (tool-call-log parity — partitions *rendering*; **reaffirmed**: the CLI remains the renderer/accounting owner), ADR **0013** (composition-root injection — the wiring site), and ADR **0014** (yield-policy owner — **unchanged**: the loop keeps the `YieldIndicator`/`RestoreIndicator` bracket). Amends nothing; supersedes nothing. | round 046 FR-009; `research.md` D9 — a project-level rule future rounds (R5 [#101](https://github.com/gosharplite/tellme/issues/101)) must cite. |
