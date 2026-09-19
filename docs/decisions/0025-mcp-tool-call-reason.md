# ADR 0025 — A tellme-owned `reason` for MCP tool calls + a universal "no reason, no go" gate

- **Status:** Accepted — **D1’s "relayed unchanged" is qualified at the provider wire by [ADR 0031](0031-provider-supported-schema-surface.md)** (round 061 / issue [#127](https://github.com/gosharplite/tellme/issues/127): a closed-wire provider receives only empirically-supported schema keywords; the declared arguments are preserved).
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0015](0015-loop-presentation-port.md) (the loop presentation port + the blank-reason predicate **single owner** this ADR's gate reuses), [ADR 0006](0006-tool-reason-fold-and-cap.md) / [ADR 0008](0008-terminal-safe-lines-and-blank-line-grouping.md) (the reason row's fold/cap/sanitize lineage), [ADR 0021](0021-ride-alongs-and-records.md) (the MCP registry re-registration the offered declaration must survive), round 032 (`specs/plans/032-mcp-client` — the MCP client this round's envelope sits on), rounds 036/039/046 (the reason row), issue [#121](https://github.com/gosharplite/tellme/issues/121) (folded into this round), round 056 (`specs/plans/056-mcp-tool-call-reason` — this ADR's round)

## Context

The round began from an operator observation — *"Why calling MCP doesn't have `[Tool Reason]`?"* — and settled a design across a working session. Two facts frame it:

1. **`[Tool Reason]` is elicited by a *declaration*, not by a prompt.** tellme's native tools declare a required `reason` property in their offered schema (`resourceSchema` / the inline `execute_command` schema), so the model is *told* to send one and does; the loop renders it (`internal/agent/agentloop.go` → `toolReason` → the `agentport.ToolLineRenderer` port). A remote **MCP** tool's schema is the **server's** and declares no `reason` — so no reason is asked for, none is sent, and no row appears. The extraction is already tool-agnostic; only the **ask** is missing.
2. **Nothing enforced the reason, even natively.** Every native tool's `Execute` ignores the `reason` key, so a reasonless native call still ran (no row, but executed).

The operator's constraints, stated directly: *"Remote server's definition should not be altered."* · *"The reason … is provided **alongside** with the MCP calling json."* · *"`{"reason":"…","MCP_PAYLOAD":"…"}` — send everything in `MCP_PAYLOAD` is cleaner."* · *"No need to touch system prompt at all."* · *"If AI needs to call MCP, I want to see the reason — this is not optional."* · *"You won't give money to someone without knowing why."*

## Decision

**D1 — Elicit `reason` the way native tools do, without touching the server's schema or the prompt.** For an MCP tool, tellme offers the model **its own declaration**:

```json
{
  "type": "object",
  "properties": {
    "reason": {"type": "string", "description": "<tellme's reason description>"},
    "MCP_PAYLOAD": <the remote server's advertised input schema, VERBATIM>
  },
  "required": ["reason"]
}
```

The server's advertised schema is **relayed unchanged** — only *positioned* as `MCP_PAYLOAD`'s subschema (`mcp.Tool.Parameters()` builds the envelope); no property is added, removed, or renamed inside it, and the **system prompt / persona is untouched**. `MCP_PAYLOAD` is not required (a server tool may take no arguments). The declared `reason` is the **same mechanism** native tools use, so there is **one** elicitation mechanism.

**D2 — Forward to the server only the payload.** `mcp.Tool.Execute` unwraps the call: it forwards **exactly the `MCP_PAYLOAD` object** to `CallTool`, and the outer keys (`reason`, `MCP_PAYLOAD`) never reach the server. An **absent `MCP_PAYLOAD`** (with a reason) is a legitimate empty payload `{}`. Any **other** top-level key, or a `MCP_PAYLOAD` that is not a JSON object, is a **shape violation** and is **refused** (the server is not contacted; a recoverable result asks the model to retry). An **unsafe inner schema** keeps round-032's skip-with-warning behaviour.

**D3 — *No reason, no go* is universal, and its owner already exists.** The reason gate is enforced for **every** tool call — native **and** MCP — at the loop's **single execution site** (`internal/agent/agentloop.go`, the `for _, tc := range resp.ToolCalls` loop, the only place a tool runs). A call without a **renderable** reason does **not execute**; the loop folds back a **recoverable** `tool`-role result instructing the model to retry with a `reason` (a nil-error result — never the terminal request-level failure; the round-032 TD1 convention). Critically, the gate **reuses the round-046 single owner** — `agentport.ToolLineRenderer.ReasonLine` returns the line **and** the `renders` decision from one evaluation of `ui.toolReasonText`, and the loop already holds that port (`AgentLoop.Lines`) and **cannot** import `internal/ui` (RULE-A). So *one* reason-presence predicate serves the gate **and** the rendering; no second predicate is defined. Gating on `renders` (not raw presence) means a reason the operator cannot see is not a reason — an **escape-only** reason is refused (recorded, RF-056-1).

**D4 — Two distinct rules, two owners, no duplication.** *A call must state a visible reason* → owned by `ReasonLine`, enforced by the loop (D3). *A call must be a valid MCP envelope* → owned by the MCP adapter (D2, an MCP-only concept). The MCP adapter MUST NOT re-implement the reason half; the loop MUST NOT know the envelope.

**D5 — What is unchanged.** Native tools' **declared** schemas (they already declare `reason` — only *enforcement* is added); native `Execute`; the reason rendering (fold/trim/sanitize/cap; blank ⇒ no row); `[Tool Action]` / the payload estimate / `turns.log`; the round-032 MCP behaviours; the system prompt; `go.mod`/`go.sum`.

## Consequences

- An MCP tool call now shows `[Tool Reason]` — MCP activity is as legible as native activity, without tellme editing a server's definition or the prompt.
- *No reason, no go* holds uniformly; the reason becomes a **purpose gate** on every action, not a display nicety.
- The reason is **tellme's field**: it is rendered and never forwarded; a server that itself declares `reason` is unaffected (its `reason` rides inside `MCP_PAYLOAD`), so the flat-key collision the pre-envelope design had is gone by construction.
- `#121` is **folded** — the native and MCP paths share one reason rule, so the round ships **one** predicate, not two.
- The gate is a **recoverable** refusal: a model that omits a reason is asked to retry, rather than the run failing (a terminal variant is a recorded forward item).
- **Boundary (recorded):** a **nil** `Lines` renderer (the round-031 assembler gate / the offline `--tool-usage` path — no prompt turn) means no gate; in production a prompt turn always has the port set.
- **Accounting residual (recorded):** a **refused** call runs no tool, so its round-026 classification (ok/error/timeout) must be pinned explicitly.

## Forward items

- **RF-056-1** — gating on the **rendered** value refuses an escape-only reason (deliberate); a future "gate on raw presence" would instead run it with no visible row.
- **RF-056-2** — **standing invariant (round-056 review TD-056-3):** a **prompt-bearing** loop assembly MUST supply a **non-nil** `Lines` renderer; only the non-prompt paths (the round-031 `agentTools()` assembler gate, the offline `--tool-usage` path) leave it nil, and those never run a loop. The gate's predicate is owned by the renderer, so a nil renderer disarms the rule — today unreachable in production (the CLI binds `dp.NewToolLines(colourOn)`, a value-type adapter), but not yet *enforced*. **Future carrier:** assert a non-nil `Lines` at the composition seam (the round-050 `LoopSpec` binding / `agent.NewLoop`), so the nil-renderer unit pin is demonstrably below the production path rather than a live escape hatch. The nil-safe default remains documented (D3).
- **RF-056-3** — **decision recorded (round-056 review TD-056-4):** a **refused** call runs no tool, so it records **no** `ToolUsage` outcome and **no** `Step`. This is deliberate — a refusal is a *gate*, not an execution — and it means the `--tool-usage` surface **cannot** distinguish a disciplined model from one that thrashed; and there is **no consecutive-refusal bound**, so a model that never adapts burns `MAX_TOOL_LOOP` (default 1000) in metered refusals. **Chosen disposition: recorded, not fixed this round.** The future control is either a `history.ToolOutcomeRefused` recorded at the gate (refusals become countable) or a consecutive-refusal bound.
- **RF-056-4** — a **terminal** (fail-the-run) refusal as an alternative to the recoverable retry.
- **RF-056-5** — the `reason` **description** wording (the ask) is tellme's own text; a future refinement is a wording change, not structural.
- **RF-056-6** — `MCP_PAYLOAD` naming is tellme's convention; a rename would be a truth change.
- **RF-056-7** — the `reason` wire key still has spellings a Go constant cannot reach (round-056 review R-056-1, reachability corrected per the fold-verification residual **R-056-d**): the loop's `json:"reason"` **struct tag** and the native schema builders' inline `properties` **template literals** (`filesystem.go`, `writer.go`, `get_tree.go`, `skills.go`, `command.go`). By contrast, the builders' **required-list** entries are ordinary Go string arguments (e.g. `resourceSchema(…, "reason", …)`) and therefore *can* take a constant. The shared `internal/domain/tools.ReasonArgKey`/`PayloadArgKey` are consumed by the MCP envelope, `internal/ui`'s argument filter, and the E2E helper; a future sweep could adopt the constant at the required-lists and generate the template literals structurally.
- **RF-056-8** — `freeformEnvelope` spells the envelope keys directly (a raw JSON literal cannot interpolate a constant); it is defined beside the constants, so the drift surface is one function.
