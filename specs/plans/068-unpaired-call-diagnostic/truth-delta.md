# Truth Delta: 068-unpaired-call-diagnostic

**Plan Package**: `specs/plans/068-unpaired-call-diagnostic`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify ESCALATED — 2 questions, BOTH RESOLVED (Q1 → A: a `[Tool …]`-class `stderr` line, terminal-gated, **not** routed to `turns.log`; Q2 → A: **informational**, the turn proceeds).** No `NEEDS CLARIFICATION` remains. The residual technical choice (the detection seam + exact rendering) defers to `/axb-technical-research` (S-6).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the `M < N` unpaired-call account is surfaced as a plain `stderr` `[Tool Warning]` diagnostic via the family-neutral `llm.UnpairedToolCalls` (adapter delegates) + a CLI gateway decorator; informational; not in `turns.log`; defensive (no live producer today). | `spec.md` US1 (FR-001…FR-007); `research.md` D1/D2/D3 |
| ADD | `docs/decisions/0038-unpaired-call-diagnostic.md` (+ index; an annotation on ADR 0037 §Forward RF-067-1) | the diagnostic + the seam. **Delivers ADR 0037 RF-067-1.** | `spec.md` A4; `research.md` D7 |
| NOOP (checked) | the OpenAI-compatible rows | untouched (the account is always empty there; I-1). | `spec.md` I-1/S-5 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `spec.md` A4 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted shape change (the diagnostic is transient; `turns.log` is a rendered-text file, not a modelled record). | `spec.md` A4 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (recorded narrowing) | `specs/truth/features/cli/**` + `dsl.md` | **NOOP** — the diagnostic has **no hermetic producer** (`agentloop.go` appends one result per call ⇒ `M == N` always), so no godog Example can drive it; an Example-less Rule is forbidden (RF-063-10 retired). Carriers are the domain/ui/cli pins (the round-059 narrowing class). | `spec.md` A5; `plan.md` §3 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive): *expected MODIFY/NOOP* — decided when the diagnostic's domain concept (if any) is fixed; `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; every new/changed step must match exactly one `DSLRow`.
