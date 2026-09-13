# Truth Delta: 009-payload-status-line

**Plan Package**: `specs/plans/009-payload-status-line`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added `Token estimator` (stdlib heuristic) + `Payload status line` (per-turn `stderr` estimate + actual, always-on, no `tellme:` prefix, injectable clock seam, observe-only). **Configuration**: added `Payload budget` (`MAX_HISTORY_TOKENS` env/config, default **1000000**, `>= 0`; negative → configuration error). **Reasoning & Provider Transport**: widened the `Provider gateway port` `Response` with the provider's reported `usage`, and `Response normalization` to extract the `usage` block. **Testing & Verification**: extended the `E2E runner`, `Local fake provider` (report/withhold `usage`), and `Pure-helper unit tests` (estimator, budget resolver, status formatting). **Not Introduced Yet**: the reference metrics line (`M:`/`H:`/`C:`/`$cost`) stays out of scope; the budget is displayed, not enforced. | Round-009 research Decisions 1–8 (Clarify Q1–Q3; user-locked default). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; the outbound provider request shape is unchanged. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked; left empty. Per-turn payload token counts are display-only and recomputed from history text on resume; the persisted session model (`history_entry`) is unchanged. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/reporting-the-payload-status.feature` | New `chat` interface feature carrying the payload-status behaviour: the pre-flight estimate line, the post-turn measured line (usage / no-usage), the answer-stream-unchanged rule, the default budget, the override, and the invalid-budget rejection. | Round-009 acceptance (`seeing-the-payload-status` + `choosing-the-payload-budget`). |
| ADD | `specs/truth/features/cli/chat/dsl.md` | Module rows — Givens: `a configured provider "{provider}" whose endpoint answers with "{answer}" and reports its usage` / `… and reports no usage` / `the payload budget is "{budget}"`; Thens: `tellme reports the estimated payload status for the turn` / `tellme reports the measured payload status for the turn` / `tellme reports no measured payload status` / `the payload status measures against a budget of {budget} tokens` / `the payload status names the active mode and model`. | Round-009 status-line contract. |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` | Added the Rule *Listing reports no payload status* and the `no payload status is reported` Then (a `-l` run emits no status line). | Round-009 — only prompt turns report a payload status. |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the status line carries **no** `tellme:` prefix, so the frozen class-phrase vocabulary stays **10**, and any shared row already resolved via the root. | No cross-module row was needed. |
