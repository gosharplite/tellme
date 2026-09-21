# Truth Delta: 078-provider-transport-retry

**Plan Package**: `specs/plans/078-provider-transport-retry`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)** (the theme/schedule are operator-given; the retryability predicate is operator-locked in-session). **`/axb-technical-research` RUN (2026-09-22)** — `research.md` D1–D9 + **ADR 0050** + `techstack.md` **MODIFY ×2 + ADD ×1** (see below). `/axb-api-plan` + `/axb-data-plan` record **NOOP**. `/axb-dsl-refine` **pending**.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Provider gateway port* | `llm.ProviderError` gains typed failure **facts** (`Status int`, `Transport bool`); the domain-owned predicate `llm.Retryable` classifies retryability without string matching | `spec.md` NFR-003; `research.md` D2 |
| MODIFY | `specs/truth/techstack.md` — *Provider output-cap truncation guard* | **Reconciles** the round-030 claim: the "tellme adds **no** retry layer (there is none to change)" clause is replaced — a truncation is **terminal / never retried**; the retry covers **transport/status** failures only | `spec.md` US2, FR-005; `research.md` D1/D8 |
| ADD | `specs/truth/techstack.md` — *Provider request retry (transient transport)* | the bounded retry policy: retryable class = transport + 429/5xx; fixed `retryDelays = {1 s, 3 s}`; ≤2 retries; exhaustion reuses the frozen phrase + exit 6; one call/usage/frame; the `stderr`-only retry line; cancellation-aware; the `TELL_ME_FORCE_RETRY_DELAY_MS` hermetic seam; a recorded divergence from the reference's resilience stack; not modelled | `spec.md` US1/US2, FR-001…FR-007; `research.md` D1–D9 |
| ADD | `docs/decisions/0050-bounded-provider-transport-retry.md` (+ index row) | the decision record: the retryable predicate, the typed classification, the single-seam decorator, the schedule/count constants, the accounting semantics, the retry diagnostic, cancellation, and the declined alternatives | `spec.md` SC-001…SC-005; `research.md` |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state change — a retried call still writes one usage record / one history entry on success; a failed attempt writes nothing. | `spec.md` 關鍵實體; `research.md` D7 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | MODIFY: a new **Rule** (a transient provider failure is retried before failing) with Examples — retry-then-succeed (exactly 2 calls), retry-exhausted (exactly 3 calls → phrase + exit 6), and a non-retryable failure (exactly 1 call) | `spec.md` US1/US2, FR-001…FR-005 |
| *(pending)* | `specs/truth/features/cli/chat/dsl.md` | MODIFY: Given rows scripting the fake provider's transient drop / fail-once-then-succeed, Then rows asserting the call count + the frozen phrase + exit 6; **and reconcile the round-030 note** ("tellme adds **no** retry layer — there is none to change") which this round makes false | `spec.md` US1/US2; `research.md` D8 |
