# Truth Delta: 047-application-import-ceiling-gate

**Plan Package**: `specs/plans/047-application-import-ceiling-gate`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: `/axb-system-analysis` has landed — `plan.md` records **0** system interfaces + **no waves**; `/axb-api-plan`, `/axb-data-plan` and `/axb-dsl-refine` are ratified as **NOOP (checked)** below (their owners inspected the surfaces and confirmed no change). The round's truth changes are the **techstack MODIFY** (Layer-discipline gate row) + the **ADR 0016 ADD**.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | The row (round-042/046 text) records the four-rule predicate (RULE-A/B/C/D) + the header-only baseline (**0**). Round 047 adds **RULE-E — the application import ceiling**: for a governed application tier (`internal/app/**`, `internal/cli`) an `internal/**` import is a violation unless in the **sanctioned set** (`internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`); the set is a **normative section of the tier table** (ADR 0011 D7), **default-deny**, and **fail-on-stale** (an unused sanctioned entry fails); violations are **deduped by edge** with RULE-A/B/C. It **rides this same target** (no new Makefile member) and reuses the baseline ratchet: the **3** currently-unsanctioned edges (`internal/cli → internal/{agent,ui,ui/tui/prompt}`) are **baselined** (RULE-A/B/C stay **0**), to be removed one-for-one by the R5 de-coupling slices. Records the scope boundary (only `internal/**` imports by the application tiers; stdlib allowed; third-party out of scope) and cites **ADR 0016**. | `truth-current`; round 047 FR-009; `research.md` D1/D3/D4/D5/D7. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the row's `verify` aggregate list already names `verify-architecture` (added in round 042). RULE-E rides that target — **no new member**, so the row is unchanged. | round 047 FR-009; `research.md` D2. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface (`specs/truth/` = only `data/`, `features/`, `techstack.md`). This round adds a build-pipeline rule (RULE-E); it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5; `research.md` D8 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes: no persisted/runtime-state change — **the baseline is a repo artifact** (a committed text file), not runtime/persisted state, and the gate reads no state at run time. | `spec.md` A5; `research.md` D8; the baseline is a repo artifact, not runtime state |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Inspected: the gate is a **dev surface** (`make verify`), not the `tellme` binary's CLI contract — no feature Rule, Example, step, or `DSLRow` changes (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the gate + its self-tests. | `spec.md` A2/A6; `research.md` D8 — the rounds 020/031/036/041/042/043/044/045/046 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0016-application-import-ceiling.md` (+ the `docs/decisions/README.md` index row) | Records **RULE-E** (the application import ceiling): the sanctioned set, default-deny, the **fail-on-stale allow-list**, the baseline/ratchet reuse (3 edges now → 0 across the R5 slices), **what the rule is *not*** (application tiers only; `internal/**` imports only — stdlib allowed, third-party out of scope), and its relation to ADR **0011** (the tier table extended, the ratchet reused — **not superseded**: RULE-A/B/C/D stand) and ADR **0013**. | round 047 FR-008; `research.md` D7 — a project-level rule the R5 de-coupling slices must cite. |
