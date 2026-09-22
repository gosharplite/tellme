# Truth Delta: 081-back-rollback-turns

**Plan Package**: `specs/plans/081-back-rollback-turns`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)** (the theme is anchored by issue [#163](https://github.com/gosharplite/tellme/issues/163); the issue's *Open design decisions* are routed to `/axb-technical-research`). Owner rows **pending** the downstream skills.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/00NN-back-rollback-turns.md` (+ `docs/decisions/README.md` index row) | the decision record: the `-b`/`--back [N]` flag; the durable store rollback capability (S-1); the clamp/`N ≤ 0`/composition decisions (S-2…S-4); the confirmation line (S-5); the generalised optional-int pre-pass (S-7, resolving ADR 0023 **RF-54-4**) | `spec.md` US1/US2, FR-001…FR-010; `research.md` |
| MODIFY | `specs/truth/techstack.md` — *CLI flag parsing* | register `-b`/`--back [N]` (default 1; a following non-integer token is a prompt); the shared optional-int pre-pass covering `-l`/`--list` and `-b`/`--back` (RF-54-4) | `spec.md` FR-001/FR-002/FR-009; `research.md` D-x |
| MODIFY | `specs/truth/techstack.md` — *Session history store* | add the durable rollback capability (atomic temp-file + `fsync` + rename; archive untouched) | `spec.md` FR-004/FR-008, NFR-001; `research.md` D-x |
| MODIFY | `specs/truth/techstack.md` — the offline-path set (`tellme performs no network access` scope) | the standalone `-b` / `-b N` joins the offline paths | `spec.md` FR-003, NFR-004; `research.md` D-x |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`History`* | add a **rollback** action/invariant (a session's last N complete turns may be removed atomically; the archive is untouched; schema unchanged) | `spec.md` FR-001/FR-004/FR-010; ADR 0041 same-PR rule |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — a rollback removes whole `history_entry` lines and adds none; the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-010; `research.md` D-x |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` (or a new `history/rolling-back-the-session-history.feature`) | new **Rules** + Examples (offline rollback of the last N turns; the default-1 form; the removed turns are the last N; the archive is untouched; the rollback-then-prompt form) | `spec.md` US1/US2, FR-001…FR-007 |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | new Given/When/Then rows (arrange K exchanges; ask tellme to roll back the last N turns; a rollback confirmation; the removed-turn assertions; the rollback-then-prompt form) + the round-081 note | `spec.md` US1/US2; `research.md` D-x |
| NOOP (checked) | `specs/truth/features/cli/**` (other modules) | Only the `history` module's feature/DSL is touched (the `cli` root still resolves the run). | `plan.md` §1 |
