# Truth Delta: 075-skill-frontmatter-block-scalars

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` initialized (2026-09-21)** — owner rows **pending**. Clarify resolved at specify time (**not escalated — 0 questions**; residual technical choices S-2/S-3/S-4 → `/axb-technical-research`).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/techstack.md` — *Skills catalog (load)* (row 54) | expected **MODIFY**: the loader resolves a YAML **block scalar** (`>`/`|`, with chomping) to its textual value — a **recorded divergence beyond the reference** (the reference's `parseSkill` is line-based and shares the limitation) | `spec.md` US1/US2, FR-001…FR-006; `research.md` D-x |
| _pending_ | `docs/decisions/00NN-*.md` (+ index) | expected **ADD**: the decision record (implementation shape S-2; folding/chomping S-3; the divergence note) | `spec.md` SC-003; `research.md` D-x |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending (expected NOOP)_ | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending (expected NOOP)_ | `specs/truth/data/**` | No persisted-state change (the catalog is not persisted). | `spec.md` NFR-003 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/chat/listing-the-available-skills.feature` (+ `chat/dsl.md`) | expected **MODIFY**: a Rule/Example for a block-scalar skill + a Given that can author one | `spec.md` US1/US2 |
