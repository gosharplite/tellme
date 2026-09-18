# Truth Delta: 050-agentloop-construction-inversion

**Plan Package**: `specs/plans/050-agentloop-construction-inversion`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: `/axb-technical-research` (D1–D12 + **ADR 0019** + the `techstack.md` MODIFY rows below) has **landed**. The api/data/dsl-refine rows are the owners' records (each inspected its surface). **Clarify round 1 CLOSED**: **Q1 → A** (port-only inversion; the `ui` wiring stays in the CLI) · **Q2 → (i)** (domain `Loop` interface + `LoopSpec` + a func-typed factory; adapter `agent.NewLoop`) · **Q3 → (a)** (the CLI keeps supplying `Lines` + the observer, domain-typed) · **Q4 → A** (RULE-E is merged-graph, so the CLI **test** surface also leaves `internal/agent` — an in-package fake `agentport.Loop`).
>
> **Round shape (Q1 → A)**: this is the **sub-slice 2** of the `cli → agent` de-coupling — the **baseline-moving** round. The RULE-E ratchet moves **2 → 1** (**two** removals: the RULE-E baseline line + the RULE-F `couplingSurface` key — ADR 0018 fold **F-4**); the `internal/cli -> internal/ui` line **and** its 19-identifier RULE-F surface stay **byte-identical**.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | **Round 050 / ADR 0019**: the RULE-E baseline figure goes **2 → 1** (the `internal/cli -> internal/agent` line is removed; only `internal/cli -> internal/ui` remains, retained by design). The **RULE-F** `couplingSurface` no longer carries the `internal/cli -> internal/agent` key (`{AgentLoop}`); the sole remaining governed edge (`→ ui` ⇒ 19 identifiers) stays tracked and byte-identical. Records the **two** ratchet removals (ADR 0018 fold F-4) and the sub-slice 2 mechanism (the `agentport.Loop`/`LoopSpec`/`LoopFactory` port + the `agent.NewLoop` adapter). | `truth-current`; round 050 FR-008; `research.md` D7/D8; review F-4. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | Notes that the loop's **construction/execution** is now inverted behind an injected **domain port** (`internal/domain/agent` declares `Loop`/`LoopSpec`/`LoopFactory`; the adapter `agent.NewLoop` builds the `AgentLoop`, bound at `cmd/tellme` through a func-typed `deps.Dependencies` field), so `internal/cli` names **no** `internal/agent` identifier. The CLI supplies `Lines`/`Observer` as domain-typed `LoopSpec` inputs; behaviour unchanged. | round 050 FR-001/FR-008; `research.md` D2/D3/D4/D6. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the `verify` aggregate already names `verify-architecture`. The round adds **no** Makefile target — the gate rides that target — so the row is unchanged. | round 050 NFR-002; `research.md` D10 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface (`specs/truth/` = `data/`, `features/`, `techstack.md`). This round is an internal construction inversion; it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5; `research.md` D9 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes: **no** persisted/runtime-state change — the refactor moves a construction (the loop object + its wiring), not data. | `spec.md` A5; `research.md` D9 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Inspected: a `cli`-side construction refactor is **not** a `tellme` CLI-contract change — no feature Rule, Example, step, or `DSLRow` changes (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the gate + the identifier-count check + the adapted unit seams. | `spec.md` A2/A6; `research.md` D9 — the rounds 020/031/036/041–049 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0019-agentloop-construction-inversion.md` (+ the `docs/decisions/README.md` index row) | Records the port-only inversion (Q1-A), the **`Loop`/`LoopSpec`/`LoopFactory`** domain port + the **`agent.NewLoop`** adapter (Q2-i), the CLI-supplied domain-typed `Lines`/`Observer` (Q3-a), the **two** ratchet removals (RULE-E line + RULE-F key; F-4), the **retained `→ ui` edge**, **what the change is *not***, and its relation to ADR 0011/0016/0017/0018/0013. | round 050 FR-007; `research.md` D7 — a structural change the `→ ui` slice must cite. |
