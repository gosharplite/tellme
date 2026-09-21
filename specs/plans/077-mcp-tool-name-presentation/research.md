# Technical Research: MCP tool-name presentation (round 077)

**Plan Package**: `specs/plans/077-mcp-tool-name-presentation`
**Spec**: `specs/plans/077-mcp-tool-name-presentation/spec.md`
**Anchor**: [#155](https://github.com/gosharplite/tellme/issues/155) · **ADR**: [0049](../../decisions/0049-mcp-tool-name-discoverability.md)
**Status**: complete (2026-09-22) — **verify-first done; a change IS warranted** (the presentation gap is real and tellme-fixable; the NOOP verdict was **not** taken).

---

## 0. The verify-first result (the issue's decisive step)

Issue #155 is **research-gated**: *"confirm whether the fallback is in play at all before implementing; a NOOP verdict is an acceptable outcome."* Verified **live** (2026-09-22, `api.githubcopilot.com/mcp/`, `initialize` + `tools/list`, bearer `$GITHUB_TOKEN`):

| Fact | Measured |
| --- | --- |
| Tools the GitHub MCP server advertises | **45** |
| Tools with an **empty** description | **0** |
| `get_me` description | `"Get details of the authenticated GitHub user. Use this when a request is about the user's own profile …"` (171 chars) — it **does not name the tool at all** |

**Conclusion — the synthesized fallback did NOT fire** in the observed run (its trigger is an empty server description). So:

- **Option 1 alone (fix the fallback) would not have changed the observed failure** — the fallback never runs for a description-bearing server. Fixing it is still a *latent-correctness* fix (tellme's own text must not advertise an uncallable name), but it is **not** the symptom's cause and **not** sufficient to satisfy #155's DoD *"make the correct callable name positively discoverable."*
- The observed mis-call came from the model's **own prior knowledge** of the GitHub MCP tool name (`get_me`) while the offered wire name is `mcp_github_get_me` — the offered declaration carries the namespaced name only in its **`name` field** (`internal/infrastructure/mcp/tool.go:60`), which nothing in the *text the model reads* corroborates.

**Decision: the presentation gap is real and tellme-fixable → ship a fix (not NOOP).**

## 1. Reference-parity check (`tell-me-go`)

- The reference namespaces identically (`mcp_<server>_<tool>`, `internal/tools/integrations/mcp/plugin.go:139-163`).
- The reference's declaration is `{Name: namespacedName, Description: t.Description, …}` — the **server's description verbatim, and *no* fallback**: an absent server description stays **empty** (no synthesized text at all).
- **So both tellme surfaces in play are tellme-only**: (a) the synthesized fallback (the reference has none), and (b) any tellme-authored note (the reference has none). Fixing either is a **recorded divergence *beyond* the reference** (robustness, cf. round 075), not parity.

## 2. Where the model-visible text is assembled (grounded, `dev` @ `ab8fab7`)

| Site | Shape |
| --- | --- |
| `internal/infrastructure/mcp/tool.go:53-56` | fallback: `"MCP tool " + def.Name + " from server " + server` — names the **bare** tool. |
| `internal/infrastructure/mcp/tool.go:60` | `name: NamespacedName(server, def.Name)` — the offered **`name`** is the callable wire name. |
| `internal/infrastructure/mcp/tool.go:76-78` | `Name()` → `t.name`; `Description()` → server text (or fallback). |
| `internal/domain/agent/result.go:75` | `llm.ToolDef{Name: t.Name(), Description: t.Description(), Parameters: t.Parameters()}` — the single projection the model sees. |
| `internal/infrastructure/llm/gemini/schema.go` | `description` is an **accepted** closed-wire keyword; the vendor-extension floor (round 061) never touches the **tool-level** `description`. |

## 3. Decisions

| # | Decision | Rationale |
| --- | --- | --- |
| **D1** | Ship a fix (not NOOP). | The verify-first result shows the gap is real (§0). |
| **D2** | **Every** offered MCP tool's description is prefixed with a short, tellme-authored **call-name note** naming the **callable wire name** (`t.name`), so the correct name is *positively discoverable* in the text the model reads. | This is the only mechanism that satisfies #155's DoD; a description-only change is **wire-shape-neutral**. |
| **D3** | The note is **added outside** the server's own text — the server's words are relayed **unchanged** (ADR 0025 D1). tellme owns the note; the server never sees it. | I-1; the issue's Option 2 is explicitly "still not touching the server text". |
| **D4** | The **synthesized fallback** (empty server description) is fixed to name the **callable wire name**, never the bare upstream name. | tellme's own text must not advertise an uncallable name (latent correctness; the reference has no fallback, so this is a beyond-reference robustness fix). |
| **D5** | The note names `t.name` (the **post-truncation** wire name), so it is always byte-identical to the name tellme actually offers — even for a 64-byte hash-truncated name. | FR-001 + edge case; the note must not lie about a truncated name. |
| **D6** | The note is **uniform** (every MCP tool, every family) and **control-free** (a fixed ASCII sentence + the sanitized name), so it cannot leak a credential and needs no per-server policy. | NFR-004; determinism. |
| **D7** | **No** system-prompt/persona change; the ask stays **declaration-carried** (ADR 0025 D4). | I-4 / A4. |
| **D8** | The change is **family-local to the MCP adapter's `Description()`** (plus the `ToolDefs` projection is untouched); the naming rule, the schema normalizer, the closed-wire projection, and the reason envelope are untouched. | NFR-002 / I-3. |

### Chosen shape (the note)

```
Call this tool as "mcp_github_get_me". <server-authored description, verbatim>
```

- **Placement: prefix** — primacy (the model reads the callable name first), and it survives a downstream length trim of the server text.
- The note's template is a **named constant** (`callNameNote`) with **one** home in `internal/infrastructure/mcp`; the name interpolated is `t.name`.
- **Fallback body** (empty server description) becomes `"MCP tool " + t.name + " from server " + server` — i.e. it now names the **callable** name (D4); the prefix note then already states it, so the fallback body is provenance + a restatement (kept explicit so the text is self-contained if a future consumer reads the body alone).

## 4. Truth & domain-model impact

- **Truth**: `specs/truth/techstack.md` — a **MODIFY** on the **MCP tool-call reason envelope** row (its sibling) / a new **MCP tool-name presentation** row: the offered declaration **positively states the callable wire name** in its description (a tellme-authored note, server text untouched), and the fallback names the callable name. **No** schema/shape change.
- **ADR**: **0049** records the verify-first result, the note, the fallback fix, the beyond-reference divergence, and the declined alternatives.
- **Domain model**: **not modelled** — the change is presentation text on an already-modelled `Tool`/`MCPTool`; no entity/invariant changes (ADR 0041 escape hatch; recorded in `plan.md` §5).

## 5. Falsifiability (the witnesses the round MUST carry)

| # | Mutation | Expected RED |
| --- | --- | --- |
| **W1** | Drop the call-name note | the unit pin + the E2E Then observing the offered declaration's description |
| **W2** | Note names `def.Name` (the bare name) instead of `t.name` | the unit pin asserting the note carries the **callable** name |
| **W3** | Revert the fallback to the bare name | the fallback unit pin (`empty description`) |
| **W4** | Truncate a long name and note the untruncated name | the truncated-name pin (note ≠ `Name()`) |

## 6. Residual risks / forward items

- **RF-077-1** — the note adds a small, steady token cost to **every** offered MCP declaration; a future truncation or a language variant is a wording change.
- **RF-077-2** — efficacy on a *weak* model is not provable hermetically (no live model in the gate); the note is a precision improvement, not a guarantee.
- **RF-077-3** — the note is a **beyond-reference divergence** (`tell-me-go` neither advertises nor falls back); recorded, not parity.
- **RF-077-4** — a server whose own description already names the namespaced tool would carry a redundant note (harmless).
- **RF-077-5** — the projection/`ToolDefs` layer is untouched, so a future non-MCP tool with a bare-name description is out of scope.

## 7. Techstack truth changes (semantic units)

| Action | Truth Spec | Summary |
| --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *MCP tool-call reason envelope* (sibling offered-declaration row) | the offered MCP declaration's **description** now states the callable wire name (tellme note; server text unchanged) and the empty-description fallback names the callable name — description-only, schema unchanged |
| ADD | `docs/decisions/0049-mcp-tool-name-discoverability.md` (+ index) | the decision record (verify-first result; note + fallback; divergence; declined alternatives) |
| NOOP | `specs/truth/` `contracts/**` | single CLI end; no OpenAPI surface |
| NOOP | `specs/truth/data/**` | no persisted-state change |
