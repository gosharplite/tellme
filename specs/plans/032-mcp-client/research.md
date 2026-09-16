# Phase 0 Research: remote MCP client (round 032)

**Topic**: how tellme gains a **remote (Streamable HTTP) Model Context Protocol client** — reading a typed `MCP_SERVERS` registry, discovering a server's tools and offering them alongside its native tools, calling them within a turn, and doing all of it **without ever letting an unreachable server stall a run**. Each decision supports `spec.md` (US1/US2/US3 · FR-001…FR-017 · NFR-001…NFR-005) and the operator-locked clarifications (**Q1 → remote HTTP only**; **Q2 → fast-fail bound + per-server `ENABLED`**; **Q3 → official SDK, confined**; **Q4 → full auth parity + reference naming**; **Q5 → MCP client only**).

**Must-ask questions (settled).** Per the AIxBDD three must-asks, all three are already written in the existing `specs/truth/techstack.md` and are **unchanged** this round: the system has a **single CLI end** (the remote MCP server is an *external dependency tellme consumes*, not a new end tellme exposes — no web frontend, no HTTP server); the BDD techstack is **`godog`** driving the interface Gherkin E2E against the built binary; the test strategy is **E2E** for the acceptance path plus fast unit tests for pure helpers. No re-ask is warranted (`Rule 2` of the must-ask rule). The one scope boundary the operator set is transport (**Q1**), not a new end.

## Decision 1: Remote Streamable HTTP only this round; stdio deferred

