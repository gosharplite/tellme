# Truth Delta: 039-terminal-safe-lines-and-turn-spacing

**Plan Package**: `specs/plans/039-terminal-safe-lines-and-turn-spacing`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md`: the **Agent tool loop** row (the decomposed `[Tool …]` log) records that the control-sequence sanitization is a **single-owned policy** applied by **every** `[Tool …]` formatter (not just `[Tool Output]`), and that the live turn output is **blank-line grouped**; the **Agent command tool (`execute_command`)** row's "siblings unsanitized" forward note is closed. **ADD** `docs/decisions/0008-terminal-safe-lines-and-blank-line-grouping.md` (governance) — which **SUPERSEDES** `docs/decisions/0007-terminal-control-sanitization.md` (its recorded scope boundary is closed; 0007's `Status` flips to `Superseded by 0008`).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/watching-the-tool-loop.feature` (a new Rule: every `[Tool …]` line is control-free; a new Rule: the live turn output is blank-line grouped) + the matching `specs/truth/features/cli/chat/dsl.md` rows.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | The terminal-control sanitization becomes a **single-owned** `internal/ui` policy (`sanitize.go`) applied by **every** `[Tool …]` formatter — `[Tool Reason]`, `[Tool Result]`, and `[Tool Action]` (keys **and** values) are now control-free like `[Tool Output]` (same class/order: fold+trim → sanitize → rune cap; the block's neutral-close restore stays `[Tool Output]`-scoped; an escape-only reason emits no row). The live turn output is **blank-line grouped** (one blank before each call's begin block — per call; one blank before the grouped tail reason block, none inside it; one blank before the post-status group). Presentation-only. | `spec.md` FR-001..FR-008; `research.md` D1–D6. |
| MODIFY | `specs/truth/techstack.md` — **Agent command tool (`execute_command`)** row | The round-038 "sibling `[Tool Reason]`/`[Tool Result]`/`[Tool Action]` values remain unsanitized — a live-issue forward item" note is **closed**: the policy is now generalized (see the Agent tool loop row), and the row points at **ADR 0008** (which supersedes ADR 0007). | `spec.md` FR-001/FR-003; `research.md` D1/D8. |
| ADD | `docs/decisions/0008-terminal-safe-lines-and-blank-line-grouping.md` | Records the generalized policy (single-owned sanitize; neutral-close stays `[Tool Output]`-scoped; sanitize-before-cap) and the blank-line grouping policy; carries D1–D6 + the two recorded reference divergences. | `spec.md` FR-003/FR-005; `research.md` D1–D6/D8/D9. |
| SUPERSEDE | `docs/decisions/0007-terminal-control-sanitization.md` — `Status` → `Superseded by 0008` | ADR 0007's recorded scope boundary ("closed on `[Tool Output]` only") is closed by this round, so 0007 is superseded rather than edited (its own immutability clause; the ADR 0005 → 0006 precedent). | `research.md` D9. |

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
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — **1** new `Given` (an escape-bearing reason fixture) + **7** new `Then` (3 control-free line assertions + 4 blank-line-position assertions) + a round-039 note (two sections: `## Given (round 039)`, `## Then (round 039)`) | ADD rows matching exactly one step each (`dsl-exact-one-match`); `dsl-single-authority` preserved (new rows, no duplication). | `spec.md` FR-001/FR-006/FR-007/FR-008; `research.md` D7. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0008-terminal-safe-lines-and-blank-line-grouping.md` (+ the `docs/decisions/README.md` index row) | Records the generalized policy (single-owned sanitize applied by every `[Tool …]` formatter; neutral-close stays `[Tool Output]`-scoped; sanitize-before-cap) and the blank-line grouping policy; carries D1–D6 and the two recorded reference divergences. | The round's generalized policy + the recorded non-change + the grouping divergence need a durable, citable home (not a frozen plan package); `research.md` D1/D8/D9. |
| SUPERSEDE | `docs/decisions/0007-terminal-control-sanitization.md` — `Status` flipped to `Superseded by 0008` | ADR 0007's recorded scope boundary ("the leak class is closed on the `[Tool Output]` surface only") is closed by this round, so 0007 is superseded rather than edited (its own immutability clause; the ADR 0005 → 0006 precedent). | `research.md` D9; the decisions README's immutability rule. |
