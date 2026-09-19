# Truth Delta: 061-mcp-schema-provider-projection

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Truth Root**: `specs/truth`
**Anchor issue**: [#127](https://github.com/gosharplite/tellme/issues/127)

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify CLOSED** (3 decisions, one at a time) — **CQ-1 → C** (layered: a `NormalizeMCPSchema` vendor-extension floor + a Gemini-adapter projection) · **CQ-2 → ii** (a named, **empirically-verified** supported-key allowlist, **default-deny**) · **CQ-3 → i** (a **new ADR 0031** amending ADR 0025 + the round-056 declaration-row MODIFY). **CQ-4 approved** — the allowlist is established by a **live probe** (`ait-comment` → `websc-dev-433809`, declaration-only calls; the table is recorded verbatim in ADR 0031; the probe is research-time, not a gate). No residual `NEEDS CLARIFICATION`; the package is ready for the owner phases.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-schema normalization* | **Round 061 (ADR 0031):** the normalizer applies a **vendor-extension floor** — `x-…`/`$schema` dropped recursively, for **every** family (S-6); the declared arguments are untouched. | `spec.md` FR-002, S-1/S-6 |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-call reason envelope* | **Round 061 (ADR 0031):** the round-056 "relayed verbatim" claim is **qualified at the provider wire** (floor for all families; projection for a closed-wire family). | `spec.md` FR-009, S-3 |
| ADD | `specs/truth/techstack.md` — *Tool-declaration schema projection (Vertex/Gemini)* | The recursive, default-deny, fail-closed projection + the measured supported surface (the probe table is in ADR 0031). | `spec.md` FR-001/FR-002/FR-004/FR-005, S-1/S-2/S-4 |
| ADD | `specs/truth/techstack.md` — *Tool-declaration schema-projection gate* | The hermetic unit pin over the production declaration path (red-capable) + the E2E carrier. | `spec.md` FR-008, S-5 |
| ADD | `docs/decisions/0031-provider-supported-schema-surface.md` (+ index row; ADR 0025 §Status pointer) | The rule + the verbatim probe table + §Forward RF-061-1…6. | `spec.md` FR-009, CQ-3 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted/in-memory domain-state change: the projection is a pure in-flight transform and the floor mutates an already-discovered schema in place. | `spec.md` A3 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — the round-056 declaration row | The provider-wire carve-out recorded on the row (the floor + the closed-wire projection; the declared arguments survive). | `spec.md` FR-001/FR-009, S-1/S-3 |
| ADD | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `chat/dsl.md`) | A new Rule *A server's argument marks never reach the provider* with 3 Examples + 5 DSL rows (2 Given, 3 Then). | `spec.md` US1/US2/US3 (`acceptance-coverage`) |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0031-provider-supported-schema-surface.md` + `docs/decisions/README.md` (+ the ADR 0025 **Status** pointer) | The provider-supported schema surface: the floor + the default-deny projection, the measured table, and §Forward RF-061-1…6. **Amends ADR 0025 D1** at the provider wire (CQ-3 → i). | `spec.md` FR-009, S-3 |

## Review folds (PR [#128](https://github.com/gosharplite/tellme/pull/128), review `5740291616`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Tool-declaration schema projection (Vertex/Gemini)* | The projection **normalizes value shapes** too (array `type` → its lone member + `nullable`; `enum` members → strings), and both new probe rows are recorded (ADR 0031 D2/D6a). | review **F-061-2** |
| MODIFY | `specs/truth/techstack.md` — *Tool-declaration schema-projection gate* | The gate reads the **owner** both directions (containment + coverage) and a native-declaration semantic-identity pin is added; the deny-lists are gone. | reviews **F-061-1**, **TD-061-2**, **RF-061-4** |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-schema normalization* | The floor is **structure-aware** (keyword positions only) and the postcondition is re-asserted. | review **B-061-1** |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — the round-061 args + containment rows | The args row is envelope-scoped (declared `MCP_PAYLOAD` properties **with descriptions**); the containment row states the **owner-set** check, declaration-scoped; the round-056 row records the value-shape normalization. | reviews **F-061-1/2/3** |
| MODIFY | `docs/decisions/0031-provider-supported-schema-surface.md` | D6a (value shapes) · D6b (structure-aware floor + re-assertion) · the shape-probe rows · §Forward RF-061-7…9. | reviews **B-061-1**, **F-061-2**, **TD-061-1/2**, nits |

## Fold-verification folds (PR [#128](https://github.com/gosharplite/tellme/pull/128), verification `5740338531`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-schema normalization* | the floor traverses `$defs`/`definitions`/`dependencies` again (definition maps are name→schema maps). | review **R-1** |
| MODIFY | `specs/truth/techstack.md` — *Tool-declaration schema projection (Vertex/Gemini)* | the coercion emits only measured shapes (`nullable` only beside a `type`; `enum` only beside a scalar type; non-object subschema → `{}`). | reviews **R-2**, **R-3** |
| MODIFY | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` | the no-mark control gains an **executed** Then (`…:18`) — the F-061-3(1) carrier, folded after verification R-4. | review **R-4** (`truth-current`) |
| MODIFY | `docs/decisions/0031-provider-supported-schema-surface.md` | D6a′ (coerced-shape rows) · D6c (non-object → `{}`) · the `$defs` traversal · RF-061-10. | reviews **R-1**, **R-2**, **R-3**, nit 2 |

| MODIFY | `specs/truth/techstack.md` (projection row) | a non-object subschema degrades to `{}` at **every** schema-node position — incl. `oneOf`/`allOf` **elements** (R-5); the `additionalProperties` bool form is measured accepted at both levels. | review **R-5**, nit 2 |

| MODIFY | `specs/truth/techstack.md` (projection + gate rows) | the single owner is now a **key + value-kind** table; the gate checks shape too. | review **V-061-1** |

| MODIFY | `specs/truth/techstack.md` (projection row) | the `type`-array coercion re-checks the substituted member; a round-trip pin asserts the projection’s output satisfies the gate. | review **W-061-1** |