# Truth Delta: 060-domain-model-and-drift-gate

**Plan Package**: `specs/plans/060-domain-model-and-drift-gate`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — **Build & Tooling** (new **Domain model** row) | Records tellme's canonical domain model: the **three** `docs/domain-model/*.modelith.yaml` sources (product · quality · environment-management) rendered to `*.modelith.md` by the **modelith** fork (`@feat/self-domain-model`); the YAML-as-source / `.md`-generated rule; the **zero-tolerance `modelith-check` drift gate** (absent binary hard-fails); the "descriptive docs, not truth — truth wins" boundary; `modelith` a dev-tool binary (not a `go.mod` dep); POSIX-only; the **ADR 0011 D10 amendment**; and the **not adopted** `modelith-drift`/`modelith-layers`. | round 060 FR-007; `research.md` D1/D2/D3/D7/D8. |
| MODIFY | `specs/truth/techstack.md` — **Task runner** row | Extends the `verify` aggregate list to `… + verify-architecture + modelith-check + lint + vulncheck` (adds the new zero-tolerance member). | round 060 FR-004/FR-007; `research.md` D5/D7. |
| MODIFY | `specs/truth/techstack.md` — **Layer-discipline gate** row (recorded-divergence line) | Corrects the "tellme adopts the Go-guard form only … no modelith toolchain" prose to record that tellme now **has** a modelith toolchain (the domain model + drift gate) while still shipping **no** `modelith-layers` architecture gate (ADR 0011 D10 amended to that extent). `truth-current`: the row no longer contradicts the new toolchain. | round 060 FR-007; `research.md` D7. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: `tellme` has a single CLI end and **no** OpenAPI/HTTP surface; this round adds a docs model + a dev gate. | `spec.md` A5; `research.md` D10 — `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`/`tool_usage_record`/`turns_log_line` | No persisted/runtime state change: the domain model is a **docs artifact** (`docs/domain-model/**`), not runtime/persisted state; no table/enum/relationship/lifecycle changes. | `spec.md` A5; `research.md` D10. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI-interface behaviour changes — the models + drift gate are a **dev/docs surface** (`make verify`), not the `tellme` binary's CLI contract; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the model + `modelith lint`/`render --check` + the drift gate. | `spec.md` A3/A5; `research.md` D10 — the round-020/031/041/042/043/055 non-BDD-tooling precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0030-domain-model-and-modelith-toolchain.md` (+ the `docs/decisions/README.md` index row) | Records the adoption of the three modelith models + the toolchain + the zero-tolerance drift gate; **amends ADR 0011 D10** (its "no modelith toolchain" position is superseded for the model + drift gate — tellme still ships no `modelith-layers`); the descriptive-docs/truth-wins boundary; the lifecycle; and the RF-060-x forward items. | round 060 FR-007; `research.md` D7 — `docs/decisions/README.md` names "a project-level rule … other artifacts (or future rounds) depend on and must be able to cite". |
