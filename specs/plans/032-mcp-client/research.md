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
- **Rationale**: this is the round's **defining** requirement, and the fix for the operator's reported pain — the reference (`tell-me-go`) builds MCP clients at the start of **every** CLI invocation and its plugin's `Register` `wg.Wait()`s on a per-server `ListTools` capped at **30 s** (`internal/tools/integrations/mcp/plugin.go` → `defaultDiscoveryTimeout = 30 * time.Second`; ADR-067: *"`TIMEOUT` still gets its own 30s `ListTools` window"*), so one offline server delayed every command (the operator's workaround was to comment `hf` out). A small fixed bound makes an unreachable server a **non-event** (`wg.Wait()` now completes in ≈ the bound, not ≈ the slowest timeout), and `ENABLED` lets the operator switch a known-flaky server off without deleting its definition (Q2 → 1+3). Gating discovery on the prompt path preserves tellme's long-standing offline guarantee (round 004/005/009). **Note (R2):** the reference is **already concurrent** (`sort.Slice` + per-server goroutines + `wg.Wait()`), so this round's value is the **smaller fixed bound** and the **prompt-path-only** gating — not concurrency per se.
- **Alternatives considered**:
  - **Cross-invocation tool cache** (load a cached tool list instantly, refresh in the background) — deferred (Q2 Option 2): gives true ~zero startup cost but adds a new persisted artifact and staleness semantics that deserve their own round; the fast-fail bound + `ENABLED` already remove the operator's pain.
  - **A single overall deadline across all servers** rather than per-server — rejected as the primary mechanism: a per-server bound composes to the same overall ceiling under concurrency while letting a *healthy* slow server still be discovered; the per-server form is simpler to reason about and to witness.

## Decision 4: Full auth parity (`auto`/`gh`/`bearer`/`basic`/`none`), credentials never logged

- **Decision**: resolve per-server credentials by auth mode (default `auto`): an explicit `TOKEN` always wins; otherwise a **GitHub-hosted** hostname falls back to a token source (`gh auth token`, then `GITHUB_TOKEN`), and any other host is anonymous; `gh` forces the token source; `bearer` uses the explicit token; `basic` sends `Authorization: Basic base64(USERNAME:TOKEN)`; `none` is anonymous. Credentials (and derived Authorization headers) are **never logged** or written to disk (FR-017).
- **Rationale**: the operator's active `github` server resolves under `auto` via its explicit `${GITHUB_TOKEN}` — no `gh` spawn needed — while `gh`/`basic` cover the realistic follow-ons (e.g. a Basic-auth Atlassian endpoint). The set is small and fully specified by the reference; trimming it would surface later as "unexpected" failures (Q4 → 1). Avoiding credential logging mirrors the reference's `mcp-token-not-logged` invariant and tellme's standing hygiene.
- **B3 fold (review):** resolving a token by spawning `gh` is itself a **managed-process consumer** that must not become a *second* unbounded stall on the same prompt path. So: (i) token resolution is bounded by the **same fixed fast-fail deadline** as discovery — on timeout/failure it warns and falls back to **anonymous** (never fails the run); (ii) it is an **injectable `tokenResolver` seam** (default = the bounded `gh` spawn; unit/E2E inject a fake so **no `gh` process is ever spawned** in tests); (iii) this is the round-024 trigger's **second managed-process consumer** (after `execute_command`), so per the recorded trigger the seam is introduced **now** — a function-typed resolver matching the reference's `tokenResolver` — with the fuller process-runner port noted as a later form. FR-020.
- **Alternatives considered**:
  - **Trimmed set (`auto`/`bearer`/`none`)** — rejected (Q4): would drop the `gh` fallback and Basic auth that the operator explicitly asked for.
  - **Minimal (`bearer`/`none`)** — rejected: loses hostname detection that the default `auto` mode already implies.

## Decision 5: Deterministic namespaced tool names (reference parity)

