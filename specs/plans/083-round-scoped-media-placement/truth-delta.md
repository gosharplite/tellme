# Truth Delta: 083-round-scoped-media-placement

**Plan Package**: `specs/plans/083-round-scoped-media-placement`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-23)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)**. **`/axb-technical-research` RUN (2026-09-23)** — `research.md` D1–D9; **ADR 0055** (+ index; **clarifies ADR 0032 D7**); `techstack.md` ×3 MODIFY. **`/axb-system-analysis` RUN** — `plan.md` (1 CLI end; api/data NOOP; not modelled). **`/axb-dsl-refine` RUN** — the `reading-a-local-image` media-round Rule + `chat/dsl.md` rows. **`/axb-tasks` + `/axb-implement` RUN** — `tasks.md`, the code, the unit pins, the E2E green.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0055-round-scoped-media-placement.md` (+ `docs/decisions/README.md` index row) | the decision record: a round's media is folded **once, after** the per-call loop (D1); single-call bytes frozen (D2); Gemini placement unchanged + the recorded cardinality consequence (D3); the loop owns the order (D4); the two-tier witness (D5); placement-only (D6); scope guard (D7); **clarifies ADR 0032 D7** | `spec.md` US1/US2, FR-001…FR-008; `research.md` D1–D9 |
| MODIFY | `specs/truth/techstack.md` — *Agent tool loop* | the loop's media-placement owner gains the round-083 clause: a round's media is folded **once, after** the per-call loop (one `user` message, call order, after all `tool` results) so the round's `tool` results are contiguous; clarifies ADR 0032 D7 (its body read round-scoped; the pre-083 implementation did not); single-call byte-identical; the fold-back paths yield no media | `spec.md` US1/US2, FR-001…FR-004/FR-007; `research.md` D1/D4/D5 |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (OpenAI-compatible)* | the OpenAI image-wire owner gains the round-083 clause: the media `user` message is **round-scoped** (one message of N `image_url` blocks after every `tool` result), so the `tool_calls` block is contiguous; pre-083 the per-call interleave 400'd for N ≥ 2; single image byte-identical | `spec.md` US1, FR-001…FR-003; `research.md` D1/D2 |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | the Gemini image-wire owner gains the round-083 clause: the adapter placement is unchanged (media after the batched function-response turn; standalone inline turns) + the **recorded cardinality consequence** — a multi-media round now emits **one** `user` turn of N `inlineData` parts | `spec.md` US2, FR-006; `research.md` D3 |
| NOOP (checked) | `specs/truth/techstack.md` — all other rows | no other technology-stack change; the round is a placement fix in the loop + records | `research.md` D6 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — media is in-flight only and was never persisted (the `history.Step` stores the tool's text result); the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-008; `research.md` D6 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/reading-a-local-image.feature` | a new Rule + Examples: a step that inspects **several** pictures is shown them **together, after every tool result** (the round completes; a single picture stays unchanged) | `spec.md` US1/US2, FR-001…FR-006 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | the round-083 note + new Given/Then rows (a multi-image read step; the tool results answered together before the pictures) | `spec.md` US1; `research.md` D4 |
| NOOP (checked) | `specs/truth/features/cli/**` (other modules) | Only the `chat` module's image feature/DSL is touched. | `plan.md` §1 |