- **Decision**: support only the **remote Streamable HTTP** MCP transport (`URL` + credentials) this round. A `COMMAND` (stdio) entry is **rejected deterministically** by configuration validation (FR-013); the local-child-process transport is a later round.
- **Rationale**: the operator's active server (`github`) is a remote URL endpoint, so the remote transport delivers the round's value end-to-end (config → discovery → registration → call) as a single vertical slice. The stdio transport is architecturally distinct — child-process spawn/lifecycle, PATH/ENV resolution, byte-for-byte env handling (the reference's ADR-069) — and roughly doubles the surface; splitting it keeps this round's risk low (Q1 → 1).
- **Alternatives considered**:
  - **Both transports in one round** — rejected (Q1): doubles the round's surface and mixes two very different failure models; the operator chose the smaller slice.
  - **stdio only** — rejected: does not enable the active `github` server; the remote transport is the higher-value slice.

## Decision 2: Adopt the official MCP Go SDK, confined behind a `tools.MCPClient` domain port

- **Decision**: use **`github.com/modelcontextprotocol/go-sdk` v1.7.0** as the MCP client implementation (the same library and version `tell-me-go` uses). The SDK is imported **only** by a new adapter package `internal/infrastructure/mcp/`, which implements a new domain port `tools.MCPClient` (`ListTools(ctx) ([]MCPToolDefinition, error)`, `CallTool(ctx, name, args) (ToolResult, error)`, `Close() error`). A Makefile gate `verify-mcp-sdk-confinement` fails if any file outside `internal/infrastructure/mcp/` imports the SDK (mirroring the reference's ADR-067 pattern and tellme's existing `verify-*` gates).
- **Rationale**: MCP's Streamable-HTTP transport carries enough protocol surface (session-id handling, `application/json` vs `text/event-stream` responses, initialization/version negotiation, typed content blocks) that a hand-rolled client carries real interop risk. The domain port keeps the SDK swappable and keeps `internal/agent`/`internal/cli` free of protocol types — the same injection discipline tellme already uses for its provider transports. This is tellme's **first protocol SDK**; recording it as a deliberate techstack decision (moving "MCP client SDK" out of *Not Introduced Yet*) keeps the truth honest (`truth-current`).
- **Alternatives considered**:
  - **Hand-roll stdlib JSON-RPC over Streamable HTTP** — rejected (Q3): consistent with tellme's "no provider SDK" house style, but the Streamable-HTTP leg (SSE parsing, session headers, version negotiation) is large enough that a hand-rolled client risks the very interop fragility this round must avoid.
  - **Adopt the SDK without a domain port** (import it directly in the tool layer) — rejected: couples the tool/agent layers to the SDK, breaking the injection pattern and the confinement gate.

## Decision 3: Discovery must never stall a run — fixed fast-fail bound + per-server `ENABLED`

- **Decision**: MCP tool discovery runs **concurrently** per enabled server, each bounded by a **small, fixed per-server deadline** (a named constant, default **3 s**) that is **independent of and much smaller than** the per-tool call timeout. A server that does not answer within the bound is **skipped with a warning**; the turn proceeds with the tools it has. Registered tools are ordered deterministically (sorted by server key). Separately, a server marked `ENABLED: false` is skipped **entirely** (no connection attempt). Discovery runs **only on the prompt-bearing turn path**, so the offline paths (`--version`, `-d`, `-l`, prompt-less `--new`, boot) make **no network contact** (FR-016).
- **Rationale**: this is the round's **defining** requirement, and the fix for the operator's reported pain — the reference builds MCP clients at the start of **every** CLI invocation and `wg.Wait()`s on a per-server `ListTools` capped at **30 s**, so one offline server delayed every command (the operator's workaround was to comment `hf` out). A small fixed bound makes an unreachable server a **non-event** (`wg.Wait()` now completes in ≈ the bound, not ≈ the slowest timeout), and `ENABLED` lets the operator switch a known-flaky server off without deleting its definition (Q2 → 1+3). Gating discovery on the prompt path preserves tellme's long-standing offline guarantee (round 004/005/009).
- **Alternatives considered**:
  - **Cross-invocation tool cache** (load a cached tool list instantly, refresh in the background) — deferred (Q2 Option 2): gives true ~zero startup cost but adds a new persisted artifact and staleness semantics that deserve their own round; the fast-fail bound + `ENABLED` already remove the operator's pain.
  - **A single overall deadline across all servers** rather than per-server — rejected as the primary mechanism: a per-server bound composes to the same overall ceiling under concurrency while letting a *healthy* slow server still be discovered; the per-server form is simpler to reason about and to witness.

## Decision 4: Full auth parity (`auto`/`gh`/`bearer`/`basic`/`none`), credentials never logged

- **Decision**: resolve per-server credentials by auth mode (default `auto`): an explicit `TOKEN` always wins; otherwise a **GitHub-hosted** hostname falls back to a token source (`gh auth token`, then `GITHUB_TOKEN`), and any other host is anonymous; `gh` forces the token source; `bearer` uses the explicit token; `basic` sends `Authorization: Basic base64(USERNAME:TOKEN)`; `none` is anonymous. Credentials (and derived Authorization headers) are **never logged** or written to disk (FR-017).
- **Rationale**: the operator's active `github` server resolves under `auto` via its explicit `${GITHUB_TOKEN}` — no `gh` spawn needed — while `gh`/`basic` cover the realistic follow-ons (e.g. a Basic-auth Atlassian endpoint). The set is small and fully specified by the reference; trimming it would surface later as "unexpected" failures (Q4 → 1). Avoiding credential logging mirrors the reference's `mcp-token-not-logged` invariant and tellme's standing hygiene.
- **Alternatives considered**:
  - **Trimmed set (`auto`/`bearer`/`none`)** — rejected (Q4): would drop the `gh` fallback and Basic auth that the operator explicitly asked for.
  - **Minimal (`bearer`/`none`)** — rejected: loses hostname detection that the default `auto` mode already implies.

## Decision 5: Deterministic namespaced tool names (reference parity)

- **Decision**: register each discovered tool under the deterministic name `mcp_<server>_<tool>`. When that exceeds the 64-byte wire maximum, truncate the tool segment (rune-wise, capped at 40) and append an 8-hex-char SHA-256 prefix of the full tool name to preserve uniqueness. Servers are registered in sorted key order so the offered set and its order are stable across runs regardless of discovery timing (FR-004/FR-007).
- **Rationale**: matching the reference's naming keeps behaviour and debugging familiar, keeps the generated names inside the OpenAI `tools[i].function.name` limit (`^[a-zA-Z0-9_-]{1,64}$`), and the sorted order preserves tellme's determinism value (the same discipline behind its spinner/log/accounting orderings).
- **Alternatives considered**:
  - **Non-namespaced names** (pass the server's tool name through) — rejected: two servers could collide, and a server tool could shadow a native tool.
  - **User-visible server prefix with no hash** — rejected: exceeds 64 bytes for long tool names and loses uniqueness.

## Decision 6: Typed `MCP_SERVERS` config + deterministic validation; no new failure class

- **Decision**: add a typed `MCP_SERVERS` block to tellme's configuration (today merely tolerated as an unknown top-level key). Per server, accept `URL`, `TOKEN`, `USERNAME`, `AUTH`, `TIMEOUT`, `ENABLED` (default true). Validate deterministically: server key `^[a-z0-9-]{1,24}$`; exactly one transport shape — a non-empty `URL` and **no** `COMMAND` this round (FR-013); a known effective auth mode; the credentials that mode requires; a non-negative `TIMEOUT`. Validation failures reuse the existing classed configuration-invalid failure surface; the round adds **no** new class phrase or exit code (FR-015). MCP tool-call outcomes ride the existing tool-result / tool-timeout paths (a tool-level error is a recoverable result; a transport failure is a classed failure).
- **Rationale**: the operator's config already carries an `MCP_SERVERS` block the reference validates this way; validating deterministically (rather than ignoring) is what makes FR-002 meaningful, and reusing the existing failure surfaces preserves tellme's frozen class-phrase/exit-code vocabulary (a hard invariant in every prior round).
- **Alternatives considered**:
  - **Ignore an invalid entry (warn+skip) rather than fail** — rejected for *validation* (key/URL/auth shape): a misconfiguration should fail loudly and deterministically; the *runtime* discovery failure is separately warn+skip (Decision 3). (The two are deliberately distinct.)
  - **Add a new MCP-specific error phrase / exit code** — rejected (FR-015): no new failure class; reuse the existing surfaces.

## Decision 7: Hermetic E2E via a fake MCP server; a manual live check at closeout

- **Decision**: exercise the acceptance path with a **fake MCP server** (`net/http/httptest`, in-process) serving the Streamable-HTTP MCP shape — mirroring tellme's existing `httptest` fake provider — so discovery, registration, calling, non-stall, and `ENABLED` are asserted hermetically. The fixed discovery bound is witnessed by configuring a fake that never answers. `make verify` stays offline. A **manual** live confirmation against a real remote MCP endpoint is a closeout step, not part of the automated gate (round-013/031 precedent).
- **Rationale**: tellme's E2E strategy is black-box against the built binary with fakes for the network edge; the unreachable-server non-stall path is exactly what a "never answers" fake proves, and it is the fault the round exists to prevent. Keeping the live check manual preserves the offline gate.
- **Alternatives considered**:
  - **A build-tagged live E2E leg** — deferred: adds harness scope for a confirmation the manual closeout check already gives.
  - **Unit-only** — rejected: the capability is end-to-end observable (offer → call → answer), so E2E is the honest acceptance layer.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add the **MCP client** rows (config, domain port, remote HTTP adapter, auth resolution, non-stall discovery, tool naming) under CLI Application / Reasoning & Provider Transport; move **"MCP client SDK"** out of *Not Introduced Yet*; add *Not Introduced Yet* bullets for the **stdio transport**, **cross-invocation tool caching**, and **MEMORY/PLUR integration**; update the Testing & Verification section (the fake MCP server; the manual live check).
- `specs/truth/features/cli/**` → **ADD** expected — the executable MCP interface truth (offering a server's tools, calling one, the non-stall behaviour, the `ENABLED` switch), authored by `/axb-dsl-refine`.
- `specs/truth/contracts/**` → **NOOP** (single CLI end; no OpenAPI surface).
- `specs/truth/data/**` → **NOOP** (no persisted state; caching deferred).

## Residual risks / forward items (not blocking)

- **First protocol SDK**: the SDK is a new third-party dependency in tellme's graph; it is walled behind the domain port (Decision 2) and its version is pinned, but a future replacement is a recorded, contained change.
- **Concrete fast-fail bound value** (default 3 s) is an RD-tunable constant; the *invariant* (small fixed bound, independent of the tool-call timeout) is the requirement, not the exact number.
- **stdio transport, cross-invocation caching, MEMORY/PLUR** — recorded forward items (Q1/Q2/Q5); each is a separate future round.
