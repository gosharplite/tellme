# Feature Specification: provider-wire projection of MCP-relayed tool schemas (round 061)

**Feature Branch**: `061-mcp-schema-provider-projection`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from **issue [#127](https://github.com/gosharplite/tellme/issues/127)**. **Clarify CLOSED** (3 decisions, one at a time — **CQ-1 → C**, **CQ-2 → ii**, **CQ-3 → i**; recorded below). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)**:

> *"Take a look."* + the live failure paste (`provider dev: provider returned status 400: Invalid JSON payload received. Unknown name "x-mcp-header" at 'tools[0].function_declarations[7].parameters.properties[0].value.properties[2].value': Cannot find field. …`)
> *"File a GitHub issue first. Open round 061-*."*

**Anchor issue**: [#127](https://github.com/gosharplite/tellme/issues/127) — *Gemini rejects MCP-relayed tool schemas — a server vendor extension (`x-mcp-header`) fails every request with a 400 (declarations are relayed verbatim)*.

**Behaviour intent**: **MODIFY (the MCP tool-declaration relay + the Gemini transport's declaration serialization).** tellme today relays a remote MCP server's advertised input schema **verbatim** into the declaration it offers the model (round 056 / ADR 0025 D1: the server schema becomes `MCP_PAYLOAD`'s subschema, unmutated). That is correct for a JSON-Schema-tolerant family and **fatal** for the Vertex/Gemini family, whose `functionDeclarations[].parameters` is parsed as the **closed** proto message `google.ai.generativelanguage.Schema`: any unknown key (the GitHub MCP server's `x-mcp-header`) is rejected and the **whole** `generateContent` call fails, so **no turn reaches the model**. This round introduces a **provider-supported-surface projection** between the relay and the wire, so a server's dialect quirks cannot sink the request — while the model still sees the server's real arguments.

---

## Grounded in the current system *(measured 2026-09-19, `dev` @ `768dbec`)*

| Site | Current shape |
| --- | --- |
| `https://api.githubcopilot.com/mcp/` — `tools/list` | **45** tools; **37** annotate top-level properties with `"x-mcp-header": "owner"` / `"repo"` (e.g. `add_comment_to_pending_review`, `add_issue_comment`, `add_reply_to_pull_request_comment`, …). Measured live over Streamable HTTP (raw `initialize` + `tools/list`, 124 068 bytes, **74** `x-mcp-header` occurrences). |
| `internal/infrastructure/mcp/schema.go` — `NormalizeMCPSchema` / `normalizeSchemaObject` | Enforces **well-formedness only** — object root, object `properties`, `required ⊆ properties`. Unknown / annotation keywords are carried through **untouched**. |
| `internal/infrastructure/mcp/tool.go` — `Parameters()` → `mcpEnvelope` | Wraps the normalized server schema **verbatim** under `MCP_PAYLOAD` (composed structurally; the server schema is a `json.RawMessage`). The map marshals `MCP_PAYLOAD` before `reason`, so `MCP_PAYLOAD` is the envelope's `properties[0]`. |
| `internal/infrastructure/llm/gemini/client.go:242-246` | `params := td.Parameters` → `map[string]any{"name":…, "description":…, "parameters": params}` under `functionDeclarations`; the raw JSON is serialized onto the wire with **no projection** onto Gemini's supported `Schema` fields. |
| `internal/infrastructure/llm/openai/client.go:156` | The same `td.Parameters` passthrough for the OpenAI-compatible family — **tolerated** (permissive server-side JSON Schema), which is why only Gemini fails. |
| `specs/truth/features/cli/chat/dsl.md:364` + `chat/using-tools-from-a-remote-mcp-server.feature` | The **round-056 "verbatim" rule**: the offered `MCP_PAYLOAD` subschema must carry the server's advertised schema **verbatim (never mutated)** — the rule this round must qualify. |
| `tell-me-go` (reference, ADR-1378 line) | Converts a discovered MCP tool's schema through a **typed per-family schema model** rather than relaying raw JSON, so a vendor annotation cannot reach the Vertex payload — the parity precedent. |
| Prior art: #64 / round 031 (`specs/plans/031-tool-schema-wellformedness/`) | The same *"one bad schema field fails the whole request"* class, fixed on the `required ⊆ properties` axis and gated over `agentTools()` — **tellme's own** tools only, never the MCP relay, and never unknown keywords. |

**Reproduction (index-perfect match).** The error path `parameters.properties[0].value.properties[2].value` (and `…[5].value`) is the envelope's `MCP_PAYLOAD` (`[0]`) → the server schema's `properties`; for the first GitHub tool alphabetically (`add_comment_to_pending_review` = `function_declarations[7]`, after the 7 native tools) the **sorted** properties are `body(0), line(1), owner(2) ← x-mcp-header, path(3), pullNumber(4), repo(5) ← x-mcp-header, …` — exactly the two reported indices. `[8]`/`[9]` are `add_issue_comment` / `add_reply_to_pull_request_comment`, likewise annotated on `owner`/`repo`.

**Consequence (current behaviour):** with an annotation-bearing MCP server enabled, **every** prompt turn under a `gemini` provider fails with `the provider request failed` (exit 6) — a total outage of the turn path for the operator's most capable provider.

---

## Settled design

| # | Decision (source) |
| --- | --- |
| **S-1** | **The projection is layered — C (CQ-1 → C, operator 2026-09-19).** Two seams, each with one concern: **(B-floor)** `NormalizeMCPSchema` drops **vendor-extension / non-standard** keywords (`x-*`, `$schema`) for **every** family — one family-agnostic home for "drop vendor noise"; **(A-guarantee)** the **Gemini** adapter projects the relayed declaration onto the **provider-supported schema surface** before serializing — the guarantee that no wire-incompatible keyword reaches Vertex (covers standard-but-unsupported keywords the floor does not). Neither seam renames, re-types, or drops a **declared argument**. |

| **S-2** | **The A-side projection is a named, empirically-verified allowlist — default-deny (CQ-2 → ii, operator 2026-09-19).** The Gemini seam keeps **only** keywords **confirmed supported** by the live Vertex/Gemini `Schema` message and **drops every other key** (so an unverified keyword cannot reach the wire, whether it is a vendor extension, a standard-but-unsupported keyword, or a future addition). The confirmed set is **established empirically during `/axb-technical-research`** (probe the live endpoint per candidate keyword: accept vs reject) and recorded in the ADR; the allowlist is the **single named owner** of the surface (S-1 / FR-007), and a hermetic unit pin covers the projection itself. |
| **S-3** | **Governance: a new ADR 0031 that amends ADR 0025 — CQ-3 → i (operator 2026-09-19).** The round writes **ADR 0031** ("the provider-supported schema surface"), carries a **forward pointer on ADR 0025** (its D1 *"relayed unchanged… no property added/removed/renamed"* is **qualified** for a provider whose wire type is closed), and **MODIFY**s the round-056 declaration row (`chat/dsl.md:364` + `using-tools-from-a-remote-mcp-server.feature`). The ADR-0030 → ADR-0011 D10 amendment pattern is the precedent; the new rule lives with the provider transports, not inside the MCP-reason ADR. §Forward records the degraded-union residual (a union expressed by `anyOf` that the wire rejects is dropped, not re-expressed) and the re-probe step on a Gemini `Schema` revision. |
| **S-4** | **The CQ-2 allowlist is established by a LIVE probe — (operator 2026-09-19, approved).** `/axb-technical-research` sends real `generateContent` calls against the operator's Vertex project (`websc-dev-433809`, provider `dev` / `gemini-3.8-flash`, the **`ait-comment`** environment sourcing its `secrets/keys`), each carrying a hand-built tool declaration with **one candidate keyword** (a declaration-only payload — no prompt work). The **accept/reject table is recorded verbatim in ADR 0031** and is the evidence for the allowlist; the probe is a **research-time** step only (the gates stay hermetic — A6). |
| **S-5** | **The regression gate is a hermetic unit pin in the existing test tree — (operator 2026-09-19).** The projection is a pure transform, so a unit pin over the declaration path is a complete proof; it rides `go test` / `make test` (already part of delivery). **No new `make verify` member** is added (the aggregate gate keeps its current size). The pin must fail (red) if the projection is removed (SC-002). |
| **S-6** | **The vendor-extension floor applies to every family — (operator 2026-09-19, confirmed).** `NormalizeMCPSchema` drops vendor-extension / non-standard keywords (`x-*`, `$schema`) for **all** servers regardless of the selected provider, so the **OpenAI-compatible family also** stops seeing the annotations — a **deliberate, cross-family** change recorded in **ADR 0031** and in the round-056 truth-row MODIFY. Justification: the annotation is **server-side metadata** (it only tells the *server* to route the argument as an HTTP header) and is meaningless to any model; no declared argument is affected (S-1). This keeps **one** home for "drop vendor noise" rather than a per-family rule. |


**Note on S-2's measured effect.** The GitHub server's 4 tools carrying `anyOf`/`additionalProperties` are covered **only if** those keys are not confirmed supported — the default-deny projection removes them, and the empirical probe decides whether they are *kept* (supported) or *dropped*. Either way no 400, and no assumption is baked in: the round **measures** the surface rather than guessing it.

**Non-negotiable invariants (proposed, not open):**

- **I-1** — The envelope semantics are unchanged: `{reason, MCP_PAYLOAD}`; only `MCP_PAYLOAD` reaches the server; a non-envelope call is still refused (round 056 / ADR 0025).
- **I-2** — The projection removes **only** keywords the selected provider's wire cannot carry; it never renames, re-types, or drops a **declared argument**.
- **I-3** — tellme's **own** native tool declarations (round 031 territory) are untouched.
- **I-4** — The **MCP server's** own definition is still what the server receives; nothing about the relay to the server changes (only what the **model** is offered).
- **I-5** — No new failure mode: a schema the projection cannot handle degrades to the existing freeform object, exactly as today.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - a Gemini turn completes with an annotation-bearing MCP server enabled (Priority: P1)

As the **operator** running tellme with a `TYPE: gemini` provider and a remote MCP server that annotates its tool schemas (the GitHub server's `x-mcp-header`), I want a prompt turn to **reach the model and answer**, so that enabling a useful MCP server does not silently disable my provider.

**Why this priority**: it is the operator's reported defect, and it is a **total** turn-path outage (not a per-tool degradation) for the most capable provider.

**Independent verification**: with the GitHub MCP server enabled and the provider set to `gemini` (a faithful fake Vertex-shaped endpoint, or a live check), a tool-offering prompt turn produces an answer and exits 0 — where today it fails with `the provider request failed` / 400. A falsifiability witness reproduces the 400 first, then the fix, then reverts.

**Acceptance Scenarios**:

1. **Given** a `gemini` provider and an enabled remote MCP server whose advertised schema carries a keyword the Gemini `Schema` cannot carry (`x-mcp-header`), **When** a prompt turn offers that server's tools, **Then** the request is accepted and the turn produces an answer (no `the provider request failed`, no 400).
2. **Given** the same configuration, **When** the offered declaration is inspected, **Then** the Gemini payload contains **no** keyword outside the provider-supported surface.
3. **Given** an MCP server with **no** annotations, **When** a turn runs, **Then** the offered declaration is **unchanged** from today.

**Functional Requirements**:

- **FR-001**: Under a `gemini` provider, a prompt turn whose offered declaration includes a tool relayed from an MCP server carrying a **provider-unsupported** schema keyword MUST complete successfully — the request MUST NOT be rejected (no `the provider request failed` / 400) (S-1).
- **FR-002**: The declaration MUST be **projected** onto the selected provider's supported schema surface before it is serialized onto the wire in `internal/infrastructure/llm/gemini/client.go`, and a **normalizer-side floor** MUST drop vendor-extension keywords for every family (S-1); the projection is **default-deny** over a **named, empirically-verified** supported-key set (S-2).
- **FR-003**: The projection MUST be derived from the **provider family** (or from an explicitly named supported-key set), not from an inline ad-hoc filter at the wire site.

### User Story 2 - the offered declaration still describes the server's real arguments (Priority: P2)

As the **operator**, I want the projected declaration to keep describing what the tool actually takes (`owner`, `repo`, their `type`/`description`/`enum`, nested objects/arrays), so the model can still call the tool correctly and the projection is a *carve-out*, not a lobotomy.

**Why this priority**: it protects the value the round-056 relay exists to provide; independent of US1 (US1 could be satisfied by a blunt drop of everything but `type`).

**Independent verification**: a unit pin drives a schema carrying both supported structure (nested `properties`, `required`, `enum`, `items`) and unsupported keywords (`x-mcp-header`, `anyOf`, `additionalProperties`), and asserts the projected result preserves the former exactly and removes only the latter.

**Acceptance Scenarios**:

1. **Given** a server schema with `owner`/`repo` properties carrying `type` + `description` + `x-mcp-header`, **When** the declaration is projected, **Then** `owner`/`repo` remain with their `type` and `description` (the annotation is the only removal).
2. **Given** a schema with `enum`, `required ⊆ properties`, and nested object properties, **When** projected, **Then** `enum`, `required`, and the nested structure are preserved unchanged.
3. **Given** a schema the projection cannot represent as an object, **When** projected, **Then** it degrades to the existing **freeform** object (today's behaviour, I-5).

**Functional Requirements**:

- **FR-004**: The projection MUST preserve every declared argument's `type`, `description`, `properties`, `required`, `enum`, and nested object/array structure (S/I-2).
- **FR-005**: The projection MUST remove **only** keywords the selected provider's wire cannot carry; it MUST NOT rename, re-type, or drop a declared argument (I-2).
- **FR-006**: An invalid / non-object / unrepresentable schema MUST degrade to the existing freeform object (I-5) — never a new failure mode.

### User Story 3 - the projection cannot silently rot (Priority: P3)

As the **maintainer**, I want the provider-supported schema surface to have **one named owner** and a **hermetic regression gate**, so the round-061 defect (and round 031's #64) cannot recur on a new axis without a red test.

**Why this priority**: recurrence prevention; it is the remedy for the *class* (three recorded instances now: #64, and this issue), not just this instance.

**Independent verification**: a hermetic gate drives the full `MCP discovery → tool declaration → Gemini payload` path with an unsupported keyword present and asserts the payload is accepted (parses against the supported surface); it must fail if the projection is removed.

**Acceptance Scenarios**:

1. **Given** the projection is deleted (witness), **When** the gate runs, **Then** it fails (red) — proving non-vacuity.
2. **Given** the projection is present, **When** the gate runs, **Then** it passes with no network and no provider credential.

**Functional Requirements**:

- **FR-007**: The provider-supported schema surface MUST have a **single named owner** (one source of truth for "what the wire can carry"), referenced by both the projection and its gate.
- **FR-008**: A **hermetic unit pin** MUST cover the MCP-relayed declaration path for at least the `x-*`-extension case (the projection is pure, so a unit pin is a complete proof); it rides the existing test suite (`go test` / `make test`) — **no new `make verify` member** (S-5) — and it MUST fail (red) when the projection is removed (SC-002).
- **FR-009**: The round-056 / ADR-0025 **"verbatim"** truth rule MUST be **MODIFY**-ed with the carve-out recorded (a `truth-delta.md` entry + **ADR 0031** amending ADR 0025), so the truth and the code agree (S-3).

---

## Edge Cases

- **A provider-unsupported *standard* keyword** (`anyOf`, `additionalProperties`, `$ref`, `$schema`, `examples`, `title`, `default`) — measured present on the GitHub server (`anyOf` ×2, `additionalProperties` ×2). **Resolved by S-2**: default-deny drops every keyword not **empirically confirmed** supported, so this class cannot 400; the probe (not an assumption) decides each key.
- **A key the probe confirms supported** (`anyOf` if Vertex accepts it) — it is **kept**, so a genuinely needed union survives.
- **A future Gemini `Schema` revision** — the verified set is the single home; re-verifying is one probe run + one list edit (no scattered filters).
- **An MCP server with no annotations** — declaration byte-identical to today (US1 scenario 3).
- **The OpenAI-compatible family** — changed **deliberately** by the floor (S-6): the vendor annotations go for it too; its declared arguments are unchanged.
- **An invalid/non-object server schema** — the existing normalizer skip (`SchemaSkippedWarning`) / freeform degradation; unchanged.
- **A disabled / unreachable MCP server** — unchanged (no declaration offered).
- **The envelope shape** — `reason` required + `MCP_PAYLOAD`; a non-envelope call refused; a non-object `MCP_PAYLOAD` refused (round 056) — all unchanged (I-1).
- **A large projected schema** — no new size bound is introduced; the existing payload/context accounting applies unchanged.
- **A nested unsupported keyword** (inside `items`/`properties` subtrees) — the projection MUST recurse, not only inspect the root.

## Key Entities

- **The relayed MCP declaration** — tellme's offered `{reason, MCP_PAYLOAD}` envelope around a server's advertised schema.
- **The provider-supported schema surface** — the named set of schema keywords the selected provider family's wire accepts (the new single-owned concept; CQ-1/CQ-2).
- **The projection** — the pure transform from a relayed declaration + a supported surface → a wire-safe declaration.
- **The regression gate** — the hermetic check proving no unsupported keyword reaches the Gemini payload.

## Success Criteria

- **SC-001**: With an annotation-bearing MCP server enabled under a `gemini` provider, a prompt turn **completes** (answer on stdout, exit 0) where today it returns the 400 / `the provider request failed`. *(Falsifiability witness: reproduce the 400 first.)*
- **SC-002**: A hermetic gate proves an unsupported keyword cannot reach the Gemini payload; removing the projection turns the gate **red**.
- **SC-003**: The projected declaration preserves every declared argument's `type`/`description`/`enum`/nested structure (a unit pin over a mixed schema).
- **SC-004**: The OpenAI-compatible family's declaration carries the server's **declared arguments** unchanged (its only change is the deliberate, ADR-recorded vendor-extension drop, S-6).
- **SC-005**: `make verify` + `go test -count=1 ./...` green (including the E2E contract); the topology/DSL audit adds **no** new findings; any dependency change is recorded.
- **SC-006**: Every changed behaviour (the projection, the truth-row MODIFY) has a reproduced falsifiability witness.

## Assumptions

- **A1**: The reference's typed per-family schema conversion (its ADR-1378 lineage) is the **parity precedent** — a per-family wire model is the intended shape, not a raw-JSON passthrough.
- **A2**: `x-mcp-header` is **server-side metadata** (it tells the *server* to route the argument as an HTTP header) and is meaningless to the model — so dropping it from the *offered* declaration is semantically free and does not weaken the tool call.
- **A3**: This is a **plain line CLI** round — `/axb-ui-plan` is skipped; `/axb-api-plan` and `/axb-data-plan` are expected **NOOP** (no HTTP surface, no persisted state).
- **A4**: Truth impact is expected in `specs/truth/features/cli/chat/dsl.md` (the round-056 declaration row) + `chat/using-tools-from-a-remote-mcp-server.feature`, and likely `specs/truth/techstack.md` (the MCP client / tool-schema rows).
- **A5**: A **new ADR (next free number: 0031)** records the provider-supported-surface rule and **amends ADR 0025** with a forward pointer (S-3); the declaration truth row is MODIFY-ed in the same round.
- **A6**: The round is testable **hermetically** (the existing fake provider + a fake MCP server already drive the declaration path in the E2E suite), so no live MCP server is required in the gate. The **one** non-hermetic step is the research-time probe (S-4), which is not part of any gate.
- **A7**: `NormalizeMCPSchema`'s existing well-formedness contract (round 031) is **not** weakened; the projection is added, not substituted.

## Out of scope (recorded forward items)

- **Re-designing the round-056 envelope / relay semantics** — the `{reason, MCP_PAYLOAD}` contract, the refusal of a non-envelope call, and the server-side relay are unchanged (I-1/I-4).
- **Re-validating tellme's own native tool schemas** — round 031 / issue #64 territory (I-3).
- **MCP discovery, consent, serial-execution, or credential semantics** — round 032 / ADR-0067 territory.
- **Schema *sanitization* as a security boundary** — this is a wire-compatibility defect, not an authorization surface (no security layer, per project direction).
- **A `--color=always`-style override or any chrome change** — rounds 054/057/058.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion) · **parallel tool calls** ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
