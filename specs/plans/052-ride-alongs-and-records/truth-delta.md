# Truth Delta: 052-ride-alongs-and-records

**Plan Package**: `specs/plans/052-ride-alongs-and-records`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: `/axb-technical-research` (D1–D7 + **ADR 0021** + the `techstack.md` rows below) has **landed**. **Clarify round 1 CLOSED**: **Q1** (one round closes #115 + #116) · **Q2 → (i)** (each `#116` record relocates to ADR 0021 §Records; #116 closes completed).
>
> **Round shape**: the two [#115](https://github.com/gosharplite/tellme/issues/115) ride-alongs (the command tool's construction-time `[Tool Output]` sink; the suggester's single-owned selection policy) + the [#116](https://github.com/gosharplite/tellme/issues/116) records' durable relocation. Both issues close on delivery.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent command tool (`execute_command`)** row | **Round 052 (closes [#115](https://github.com/gosharplite/tellme/issues/115) R-2; ADR 0021)**: the `[Tool Output]` sink is **constructor-injected** (`NewCommandTool(sink domaintools.OutputSink)`); the `toolOutputBox` pointer indirection, the `BindToolOutput` function, and the `deps.Dependencies.BindToolOutput` seam are **removed**; the agent registry is built through one sink-aware builder (`agentTools()` stays parameterless + read-free, assembling with a **nil** sink; the prompt path passes `prog.ToolOutput`). The block's rendering behaviour is unchanged. | `spec.md` US1 / FR-001…FR-004; `research.md` D1/D2. |
| MODIFY | `specs/truth/techstack.md` — **Interactive TUI prompt (`-i`)** row | **Round 052 (R-1; ADR 0021)**: the selection policy gets a single named owner — `suggester.set(items []string, cursor int)` takes the cursor explicitly and no longer resets it internally; the refresh caller passes the `noChoice` sentinel, so the reset is a **caller** decision (rendering byte-unchanged). | `spec.md` US2 / FR-005; `research.md` D3. |
| MODIFY | `specs/truth/techstack.md` — **Prompt suggestion engine** row | **Round 052 (R-1; ADR 0021)**: the selection **reset** is owned by the caller (`set(items, cursor)`; the refresh passes `noChoice`) — no hidden reset inside `set`. | `spec.md` US2 / FR-005; `research.md` D3. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the `verify` aggregate already names `verify-architecture`. The round adds **no** Makefile target — the gate rides that target. | `spec.md` FR-008; `research.md` D7. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and no OpenAPI/HTTP surface. The round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`: no persisted/runtime-state change (a construction-seam refactor + a caller-owned cursor). | `spec.md` A5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/**/dsl.md` | Inspected: a `cli`-side refactor is not a tellme CLI-contract change (no Rule/Example/step/`DSLRow` change; the `[Tool Output]` literals + the suggestion rendering stay byte-identical). | `spec.md` A2/A6; the rounds 020/031/036/041–051 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0021-ride-alongs-and-records.md` (+ the `docs/decisions/README.md` index row) | Records D1–D4 (the construction seam + `agentTools()` parameterless; `BindToolOutput`/`toolOutputBox` deletion; `set(items, cursor)`; the records' relocation) and **hosts the three relocated `#116` records** (permanent E2E narrowing · one concurrent block · `End`-while-write-stalled) with their provenance, plus a §Forward (RF-52-1…3). | `spec.md` US3 / FR-006/FR-007; `research.md` D4/D6 |
