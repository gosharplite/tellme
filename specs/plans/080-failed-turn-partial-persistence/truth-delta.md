# Truth Delta: 080-failed-turn-partial-persistence

**Plan Package**: `specs/plans/080-failed-turn-partial-persistence`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)** (the theme is anchored by issue [#161](https://github.com/gosharplite/tellme/issues/161); the issue's *Open Design Decisions* are routed by the operator to `/axb-technical-research`). Owner rows **pending** the downstream skills.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Interrupted-turn persistence (operator)* | the row is **broadened** to *Partial-turn persistence (operator interruption or failure)*: a failed turn with ≥1 completed step is persisted with a class-specific synthetic answer, then the failure surface is reported unchanged (phrase + exit 6/7) with the informational keep line; the failed-turn `Append` is best-effort; the interruption predicate broadened to `ctx.Err() != nil \|\| errors.Is(err, context.Canceled)` (folds hole #2) | `spec.md` US1/US2, FR-001…FR-004; `research.md` D1–D7 |
| MODIFY | `specs/truth/techstack.md` — *Session history store* | the round-079 append-after-complete exception is **extended** to a FAILED turn (one entry closed with a failure synthetic answer; zero steps ⇒ nothing) | `spec.md` FR-001/FR-005; `research.md` D1 |
| ADD | `docs/decisions/0052-failed-turn-partial-persistence.md` (+ `docs/decisions/README.md` index row) | the decision record: persist a failed turn's completed steps; scope (B) any failed turn; the unchanged failure surface; the class-specific synthetic answers; the broadened predicate (hole #2); the best-effort append; the records | `spec.md` SC-001…SC-005; `research.md` |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`history-append-after-complete`* | extended the exception from *operator-interrupted* to *operator-interrupted **or failed*** (≥ 1 completed step ⇒ persisted with a class-specific synthetic close; zero steps ⇒ not persisted) | `spec.md` FR-001/FR-005; ADR 0041 same-PR rule; `research.md` D7 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`turn-answer-stored-verbatim`* | extended the synthetic-answer exception to a failed turn | `spec.md` FR-002; ADR 0041 same-PR rule; `research.md` D7 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — the synthetic answers are the existing `answer` field; the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-007; `research.md` D1 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/remembering-the-conversation.feature` | 3 new **Rules** + 4 Examples (a provider-keeps-failing turn keeps its steps, exits 6; a rejected-outright equivalent; the session continues from the kept turn; a zero-step failure writes nothing) | `spec.md` US1/US2, FR-001…FR-005 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **+3** Given rows (read-then-always-drop; read-then-reject; an arranged failed exchange) + **+4** Then rows (the failed turn's step count; the turn-failure answer; the keep line; an empty history) + the round-080 note | `spec.md` US1/US2; `research.md` D7 |
| NOOP (checked) | `specs/truth/features/cli/**` (other modules) | No other CLI module's feature/DSL is touched. | `plan.md` §1 |
