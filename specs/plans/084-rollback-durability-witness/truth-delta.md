# Truth Delta: 084-rollback-durability-witness

**Plan Package**: `specs/plans/084-rollback-durability-witness`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-23)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)**. **`/axb-technical-research` RUN (2026-09-23)** — `research.md` D1–D9; **ADR 0056** (+ index; **witnesses ADR 0053**, whose `Status` gains the pointer); `techstack.md` MODIFY (Rule 6 correction); `docs/domain-model/**` MODIFY + re-render. `/axb-system-analysis`, `/axb-dsl-refine`, `/axb-tasks`, `/axb-implement` — pending.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0056-rollback-durability-witness.md` (+ `docs/decisions/README.md` index row) | the decision record: resolve by **witnessing** the clause (D1); the **unexported `durableFS` seam** `Sync`/`Rename`/`SyncDir` (D2); **effect-strength calibration** (D3); the directory `fsync` as **accepted-unwitnessed** (D4); the witness pins (D5); the truth-surface corrections (D6); records + scope (D7/D8); RF-084-1…5 §Forward | `spec.md` US1/US2, FR-001…FR-005; `research.md` D1–D9 |
| MODIFY | `docs/decisions/0053-back-rollback-turns.md` — `Status` | the round-084 **forward pointer** (its durability clause is witnessed + calibrated by ADR 0056); the ADR body stays history | `spec.md` FR-004/FR-005; `research.md` D6 |
| MODIFY | `specs/truth/techstack.md` — *Session history store* | removes the **behaviour guarantee** ("…so a crash mid-rollback leaves the prior `history.jsonl` intact") and keeps the **mechanism** (same-directory temp file + `File.Sync` before atomic `os.Rename`), pointing at interface truth / ADR 0056 — `techstack.md` records technology only (`axb-technical-research` **Rule 6**) | `spec.md` FR-004; `research.md` D5 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — invariant `history-rollback-removes-complete-turns` | the durability sentence is **calibrated** to the witnessed effect (temp file, `fsync` **before** the atomic rename) and drops the un-calibrated "crash mid-rollback" over-claim; the observable clauses (removes the last N complete turns atomically, clamps, MUST NOT touch the archive, survivors byte-unchanged) are unchanged; re-rendered | `spec.md` US2, FR-004/FR-005; `research.md` D3/D5 (the model is load-bearing — ADR 0041) |
| NOOP (checked) | `specs/truth/techstack.md` — all other rows | no other technology-stack change; the round adds a test-only seam defaulting to `os` (no new dependency) | `research.md` D8 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | | | |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | | | |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | | | |
