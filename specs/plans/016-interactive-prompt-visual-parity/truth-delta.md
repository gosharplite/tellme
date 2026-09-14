# Truth Delta: 016-interactive-prompt-visual-parity

**Plan Package**: `specs/plans/016-interactive-prompt-visual-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (placeholder) | `specs/truth/techstack.md` | To be recorded by `/axb-technical-research`: the strict-parity `-i` prompt chrome (border/padding/placeholder/height/width + styled suggestion list) and the debounced, cancelable suggestion pipeline. | Round 016 — align the `-i` prompt surface with the reference; no new dependency. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (placeholder) | `specs/truth/contracts/**` | To be recorded by `/axb-api-plan` (expected `NOOP` — single CLI end, no tellme-owned API surface). | Round 016 changes a terminal surface only. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (placeholder) | `specs/truth/data/**` | To be recorded by `/axb-data-plan` (expected `NOOP` — the shared prompt log and all persisted state are unchanged). | Round 016 adds no state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (placeholder) | `specs/truth/features/cli/**` | To be recorded by `/axb-dsl-refine`: retire the dashboard rule from `chat/using-the-interactive-prompt.feature`; add/adjust the rows pinning the parity chrome and the `Tab`-inserts-selection behaviour; the interface root is expected unchanged. | Round 016 — strict parity (the dashboard header is removed; the chrome + insertion behaviour become executable truth). |
