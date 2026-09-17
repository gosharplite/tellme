# Truth Delta: 035-spinner-tail-residue

**Plan Package**: `specs/plans/035-spinner-tail-residue`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Turn progress spinner (operator)** row | Added the round-035 clause: the per-call **tail** (grouped `[Tool Reason]` + measured payload + metrics + `Ready`) is written with the indicator **yielded** — a *phase-boundary* yield that synchronously clears the spinner **before** the tail's first line and does **not** resume it (the next waiting phase re-activates), so the tail's lines begin on their own rows and no frame residue survives into the finished turn. The tool-log lines / `[Tool Output]` keep their clear+write+resume yields; the deferred final tail is unchanged. | Issue #72; `spec.md` FR-001/FR-002; `research.md` D1/D2/D6. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; this round changes only a `stderr` diagnostic write. | `contract-authoritative` holds vacuously; `spec.md` FR-006. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry` / `history_step` / `usage_record` and the `output/<mode>/tokens.log` cadence | No persisted-state change: the fix alters only **when/how the tail is written to the terminal** (a spinner clear before the write). The usage record shape and the one-`AppendBatch`-per-turn cadence are untouched. | `spec.md` FR-005/FR-006; `research.md` D6. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` | Added `Rule: The spinner leaves no residue on a tool-using turn` with the Example `The turn runs a tool and then answers` — the gate (`the diagnostics are shown at a terminal`) + a scripted tool round (`… asks tellme to read "notes.txt" and then answers with "all good"`) + the whole-stream residue row `the run shows no progress spinner`. No new `DSLRow` (3 Given / 1 When / 2 Then all reuse existing rows). | Issue #72; `spec.md` FR-001/FR-002/FR-004; `research.md` D3/D4. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added a round-035 note: the per-call tail is written with a **phase-boundary** spinner yield (synchronous clear before the tail, no immediate resume; the next waiting phase re-activates); this makes the whole-stream residue row (`the run shows no progress spinner`) hold for the gated tool-tail path; witnessed E2E + a unit ordering pin. Class-phrase vocabulary unchanged. | `research.md` D1/D3/D4. |
