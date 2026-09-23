# Truth Delta: 082-listing-backward-turn-indices

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-23)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)**. **`/axb-technical-research` RUN (2026-09-23)** — `research.md` D1–D8; **ADR 0054** (+ index; **amends ADR 0045**); the `techstack.md` *Session lifecycle flags* row MODIFY. **`/axb-system-analysis` RUN** — `plan.md` (1 CLI end; api/data NOOP; not modelled). **`/axb-dsl-refine` RUN** — the `history` listing header Rule + `history/dsl.md` rows. **`/axb-tasks` + `/axb-implement` RUN** — `tasks.md`, the code, the unit pins, the E2E green.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0054-list-backward-turn-indices.md` (+ `docs/decisions/README.md` index row) | the decision record: the backward turn index in the `-l` role headers (`[USER] - N` / `[MODEL] - N`); the true-distance computation on the full entry list (D1/D2); the label + `<= 0` fallback (D3); the whole-label colour unit (D4); presentation-only + scope guard (D5…D7); **amends ADR 0045** | `spec.md` US1/US2, FR-001…FR-008; `research.md` D1–D8 |
| MODIFY | `specs/truth/techstack.md` — *Session lifecycle flags* | the `-l` listing presentation owner gains the round-082 clause: the role headers carry the backward turn index (`[USER] - N` / `[MODEL] - N`; same index for both messages of a turn; true distance; whole-label colour unit; the last-N-**messages** selection unchanged) | `spec.md` US1/US2, FR-001…FR-006; `research.md` D1–D8 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — the round reads `history.jsonl` and formats a header; the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-008; `research.md` D5/D6 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | a new Rule + Examples: the listed role headers carry the backward turn index (`[USER] - N` / `[MODEL] - N`; the most recent turn is `- 1`; both messages of a turn share the index; a partial listing keeps the true distance; the whole label is accented only on a terminal) | `spec.md` US1/US2, FR-001…FR-006 |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | the round-082 note + updated/added Then rows (the header sequence `[USER] - K` …; the true-distance leading message; the whole-label accent) | `spec.md` US1/US2; `research.md` D2/D4 |
| NOOP (checked) | `specs/truth/features/cli/**` (other modules) | Only the `history` module's feature/DSL is touched (the `cli` root still resolves the run). | `plan.md` §1 |
