# Feature Specification: MCP tool-call reason — a tellme-owned `reason` alongside the server payload (round 056)

**Feature Branch**: `056-mcp-tool-call-reason`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify round 1 IN PROGRESS (asked one question at a time): Q1 → Option A (LOCKED — the offered envelope carries the server's schema verbatim as `MCP_PAYLOAD`); Q2 OPEN (un-wrapped / legacy call disposition).** No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19)**: a design conversation that began from *"Why calling MCP doesn't have `[Tool Reason]`?"* and settled the following operator intent:

> *"Remote server's definition should not be altered."* · *"The `reason` which tellme needs to render is provided **alongside** with the MCP calling json."* · *"`{"reason":"…","MCP_PAYLOAD":"…"}` — send everything in `MCP_PAYLOAD` is cleaner."* · *"No need to touch system prompt at all."* · *"If AI needs to call MCP, I want to see the reason — this is not optional."*

**Behaviour intent**: **MODIFY (CLI-visible diagnostic chrome + the model→tool call contract for MCP tools).** Today an MCP tool call renders no `[Tool Reason]` row, because the reason is elicited only by a tool schema that declares a `reason` property — tellme's own `resourceSchema` does that for native tools, but an MCP tool's schema is the remote server's and declares no `reason`. This round makes tellme **request a `reason` alongside the MCP payload**, render it as `[Tool Reason]`, and forward to the server **only the server's own payload** — **without editing the remote server's tool definition and without changing the system prompt**.

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
| **S-5** | **MCP-only.** Native tools are unchanged (they already declare a mandatory `reason` via `resourceSchema`/the inline command schema). |
| **S-6** | Because the reason is tellme's own field (outside `MCP_PAYLOAD`), a server that *itself* declares a `reason` argument is unaffected: its `reason` (if the model supplies it) rides **inside `MCP_PAYLOAD`**, distinct from tellme's outer `reason`. |

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
| **Q2** | **How is a call that does *not* use the envelope handled** — e.g. a bare `{"sku":"A1"}` with no `MCP_PAYLOAD` (a server-advertised call, a legacy fixture, or a model that ignored the shape)? Options: **(A)** accept it as a **legacy/pass-through** payload (forward the object as the server args; render a reason only if a top-level `reason` happens to be present); **(B)** treat it as a **recoverable error** fed back to the model to retry ("call the tool with `reason` and `MCP_PAYLOAD`"); **(C)** treat a missing `MCP_PAYLOAD` as an empty `{}` payload. **Recommended: (A)** — keeps already-working server calls alive; the reason row stays best-effort for that path. | ⏳ **OPEN** |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — the remote server's advertised schema is **byte-identical** in what tellme relays (never mutated, never extended); the server receives **only** the payload object (S-1/S-3).
- **I-2** — `reason` is **tellme's field**: it is rendered by tellme and **never forwarded** to the server (S-3).
- **I-3** — the change is **MCP-only**; `resourceSchema`-backed native tools and the system prompt are untouched (S-4/S-5).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - an MCP tool call shows its reason (Priority: P1)

As the **operator**, when the model calls a tool on a remote MCP server, I want to see **why** — a `[Tool Reason]` row, exactly as for a native tool — so MCP activity is as legible as local tool activity.

**Why this priority**: it is the entire point of the round (the operator's *"I want to see the reason — this is not optional"*).

**Independent verification**: on a scripted MCP tool call that carries a `reason` alongside its payload, the captured `stderr` carries a `[HH:MM:SS] [Tool Reason] <reason>` line, and the run exits 0.

**Acceptance Scenarios**:

1. **Given** a remote MCP server offering a tool, **When** the model calls that tool with a `reason` alongside the server payload, **Then** tellme renders a `[Tool Reason] <reason>` row for the call (the same row a native tool produces) and the call still executes.
2. **Given** the same call, **When** it completes, **Then** the reason is also carried in the grouped post-call tail (the round-034 rendering), consistent with a native tool.
3. **Given** tellme offers the MCP tool to the model, **When** the offered declaration is inspected, **Then** its shape is the model-visible envelope that **requests** `reason` (S-4/Q1), with the server's advertised schema carried **verbatim** inside it (I-1).

**Functional Requirements**:

- **FR-001**: tellme MUST offer each discovered MCP tool to the model with a declaration that **requests a `reason`** (a declared `reason` property), so the model is told to provide one — **without** a system-prompt/persona change (S-4).
- **FR-002**: tellme MUST render the MCP call's `reason` as a `[Tool Reason]` row on the same surfaces a native tool uses (begin line + grouped tail), via the existing single-owned reason path (`agentport.ToolLineRenderer` / `ui.ToolLineRenderer.ReasonLine`) — **no** MCP-specific renderer.
- **FR-003**: the reason MUST be **mandatory** for an MCP call (S-2): the offered declaration marks `reason` required; a call that omits it is handled per clarify **Q2**.
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

## Edge Cases

- **Un-wrapped / legacy call** — the model (or an existing scripted fixture) sends the server args at top level with no `MCP_PAYLOAD` ⇒ clarify **Q2**.
- **A server that itself declares `reason`** — its `reason` lives **inside `MCP_PAYLOAD`**, distinct from tellme's outer `reason` (S-6); no collision by construction. (The old flat-key design had this collision; the envelope removes it.)
- **Blank reason** — the round-036 blank-reason predicate already suppresses an empty/whitespace-only reason row; the envelope does not change that (a blank reason renders no row, the call still executes).
- **A non-object `MCP_PAYLOAD`** (e.g. a string) — must be handled deterministically (degrade to `{}` or a recoverable error; pinned in the implementation, matching clarify Q2's disposition).
- **Empty payload** — `"MCP_PAYLOAD":{}` is valid (a server tool with no arguments).
- **The offered-declaration well-formedness invariant** — the envelope MUST keep `required ⊆ properties` (round-031/#64), and the inner server schema keeps its own well-formedness (round-032 FR-019).
- **`[Tool Action]` / payload estimate / `turns.log`** — the action line already excludes the `reason` key and lists the remaining args; the estimate counts the offered declarations; `turns.log` inherits the reason row via the chrome tee. These MUST stay consistent with the new declaration shape.
- **Native tools** — unchanged (S-5); a native `read_files`/`write_file` call is byte-identical to today.
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
- **SC-005**: The elicit-and-render path is **declaration-carried** (no system-prompt change) and **MCP-only** (native tools untouched) — demonstrable, not just stated.

## Assumptions

- **A1**: Scope is the **remote (Streamable HTTP) MCP** tool path only; stdio stays deferred (round-032 Q1); native tools and the system prompt are untouched.
- **A2**: Truth impact is expected in `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `chat/dsl.md`) and `specs/truth/techstack.md` (the *MCP client* / *Agent tool loop* rows); `/axb-api-plan` and `/axb-data-plan` are **NOOP**. `/axb-spec-by-example` is **NOT** NOOP (the row is operator-visible) and `/axb-dsl-refine` is **NOT** NOOP.
- **A3**: An **ADR (0025)** records the envelope + the untouched-server-definition invariant + the elicitation mechanism, with a §Forward.
- **A4**: The existing round-032 MCP E2E fixtures that script `Arguments: "{}"` are updated by the implementation half (to the envelope or the Q2 disposition) — a test/`specs/truth/**` change owned by the later phases, not this skill.
- **A5**: The reason is **tellme's own field**; forwarding it to the server is explicitly **not** intended (I-2).

## Out of scope (recorded forward items)

- **Altering the remote server's definition** — never (the operator's constraint).
- **A system-prompt / persona change** — explicitly excluded (S-4).
- **Native-tool changes** — they already carry `reason`.
- **A hard gate that refuses a reasonless call** — only if clarify **Q2** selects option (B).
- The locked exclusions: no security/consent layer · no Windows · sequential tool calls ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
