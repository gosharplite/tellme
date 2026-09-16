# Truth Delta: 034-tool-call-log-parity

**Plan Package**: `specs/plans/034-tool-call-log-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`; owner rows filled after the grill-round (PR #70) fold. Each truth owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | The `stderr` tool-loop log becomes the reference's decomposed shape: `[Tool Engine] Step i/M` (once per **executed** round, at the execution site — G9), `[Tool Reason]`, `[Tool Action] <tool>(sorted k: v)` (reason excluded; value ≤189 runes; `json.Number` rendering — G3/G7), `[Tool Result] <tool>: <snippet>` (≤200 runes — G3), a live `[Tool Output]` block for shell-class non-`output_file` calls (bounded-and-stopped — G4/G8), and the round's reasons re-emitted grouped at each call's **post-call tail** (G6). Round 022's single-line `[Tool]` shape and its blank-line rule are **superseded**; the round-022 no-reason behaviour is retired. | Grill folds G1–G10; `spec.md` FR-001–FR-012. |
| MODIFY | `specs/truth/techstack.md` — **Turn chrome (operator)** row | The chrome **cadence** becomes a status **frame per AI-endpoint call** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload), with the frame number advancing within a prompt (`prior-calls + k` — G6); the round-018/027 line **formats** are unchanged (Q4). | FR-008, FR-009; G5/G6. |
| MODIFY | `specs/truth/techstack.md` — **Post-turn status lines (operator)** row | The metrics line + `╰─⠿ Ready` summary move to the **per-call tail** (each call's own usage), with the **final** call's tail deferred past the answer (G5); the session field is a **recorded display-only divergence** on failed turns / a final call reporting no usage (G2); persistence stays one `AppendBatch` per turn. | FR-008, FR-010b; G2/G5. |
| MODIFY | `specs/truth/techstack.md` — **Turn progress spinner (operator)** row | The spinner is **yielded once per call** (single writer) around a `[Tool Output]` block, and the block renders even when the spinner is gated off (G8); the turn-scoped elapsed is unchanged. | FR-012; G8. |
| MODIFY | `specs/truth/techstack.md` — **Agent command tool (`execute_command`)** row | The child's stdout/stderr are also streamed live as `[Tool Output]` lines up to the round-024 byte budget, after which the process group is stopped; `output_file` calls emit no block; a trailing partial line is dropped (G4). | FR-010, FR-011; G4/G8. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (re-derived, checked) | `specs/truth/contracts/**` | Inspected — tellme has a single CLI end and **no** OpenAPI/HTTP surface; this round authors no request/response contract (it reshapes `stderr` diagnostics only). | `contract-authoritative` holds vacuously; `spec.md` FR-016. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (re-derived, checked) | `specs/truth/data/data-model.dbml` — inspected `usage_record` (the per-call `{timestamp, provider, model, cached/prompt/response/total/thinking tokens, cost}` record) and the `output/<mode>/tokens.log` file cadence | The **persistence cadence and record shape are unchanged**: this round keeps **one `AppendBatch` per turn** and the persistence gate on the final call's `Reported` flag (the one-append-per-call alternative the grill raised was **rejected** — it would change failed-turn durability). The round changes only **when the tail is rendered** (per call + final deferral) and its **display-only** session field (G2). | FR-010b; G2 — **re-derived after the grill** (the earlier blanket NOOP, which never examined the accounting path, is withdrawn). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` | The tool-loop Rules reshaped to the decomposed rendering: `Step i/M`; reason + action (sorted args, rune-capped); grouped post-call reasons; result (rune-capped); the live `[Tool Output]` block (bounded-stopped; `output_file` → none; trailing partial dropped); retirement of the round-022 no-reason Rule and the blank-line Rule. | FR-001–FR-012. |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-turn.feature` | The frame Rule re-anchored: frames number one per AI-endpoint call (`Turn N` advancing within a prompt); the "frame's last line is the pre-flight payload line" clause re-anchored to the **last frame**. | FR-008, FR-009; G5/G6. |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature` | The tail is per call with the **final** call deferred past the answer; the trailing Rule preserved; a new note that the metrics line reflects each call (the last reflecting the last request). | FR-008; G5/G6. |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` | The spinner yields once per call around a `[Tool Output]` block; the block renders when the spinner is gated off. | FR-012; G8. |
| MODIFY | `specs/truth/features/cli/chat/reporting-the-payload-status.feature` | A per-call pre-flight clause (a pre-flight line per frame, monotone within the turn). | FR-010a; G10. |
| MODIFY | `specs/truth/features/cli/chat/estimating-the-wire-payload.feature` | The estimate row re-anchored to the **first** frame (the messages-vs-persona+tools comparison stays valid there). | FR-010a; G10. |
| MODIFY | `specs/truth/features/cli/chat/failing-the-tool-loop.feature` | The bound-reached witness extended: N frames, `M` `Step i/M` lines, one **engine-less** final frame, `the tool request failed`, exit 7, nothing persisted. | FR-001; G9. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Rows for: per-call frames; the k-th frame number; the last frame preceding the answer; engine-line suppression on the bound-reached call; the sorted-args string; the rune-capped value; the dropped trailing partial line; the `output_file` no-block rule; retirement of the round-022 rows; the loop-bound unit note. | FR-001–FR-012. |
