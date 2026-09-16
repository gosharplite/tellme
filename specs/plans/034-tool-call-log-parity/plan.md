# Plan — round 034 (tool-call log parity)

Plan package: `specs/plans/034-tool-call-log-parity`

## Theme

Re-cut tellme's tool-call diagnostic rendering to the reference's decomposed shape and the turn output to the reference's **per-AI-endpoint-call** frame cadence — a **rendering-only** round (no capability, no contract, no persisted-state change; `spec.md` FR-016).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (terminal diagnostics)** | `cli` | The subject of the round: what tellme writes to `stderr` during a tool-using turn (the tool-call lines, the per-call status frames, the `[Tool Output]` stream). `stdout` (the answer) stays byte-exact. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change**.

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` MODIFIES the affected `chat` features + `dsl.md`.

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (re-derived, checked).** Inspected `specs/truth/data/data-model.dbml`'s `usage_record` entity + the `output/<mode>/tokens.log` cadence: the round keeps **one `AppendBatch` per turn** and the final call's `Reported` gate, so the persisted record shape/cadence are unchanged (the per-call tail is display-only + a recorded divergence — `spec.md` FR-010b). The earlier blanket NOOP was withdrawn by the grill fold (G2).
- **`/axb-ui-plan` → skipped.** A line-oriented CLI diagnostic: no web/HTML prototype; terminal interactions live directly in the CLI interface Gherkin (the standing CLI streamlining).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must MODIFY the executable CLI truth to carry the revised rendering (`spec.md` FR-001–FR-012, folds G1–G10):

- `chat/watching-the-tool-loop.feature` — the decomposed `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` rendering; the grouped post-call reasons; retirement of the round-022 no-reason + blank-line Rules.
- `chat/presenting-the-turn.feature` — the per-AI-call **frame** cadence; the `N` advancing within a prompt; the "last frame preceding the answer" re-anchor.
- `chat/presenting-the-post-turn-status.feature` — the **per-call tail** with the **final** call's tail trailing the answer; the trailing Rule preserved; the metrics-line-per-call clause.
- `chat/presenting-the-progress-spinner.feature` — the single-writer, once-per-call yield around a `[Tool Output]` block.
- `chat/reporting-the-payload-status.feature` + `chat/estimating-the-wire-payload.feature` — the per-call pre-flight estimate (the estimate row re-anchored to the first frame).
- `chat/failing-the-tool-loop.feature` — the bound-reached witness (N frames, `M` engine lines, one engine-less final frame, failure, exit 7, nothing persisted).
- `chat/dsl.md` — the new/changed rows and the retirements.

`truth-delta.md` carries the explicit MODIFY rows; `specs/truth/techstack.md` is MODIFY-owned by `/axb-technical-research` (already folded).
