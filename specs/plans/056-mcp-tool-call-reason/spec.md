# Feature Specification: MCP tool-call reason — a tellme-owned `reason` alongside the server payload (round 056)

**Feature Branch**: `056-mcp-tool-call-reason`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify round 1 CLOSED (asked one question at a time): Q1 → Option A** (tellme's own offered envelope: a required top-level `reason` + `MCP_PAYLOAD` carrying the server's advertised schema **verbatim**); **Q2 → Option B, STRICT** (an MCP call that is not a valid envelope — missing/blank `reason`, or a non-object `MCP_PAYLOAD` — is **refused**; the server is never contacted; a recoverable result asks the model to retry). **Scope extended (operator, 2026-09-19, decision (ii)):** the Q2 → B refusal is **generalised to a universal, single-owned *no reason, no go* gate** for **every** tool call — native **and** MCP; issue **[#121](https://github.com/gosharplite/tellme/issues/121) is folded into this round** (closed as superseded). No open `NEEDS CLARIFICATION`. No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19)**: a design conversation that began from *"Why calling MCP doesn't have `[Tool Reason]`?"* and settled the following operator intent:

> *"Remote server's definition should not be altered."* · *"The `reason` which tellme needs to render is provided **alongside** with the MCP calling json."* · *"`{"reason":"…","MCP_PAYLOAD":"…"}` — send everything in `MCP_PAYLOAD` is cleaner."* · *"No need to touch system prompt at all."* · *"If AI needs to call MCP, I want to see the reason — this is not optional."*

**Behaviour intent**: **MODIFY (CLI-visible diagnostic chrome + the model→tool call contract — universal).** Today an MCP tool call renders no `[Tool Reason]` row (its schema is the remote server's and declares no `reason`), and **no** tool call is *enforced* to state a reason (native tools are declaration-only). This round (a) makes tellme **request a `reason` alongside the MCP payload**, render it as `[Tool Reason]`, and forward to the server **only the server's own payload** — **without editing the remote server's tool definition and without changing the system prompt**; and (b) makes the *reason required* rule **universal and enforced** — **no reason, no go** for **every** tool call (native **and** MCP), via **one** owner.

---

## Grounded in the current system *(measured 2026-09-19, `dev` @ `f9907ea`)*

| Site | Current shape |
| --- | --- |
| `internal/agent/agentloop.go` → `toolReason(arguments)` | The loop extracts a **top-level `reason`** from the model's call JSON (`{"reason":…}`) and ignores the tool's origin — native or MCP alike. **The extraction is already tool-agnostic**; the reason is rendered from the *call arguments*, not from the tool definition. |
| `internal/agent/agentloop.go` → `logAction` / `reasonsOf` | Emit `[Tool Reason]` (begin line) and the grouped post-call tail **only when the reason renders** (`ui.ToolLineRenderer.ReasonLine` → non-blank after fold/trim). **No reason ⇒ no row.** |
| `internal/infrastructure/tools/filesystem.go` → `resourceSchema(...)` | Builds **every tellme-owned tool's** offered schema and declares a **mandatory `reason`** property (plus `max_output_tokens`/`timeout`) — so native tools are *told* to send a reason, and do. `execute_command` declares `reason` inline. (The round-031/#64 fix; `required ⊆ properties`.) |
| `internal/infrastructure/mcp/schema.go` → `NormalizeMCPSchema` | Verifies/normalizes a **remote** server's advertised schema for well-formedness only (`object` root, `required ⊆ properties`) — it **never adds** a `reason` property (no `reason` string anywhere under `internal/infrastructure/mcp/**`). |
| `internal/infrastructure/mcp/tool.go` → `Tool` | Holds the server's `description` + normalized `parameters` (offered to the model as-is; only an **empty** description gets a generated fallback). `Execute` decodes the call's argument JSON into a `map[string]any` and hands **the whole map** to `client.CallTool(ctx, t.tool, args)`. |
| `internal/infrastructure/llm/{openai,gemini}/client.go` | On the wire a tool call is `{name, arguments}` only (`functionCall{name, args}` for Gemini); the tool **declaration** (`tools[]`) is `{name, description, parameters}` built from `agentport.ToolDefs(registry)` — i.e. from the `llm.ToolDef{Name,Description,Parameters}` the registry exposes. The **schema flows tellme→model only**; the model's reply carries no schema. |
| `internal/domain/agent/loop.go` / `internal/agent/agentloop.go` | The turn flow: `Registry` → `toolDefs()` → `llm.Request{Tools: …}` → model reply `ToolCalls[]` → `Execute(arguments)` → result folded back. |

**Consequence (the observed behaviour):** an MCP tool call is scripted as `Arguments: "{}"` in the round-032 E2E (`tests/e2e/steps/step_r032_t013.go`), so no reason is present and no `[Tool Reason]` row is emitted — consistent with `specs/truth/techstack.md` (*Agent tool schemas* row scopes the mandatory `reason` to tellme-owned schemas) and with the recorded MCP feature (which asserts the call, not a reason row).

---

## Settled design *(operator-confirmed this session — the plan is built to this shape)*

| # | Decision |
| --- | --- |
| **S-1** | The **remote MCP server's tool definition** (name, description, advertised input schema) and the **payload it receives** are **never altered**: the server receives only its own arguments object. (The operator's hard constraint.) |
| **S-2** | tellme **requests a `reason`** from the model *alongside* the server payload, and renders it as `[Tool Reason]` for the MCP call. The reason is **not optional** — it is part of the declared call shape, not a best-effort hint. |
| **S-3** | The model's MCP call JSON is `{"reason": "<why>", "MCP_PAYLOAD": {<server args>}}`. tellme renders `reason` and forwards **everything inside `MCP_PAYLOAD`** to `CallTool`; the outer keys (`reason`, `MCP_PAYLOAD`) are **not** sent to the server. |
| **S-4** | The ask lives **in the tool declaration tellme offers the model** (the `tools[]` entry the model reads) — **no system-prompt / persona change**, and **no edit to the server's advertised schema**. This is the same mechanism by which native tools are asked for a reason (a declared property), applied to MCP. |
| **S-5** | **Universal rule, one owner (operator decision (ii), 2026-09-19).** The *no reason, no go* rule applies to **every** tool call — native **and** MCP. Native tools keep their **declared** `reason` (`resourceSchema`/the inline command schema, unchanged) and now also become **enforced** (they were declaration-only). The rule has **one owner** — a single reason-required predicate consumed by the loop's gate; the MCP envelope check is MCP-only and layered on top. |
| **S-6** | Because the reason is tellme's own field (outside `MCP_PAYLOAD`), a server that *itself* declares a `reason` argument is unaffected: its `reason` (if the model supplies it) rides **inside `MCP_PAYLOAD`**, distinct from tellme's outer `reason`. |
| **S-7** | The **reason gate** is enforced at the loop's single execution site: a call whose top-level `reason` is missing/blank is **refused** — the tool does **not** execute — and the model receives a **recoverable result** asking it to retry with a `reason`. This mirrors the MCP refusal (Q2 → B) and makes the reason a *purpose gate* on **every** action (the operator's *"you won't give money to someone without knowing why"*). *(Folds [#121](https://github.com/gosharplite/tellme/issues/121).)* |

> **Q1 has since LOCKED → Option A** (below): tellme offers **its own declaration** for an MCP tool whose parameters are the envelope `{reason (required), MCP_PAYLOAD}`, with the remote server's advertised schema carried **verbatim** as `MCP_PAYLOAD`'s subschema. The only remaining open question is **Q2** (un-wrapped / legacy call disposition).

**The offered declaration (locked by Q1 → A).** For an MCP tool, tellme offers the model:

```json
{
  "type": "object",
  "properties": {
    "reason": { "type": "string", "description": "<tellme's reason description>" },
    "MCP_PAYLOAD": <the remote server's advertised input schema, VERBATIM>
  },
  "required": ["reason"]
}
```

The server's advertised schema is **relayed unchanged** (only *positioned* inside `MCP_PAYLOAD` — no property added, removed, or renamed); `reason` is tellme's **own** declared property (required), exactly the mechanism native tools use. The model's call is therefore `{"reason":"…","MCP_PAYLOAD":{…}}`; tellme renders `reason` and forwards **the `MCP_PAYLOAD` contents** to `CallTool`. `MCP_PAYLOAD` is **not** marked required (a server tool may take no arguments).

---

## Clarify round 1 *(asked one question at a time; ≤ 3 questions; ≤ 5 per session)*

| # | Question | Status |
| --- | --- | --- |
| **Q1** | **How is `reason` elicited given "no prompt change" — i.e. what exactly does tellme offer the model for an MCP tool?** Options: **(A)** tellme offers its **own** declaration for the MCP tool whose parameters are the envelope `{reason (required), MCP_PAYLOAD}` — the server's advertised schema is carried **verbatim** as `MCP_PAYLOAD`'s subschema, so the model sees the real structure; **(B)** offer the server's schema **byte-verbatim at top level** and elicit `reason` some other way (would re-open the no-prompt constraint); **(C)** other. | ✅ **LOCKED → Option A** (operator, 2026-09-19): tellme's **own** declaration is offered — top-level `reason` (required, tellme-owned) + `MCP_PAYLOAD` carrying the server's advertised schema **verbatim**; the server's definition is never mutated (only positioned as `MCP_PAYLOAD`'s subschema); no system-prompt change. |
| **Q2** | **How is a call that does *not* use the envelope handled** — e.g. a bare `{"sku":"A1"}` with no `MCP_PAYLOAD` (a server-advertised call, a legacy fixture, or a model that ignored the shape)? Options: **(A)** accept it as a **legacy/pass-through** payload (forward the object as the server args; render a reason only if a top-level `reason` happens to be present); **(B)** treat it as a **recoverable error** fed back to the model to retry ("call the tool with `reason` and `MCP_PAYLOAD`"); **(C)** treat a missing `MCP_PAYLOAD` as an empty `{}` payload. | ✅ **LOCKED → Option B (STRICT)** (operator, 2026-09-19): *"If a model tries to call MCP with no reason, no go. You won't give money to someone without knowing why."* — an MCP call whose arguments are **not a valid envelope** (a missing/blank `reason`, or a `MCP_PAYLOAD` that is not an object) is **refused**: the server is **never contacted**, and tellme returns a **recoverable result** instructing the model to retry with the envelope. An **absent `MCP_PAYLOAD` with a valid `reason`** is a legitimate **empty payload** (`{}`). No legacy/flat pass-through. |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — the remote server's advertised schema is **byte-identical** in what tellme relays (never mutated, never extended); the server receives **only** the payload object (S-1/S-3).
- **I-2** — `reason` is **tellme's field**: it is rendered by tellme and **never forwarded** to the server (S-3).
- **I-3** — the **system prompt** is untouched (S-4). The **reason gate is universal** (native **and** MCP, S-5/S-7/I-5); native tools' **declared** schemas are unchanged (they already declare `reason`) — only *enforcement* is added.
- **I-4** — an MCP call is **refused unless it is a valid envelope** (a non-blank `reason` + `MCP_PAYLOAD` absent-or-an-object); a refused call **never reaches the server** (Q2 → B — the operator's "no go" / the purpose gate for external calls). A stray top-level key other than `reason`/`MCP_PAYLOAD` is a shape violation and is **not** forwarded.
- **I-5** — ***no reason, no go* holds for every tool call** (native **and** MCP) via **one** reason-required predicate (single owner): a call whose top-level `reason` is missing/blank **does not execute** and receives a recoverable retry result (S-5/S-7). The rule MUST NOT be implemented twice (the MCP adapter must not re-implement the reason-presence half).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - an MCP tool call shows its reason (Priority: P1)

As the **operator**, when the model calls a tool on a remote MCP server, I want to see **why** — a `[Tool Reason]` row, exactly as for a native tool — so MCP activity is as legible as local tool activity.

**Why this priority**: it is the entire point of the round (the operator's *"I want to see the reason — this is not optional"*).

**Independent verification**: on a scripted MCP tool call that carries a `reason` alongside its payload, the captured `stderr` carries a `[HH:MM:SS] [Tool Reason] <reason>` line, and the run exits 0.

**Acceptance Scenarios**:

1. **Given** a remote MCP server offering a tool, **When** the model calls that tool with a `reason` alongside the server payload, **Then** tellme renders a `[Tool Reason] <reason>` row for the call (the same row a native tool produces) and the call still executes.
2. **Given** the same call, **When** it completes, **Then** the reason is also carried in the grouped post-call tail (the round-034 rendering), consistent with a native tool.
3. **Given** tellme offers the MCP tool to the model, **When** the offered declaration is inspected, **Then** its shape is the model-visible envelope that **requests** `reason` (S-4/Q1 → A), with the server's advertised schema carried **verbatim** inside it (I-1).
4. **Given** the model calls an MCP tool with arguments that are **not a valid envelope** (no `reason`, a blank `reason`, or a non-object `MCP_PAYLOAD`), **When** tellme handles the call, **Then** the remote server is **never contacted** and tellme returns a **recoverable result** telling the model to retry with `{"reason":…,"MCP_PAYLOAD":…}`; a later conforming call then executes normally (Q2 → B).

**Functional Requirements**:

- **FR-001**: tellme MUST offer each discovered MCP tool to the model with a declaration that **requests a `reason`** (a declared `reason` property), so the model is told to provide one — **without** a system-prompt/persona change (S-4).
- **FR-002**: tellme MUST render the MCP call's `reason` as a `[Tool Reason]` row on the same surfaces a native tool uses (begin line + grouped tail), via the existing single-owned reason path (`agentport.ToolLineRenderer` / `ui.ToolLineRenderer.ReasonLine`) — **no** MCP-specific renderer.
- **FR-003**: the reason MUST be **mandatory** for an MCP call (S-2/Q2 → B): the offered declaration marks `reason` required, **and** tellme **enforces** it at call time — an MCP call that is not a valid envelope (a missing/blank `reason`, or a `MCP_PAYLOAD` that is not an object) MUST be **refused**: the remote server MUST NOT be contacted, and tellme MUST return a **recoverable result** instructing the model to retry with `{"reason":…,"MCP_PAYLOAD":…}` (the in-turn recoverable-error path, a nil-error result — never the terminal request-level failure). An **absent `MCP_PAYLOAD` with a valid `reason`** is a legitimate **empty payload** (`{}`).
- **FR-004**: tellme MUST forward to the remote server **only the contents of `MCP_PAYLOAD`** — the outer `reason`/`MCP_PAYLOAD` keys MUST NOT reach `CallTool` (S-3/I-2).

### User Story 2 - the remote server's definition and payload are untouched (Priority: P1)

As the **operator**, I want the remote MCP server to see **exactly its own definition and its own arguments** — never a tellme invention — so that transparency to the server is preserved and a strict server that validates arguments against its advertised schema can never be broken by tellme's `reason`.

**Why this priority**: it is the operator's hard constraint (*"the server's schema should not be edited at all"*) and the safety property that makes the round shippable.

**Independent verification**: the E2E fake records the arguments it received — exactly `MCP_PAYLOAD`, no `reason`/envelope keys; and the offered schema carries the server's advertised schema unchanged.

**Acceptance Scenarios**:

1. **Given** a scripted MCP call `{"reason":"…","MCP_PAYLOAD":{"sku":"A1"}}`, **When** tellme executes it, **Then** the server recorded **exactly** `{"sku":"A1"}` as its arguments (no `reason`, no `MCP_PAYLOAD` wrapper).
2. **Given** a server advertising an input schema, **When** tellme offers the tool, **Then** the server's advertised schema appears **verbatim** in tellme's declaration (S-1/I-1) — no added, removed, or renamed property inside it.
3. **Given** a server that advertises a **malformed** schema, **When** the round runs, **Then** the round-032 skip-with-warning behaviour is **unchanged** (a malformed schema is still not offered).

**Functional Requirements**:

- **FR-005**: tellme MUST NOT modify the remote server's advertised tool definition (name, description, input schema) and MUST relay that schema **verbatim** in what it offers the model (S-1/I-1).
- **FR-006**: the arguments handed to the MCP client MUST be **exactly** the `MCP_PAYLOAD` object (S-3/I-2); tellme MUST NOT inject `reason` (or any envelope key) into the server payload.
- **FR-007**: the round-032 MCP behaviours MUST be preserved: namespaced naming, discovery fast-fail, disabled servers never contacted, malformed-schema skip, token resolution, and a failed call not aborting the run.

---

### User Story 3 - no reason, no go — for every tool call (Priority: P1)

As the **operator**, I want **no** tool call to execute without a stated reason — native or MCP — so that every action the model takes is justified, exactly as for a remote call.

**Why this priority**: it is the operator's headline principle (*"You won't give money to someone without knowing why"*) applied uniformly; leaving native tools declaration-only would be an asymmetry the round-056 MCP gate would otherwise create. *(Folds [#121](https://github.com/gosharplite/tellme/issues/121) — decision (ii).)*

**Independent verification**: a scripted **native** tool call with no `reason` does **not** execute (the side effect is absent) and the model receives a recoverable retry result; a conforming native call is byte-identical to today.

**Acceptance Scenarios**:

1. **Given** a native tool call whose arguments carry **no** `reason` (or a blank/whitespace-only one), **When** tellme handles the call, **Then** the tool **does not execute** and tellme returns a **recoverable result** instructing the model to retry with a `reason`.
2. **Given** a retry that includes a valid `reason`, **When** tellme handles it, **Then** the tool executes and its `[Tool Reason]` row renders exactly as today.
3. **Given** the reason-required rule, **When** its implementation is inspected, **Then** there is **one** owner — a single reason-required predicate consumed by the loop's gate for **every** call; the MCP path adds only its `MCP_PAYLOAD` envelope check (no second reason-presence predicate).

**Functional Requirements**:

- **FR-008**: tellme MUST refuse a tool call (native **or** MCP) whose top-level `reason` is missing or blank: the tool MUST NOT execute, and the model MUST receive a **recoverable result** (the in-turn recoverable-error path, a nil-error result — never the terminal request-level failure) instructing it to retry with a `reason`. (S-5/S-7; Q2 → B generalised; folds [#121](https://github.com/gosharplite/tellme/issues/121).)
- **FR-009**: the reason-required rule MUST have **one owner** — a single reason-presence predicate (a domain-tier check, since the loop cannot import `internal/ui`) consumed by the loop's gate for every call; the MCP adapter MUST NOT re-implement the reason-presence half (it adds **only** the MCP-envelope check). The rule MUST NOT be split/duplicated across layers.
- **FR-010**: a call **with** a valid `reason` MUST behave exactly as today for native tools (execute + render the `[Tool Reason]` row). The round-022 **schema-nonconforming fixture** (a `read_files` call scripted **without** a reason — `specs/truth/features/cli/chat/dsl.md:43`) MUST be updated to the new contract, and the interface truth re-aligned.

---

## Edge Cases

- **Un-wrapped / legacy call** — a call that is not a valid envelope is **refused** (Q2 → B): the server is never contacted and the model receives a recoverable retry instruction. There is **no** legacy/flat pass-through (the round-032 fixtures that script `Arguments: "{}"` are updated in the implementation half — A4).
- **A server that itself declares `reason`** — its `reason` lives **inside `MCP_PAYLOAD`**, distinct from tellme's outer `reason` (S-6); no collision by construction. (The old flat-key design had this collision; the envelope removes it.)
- **Blank reason** — a blank/whitespace-only `reason` is **now a refusal** (FR-008): the tool does not execute. The round-036 blank-reason predicate still governs *rendering* when a reason **is** present; the gate's rule (presence) and the renderer's rule (rendering) are distinct concerns, and their relationship is pinned in `research.md`. **Recorded edge:** an escape-only-but-non-blank reason passes the gate yet renders no row (a reason that is present but invisible) — disposition pinned in research (gate on presence, or on the rendered value).
- **A non-object `MCP_PAYLOAD`** (e.g. a string) — must be handled deterministically (degrade to `{}` or a recoverable error; pinned in the implementation, matching clarify Q2's disposition).
- **Empty payload** — `"MCP_PAYLOAD":{}` is valid (a server tool with no arguments).
- **The offered-declaration well-formedness invariant** — the envelope MUST keep `required ⊆ properties` (round-031/#64), and the inner server schema keeps its own well-formedness (round-032 FR-019).
- **`[Tool Action]` / payload estimate / `turns.log`** — the action line already excludes the `reason` key and lists the remaining args; the estimate counts the offered declarations; `turns.log` inherits the reason row via the chrome tee. These MUST stay consistent with the new declaration shape.
- **Native tools** — declared schemas unchanged; **enforcement added** (S-5/S-7): a native call with no/blank `reason` is **refused**. The round-022 schema-nonconforming `read_files`-without-reason fixture is **updated** (FR-010).
- **stdio MCP transport** — out of scope (tellme ships remote Streamable HTTP only, round-032 Q1).

## Key Entities

- **The model-visible MCP tool declaration** — the `tools[]` entry tellme offers for an MCP tool; it now requests `reason` and carries the server's schema (Q1 shapes the naming/carrying).
- **The MCP call JSON** — the model's reply arguments for an MCP tool: `{"reason":…, "MCP_PAYLOAD":…}`.
- **The MCP payload** — the server's own arguments object (`MCP_PAYLOAD` contents), forwarded verbatim.
- **The reason row** — the existing `[Tool Reason]` chrome (unchanged; now reachable for MCP calls).

## Success Criteria

- **SC-001**: On a scripted MCP tool call carrying a `reason`, the captured `stderr` carries a `[Tool Reason] <reason>` row (and the grouped tail carries it too), and the run exits 0 — the operator's *"I want to see the reason"* satisfied.
- **SC-002**: The remote server records **exactly** the `MCP_PAYLOAD` contents as its arguments — no `reason`, no envelope keys (witnessed against the E2E fake's recorded call).
- **SC-003**: The server's advertised input schema appears **verbatim** in the offered declaration; no round-032 MCP behaviour regresses (all existing MCP scenarios stay green).
- **SC-004**: The round's own gates hold: `gofmt`/`go vet` clean · `make verify` **OK** · `go test -count=1 ./...` green (all Examples) · native-tool rendering byte-identical · `go.mod`/`go.sum` unchanged.
- **SC-005**: The reason-required rule is **declaration-or-prompt-free for the ask** (no system-prompt change) and **single-owned** across native **and** MCP calls (I-5/FR-009) — demonstrable, not just stated.
- **SC-006**: A non-conforming MCP call is **refused without contacting the server** (witnessed against the E2E fake's recorded calls: zero calls), and a subsequent conforming call succeeds — the "not optional" guarantee (Q2 → B).
- **SC-007**: A **native** tool call with no/blank `reason` does **not** execute (witnessed: the tool's side effect is absent) and the model receives a recoverable retry result; a conforming native call is unchanged (FR-008/FR-010 — the [#121](https://github.com/gosharplite/tellme/issues/121) fold).
- **SC-008**: The reason-required rule is **single-owned** — demonstrable: **one** predicate drives the gate for native **and** MCP calls (FR-009) — not two site-local copies.

## Assumptions

- **A1**: The **MCP** transport scope is the **remote (Streamable HTTP)** path only (stdio stays deferred, round-032 Q1). Native tools' **declared** schemas and the **system prompt** are untouched; native **enforcement** is added (US3, decision (ii)).
- **A2**: Truth impact is expected in `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` **and** a new `chat` Rule for the universal reason gate (+ `chat/dsl.md`), and `specs/truth/techstack.md` (the *MCP client* / *Agent tool loop* / *Agent tool schemas* rows); `/axb-api-plan` and `/axb-data-plan` are **NOOP**. `/axb-spec-by-example` is **NOT** NOOP (the row is operator-visible) and `/axb-dsl-refine` is **NOT** NOOP.
- **A3**: An **ADR (0025)** records the envelope + the untouched-server-definition invariant + the elicitation mechanism **and** the universal, single-owned reason gate (the [#121](https://github.com/gosharplite/tellme/issues/121) fold), with a §Forward.
- **A4**: The existing round-032 MCP E2E fixtures that script `Arguments: "{}"` **and** the round-022 schema-nonconforming native fixture are updated by the implementation half (to the new contract) — `specs/truth/**` changes owned by the later phases, not this skill.
- **A5**: The reason is **tellme's own field**; forwarding it to the server is explicitly **not** intended (I-2).

## Out of scope (recorded forward items)

- **Altering the remote server's definition** — never (the operator's constraint).
- **A system-prompt / persona change** — explicitly excluded (S-4).
- **Native-tool changes** — they already carry `reason`.
- **The hard gate** — **IN SCOPE, universal** (Q2 → B generalised, decision (ii)): **no reason, no go** for **every** tool call (FR-003 the MCP envelope; FR-008 native + universal). **[#121](https://github.com/gosharplite/tellme/issues/121) is folded into this round** (closed as superseded). The gate is a *recoverable* refusal (the model may retry), not a terminal failure; a future option to hard-fail a run instead is a recorded forward item.
- The locked exclusions: no security/consent layer · no Windows · sequential tool calls ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
