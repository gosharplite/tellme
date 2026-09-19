# System Analysis Plan — round 056 (`056-mcp-tool-call-reason`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/056-mcp-tool-call-reason/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md                    # ✓ done (/axb-technical-research)
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # ✓ done (/axb-spec-by-example) — 3 journeys:
│       ├── explaining-a-remote-tool-call.feature
│       ├── keeping-the-remote-server-untouched.feature
│       └── refusing-a-tool-call-without-a-reason.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (MCP Client + Agent tool loop + Agent tool schemas) ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — MODIFY (this phase's handoff; NOT done yet)

docs/decisions/
├── 0025-mcp-tool-call-reason.md   # governance — ADD ✓ done
└── README.md                      # index row ✓ done
```

*(No `contracts/**` — `/axb-api-plan` **NOOP**. No `data/**` — `/axb-data-plan` **NOOP**. No `ui/**`
artifact — no new terminal screen; the change refines existing `stderr` tool chrome + the wire call
shape.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
internal/infrastructure/mcp/tool.go     # CHANGED — build the offered envelope (reason required + MCP_PAYLOAD = the server schema verbatim); Execute forwards ONLY MCP_PAYLOAD; refuse a malformed envelope
internal/infrastructure/mcp/*_test.go   # CHANGED/NEW — envelope + payload-purity + verbatim-schema unit pins
internal/agent/agentloop.go             # CHANGED — the universal no-reason-no-go gate at the single execution site (reuses the injected Lines.ReasonLine owner); nil-error recoverable refusal
internal/agent/*_test.go                # CHANGED/NEW — gate pins (fake renderer: renders=false ⇒ refusal)
tests/e2e/steps/*                       # CHANGED/NEW — MCP-call fixtures carry the envelope; a native reasonless fixture for the refusal; stepdefs for the 3 new Rules
specs/truth/features/cli/chat/**        # MODIFY — /axb-dsl-refine (next phase)
docs/decisions/0025-mcp-tool-call-reason.md   # NEW — ADR 0025; index row ✓ done
specs/truth/techstack.md                # MODIFY ✓ done
go.mod / go.sum                         # unchanged — no dependency change
cmd/tellme/** , internal/cli/**         # unchanged — no CLI flag/behaviour change
```

**Structure Decision**: Round 056 is an **MCP-call contract + tool-loop enforcement** change, all
inside the **existing CLI end** (the `stderr` tool chrome and the model→tool call shape). It (a)
changes the **declaration** tellme offers the model for an MCP tool (a tellme-owned `reason` +
`MCP_PAYLOAD` carrying the server's schema **verbatim** — D1) and the **arguments** it forwards to the
server (only `MCP_PAYLOAD` — D2); and (b) adds the **universal *no reason, no go* gate** at the loop's
single execution site (D3, reusing the round-046 `ReasonLine` owner). No new system boundary is
introduced; the truth changes are `specs/truth/techstack.md` (three rows) + the CLI interface truth
(`/axb-dsl-refine`) + the governance **ADR 0025**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (`cli`) that round 056
refines: the prompt path's tool-call handling and its `stderr` tool chrome. It is the same interface
the MCP-client round (032) and the tool-call-log rounds (034/039) already own; the round **changes
existing behaviour** on it (the MCP tool declaration, the forwarded arguments, and a new refusal), so
it is a **CLI contract change**, not a new boundary.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface). The _remote MCP server_ is
>   an **external dependency** consumed through the existing `tools.MCPClient` port — it is not
>   tellme's API, and tellme's OWN interface to it is unchanged by this round (tellme still speaks the
>   server's declared tool contract; the envelope is tellme↔model, and only `MCP_PAYLOAD` reaches the
>   server).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state model change; the call envelope is
>   transient, and the reason is tellme's field, never persisted to the server).
> - `/axb-dsl-refine` = **MODIFY** (the CLI end's contract owner): new/updated executable Rules in
>   `specs/truth/features/cli/chat/**` (the MCP feature gains the reason/payload Rule; a Rule for the
>   universal gate; the round-022 schema-nonconforming `read_files`-without-`reason` fixture is
>   updated) + `chat/dsl.md` rows.
> - `/axb-ui-plan` = **skipped** (no new terminal screen / keybinding / state transition; only
>   existing `stderr` chrome content changes — a line that now also appears for MCP calls).

### Analysis Wave schedule

**1 wave** (cli-only, no dependencies to order):

| Wave | Interface | Kind | Carried by |
| --- | --- | --- | --- |
| **W1** | The CLI end — MCP tool-call reason + the universal reason gate | `cli` | **`/axb-dsl-refine`** (the CLI contract owner; carried to it at delivery) |

No API wave, no data wave (both NOOP). No inter-wave ordering is required (a single interface).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — **delegated** (the CLI end's contract owner): write the executable Rules
   for (i) the MCP tool-call reason + payload purity, (ii) the server's definition relayed verbatim,
   and (iii) the universal *no reason, no go* refusal; update the round-022 schema-nonconforming
   fixture; add/retarget `chat/dsl.md` rows.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for `/axb-dsl-refine`)*: plan package
`specs/plans/056-mcp-tool-call-reason`; truth root `specs/truth`; truth-delta
`specs/plans/056-mcp-tool-call-reason/truth-delta.md`; interface: **CLI end** (`chat` module);
analysis focus: (1) an MCP tool call renders `[Tool Reason]` (the reason is tellme's; the server
receives only `MCP_PAYLOAD`); (2) the server's advertised schema is relayed verbatim (positioned as
`MCP_PAYLOAD`'s subschema); (3) an envelope-less MCP call is refused; (4) a native **or** MCP call
with no renderable reason is refused (recoverable retry); (5) the round-022 schema-nonconforming
`read_files`-without-`reason` fixture is updated to the new contract.

---

### The invariants this round introduces (normative: ADR 0025 + the loop/the MCP adapter)

| # | Invariant | Carrier |
| --- | --- | --- |
| **I-1** | **Elicitation is declaration-carried** — an MCP tool is offered with tellme's own declaration: a required `reason` + `MCP_PAYLOAD` = the server's advertised schema **verbatim**; **no** system-prompt change | `mcp.Tool.Parameters()`; `research.md` D1; ADR 0025 D1 |
| **I-2** | **The server's definition is never mutated** and the server receives **only** the `MCP_PAYLOAD` object (outer `reason`/`MCP_PAYLOAD` never forwarded; an absent payload = `{}`; a shape violation ⇒ refused, server not contacted) | `mcp.Tool.Execute`; `research.md` D3; ADR 0025 D2 |
| **I-3** | ***No reason, no go* is universal and single-owned** — at the loop's single execution site, a call whose reason does not render is refused (recoverable nil-error result; the tool does not execute); the gate **reuses** the round-046 `ReasonLine` owner (no second predicate); the MCP adapter adds only the envelope check | `internal/agent/agentloop.go`; `research.md` D2/D4; ADR 0025 D3 |
| **I-4** | **The reason is tellme's field** — rendered by tellme, never forwarded; a server that itself declares `reason` is unaffected (its `reason` rides inside `MCP_PAYLOAD`) | `mcp.Tool.Execute`; `research.md` D3; ADR 0025 D2 |
| **I-5** | **Unchanged** — native declared schemas; native `Execute`; the reason rendering (fold/trim/sanitize/cap; blank ⇒ no row); `[Tool Action]`/the payload estimate/`turns.log`; the round-032 MCP behaviours; the system prompt; `go.mod`/`go.sum` | `research.md` D5; ADR 0025 D5 |

### Gating blockers

*(none — the operator locked the design (S-1…S-7) and answered clarify Q1 → A and Q2 → B one at a
time, and confirmed the scope extension (decision (ii): the gate is universal; #121 folded). No open
decision gates the round.)*
