# ADR 0031 — The provider-supported schema surface: a tool declaration is projected onto the wire's closed type

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0025](0025-mcp-tool-call-reason.md) (the `{reason, MCP_PAYLOAD}` envelope whose D1 *"the server's schema is relayed unchanged"* this ADR **qualifies at the provider wire**), [ADR 0007](0007-terminal-control-sanitization.md)/[ADR 0008](0008-terminal-safe-lines-and-blank-line-grouping.md) (the "sanitize at the presentation seam" lineage this mirrors at the declaration seam), round 032 (`specs/plans/032-mcp-client` — the MCP relay + `NormalizeMCPSchema`), round 031 / issue [#64](https://github.com/gosharplite/tellme/issues/64) (the same *"one bad schema field fails the whole request"* class, fixed on the `required ⊆ properties` axis), issue [#127](https://github.com/gosharplite/tellme/issues/127) (the operator-observed defect), round 061 (`specs/plans/061-mcp-schema-provider-projection` — this ADR's round)

## Context

tellme relays a remote MCP server's advertised input schema into the declaration it offers the model (round 056 / ADR 0025 **D1**: the server's schema is *"relayed unchanged"* as `MCP_PAYLOAD`'s subschema). That is correct for a **JSON-Schema-tolerant** provider (the OpenAI-compatible family) and **fatal** for a provider whose tool-declaration reader is a **closed type**.

Vertex/Gemini parses `functionDeclarations[].parameters` as the proto message `google.ai.generativelanguage.Schema`. Any key that message does not define is rejected — and because the tools array is one object, the API fails the **whole** `generateContent` call:

```
tellme: the provider request failed: provider dev: provider returned status 400:
Invalid JSON payload received. Unknown name "x-mcp-header" at
'tools[0].function_declarations[7].parameters.properties[0].value.properties[2].value': Cannot find field.
```

**Measured context (2026-09-19, `dev` @ `768dbec`).** The GitHub MCP server advertises **45** tools; **37** annotate arguments with `"x-mcp-header": "owner"|"repo"` (74 occurrences). The error path is index-perfect: `properties[0]` is the envelope's `MCP_PAYLOAD`; its `properties[2]` is `owner` for `add_comment_to_pending_review` (= `function_declarations[7]`, after tellme's 7 native tools) — with the sorted property order `body(0), line(1), owner(2) ← mark, path(3), pullNumber(4), repo(5) ← mark`. The result was a **total outage of the turn path** for the operator's most capable provider whenever that server was enabled. `NormalizeMCPSchema` (round 032) enforces *well-formedness* only — it has no notion of the provider's supported surface, so nothing stood between a third-party dialect and the wire.

Two truths therefore held at once and only one was checked: (1) the relay must be verbatim; (2) the wire's type is closed. This ADR reconciles them.

## Decision

**D1 — Two seams, one concern each: a family-agnostic floor and a provider-side projection.** The operator settled clarify **CQ-1 → C**:

