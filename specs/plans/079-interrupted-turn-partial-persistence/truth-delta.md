# Truth Delta: 079-interrupted-turn-partial-persistence

**Plan Package**: `specs/plans/079-interrupted-turn-partial-persistence`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)**. **`/axb-technical-research` RUN (2026-09-22)** — `research.md` (D1–D10) + **ADR 0051** + `techstack.md` **MODIFY ×1 + ADD ×1** + the `docs/domain-model/**` invariant amendment (recorded here; the model is not a `specs/truth/**` artifact). `/axb-dsl-refine` **pending**. `/axb-api-plan` + `/axb-data-plan` record **NOOP**.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Session history store* | the append-after-complete rule is **qualified**: an operator-interrupted turn with **>= 1** completed step is appended as one entry closed with the synthetic `history.InterruptedTurnAnswer`; **zero** steps writes nothing; `calls` counts only the completed inference rounds; the JSON shape is unchanged | `spec.md` US1/US2, FR-001…FR-003, FR-007; `research.md` D1/D8 |
| ADD | `specs/truth/techstack.md` — *Interrupted-turn persistence (operator)* | the interrupted-turn policy: keep completed steps, close with the synthetic answer, exit **0** + a `stderr` informational line, zero-step => today's surface unchanged, typed detection, no usage persistence, the frozen vocabulary/exit set unchanged, the in-flight child kill untouched; a recorded divergence from the reference (which discards the partial turn) | `spec.md` US1/US2, FR-004…FR-008, NFR-001…NFR-004; `research.md` D1–D9 |
| ADD | `docs/decisions/0051-interrupted-turn-partial-persistence.md` (+ `docs/decisions/README.md` index row) | the decision record: persist-with-synthetic-answer, typed detection, the constant, always-on (no toggle), the exit-code/diagnostic choice, no usage persistence, the step trigger, the records, the hermetic E2E seam | `spec.md` SC-001…SC-005; `research.md` |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`history-append-after-complete`* | scoped the exception: an operator-interrupted turn with >= 1 completed step is *completed* by a synthetic closing answer and persisted; zero steps is not persisted | `spec.md` FR-001/FR-005; ADR 0041 same-PR rule; `research.md` D10 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`turn-answer-stored-verbatim`* | added the synthetic-answer exception for an operator-interrupted turn | `spec.md` FR-002; ADR 0041 same-PR rule; `research.md` D10 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — the synthetic answer is the existing `answer` field; no new table/column; the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-007; `research.md` D8 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/chat/remembering-the-conversation.feature`, `specs/truth/features/cli/chat/dsl.md` | a new **Rule** (the interrupted turn) + Examples; the new Given/Then rows | `spec.md` US1/US2 |
