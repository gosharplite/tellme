# Truth Delta: 025-spinner-width-safety

**Plan Package**: `specs/plans/025-spinner-width-safety`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Updated the **Turn progress spinner (operator)** row: the tool-phase label is now **bounded** for several tools (` Executing tools [<first> and <N-1> more]...`; the single-tool / no-names forms unchanged) and the presenter **tracks the rendered-row count** of its last frame and erases **every** occupied row on redraw/clear (a width-safe clear, so a soft-wrapped frame leaves no residue). The terminal width comes from an injected `columns` seam (default `golang.org/x/term.GetSize` on the `stderr` fd; the diagnostic `TELL_ME_FORCE_STDERR_COLS` seam drives it in E2E/unit). | Round-025 research Decisions 1–4: both defects from issue #55 — the unbounded several-tool label (clipped resource segment) and the single-row clear (wrap defeats the round-019 teardown contract). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the spinner authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked — the spinner is ephemeral runtime presentation; no new persisted state, entity, field, or lifecycle. | No record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` | The several-tools Example's `Then` becomes the **bounded** `the progress spinner names the first tool and counts the remaining tools`; added the Rule `The spinner leaves no residue on a terminal narrower than its line` (+ a narrow-terminal Example asserting the row-aware clear + the existing teardown Then). | FR-001/FR-004/FR-005 — the bounded label and the residue-free clear. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Bounded the several-tool Then row (`the progress spinner names the first tool and counts the remaining tools` — the retired row enumerated every name); added the Given `the diagnostics are shown at a terminal narrower than the indicator line` and the Then `the progress indicator is cleared from every row it occupied`; added the round-025 note. | The bound + the row-aware clear need executable rows; the retired enumeration row is replaced (no dangling row). |
| NOOP | `specs/truth/features/cli/dsl.md` | Checked — the narrow-terminal Given is used only by the `chat` module, so it lives in `chat/dsl.md` (not the interface root); the root DSC vocabulary is unchanged. | `dsl-single-authority` — the root keeps only cross-module rows. |
