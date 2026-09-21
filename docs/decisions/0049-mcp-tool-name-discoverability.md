# ADR 0049 — Make the callable MCP wire name discoverable in the offered declaration

**Status**: Accepted (round 077)

**Date**: 2026-09-22

**Related**: issue [#155](https://github.com/gosharplite/tellme/issues/155) · issue [#154](https://github.com/gosharplite/tellme/issues/154) / [ADR 0048](0048-recoverable-unknown-tool-name.md) (the loop-side unknown-name fix this complements) · [ADR 0025](0025-mcp-tool-call-reason.md) (the tellme-owned offered declaration; D1 — the server's definition is never mutated; D4 — no prompt change) · [ADR 0031](0031-provider-supported-schema-surface.md) (the closed-wire schema projection) · round 032 (the `mcp_<server>_<tool>` naming contract) · `specs/truth/techstack.md` §MCP Client · round 077 (`specs/plans/077-mcp-tool-name-presentation`).

## Context

tellme offers a remote MCP tool to the model under the deterministic **namespaced wire name** `mcp_<server>_<tool>` (`internal/infrastructure/mcp/naming.go`); the offered declaration's **`name`** is that namespaced name. But the text the model *reads* does not corroborate it:

1. **The synthesized fallback names the BARE tool** when the server ships no description — `"MCP tool " + def.Name + " from server " + server` → `MCP tool get_me from server github` — advertising an **uncallable** name. *(Pre-change location on `dev` @ `ab8fab7`: `internal/infrastructure/mcp/tool.go:53-56`; post-change the fallback body is built in `NewTool` and cites the callable name.)*
2. **Nothing advertises the `mcp_<server>_<tool>` convention** (ADR 0025 D4: no system-prompt change) — namespacing is an *unadvertised* wire convention; a weak model that recalls the bare upstream name (`get_me`) may emit it. *(Pre-change `tool.go:60` was the assignment `name: NamespacedName(server, def.Name)`.)*

Observed live (butler / `deepseek-flash` on `misc`): the model emitted `get_me` while the callable name is `mcp_github_get_me`.

**Verify-first (the issue's decisive step, performed live 2026-09-22 against `api.githubcopilot.com/mcp/`):** the GitHub server advertises **45** tools, **0** of them with an empty description; `get_me`'s description is `"Get details of the authenticated GitHub user. …"` and **names no tool at all**. So:

- the **fallback did NOT fire** in the observed run (its trigger is an empty server description) — fixing it alone would not have changed the failure;
- the mis-call came from the model's **own prior knowledge**, while the only place the callable name appears is the declaration's `name` field.

**Reference check:** `tell-me-go` namespaces identically and sets `Description: t.Description` verbatim with **no fallback** — so *both* tellme-only surfaces in play (the synthesized fallback; any tellme-authored note) are tellme additions, not parity.

## Decision

1. **Ship a fix (not NOOP) (D1).** The verify-first result shows the presentation gap is real and tellme-fixable.

2. **The offered declaration positively states the callable wire name (D2).** Every offered MCP tool's **description** is **prefixed** with a short, tellme-authored **call-name note** naming the callable wire name:

   ```
   Call this tool as "mcp_github_get_me". <server-authored description, verbatim>
   ```

   The note's template is a named constant with **one** home in `internal/infrastructure/mcp`. Placement is a **prefix** (the model reads the callable name first, and it survives a downstream trim of the server text). This is the mechanism that satisfies #155's DoD *"make the correct callable name positively discoverable."*

3. **The server's own text is added-to, never rewritten (D3).** The server's description is relayed **unchanged**; the note lives **outside** it (ADR 0025 D1 — the server's definition is never mutated; the offered declaration is tellme's own).

4. **The synthesized fallback names the CALLABLE name (D4).** With an empty server description the fallback becomes `"MCP tool " + t.name + " from server " + server` — tellme's own text must never advertise an uncallable name (a latent-correctness fix; the reference has no fallback at all, so this is a beyond-reference robustness change).

5. **The note names `t.name`, the post-truncation wire name (D5).** For a name truncated for the 64-byte wire maximum, the note states the *actual* offered name (`Name()`), never the untruncated one — the note can never lie.

6. **Uniform + control-free (D6).** The note is the same fixed ASCII sentence for **every** MCP tool and **every** family, interpolating the name only — no credential can leak, no per-server policy is needed. Two precisions: the interpolated name is assumed already **wire-grammar-validated** by the discovery path (`discovery.go` validates before `NewTool`; the exported `NewTool` itself validates nothing — a precondition, not a formatter invariant); and the note is built with Go's `%q` (**Go**-escaping, not JSON-escaping — safe here precisely because the interpolated value is tellme-derived and the whole declaration is later JSON-marshalled as a plain string value, the round-056 R-056-2 lineage).

7. **Description-only; wire-shape-neutral (D3/I-3).** The note rides the `description` field, which is an **accepted** closed-wire keyword (ADR 0031); the vendor-extension floor and the closed-wire projection never touch the tool-level `description`. The offered **schema** bytes are unchanged.

8. **No prompt change; not modelled (D7).** The ask stays **declaration-carried** (ADR 0025 D4). The change is presentation text on an already-modelled `Tool`; `docs/domain-model/**` is unchanged (ADR 0041 escape hatch; `plan.md` §5).

9. **Recorded divergence *beyond* the reference (D8).** `tell-me-go` neither advertises the convention nor synthesizes a fallback; both are tellme-side robustness improvements, recorded as such (cf. round 075).

## Consequences

- The model reads the **callable** wire name in the offered declaration, so a bare upstream name is less likely to be chosen; a model that still slips is caught by the round-076 recoverable fold-back (which lists the available wire names).
- tellme's own text (the fallback) no longer advertises an uncallable name.
- The offered **schema** is unchanged (the envelope's shape and its `MCP_PAYLOAD` subschema are untouched): **zero *schema-shape* diff** — the only wire-text delta is the description prefix.
- A small, steady token cost is added to **every** offered MCP declaration (the note, ≈38 bytes ≈ 10 tokens per declaration per request — ≈450 tokens/request for a 45-tool server, every turn); the reference does not pay it (a recorded divergence).
- Efficacy on a *weak* model is not hermetically provable (no live model in the gate); the note is a precision improvement, not a guarantee (RF-077-2).

## Forward

- **RF-077-1** — the note is a small steady token cost per MCP declaration (≈38 bytes ≈ 10 tokens, i.e. ≈450 tokens/request for a 45-tool server, every turn); a future length budget or a localized wording is a wording change, not structural.
- **RF-077-2** — hermetically unprovable model-efficacy; the note improves precision, it does not guarantee compliance.
- **RF-077-3** — recorded divergence *beyond* the reference (which neither advertises nor falls back).
- **RF-077-4** — a server whose own description already names the namespaced tool would carry a redundant (harmless) note.
- **RF-077-5** — the `ToolDefs` projection / other tool kinds are untouched; a future non-MCP tool with a bare-name description is out of scope.
- **RF-077-6** — the note's **literal wording** is deliberately **un-pinned** (fold-verification R-2): the pinned contract is *the quoted callable wire name appears in the description* (unit + E2E + the `chat/dsl.md` semantics), so a future wording change to `callNameNoteFormat` is **silent** — the authority for the wording is this ADR + the `dsl.md` prose, not a golden string.
