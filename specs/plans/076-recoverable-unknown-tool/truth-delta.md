# Truth Delta: 076-recoverable-unknown-tool

**Plan Package**: `specs/plans/076-recoverable-unknown-tool`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **skeleton (initialized by `/axb-specify`, 2026-09-22)** — owner rows pending. Clarify resolved at specify time (**0 questions** — not escalated; see `spec.md` §Clarify strategy). Expected owners: `/axb-technical-research` (the fold-back rule + the FR-010 phrase/exit-7 **narrowing** + an ADR + `techstack.md` row), `/axb-dsl-refine` (the E2E interface truth under `specs/truth/features/cli/chat/**`); `/axb-api-plan` + `/axb-data-plan` expected `NOOP`.
>
> **Anchor**: [#154](https://github.com/gosharplite/tellme/issues/154) — the round's DoD.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/techstack.md` — *(TBD row)* | _pending_ — the unknown-tool-name fold-back rule; the FR-010 phrase/exit-7 **narrowing** (bound reached / cap exhausted / no tools registered) | `spec.md` US1/US2, FR-001…FR-006; `research.md` D-x |
| _pending_ | `docs/decisions/00NN-*.md` (+ index row) | _pending_ — the ADR recording the recoverable-fold-back decision, the per-turn cap, and the narrowed incomplete contract | `spec.md` SC-003; `research.md` D-x |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | _expected NOOP_ — single CLI end; no OpenAPI/HTTP surface | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/**` | _expected NOOP_ — no persisted-state change (a per-turn counter is runtime-only) | `spec.md` 關鍵實體 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/chat/**` + `chat/dsl.md` | _pending_ — an interface journey that drives the loop to request an unknown tool name and asserts the fold-back + continuation, and the per-turn cap termination | `spec.md` US1/US2, FR-001…FR-006 |
