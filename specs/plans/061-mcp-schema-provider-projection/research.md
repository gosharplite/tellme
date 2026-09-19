# Research — Provider-wire projection of MCP-relayed tool schemas (round 061)

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Anchor issue**: [#127](https://github.com/gosharplite/tellme/issues/127)
**Truth owner**: `/axb-technical-research` — updates `specs/truth/techstack.md`

## Problem (grounded)

`internal/infrastructure/llm/gemini/client.go:242-246` serialized each offered tool's `Parameters` (raw JSON) straight into `functionDeclarations[].parameters`. Vertex parses that field as the **closed** proto `google.ai.generativelanguage.Schema`, so **any** unknown key 400s the **whole** request (`the provider request failed`, exit 6). A remote MCP server's annotations (the GitHub server's `x-mcp-header`, 37/45 tools) are exactly such keys. `mcp.NormalizeMCPSchema` enforced well-formedness only (round 032), and round 031's gate covered tellme's own tool set only.

Reproduction is in the spec (index-perfect error path) and in the issue.

## Decisions taken upstream (`/axb-clarify`, one at a time → `spec.md` S-1…S-7)

| # | Decision |
| --- | --- |
| S-1 | **Layered**: a family-agnostic **floor** (`mcp.NormalizeMCPSchema` drops `x-…`/`$schema`) **+** a Gemini-side **projection**. |
| S-2 | The projection is **default-deny** over a **named, empirically-verified** allowlist (single owner). |
| S-3 | Governance: **ADR 0031 amends ADR 0025 D1** + the round-056 declaration truth row is MODIFY-ed. |
| S-4 | The allowlist is established by a **live probe** (operator-approved): declaration-only `generateContent` calls from the `ait-comment` env against project `websc-dev-433809`. |
| S-5 | The regression gate is a **hermetic unit pin**; **no new `make verify` member**. |
| S-6 | The floor reaches **every** family (the tolerant family's declaration changes too — deliberate, recorded). |
| S-7 | The projection is **silent** (ADR-recorded only). |

## The probe (measured, 2026-09-19)

Harness: `POST {url}/{model}:generateContent` with `contents` = one trivial user turn, `generationConfig.maxOutputTokens = 8`, and a **single** `functionDeclarations` entry whose `parameters` carries the keyword under test (property-level for property keywords, root-level for root keywords; the control probe carries none). Model `gemini-3.8-flash`; project `websc-dev-433809`; token via the operator's ADC/SA credential. Results (verbatim table in **ADR 0031 §D2**):

- **accepted** — `type`, `description`, `properties`, `required`, `items`, `enum`, `format`, `title`, `default`, `nullable`, `pattern`, `minimum`, `maximum`, `minLength`, `maxLength`, `minItems`, `maxItems`, `oneOf`, `allOf`, `additionalProperties` (root + property), `propertyOrdering`, `type: "null"`, a property with only `description`.
- **rejected** — any `x-…`, `$schema` (root), `$ref`, `$defs`, `definitions`, `const`, `examples`, `deprecated`, `readOnly`, `writeOnly`, `multipleOf`, `uniqueItems`; an unknown key at a **nested** level; and `anyOf` **whenever another schema keyword sits on the same node** (*"when using any_of, it must be the only field set"*).
- `anyOf` **alone** is accepted — **not** allowlisted (shape a projection cannot guarantee; ADR 0031 D3/§Forward).

**Design consequence**: the projection is **recursive** (nested marks/kets are rejected too) and **fail-closed** (a non-object schema → the freeform object).

## Alternatives considered

| Option | Verdict |
| --- | --- |
| **(A) Gemini-only projection** | Rejected (CQ-1): leaves "drop vendor noise" with no home and duplicates the rule per family. |
| **(B) Normalizer strip only** | Rejected (CQ-1): does not cover standard-but-unsupported keywords (`anyOf`, `const`, …) that also 400. |
| **(C) Layered floor + projection** | **Chosen** (CQ-1 → C, S-1). |
| **(D) Discovery-time skip of an offending tool** | Rejected: silently removes tools and makes tool availability provider-dependent. |
| Deny-list (drop only the known-bad keys) | Rejected (CQ-2 → ii): rots on the next provider revision; default-deny is the measured rule. |

## Reference note

`tell-me-go` converts a discovered MCP schema through a **typed per-family schema model**, so a vendor annotation cannot reach the Vertex payload — the same shape as (C)'s projection, but built as a full converter. tellme keeps its raw-JSON relay and adds the projection at the transport (smaller change, same guarantee).

## Risks / residuals (→ ADR 0031 §Forward)

- `anyOf` unions are dropped (RF-061-1) · the allowlist is a probe snapshot with no CI re-probe (RF-061-2) · the projection is silent (RF-061-3) · a native declaration using a rejected keyword would be trimmed silently (RF-061-4) · the floor changes the tolerant family's declaration (RF-061-5) · the probe ran against one model (RF-061-6).

## Truth updated by this phase

| Action | Spec | Summary |
| --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-schema normalization* | the vendor-extension floor (recursive, every family) |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-call reason envelope* | the "verbatim" claim qualified at the provider wire |
| ADD | `specs/truth/techstack.md` — *Tool-declaration schema projection (Vertex/Gemini)* | the projection + the measured surface |
| ADD | `specs/truth/techstack.md` — *Tool-declaration schema-projection gate* | the hermetic pin (S-5) |
| ADD | `docs/decisions/0031-provider-supported-schema-surface.md` (+ index row; ADR 0025 pointer) | the rule + the probe table |
