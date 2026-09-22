# Truth Delta: 081-back-rollback-turns

**Plan Package**: `specs/plans/081-back-rollback-turns`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)** (the theme is anchored by issue [#163](https://github.com/gosharplite/tellme/issues/163); the issue's *Open design decisions* are routed to `/axb-technical-research`). **`/axb-technical-research` RUN** (2026-09-22) — `research.md` D1–D9; **ADR 0053**; the `techstack.md` rows; the `docs/domain-model` `History` rollback action + `Session` offline invariant (re-rendered; `modelith-check` green). **`/axb-dsl-refine` RUN** (2026-09-22) — the new `history` feature + the `history/dsl.md` rows + the root `cli/dsl.md` offline-path scope. **`/axb-system-analysis` + `/axb-tasks` + `/axb-implement` RUN** — `plan.md`, `tasks.md` (all `[X]`), the code, the unit pins, and the E2E green.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0053-back-rollback-turns.md` (+ `docs/decisions/README.md` index row) | the decision record: the `-b`/`--back [N]` flag; the durable store rollback capability (S-1); the clamp/`N ≤ 0`/composition decisions (S-2…S-4); the confirmation line (S-5); the generalised optional-int pre-pass (S-7, resolving ADR 0023 **RF-54-4**) | `spec.md` US1/US2, FR-001…FR-010; `research.md` |
| MODIFY | `specs/truth/techstack.md` — *CLI flag parsing* | register `-b`/`--back [N]` (default 1; a following non-integer token is a prompt); the shared optional-int pre-pass covering `-l`/`--list` and `-b`/`--back` (RF-54-4) | `spec.md` FR-001/FR-002/FR-009; `research.md` D1-D9 |
| MODIFY | `specs/truth/techstack.md` — *Session history store* | add the durable rollback capability (atomic temp-file + `fsync` + rename; archive untouched) | `spec.md` FR-004/FR-008, NFR-001; `research.md` D1-D9 |
| MODIFY | `specs/truth/techstack.md` — the offline-path set (`tellme performs no network access` scope) | the standalone `-b` / `-b N` joins the offline paths | `spec.md` FR-003, NFR-004; `research.md` D1-D9 |
| MODIFY | `specs/truth/techstack.md` — *Prompt input* | the dispatch-precedence owner row gains the `-b` tier (`… → -l → -b → -t → …`) and `-b`/`--back` joins the never-reads-stdin list (review fold **F-081-5**) | `spec.md` NFR-004; ADR 0053 D7 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — *`History`* | add a **rollback** action/invariant (a session's last N complete turns may be removed atomically; the archive is untouched; schema unchanged) | `spec.md` FR-001/FR-004/FR-010; ADR 0041 same-PR rule |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state **shape** change — a rollback removes whole `history_entry` lines and adds none; the `history_entry`/`history_step` DBML is unchanged. | `spec.md` FR-010; `research.md` D1-D9 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/rolling-back-the-session-history.feature` (**NEW**) | new **Rules** + Examples (offline rollback of the last turn / the last N; a tool-using exchange survives while a later plain exchange is undone; rolling back more than the session holds clears it; the archive is untouched; the rollback-then-prompt form; a `0` count and a `-b`×`--new` refusal) | `spec.md` US1/US2, FR-001…FR-007 |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | **+1** Given row (a plain exchange after a tool-using one) + **+5** When rows (roll back the last turn / the last N / N turns (invalid) / roll back-then-ask / roll back-and-start-fresh) + **+6** Then rows (the exact surviving count; the 1-turn report; the N-turn report; the surviving tool step; the empty archive; the last persisted exchange) + the round-081 note | `spec.md` US1/US2; `research.md` |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) | the `tellme performs no network access` row's **offline-path scope** gains the standalone `-b`/`--back [N]` | `spec.md` FR-003, NFR-004 |
| NOOP (checked) | `specs/truth/features/cli/**` (other modules) | Only the `history` module's feature/DSL is touched (the `cli` root still resolves the run). | `plan.md` §1 |
