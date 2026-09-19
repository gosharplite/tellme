# Technical Research: MCP tool-call reason + a universal "no reason, no go" gate (round 056)

**Plan Package**: `specs/plans/056-mcp-tool-call-reason`
**Truth Root**: `specs/truth`
**Owner**: `/axb-technical-research`
**Status**: complete — clarify round 1 CLOSED (**Q1 → A**: the offered envelope with the server's schema **verbatim** as `MCP_PAYLOAD`; **Q2 → B, STRICT**: an envelope-less MCP call is refused). **Scope extended by operator decision (ii):** the refusal generalises to a **universal, single-owned** reason gate for **every** call (native + MCP); **[#121](https://github.com/gosharplite/tellme/issues/121) is folded into this round**.

> No measured claims are made in this round (it is a design round; no numbers are asserted). Every code reference below is a **static read** at `dev` @ `f9907ea` (2026-09-19).

---

## D1 — Elicitation: tellme offers its **own** declaration for an MCP tool (Q1 → A)

**Decision.** For a discovered MCP tool, the **model-visible declaration** — the `parameters` the registry exposes via `agentport.ToolDefs` (which reads `Tool.Parameters()`; `internal/agent/agentloop.go:168`) — is tellme's **own** object:

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

The remote server's advertised schema is **relayed unchanged** — only *positioned* inside `MCP_PAYLOAD` (`mcp.Tool.Parameters()`, `internal/infrastructure/mcp/tool.go`, becomes the envelope builder; the stored normalized `parameters` become the envelope's `MCP_PAYLOAD` subschema). `MCP_PAYLOAD` is **not** required (a server tool may take no arguments).

**Rationale.** The declared `reason` is the **same mechanism** by which tellme's native tools are asked for a reason (`resourceSchema` declares a required `reason` — `internal/infrastructure/tools/filesystem.go:60`), so the elicitation is **one mechanism, not two**. It needs **no** system-prompt change (S-4) and **no** edit to the server's schema (S-1). The envelope *positioning* is the only thing tellme does to the server's schema, and it moves no property inside it.

**Alternatives considered.**
- *(B) Offer the server's schema byte-verbatim at top level + elicit `reason` elsewhere* — rejected: the only "elsewhere" left is the system prompt (excluded by S-4) or a tool-description edit (a server-owned field — the grey zone the operator closed).
- *(A′) Append a `reason` property to the server's advertised schema in place* — rejected: that **mutates the server's definition** (S-1), the operator's hard constraint.
- *(Prompt-only elicitation)* — rejected: soft/best-effort, and it would not make the reason **non-optional** (S-2).

