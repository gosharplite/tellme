# Truth Delta: 078-provider-transport-retry

**Plan Package**: `specs/plans/078-provider-transport-retry`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + this checklist initialized. Clarify **not escalated (0 questions)**: the theme and the retry schedule are operator-given, and the retryability predicate (transport + HTTP 429/5xx) is **operator-locked** in-session; the residual choices (mechanism / seam / typed classification / hermetic test seam / record shape) are `/axb-technical-research` decisions. **Truth owners not yet run** — the tables below are empty until `/axb-technical-research` and `/axb-dsl-refine` record their entries.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — provider-failure rows (incl. the round-030 *Provider output-cap truncation guard* row, which currently asserts "tellme adds **no** retry layer") | MODIFY: record the bounded transport-retry policy (retryable class = transport + 429/5xx; schedule 1 s → 3 s; ≤2 retries; exhaustion reuses the frozen phrase + exit 6), and **reconcile** the "no retry layer" statement | `spec.md` US1/US2, FR-001…FR-007; `research.md` |
| *(pending)* | `docs/decisions/0050-*.md` (+ index row) | ADD: the decision record — retryability predicate, typed classification, the single-seam decorator, the accounting semantics, the hermetic test seam, declined alternatives | `spec.md` SC-001…SC-005 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 *(to be recorded)* |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/**` | Likely **NOOP** — the retry changes provider-failure handling, not persisted state (a retried call still writes one usage record / one history entry on success). Confirm in `plan.md` §5. | `spec.md` 關鍵實體 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | MODIFY: a new **Rule** (transient provider failures are retried before failing) with Examples — retry-then-succeed (exactly 2 calls), retry-exhausted (exactly 3 calls → phrase + exit 6), and a non-retryable failure (exactly 1 call) | `spec.md` US1/US2, FR-001…FR-005 |
| *(pending)* | `specs/truth/features/cli/chat/dsl.md` | MODIFY: Given rows scripting the fake provider's transient failure / fail-once-then-succeed, and Then rows asserting the call count + the frozen phrase + exit 6 | `spec.md` US1/US2 |