- **Decision**: register each discovered tool under the deterministic name `mcp_<server>_<tool>`. When that exceeds the 64-byte wire maximum, the tool segment is truncated to a **derived budget** — `maxPrefixLen = 64 − len("mcp_") − len(server) − len("_") − len("_") − len(hash8)` (= `50 − len(server)`), **capped at 40, floored at 0** — and a stable 8-hex-char SHA-256 prefix of the full tool name is appended. The result is always ≤ 64 bytes: the `{1,24}` server-key bound exists **precisely** so the fixed segments leave room. Servers are registered in sorted key order so the offered set and its order are stable across runs regardless of discovery timing (FR-004/FR-007). **TD3 fold (review):** the truncation budget is **derived** (not a flat 40) and measured in **bytes (`len()`), not runes** (the wire limit is bytes) — pinned by a `server=24 + very long tool` unit.
- **Rationale**: matching the reference's naming keeps behaviour and debugging familiar, keeps the generated names inside the OpenAI `tools[i].function.name` limit (`^[a-zA-Z0-9_-]{1,64}$`), and the sorted order preserves tellme's determinism value (the same discipline behind its spinner/log/accounting orderings).
- **Alternatives considered**:
  - **Non-namespaced names** (pass the server's tool name through) — rejected: two servers could collide, and a server tool could shadow a native tool.
  - **User-visible server prefix with no hash** — rejected: exceeds 64 bytes for long tool names and loses uniqueness.

## Decision 6: Typed `MCP_SERVERS` config + deterministic validation; no new failure class

- **Decision**: add a typed `MCP_SERVERS` block to tellme's configuration (today merely tolerated as an unknown top-level key). Per server, accept `URL`, `TOKEN`, `USERNAME`, `AUTH`, `TIMEOUT`, `ENABLED` (default true). Validate deterministically: server key `^[a-z0-9-]{1,24}$`; a non-empty `URL`; a known effective auth mode; the credentials that mode requires; a non-negative `TIMEOUT`. Validation failures reuse the existing classed configuration-invalid failure surface; the round adds **no** new class phrase or exit code (FR-015). **TD4 fold (review):** consistent with round 003's deliberate **tolerant top-level decode** (`specs/plans/003-provider-registry-completeness/research.md:24`; `internal/config/config.go` keeps `KnownFields(true)` on `Config` as a documented, deferred seam), (a) **unmodelled sub-keys** under a server — `REQUIRES_CONSENT`, `ARGS`, `DIR`, `ENV` (carried by real `tell-me-go` blocks) — are **tolerated (ignored)**, and (b) a **`COMMAND`-shaped (stdio) entry is warn+skipped, not fatal** — so a real-world config never makes tellme refuse to start (which would invert the round's purpose: "wait 30 s" → "tellme won't start"). Only a malformed **remote** entry is a classed, fatal configuration-invalid failure. **TD7:** the divergence is recorded (below). MCP tool-call outcomes ride the existing tool-result paths (see Decision 9).
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

## Decision 8: MCP tool schemas are normalized/verified before offering (B1)

- **Decision**: before a server's tool is offered to the model, its `InputSchema` is **normalized and verified** (FR-019): a non-object / absent / unparseable schema is coerced to a freeform object, and the offered schema MUST be a JSON object declaring its `properties` with `required ⊆ properties`. A schema that cannot be made safe is **skipped with a warning** (the tool is not offered; the server's other tools still are) — tellme never offers a third-party schema verbatim that can 400 a strict provider. The round-031 / issue #64 well-formedness invariant's home is **extended to cover this normalized path** via a `[UNIT]` pin on the normalizer's output (the static `agentTools()` gate cannot see dynamically registered tools).
- **Rationale**: dynamic MCP tools reopen the exact #64 vector from the **outside** — a remote, non-reviewable author advertising `required` without a `properties` entry (or a non-object schema) produces the same "strict provider 400s the whole request" defect. The reference guards it (`convertSchema` + *"Unsupported schema combinators gracefully degrade to freeform arguments"*); tellme must too. Per-tool skip (not whole-server skip) keeps the turn alive and preserves the server's good tools.
- **Alternatives considered**:
  - **Offer the schema verbatim** — rejected: reopens #64 from the outside.
  - **Drop only the offending `required` entry** — rejected: silently changes the tool contract (the model may omit a mandatory argument); skip+warn is honest.
  - **Skip the whole server on a bad schema** — rejected: one bad tool would lose the server's good tools.

## Decision 9: MCP tool-call failures are recoverable, never abort the run (TD1)

- **Decision**: an MCP tool-call outcome is fed back like any tool result and **never aborts the turn** (FR-018): a **tool-level error** (`isError: true`) → a recoverable tool result carrying the error text; a **transport/connection failure** → likewise a recoverable tool result (the loop continues to a final answer). Neither is the terminal `the tool request failed` class phrase (reserved, unchanged, for an **unregistered tool name** loop abort).
- **Rationale**: resolves the tension the review found between Decision 6's original "a transport failure is a classed failure" and FR-009 ("a skipped or failed server MUST NOT abort the run") — the round's own motivation (a flaky server is a non-event) applies at **call time**, not only discovery. Matches the reference (*"`isError: true` returns error text with nil Go error for in-turn LLM recovery"*).
- **Alternatives considered**:
  - **Abort the run on a transport failure** (Decision 6's original wording) — rejected: contradicts FR-009; a transient mid-turn failure would kill the run.
  - **Recoverable only for `isError`; transport aborts** — rejected: same contradiction.

## Decision 10: Effective MCP tool-call timeout = the resource contract (TD5)

- **Decision**: an MCP tool's call timeout uses the existing **tool resource contract** (round 024): a resolved default (tellme's non-shell default, **30 s**), overridable by the server's `TIMEOUT`, and **clamped by the contract's ceiling** (`effectiveBudget ÷ 2` seconds; today 7200 s). A server-set `TIMEOUT` (e.g. the reference's `300`) MUST NOT defeat the ceiling. Unit-pinned (FR-021).
- **Rationale**: FR-014 requires MCP tools to honour the same bounding as native tools; leaving the default/clamp unstated would let a server config bypass tellme's resource contract.
- **Alternatives considered**:
  - **Reference default (300 s), unclamped** — rejected: bypasses the round-024 ceiling; inconsistent with native tools.
  - **No default (require `TIMEOUT`)** — rejected: worse ergonomics, no reference parity.

## Decision 11: The e2e fake MCP server is SDK-built inside the confined package; diagnostics deferred (B2, TD8, TD7)

- **Decision (B2)**: the hermetic e2e fake MCP server is built **with the SDK's server API** as an **exported test helper inside the confined package** — `internal/infrastructure/mcp/mcptest/` — mirroring the reference (whose fake lives inside `internal/infrastructure/mcp/` and *"starts a real SDK MCP server (same wire…)"*). The confinement gate (T007) covers **production and test files** and allows `internal/infrastructure/mcp/**` (including `mcptest/`). No hand-rolled Streamable-HTTP wire; the gate and the fake are **no longer mutually exclusive**.
- **Decision (TD8)**: `-d` diagnostics do **not** report MCP servers this round (discovery is prompt-path-only); a non-dialing MCP diagnostic is a recorded forward item — not silently omitted.
- **Decision (TD7)**: **recorded divergences** — `ENABLED` is a tellme **addition** (the reference has no such key); `REQUIRES_CONSENT` is **dropped** (tellme has no consent layer — a settled exclusion). Recorded, not silent (round-028 ADR-0004 precedent).
- **Rationale**: building the fake with the SDK keeps the wire correct (tests tellme, not a hand-rolled protocol) and locates it where the SDK is permitted, so T007 and T008 are consistent by construction.
- **Alternatives considered**:
  - **Hand-rolled Streamable-HTTP fake in `tests/e2e/`** — rejected: brittle, tests the fake, forces a gate exception.
  - **A gate allow-list for `tests/e2e/`** — rejected: widens the confinement surface; the SDK-built helper inside the package is narrower.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add the **MCP client** rows (config, domain port, remote HTTP adapter, credential resolution incl. the bounded/seam token resolver, non-stall discovery, deterministic naming, **MCP tool-schema normalization/well-formedness**, and the **timeout default/clamp**) under CLI Application; move **"MCP client SDK"** out of *Not Introduced Yet*; add *Not Introduced Yet* bullets for the **stdio transport**, **cross-invocation tool caching**, **MEMORY/PLUR integration**, and the **MCP `-d` diagnostic**; update the Testing & Verification section (the **SDK-built** fake MCP server in `internal/infrastructure/mcp/mcptest/`; the manual live check); record the **divergences** (`ENABLED` addition; `REQUIRES_CONSENT` dropped).
- `specs/truth/features/cli/**` → **ADD** expected — the executable MCP interface truth (offering a server's tools, calling one, the non-stall behaviour, the `ENABLED` switch), authored by `/axb-dsl-refine`.
- `specs/truth/contracts/**` → **NOOP** (single CLI end; no OpenAPI surface).
- `specs/truth/data/**` → **NOOP** (no persisted state; caching deferred).

## Residual risks / forward items (not blocking)

- **First protocol SDK**: the SDK is a new third-party dependency in tellme's graph; it is walled behind the domain port (Decision 2) and its version is pinned, but a future replacement is a recorded, contained change.
- **Concrete fast-fail bound value** (default 3 s) is an RD-tunable constant; the *invariant* (small fixed bound, independent of the tool-call timeout) is the requirement, not the exact number.
- **stdio transport, cross-invocation caching, MEMORY/PLUR** — recorded forward items (Q1/Q2/Q5); each is a separate future round.
- **MCP `-d` diagnostic (TD8)** — out of scope this round (discovery is prompt-path-only); a non-dialing report is a recorded forward item.
- **Recorded divergences (TD7)** — `ENABLED` is a tellme addition; `REQUIRES_CONSENT` is dropped (no consent layer).
- **Acceptance-carrier completeness (TD6)** — the interface truth carries the multi-server, alongside-native-tools, default-on, malformed-schema, and tool-failure journeys; any residual gap is recorded explicitly (round-031 ARCH-5 precedent), never silently dropped.
