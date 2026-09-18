# Truth Delta: 048-cli-tui-prompt-decoupling

**Plan Package**: `specs/plans/048-cli-tui-prompt-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: `/axb-specify` has landed the skeleton; clarify **Q1 → B** (the **TUI prompt** edge), **Q2 → A** (the port lives in `internal/domain/**`), **Q3 → A** (F-4 folded in) are **locked**. No truth owner has run yet — the rows below are **provisional expectations**, to be finalised/ratified by each owner at `/axb-technical-research`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Records the new RULE-E baseline figure **2** (the `internal/cli → internal/ui/tui/prompt` edge removed), citing the round's new ADR. | `truth-current`; round 048 FR-007. |
| MODIFY (expected) | `specs/truth/techstack.md` — an architectural row naming the inverted seam (CLI Application *Interactive TUI prompt* / *Composition root*) | Restated to the injected-port reality (the direct `internal/cli → internal/ui/tui/prompt` import is replaced by an injected port wired at `cmd/tellme`). | round 048 FR-007 — only if the seam is named in a truth row. |
| NOOP (expected) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | The `verify` aggregate already names `verify-architecture`; the round adds no target. | round 048 FR-007. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/` (**no `contracts/**`**) | tellme has a single CLI end and no OpenAPI/HTTP surface; this round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/data/data-model.dbml` | No persisted/runtime-state change — the refactor moves wiring, not data. | `spec.md` A5. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | A `cli`-reader refactor is not a CLI-contract change — no feature Rule, Example, step, or `DSLRow` change; the topology audit is unchanged. | `spec.md` A2/A6. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD (expected) | `docs/decisions/NNNN-cli-tui-prompt-decoupling.md` (+ the `docs/decisions/README.md` index row) | Records the de-coupling: the injected port, its home tier, the behaviour-preservation claim, the baseline movement **3 → 2**, and its relation to ADR 0011/0016/0013 + **what the change is *not***. | round 048 FR-006. |
