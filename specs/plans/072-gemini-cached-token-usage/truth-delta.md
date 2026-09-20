# Truth Delta: 072-gemini-cached-token-usage

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-20)** — `research.md` D1–D7 + **ADR 0044** + `techstack.md` MODIFY ×2. Clarify **not escalated** (0 questions). `/axb-api-plan` + `/axb-data-plan` record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the adapter decodes `cachedContentTokenCount` → `Usage.CachedTokens` and `thoughtsTokenCount` → `Usage.ThinkingTokens` (floored at 0); the disjointness is family-specific (no subtraction for Gemini) | `spec.md` US1/US2, FR-001/FR-002; `research.md` D1–D4 |
| MODIFY | `specs/truth/techstack.md` — *Post-turn status lines* | the Gemini family now supplies `H`/`Th`, so `M`/`H`/`Th`, the cost, the session totals, and `tokens.log` are correct for it too (was `H: 0`, ~10× over-charge) | `spec.md` US3, FR-007 |
| ADD | `docs/decisions/0044-gemini-cached-token-usage.md` (+ index row) | the Gemini usage decode + the family-specific `thoughtsTokenCount` disjointness rule | `spec.md` SC-002/SC-004; `research.md` D7 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — `UsageRecord` | the cached/thinking counts are provider-reported and family-decoded (one sentence); rendered via `make modelith-render` | ADR 0041 (the model is load-bearing) |
| NOOP (checked) | the pricing/`ComputeCost` rows | the cost formula is unchanged; only the input is corrected | `spec.md` I-2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | The persisted `UsageRecord` **shape** (`tokens.log`) is unchanged; only its values become correct. | `spec.md` I-5; `plan.md` §1 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature` | ADD a Rule + Example: a Vertex turn that reused most of its input and reasoned reports `1026 missed, 389538 cached, 100 completed, 4096 reasoning` | `spec.md` US1/US2, FR-007 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | ADD the round-072 Gemini-detailed-usage Given row + note | `spec.md` FR-010-analogue |
| NOOP (checked) | `specs/truth/features/cli/dsl.md` (interface root) | no new cross-module row; the new step is module-scoped | `plan.md` W3 |