- **Floor (all families)** — `mcp.NormalizeMCPSchema` drops **vendor-extension / non-standard** keywords (`x-…`, `$schema`) **recursively**, for every provider. The annotation is **server-side plumbing** (it tells the *server* to route the argument as an HTTP header) and is meaningless to any model, so removing it is semantically free — and gives "drop vendor noise" **one** home rather than a per-family rule (S-6, operator-confirmed: a tolerant family's declaration changes too, deliberately).
- **Projection (closed-wire families)** — the Gemini transport projects every offered declaration onto the provider's **supported schema surface** before serializing (`gemini.projectSchema`, applied in `buildToolDeclarations`). This is the **guarantee** that no wire-incompatible keyword reaches the wire, including standard-but-unsupported ones the floor does not cover.

**D2 — Default-deny over a named, empirically-verified allowlist.** The projection keeps a keyword **only** if it is in `supportedSchemaKeys` — the single named owner the projection *and* its regression pin read (the round's FR-007). The set is **measured, not assumed**: the round probed the live endpoint with declaration-only `generateContent` calls (`gemini-3.8-flash`, project `websc-dev-433809`, 2026-09-19) and recorded the table verbatim:

| Keyword (placement) | Result |
| --- | --- |
| `type`, `description`, `properties`, `required`, `items` | **accepted** |
| `enum`, `format`, `title`, `default`, `nullable`, `pattern` | **accepted** |
| `minimum`, `maximum`, `minLength`, `maxLength`, `minItems`, `maxItems` | **accepted** |
| `oneOf`, `allOf`, `additionalProperties` (root and property level) | **accepted** |
| `propertyOrdering` (Gemini-native) | **accepted** |
| `type: "null"` | **accepted** |
| a property carrying only `description` (no `type`) | **accepted** |
| `x-mcp-header`, `x-anything`, `x-google-*` (any vendor extension) | **rejected** (unknown name) |
| `$schema` (root), `$ref`, `$defs`, `definitions` | **rejected** |
| `const`, `examples`, `deprecated`, `readOnly`, `writeOnly`, `multipleOf`, `uniqueItems` | **rejected** |
| `anyOf` **with any other schema keyword on the same node** | **rejected** (*"when using any_of, it must be the only field set"*) |
| `anyOf` **alone** (no sibling schema keyword) | accepted — **not** allowlisted (see D3) |
| an unknown key at a **nested** level | **rejected** (⇒ the projection must recurse) |
| a declaration with **no** extra keyword (control) | accepted |

**D3 — `anyOf` is dropped, deliberately.** Its acceptance is **shape-dependent** (sole-keyword only), a shape a projection composing arbitrary third-party schemas cannot guarantee; default-deny therefore excludes it. The cost is a recorded narrowing (a union degrades to a looser declaration — the argument's `description` remains), and the enrichment idea (re-express a sole-field union) is a §Forward item.

**D4 — The projection is a carve-out, not a lobotomy.** It removes **only** keywords: every declared argument keeps its `type`, `description`, `enum`, numeric/string/array bounds and nested structure. A property left holding only `description` stays as a **type-omitted** property (round 031 established that Gemini accepts a type-less property; dropping the argument itself would be a regression).

**D5 — Fail closed.** An unparseable, non-object or unrepresentable schema degrades to the freeform object (`{"type":"object","properties":{}}`) — a declaration the wire always accepts. No new failure mode is introduced; a schema the projection cannot handle is the round-032/README degradation, not an error.

**D6 — Silent.** The run emits **no** notice for a projected-away keyword (the operator's report was a *failure*, not a visibility gap): the rule is documented here, and the only existing channel (MCP discovery warnings) runs earlier than the projection and cannot faithfully count adapter-side drops. An aggregated notice is a recorded §Forward item.

**D7 — What is unchanged.** The round-056 envelope semantics (`{reason, MCP_PAYLOAD}`; only the payload reaches the server; a non-envelope call is refused); tellme's **native** tool declarations (round 031 territory — they contain only supported keys and pass the projection untouched); the MCP server's own definition (the server still receives its own arguments — the projection changes only what the **model** is offered); the OpenAI-compatible family's reader (permissive; unaffected by the projection, changed only by the floor); the E2E/gate topology (the projection's carrier is a hermetic unit pin — **no** new `make verify` member, S-5).

## Consequences

- A `gemini` turn completes with an annotation-bearing MCP server enabled — the reported outage closes, and the fix is a **rule** (default-deny over a measured surface) rather than a one-off filter for `x-mcp-header`.
- The "relayed verbatim" truth row gains an explicit **carve-out** at the provider wire, so truth and code agree (ADR 0025's D1 is qualified, not rewritten — a forward pointer, the ADR-0030 → ADR-0011 D10 pattern).
- The supported surface has **one home**; re-verifying it after a provider schema revision is one probe run plus one list edit, and the regression pin fails if the projection is removed.
- A third-party dialect can no longer sink a request: the worst case is a dropped keyword, i.e. a slightly looser declaration.

## Forward items

- **RF-061-1** — `anyOf` (and `oneOf`/`allOf` subtrees) are dropped/left as-is; a union expressed by `anyOf` is **not** re-expressed (the model sees the description only). Enrichment idea: when `anyOf` is the sole schema keyword, keep it (the measured accepted shape) — needs its own probe + pin.
- **RF-061-2** — the allowlist is a **probe snapshot**: a future Gemini `Schema` revision (new keyword, or a change to the `anyOf` rule) is not detected automatically. A periodic/CI probe is *not* adopted (it needs a live credential); the ADR records the re-probe step.
- **RF-061-3** — the projection is **silent** (D6); an aggregated "N keywords projected for provider X" notice on the MCP/discovery channel is a future option.
- **RF-061-4** — the projection is applied to **every** declaration the Gemini adapter sends (native ones pass through unchanged by construction); a native schema using a keyword the probe rejects would be silently trimmed — a future "declaration lint" could assert the native set directly (round 031's `agentTools()` gate covers `required ⊆ properties` only).
- **RF-061-5** — the floor's effect on the OpenAI-compatible leg is a **deliberate** declaration change (S-6); a future per-family floor would need a new named owner (today: one home by design).
- **RF-061-6** — the probe was run against one project/model (`gemini-3.8-flash`); the surface is read as model-independent (the `Schema` message is a Vertex API type), but that assumption is recorded, not proven.
