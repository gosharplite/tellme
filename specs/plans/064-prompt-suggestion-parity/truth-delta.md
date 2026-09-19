# Truth Delta: 064-prompt-suggestion-parity

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **Clarify CLOSED — Q1 → A** (pool-only: deepen the recent-prompt candidate pool to the newest 50; the surfaced cap stays 10) locked by the operator (2026-09-20). Owner rows: `research`/`api`/`data` **recorded**; `dsl-refine` rows **recorded** (the phase has run). Implementation: `promptPoolDepth = 50` split from `maxSuggestions = 10` (`internal/app/suggestions/service.go`); unit pin + E2E carrier; gates green.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Prompt suggestion engine* | the row now states the recent-prompt **candidate pool** is the newest **50** distinct prompts (`promptPoolDepth`), the surfaced list stays **capped at 10** (`maxSuggestions`), and the two are **distinct single-owner constants**; corrected the stale "active session" prompt-source clause (the user-global shared log is the only prompt source) and dropped "session" from the mechanism cell. | `spec.md` FR-001…FR-005, S-1/S-3 (Q1 → A); `research.md` D1/D2/D3 |
| ADD | `docs/decisions/0034-prompt-suggestion-pool-depth.md` (+ the index row) | the decision: the pool deepens to the newest 50 (the reference's `LoadTopN(ctx, 50)`), the cap stays 10, the kept divergences (tool source, `~/.tellme/` log, no `WorkspacePolicy`, no compaction) and the unadopted forward items (empty-query-first-5, session source). | `spec.md` S-1/S-4/S-5; `research.md` D6 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the shared prompt-log record (`prompt_log_entry`) is unchanged; only the read depth into it moves. | `spec.md` A3 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/prompting-with-suggestions.feature` | ADD-ed the round-064 depth Rule (*The interactive prompt searches a deep window of recent prompts*) + its Example (a match older than the newest 10 is still offered), reusing the existing Given/Then sentences plus the new deep-log Given. | `spec.md` US1; acceptance `finding-a-recent-prompt-beyond-the-shallow-window.feature` (`acceptance-coverage`) |
| ADD | `specs/truth/features/cli/chat/dsl.md` — `the shared prompt log holds {count} newer prompts about other topics` (Given) + a round-064 note | the depth-distinguishing fixture Given (append `{count}` unrelated records so a match falls beyond the newest-10 window but inside the newest-50 pool) and the module note (the deepened pool, the unchanged cap, the kept disciplines). **No new Then row** — the existing `the interactive prompt offers the recent prompt "{prompt}"` row is reused. | `spec.md` FR-001/FR-002/FR-004 (`dsl-exact-one-match`; each new step matches one row) |
| NOOP (checked) | the interface-root `dsl.md` + the other `chat/**` features | no cross-module row changed; the workspace/tool/accept/over-long Rules are untouched (depth-only change). | `spec.md` S-2/I-2 (`dsl-single-authority` holds) |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): expected unchanged unless a suggestion-source fact's wording changes; `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
