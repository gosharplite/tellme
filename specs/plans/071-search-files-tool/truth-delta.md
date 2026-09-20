# Truth Delta: 071-search-files-tool

**Plan Package**: `specs/plans/071-search-files-tool`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-20)**. Clarify **folded** (Q1 → A literal/`is_regex`; Q2 → A binary/line-token skip, no ignore list; Q3 → A budget + 100 cap + 500 trim, sorted). `/axb-api-plan` + `/axb-data-plan` record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — *In-file content search tool (`search_files`)* | a ninth agent tool: a bounded, deterministic in-file content search (literal default + `is_regex`; `path:line: text` sorted path-then-line; the byte budget + a 100-match cap + a 500-char trim; binary skipped; no directory ignore list; no `SafePath`); three recorded divergences from the reference | `spec.md` US1/US2/US3, FR-001…FR-009; `research.md` D1–D7 |
| MODIFY | `specs/truth/techstack.md` — *Agent tool schemas* | the shared `resourceSchema`-backed enumeration gains `search_files` (the ninth tool) | `spec.md` FR-005/I-7 |
| MODIFY | `specs/truth/techstack.md` — *Tool-usage accounting* | the recordable union (`--tool-usage`) gains `search_files` | `spec.md` FR-006/I-6 |
| ADD | `docs/decisions/0043-search-files-tool.md` (+ index row) | the tool, the design-intent justification, the three divergences, and the Q1–Q3 locked choices | `spec.md` SC-004; `research.md` D7 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — `Tool` entity | the surface now names the reader family incl. `search_files` (nine tools with the vision-gated `read_image`); rendered via `make modelith-render` | ADR 0041 (the model is load-bearing); `modelith-check` green |
| NOOP (checked) | every other techstack row | no config key, no wire change, no new dependency | `spec.md` I-3/I-8 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-record change; the tool result is in-flight only. | `plan.md` §1 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/searching-file-contents.feature` | a new executable journey: 4 Rules (a query found across files; a no-match result; an over-budget search trimmed; an invalid pattern reported) | `spec.md` US1/US2/US3, FR-010 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | 4 Given rows + 5 Then rows (round-071 section) + the round-071 note | `spec.md` FR-010 |
| MODIFY | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` | the offered base set is **eight** (adds `search_files`); the vision-gated form is nine | I-5/I-6; the single-sourced `registeredToolNames()` |
| NOOP (checked) | `specs/truth/features/cli/dsl.md` (interface root) | no new cross-module row; the new steps are module-scoped | `plan.md` W3 |
