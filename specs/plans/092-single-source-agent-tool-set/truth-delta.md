# Truth delta — round 092 `092-single-source-agent-tool-set`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop` proves the
area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*Composition root (dependency injection)* row) | Add a one-line note: the **base** agent-tool composition (readers · `search_files` · write pair · `execute_command` · `list_skills`) is **single-owned** by `internal/infrastructure/tools.NewAgentBaseTools` (round 092; **ADR 0062**), consumed by the composition root (`cmd/tellme.agentTools()`) and the e2e harness (`tests/e2e/steps.registeredToolNames()`) so they cannot drift; the `vision` gate + the family-aware ceiling resolution stay in the root (ADR 0039 D2/D3). | Record the placement decision (the round's design question) on the techstack surface; **no technology change**. |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. | No new dependency/technology (I-3). |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**` (incl. `chat/dsl.md` the offered-set `集合` cell, `chat/offering-the-agent-tools.feature`) | Checked — **no change**. The offered set + offer order are **byte-identical**; no `.feature` step text changes. | Behaviour identity (I-2 / NFR-001); a composition-placement refactor does not touch executable truth. |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | Refactor-only round. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0062-*.md` | **ADR 0062** — single-source the base agent-tool composition in `internal/infrastructure/tools`; the root keeps the capability policy + registry construction. | The placement decision the issue asked to record (FR-001 / DoD). |
| MODIFY | `docs/decisions/README.md` | Add the **ADR 0062** index row (status `Accepted`). | `verify-adr-index` (`adr-index-consistent`). |
| ADD | `internal/infrastructure/tools/agentbase.go` | The canonical base-set owner `NewAgentBaseTools(sink)`. | FR-001 single source. |
| MODIFY | `cmd/tellme/deps.go`, `tests/e2e/steps/tool_usage.go` | Delegate the base-set composition to the canonical owner; the capability gate / union unchanged. | FR-001 — derive, not re-implement. |
| ADD | `cmd/tellme/deps_agentbase_test.go` (carrier), `tests/e2e/steps/tool_usage_test.go` (carrier) | Binding carriers: the production base set and the e2e enumerator each equal the canonical owner (W-1/W-3). | FR-002/FR-003 — the round-090 equality-of-surfaces carrier shape. |
| NOOP | `docs/domain-model/**` | Checked — **not modelled** (ADR 0041 escape hatch). The model already records the tool surface; a composition-placement refactor changes no entity/invariant/scenario. | A refactor touches no modelled behaviour. |
