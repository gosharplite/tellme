# Truth Delta: 036-tool-reason-sanitize

**Plan Package**: `specs/plans/036-tool-reason-sanitize`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | The per-call `[Tool Reason]` now folds (`\n`/`\r` → space) + trims the model-authored reason and caps it at **200 rendered runes** (one U+2026 counted inside the cap, rune-safe, evaluated on the folded value), exactly like the result snippet, and emits **no** reason line when the folded reason is blank (empty/whitespace-only). The cap set grows from `{189, 200}` to `{189, 200, 200}`. Fix restores round-022 **B1**'s fold guarantee that round 034 dropped **unrecorded**; the drop + this round's decision are recorded in the new **ADR 0006** (`docs/decisions/0006-tool-reason-fold-and-cap.md`). | Issue #74; `spec.md` FR-001/FR-002/FR-005; `research.md` D1/D2/D3/D5. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; this round changes only a `stderr` diagnostic rendering (the `[Tool Reason]` line). | `contract-authoritative` holds vacuously; `spec.md` FR-007; `plan.md`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry` / `history_step` / `usage_record` and the `output/<mode>/tokens.log` cadence | No persisted-state change: the fix alters only **how the model-authored reason is rendered** to the terminal (fold + trim + cap on the `[Tool Reason]` line) — no field, record shape, or cadence moves. | `spec.md` FR-007; `research.md` D7; `plan.md`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — the `chat/dsl.md` reason row is owned here)_ | `specs/truth/features/cli/chat/dsl.md` | — | — |
