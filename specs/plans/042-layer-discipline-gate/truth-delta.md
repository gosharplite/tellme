# Truth Delta: 042-layer-discipline-gate

**Plan Package**: `specs/plans/042-layer-discipline-gate`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling** table | Adds a **Layer-discipline gate** row: the pinned layer ranking; the **two-part predicate** (A direction · B application-target · C domain purity · D default-deny); the `go list`-based guard anchored to the module root and evaluated over the **`CROSS_TARGETS` union**; the committed sorted **baseline** (8 known violations) with the **fail-on-stale** ratchet; the acyclicity assertion; the tag-scope residual; a citation of **ADR 0011** as the normative host; and the `modelith-layers` recorded divergence. | round 042 FR-010; `research.md` D1–D12. |
| MODIFY | `specs/truth/techstack.md` — **Task runner** row | Extends the `verify` aggregate list to `verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + verify-architecture + lint + vulncheck` (adds the new member; also corrects the pre-existing round-032 omission of `verify-mcp-sdk-confinement`). | round 042 FR-010/FR-011; `research.md` D7. |
| MODIFY | `specs/truth/techstack.md` — **Layer-discipline gate** row (invocation) | Corrects the recorded invocation to the shipped one: `go vet -tags=arch ./tools/arch` + `golangci-lint run --build-tags=arch ./tools/arch/...` + **`go test -count=1`**` -tags=arch -run TestVerifyRealArchitecture ./tools/arch` (the earlier text omitted `-count=1`, i.e. documented the vacuous cached path review F-1 fixed; the tagged lint also covers the build-tagged guard). `truth-current`: the row now reflects the shipped mechanism. | PR #95 fold-review #3, Fold 1; `research.md` D2 (corrected in the same commit). |

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
| ADD | `docs/decisions/0011-layer-discipline-gate.md` (+ the `docs/decisions/README.md` index row) | Records the **layer rule** (the two-part predicate A–D: direction + application target rule + domain purity + default-deny) and the **baseline policy** (a fail-on-stale ratchet), the enumeration anchor, the `CROSS_TARGETS` union, the acyclicity assertion, the deterministic format, and the worked **8-entry** baseline; no existing ADR is superseded. | round 042 FR-010; `research.md` D7 — `docs/decisions/README.md` names "a project-level rule … other artifacts (or future rounds) depend on and must be able to cite": R2–R4 must cite the rule. *(Replaces the earlier "no ADR" position — review RF-1.)* |
