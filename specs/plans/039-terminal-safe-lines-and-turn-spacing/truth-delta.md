# Truth Delta: 039-terminal-safe-lines-and-turn-spacing

**Plan Package**: `specs/plans/039-terminal-safe-lines-and-turn-spacing`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md`: the **Agent tool loop** row (the decomposed `[Tool …]` log) records that the control-sequence sanitization is a **single-owned policy** applied by **every** `[Tool …]` formatter (not just `[Tool Output]`), and that the live turn output is **blank-line grouped** (a blank before each call's begin block, before the grouped tail reason block, and before the post-status group). **AMEND** `docs/decisions/0007-terminal-control-sanitization.md` (governance, in place): broaden the policy scope from the one `[Tool Output]` formatter to all four `[Tool …]` formatters and record the `[Tool Output]`-scoped neutral-close decision (Q4).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/watching-the-tool-loop.feature` (a new Rule: every `[Tool …]` line is control-free; a new Rule: the live turn output is blank-line grouped) + the matching `specs/truth/features/cli/chat/dsl.md` rows.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | The terminal-control sanitization becomes a **single-owned** `internal/ui` policy applied by **every** `[Tool …]` formatter (`Reason` / `Result` / `Action` / `Output`); the reason/result/action lines are now control-free too. The live turn output is **blank-line grouped** (one blank before each call's begin block, before the grouped tail reason block, before the post-status group). Presentation-only. | `spec.md` FR-001..FR-008; `research.md` D1–D8. |
| AMEND | `docs/decisions/0007-terminal-control-sanitization.md` | Broaden the recorded policy scope from the one `[Tool Output]` formatter to all four `[Tool …]` formatters; record the `[Tool Output]`-scoped neutral-close decision (Q4) and the reference divergence check. | `spec.md` FR-003/FR-005; `research.md` D1–D5. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round changes only `stderr`-bound **presentation** (sibling sanitization + blank-line grouping). | `contract-authoritative` holds vacuously; `spec.md` FR-009; `plan.md`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record` and the `~/.tellme/*.jsonl` record shapes | No persisted-state change: the affected text is streamed to `stderr` and never persisted; blanks and stripped escapes are presentation-only. | `spec.md` FR-009/FR-011; `research.md` D8; `plan.md`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — new `Rule: Every tool-loop line is free of terminal control sequences` | ADD a Rule with Examples: an escape-bearing reason yields a control-free `[Tool Reason]` line; an escape-bearing result yields a control-free `[Tool Result]` line; an escape-bearing argument yields a control-free `[Tool Action]` line. | `spec.md` FR-001/FR-002/FR-005; `research.md` D1–D4. |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — new `Rule: The live turn output is grouped by blank lines` | ADD a Rule with Examples: a multi-call round puts a blank before each call's begin block; a reason-less call's action line is still preceded by a blank; the grouped tail reason block and the post-status group are each preceded by exactly one blank. | `spec.md` FR-006/FR-007/FR-008; `research.md` D6/D7. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — new `Then` rows (control-free line; blank-line positions) + a round-039 note | ADD rows matching exactly one step each (`dsl-exact-one-match`); `dsl-single-authority` preserved (new rows, no duplication). | `spec.md` FR-001/FR-006/FR-007/FR-008; `research.md` D5. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| AMEND | `docs/decisions/0007-terminal-control-sanitization.md` (+ the `docs/decisions/README.md` index row if the summary changes) | Extends the ADR's scope: the policy is now owned by all four `[Tool …]` formatters, and the neutral-close restore remains `[Tool Output]`-scoped (Q4). | The round's generalized policy + the recorded non-change need to live on the existing authoritative ADR (not a frozen plan package). |
