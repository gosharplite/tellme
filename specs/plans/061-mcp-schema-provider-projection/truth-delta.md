# Truth Delta: 061-mcp-schema-provider-projection

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Truth Root**: `specs/truth`
**Anchor issue**: [#127](https://github.com/gosharplite/tellme/issues/127)

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify CLOSED** (3 decisions, one at a time) — **CQ-1 → C** (layered: a `NormalizeMCPSchema` vendor-extension floor + a Gemini-adapter projection) · **CQ-2 → ii** (a named, **empirically-verified** supported-key allowlist, **default-deny**) · **CQ-3 → i** (a **new ADR 0031** amending ADR 0025 + the round-056 declaration-row MODIFY). No residual `NEEDS CLARIFICATION`; the package is ready for the owner phases.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/techstack.md` — MCP client / tool-schema rows | Anticipated: record the provider-supported schema surface + the projection. | `spec.md` FR-002/FR-007; A4 |
| (pending) | `docs/decisions/00NN-*.md` (+ index row) | Anticipated: **ADR 0031** (the provider-supported schema surface) **or** an ADR-0025 amendment. | `spec.md` FR-009; CQ-3 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/` (**no `contracts/**`**) | Anticipated NOOP: single CLI end, no OpenAPI/HTTP surface change. | `spec.md` A3 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/**` | Anticipated NOOP: no persisted/in-memory domain-state change (the projection is a pure in-flight transform). | `spec.md` A3 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/chat/dsl.md:364` + `chat/using-tools-from-a-remote-mcp-server.feature` | Anticipated **MODIFY**: the round-056 "verbatim" declaration rule gains the provider-wire carve-out (unsupported keywords are projected away for a family whose wire cannot carry them; the declared arguments survive). | `spec.md` FR-001/FR-009 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `docs/decisions/` (next free number **0031**) + `docs/decisions/README.md` | Anticipated **ADD** or an **ADR 0025** amendment — the provider-supported schema surface + the projection rule; §Forward for the recorded residuals (e.g. the standard-but-unsupported keyword class if CQ-2 → i). | `spec.md` FR-009; CQ-3 |