**Well-formedness.** The envelope satisfies `required ⊆ properties` (round-031 / issue #64) by construction; the **inner** server schema keeps its own round-032 normalization/verification (`NormalizeMCPSchema`), and an unsafe inner schema is still **skipped with a warning** (round-032 FR-019) — the envelope does not weaken that.

---

## D2 — The reason gate: **universal**, and a **single owner already exists** (Q2 → B generalised; folds #121)

**Decision.** *No reason, no go* is enforced for **every** tool call — native **and** MCP — at the loop's **single execution site** (`internal/agent/agentloop.go:129-140`, the `for _, tc := range resp.ToolCalls` loop, the *only* place any tool runs). The gate: if the call does not state a **renderable** reason, the tool **does not execute**; the loop folds back a **recoverable result** (a nil-error `tool`-role message, the round-032-TD1 in-turn recoverable shape) instructing the model to retry with a `reason`.

**The single owner is the round-046 predicate.** The loop already holds the renderer (`AgentLoop.Lines`, an `agentport.ToolLineRenderer`) and already asks it for the decision: `Lines.ReasonLine(t, reason)` returns **both** the line and the `renders` bool from **one** evaluation of `ui.toolReasonText` (`internal/ui/toolrenderer.go:46`; ADR 0015). The gate **reuses that owner** — `renders == false ⇒ refuse` — so:

- there is **one** reason-presence predicate in the system (no new predicate, no second definition);
- the gate and the rendering cannot disagree (they are the same call);
- the MCP adapter does **not** re-implement the reason half (FR-009) — it adds only its own `MCP_PAYLOAD` envelope check (D3).

**Rationale.** FR-009/SC-008 demand one owner; the loop **cannot** import `internal/ui` (RULE-A — the round-046 decoupling), so a *new* domain predicate would have been needed **and** would duplicate `ReasonLine`. Asking the injected port (already held, already single-owned) is the only option that adds **no** second predicate.

**Consequences of gating on `renders` (not merely on raw presence).** An **escape-only** reason (present, but blank after the round-039 sanitize) renders **no** row and therefore **fails** the gate — consistent with the operator's intent (*the operator must be able to see why*): a reason that cannot be seen is not a reason. This resolves the spec's **blank-reason edge** in favour of *gate on the rendered value*.

**Nil-renderer caveat (recorded).** `Lines == nil` is the documented nil-safe "no tool lines" default (the round-031 assembler gate + the offline `--tool-usage` path — *no prompt turn runs*). With a nil renderer the loop has no predicate to ask, so **the gate is not applied**; in production a prompt turn always has `Lines` set (`ui.ToolLineRenderer{}` via `deps`, `cmd/tellme`). Recorded as a known boundary of the owner-based design.

**Alternatives considered.**
- *(New domain predicate, e.g. `domain/agent.ReasonRequired(args)`)* — rejected: duplicates the round-046 owner and re-implements `toolReasonText`; a second definition is exactly what FR-009 forbids.
- *(Per-tool enforcement, in each native `Execute` + the MCP adapter)* — rejected: N implementations of one rule, guaranteed to drift.
- *(Terminal error instead of a recoverable retry)* — rejected: one non-conforming turn would fail the whole run; the recoverable result lets the model self-correct (the round-032 TD1 convention). **Recorded forward item** as a future option.

---

## D3 — The MCP adapter: unwrap the envelope, forward **only** the payload; refuse a malformed envelope

**Decision.** `mcp.Tool.Execute` (`internal/infrastructure/mcp/tool.go`) becomes:

1. Decode the call's argument JSON into a `map[string]any` (as today).
2. **Envelope check (MCP-only):** require **no** top-level key other than `reason`/`MCP_PAYLOAD`; require `MCP_PAYLOAD`, when present, to be a **JSON object**. A violation is a **refused** call — the server is **not** contacted; a recoverable result asks the model to retry with the envelope.
3. **Forward exactly the `MCP_PAYLOAD` object** to `client.CallTool` — the outer keys (`reason`, `MCP_PAYLOAD`) MUST NOT be forwarded.
4. **An absent `MCP_PAYLOAD` with a passed reason = an empty payload `{}`** (a server tool with no arguments).

**Rationale.** S-1/S-3/I-1/I-2: the server receives **only its own arguments**, and its definition is never edited. The envelope check is genuinely **MCP-local** (an MCP-only concept), so it lives with the MCP adapter, while the **reason** half stays with the loop (D2) — the split FR-009 requires.

**Ordering (why the server is never contacted for a reasonless call).** The loop gate (D2) runs **before** `Execute`, so a reasonless MCP call never reaches the adapter; and the adapter's envelope check runs **before** `CallTool`, so a malformed call never reaches the server. Either refusal ⟹ **zero** server calls — the round's observable guarantee (SC-006).

**Alternatives considered.**
- *(Accept a bare/flat call as legacy pass-through)* — rejected (clarify Q2 → B): it would let a reasonless or non-enveloped call through.
- *(Strip `reason` and forward the rest, flat)* — rejected: it re-introduces the flat-key collision (a server that itself declares `reason`) and implies a shape the operator rejected in favour of the explicit envelope.

---

## D4 — Predicate ownership & layering (the layered statement)

| Concern | Owner | Where |
| --- | --- | --- |
| *Is there a reason the operator can see?* | `agentport.ToolLineRenderer.ReasonLine` (**existing**, ADR 0015) | asked by the loop's gate (D2) |
| *Is this a valid MCP envelope?* | the MCP adapter | `internal/infrastructure/mcp/tool.go` (D3) |
| *Render the reason* | `internal/ui` (`ReasonLine`/`FormatToolReason`) | unchanged |
| *Extract the reason for display* | `internal/agent` (`toolReason`) | unchanged |

One rule (*a call must state a visible reason*), one owner (`ReasonLine`), one enforcement site (the loop); the MCP-only envelope is a **second, distinct** rule with its own owner (the adapter). No predicate is defined twice.

---

## D5 — What is deliberately **unchanged**

| Element | State | Why |
| --- | --- | --- |
| System prompt / persona (`PERSON`) | untouched | S-4 — the ask is declaration-carried |
| Native tools' **declared** schemas (`resourceSchema`, `execute_command`) | unchanged | they already declare a required `reason`; only *enforcement* is added (S-5) |
| Native tools' `Execute` | unchanged | they already ignore `reason`; the loop forwards args as today |
| The MCP transport, naming, discovery fast-fail, disabled-server skip, malformed-schema skip, token resolution, failed-call-no-abort | unchanged | round-032 FR-007 |
| `[Tool Action]` (already `reason`-excluded) · payload estimate · `turns.log` chrome | unchanged semantics | round 034/039/053 contracts |
| The reason **rendering** (fold/trim/sanitize/cap; blank ⇒ no row) | unchanged | rounds 036/039/046 |
| `go.mod` / `go.sum` | unchanged | no dependency change |

---

## D6 — Truth impact & governance (recorded via `truth-delta.md`)

`specs/truth/techstack.md` **MODIFY**: a **new MCP Client row** (*MCP tool-call reason envelope*) + the **Agent tool loop** row (the universal gate) + the **Agent tool schemas** row (enforcement note). `specs/truth/features/cli/chat/**` **MODIFY** (owned by `/axb-dsl-refine`): the MCP feature gains a Rule (reason shown; server payload only; envelope-less refused) and a new Rule for the native refusal; the round-022 schema-nonconforming `read_files`-without-`reason` fixture is updated. `/axb-api-plan` + `/axb-data-plan` are **NOOP**. Governance: **ADR 0025** + the `docs/decisions/README.md` index row.

---

## D7 — Planned witnesses (falsifiability)

- **(a) MCP reason row**: a scripted MCP call carrying a reason ⇒ a `[Tool Reason]` row on `stderr`; freeze the envelope ⇒ the row disappears (E2E carrier: the new `chat` Rule).
- **(b) Server payload purity**: the E2E fake records the received arguments ⇒ exactly the payload (no `reason`, no envelope); inject `reason` into the forwarded map ⇒ the fixture reds.
- **(c) Server definition verbatim**: the offered declaration's `MCP_PAYLOAD` subschema equals the server's advertised schema byte-for-byte (unit pin in `internal/infrastructure/mcp`).
- **(d) Refusal — MCP (envelope)**: an envelope-less MCP call ⇒ **zero** recorded server calls + a recoverable result; remove the adapter's check ⇒ the fixture reds.
- **(e) Refusal — universal (reason)**: a native call with no/blank `reason` ⇒ the tool's side effect is **absent** + a recoverable result; drop the loop gate ⇒ the fixture reds.
- **(f) Single owner**: the gate and the renderer share `ReasonLine` — a unit pin over the loop seat asserts the refusal decision comes from the injected renderer (a fake renderer returning `renders=false` yields a refusal).

---

## Residual risks (recorded)

- **RF-056-1** — gating on `renders` means an **escape-only** reason is refused (deliberate, D2); a future "gate on raw presence" would instead let it run with no visible row.
- **RF-056-2** — `Lines == nil` ⇒ **no gate** (D2's nil-renderer caveat); a non-production assembly only.
- **RF-056-3** — the refusal is a **recoverable** result counted by round-026 as neither `ok`/`error`/`timeout` from a tool call (no tool runs); the accounting classification of a **refused** call must be pinned (the round-026 forward item's neighbourhood).
- **RF-056-4** — a **terminal** (fail-the-run) refusal instead of a retry is a future option (D2 alternatives).
- **RF-056-5** — the envelope's `reason`-description wording (the *ask*) is tellme's; a future refinement is a `/axb-dsl-refine`/wording change, not structural.
- **RF-056-6** — `MCP_PAYLOAD` naming is tellme's convention; a future rename is a truth change (the naming was chosen for readability over an ownership-prefixed name).
