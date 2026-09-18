# Truth Delta: 046-blank-reason-owner-and-presentation-decoupling

**Plan Package**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status: skeleton** — initialized by `/axb-specify`; the owner rows are appended as each truth-owner skill runs. `/axb-technical-research` runs **after** clarify round 1 (**C-R4-1 … C-R4-4**), since the seam shape determines the techstack MODIFY text.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — clarify round 1)_ | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row (expected MODIFY) | The row names the loop's tool-line rendering / the `internal/ui` coupling; under R4 the loop owns no presentation, so the row is rewritten to name the new seam. | `truth-current`; round 046 FR-010; `spec.md` §Truth obligations. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/` (**no `contracts/**` directory exists**) | Expected NOOP — tellme has a single CLI end and **no** OpenAPI/HTTP surface; the round relocates an internal presentation seam and authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A4. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/data-model.dbml` | Expected NOOP — no persisted/runtime state change (an in-memory presentation ownership move). | `spec.md` A4. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Expected NOOP — no user-facing CLI interface behaviour changes (behaviour-preserving); no feature Rule, Example, step, or `DSLRow` added or changed. **Stale-row guard (to be measured):** the round must `grep` `specs/truth/**` for the moved seam names (e.g. `ToolReasonRenders`, `FormatTool*` paths) and record the measured hit count — a renamed seam can leave a stale row passing green (the audit checks feature → row only). | `spec.md` A1/A5. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — C-R4-4)_ | `docs/decisions/0015-*.md` (+ the `docs/decisions/README.md` index row) | Expected ADD — records the presentation-ownership rule (the agent loop emits semantics; the presenter owns the tool-line formatting and the blank-reason predicate) and states its relation to ADRs 0005 (D1 partitions *rendering*), 0013 (composition root), and 0014 (yield-policy owner). | round 046 FR-009; a project-level rule future rounds must cite. |
