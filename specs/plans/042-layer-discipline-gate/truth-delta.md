# Truth Delta: 042-layer-discipline-gate

**Plan Package**: `specs/plans/042-layer-discipline-gate`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling** table | Adds a **Layer-discipline gate** row: the pinned layer ranking, the `go list`-based import-direction guard (`Makefile` `verify-architecture`), the committed sorted **baseline** (8 known violations), the **fail-on-stale** ratchet policy, hermetic/host-independent/stdlib-only, and the `modelith-layers` recorded divergence. | round 042 FR-009; `research.md` D1–D3/D5–D8. |
| MODIFY | `specs/truth/techstack.md` — **Task runner** row | Extends the `verify` aggregate list to `verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + verify-architecture + lint + vulncheck` (adds the new member; also corrects the pre-existing round-032 omission of `verify-mcp-sdk-confinement`). | round 042 FR-009/FR-010; `research.md` D5. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round adds a build-pipeline gate + a repo baseline file; it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A6; `research.md` D9. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`/the `~/.tellme/*.jsonl` shapes | No persisted/openruntime state change: the **baseline** is a **repo artifact** (a committed text file), not runtime/persisted state, and the gate reads nothing at run time. | `spec.md` A6; `research.md` D9. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI interface behaviour changes — the gate is a **dev surface** (`make verify`), not the `tellme` binary's CLI contract; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the gate + its self-test. | `spec.md` A3/A6; `research.md` D9 — the round-020/031/041 non-BDD-tooling precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked, recorded) | `docs/decisions/*` + `docs/decisions/README.md` | Inspected: R1 settles **no** decision that other artifacts must cite as a superseding rule; the layer ranking + baseline policy live in `techstack.md`. The [#92](https://github.com/gosharplite/tellme/issues/92) ADR obligation attaches to **R3**'s yield policy, not here. | `spec.md` A5; `research.md` D8 — deferring keeps R1 minimal. |
