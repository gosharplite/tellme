# Feature Specification: remote MCP client (round 032)

**Feature Branch**: `032-mcp-client`

**Created**: 2026-09-16

**Status**: Draft — scope resolved from an operator request. Clarify locked in-session (Q1–Q5; see below).

**Input**: Operator request: *"Let tellme support MCP."* The operator also reported the concrete pain that motivated the round: *"earlier today hf was off-line, and tell-me-go has to wait for the timeout at the beginning of every command. Really bad!"*

**Scope note**: this is a **new-capability** round — tellme gains a **remote** Model Context Protocol (MCP) **client** so tools exposed by a configured MCP server become available to the agent. It is deliberately the **remote Streamable HTTP** transport only; the **local stdio** transport (`COMMAND`) and the automatic **MEMORY / PLUR** integration are deferred to later rounds. The reference implementation (`tell-me-go`) is the capability benchmark; tellme re-specifies the same capability with its own discipline — in particular it **must not** inherit the reference's startup-stall defect an unreachable server causes.

**Clarify (locked in-session, one decision at a time)**:

- **Q1 → 1** — transport scope: **remote Streamable HTTP only** (stdio `COMMAND` deferred).
- **Q2 → 1+3** — an unreachable/slow server must never stall a turn: a **fixed, small fast-fail discovery bound** (never the tool-call timeout) **plus** a per-server **`ENABLED`** switch (no cross-invocation cache this round).
- **Q3 → 1** — implementation: adopt the official **MCP Go SDK**, confined to an infrastructure adapter behind a new `tools.MCPClient` domain port (a `verify-mcp-sdk-confinement`-style gate).
- **Q4 → 1** — auth: **full parity** (`auto`/`gh`/`bearer`/`basic`/`none`); tool naming mirrors the reference (`mcp_<server>_<tool>`, 64-byte hash-truncated, sorted registration, fail-soft).
- **Q5 → 1** — scope: **MCP client only**; stdio + MEMORY/PLUR + cross-invocation caching are out.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consume tools from a configured remote MCP server (Priority: P1)

As an operator who runs tellme against a remote MCP server, I want tellme to discover that server's tools and offer them to the model alongside its native tools, so the agent can call them within a turn and feed their results back into the conversation.

**Why this priority**: This is the capability itself — the whole point of the round. Without it, nothing else in the round has value. It is the first value to prove and the foundation the other stories build on.

**Independent verification**: point tellme at a reachable MCP server (a hermetic fake in tests), run a prompt whose model answer requests one of the server's tools, and confirm the tool is executed on the server, its result is fed back, and the turn completes with the final answer.

**Acceptance Scenarios**:

1. **Given** a configured, reachable remote MCP server, **When** tellme starts a prompt-bearing turn, **Then** the server's tools are offered to the model (under namespaced names) alongside tellme's native tools.
2. **Given** the model requests one of the server's tools during a turn, **When** tellme executes the request, **Then** the tool runs on the configured server and its result is fed back into the turn like a native tool result, and the turn completes with a final answer.
3. **Given** a server configured with an explicit token, **When** tellme calls the server, **Then** the request carries the resolved credentials (and an anonymously-configured server carries none).

**Functional Requirements (FR)**:

- **FR-001**: tellme MUST load a remote MCP server registry from its existing configuration (`MCP_SERVERS`), supporting at least `URL`, `TOKEN`, `USERNAME`, `AUTH`, `TIMEOUT`, and `ENABLED` per server.
- **FR-002**: tellme MUST validate each **remote** MCP server entry deterministically — server key format (`^[a-z0-9-]{1,24}$`), a non-empty `URL`, a known auth mode and the credentials that mode requires, and a non-negative `TIMEOUT` — and MUST reject an invalid **remote** entry with a stable, classed failure. **Transport shape is classified first (FR-013):** a `COMMAND`-shaped entry is warn+skipped and is **never** evaluated as a remote entry, so FR-002's remote validation does not make it fatal. **Unmodelled sub-keys** under a server (e.g. a real-world reference's `REQUIRES_CONSENT`/`ARGS`/`DIR`/`ENV`) MUST be **tolerated** (ignored), consistent with round 003's tolerant top-level decode, so an existing `tell-me-go` configuration never makes tellme refuse to start.
- **FR-003**: On a prompt-bearing turn, tellme MUST discover the tools a configured, enabled remote MCP server offers and offer them to the model **in addition to** its native tools.
- **FR-004**: Each discovered tool MUST be offered under a **deterministic, namespaced** name derived from the server key and the tool name, and MUST NOT collide with native tool names or with another server's tools.
- **FR-005**: When the model requests a discovered MCP tool, tellme MUST execute that tool on the configured server and feed the result back into the turn, subject to the same per-tool bounding as native tools.
- **FR-006**: tellme MUST resolve server credentials by auth mode — `auto` (explicit token wins, else a GitHub-hosted endpoint falls back to a token source, else anonymous), `gh`, `bearer`, `basic`, `none` — and attach them to every request to that server; a resolved token MUST NOT be logged.
- **FR-007**: Registered MCP tools MUST be ordered deterministically (stable across runs regardless of discovery timing).

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The MCP integration MUST be exercisable **hermetically** (offline, no external network) in the ordinary test suite via a fake MCP server.

---

### User Story 2 - An unreachable or slow MCP server never stalls the run (Priority: P2)

As an operator whose configured MCP server is offline or hanging, I want tellme to skip it quickly and complete my command normally, so that a dead server never makes every run wait on a timeout (the pain that forced me to comment the server out).

**Why this priority**: This is what makes US1 usable. The reference blocks startup until the slowest server answers (a 30 s discovery cap, with the handshake bounded by the per-server timeout); one offline server therefore delayed *every* command. tellme must bound this to a small fixed window. It depends on US1 existing, so it ranks below it — but it is a first-class requirement, not a nicety.

**Independent verification**: configure a server that never answers; run a prompt; confirm the run completes, the dead server is skipped with a warning, and the discovery-attributable startup delay stays within the fixed bound.

**Acceptance Scenarios**:

1. **Given** a configured server that never responds, **When** I run a prompt-bearing turn, **Then** tellme begins the turn within the fixed discovery bound (skipping that server with a warning) and completes instead of hanging on a long timeout.
2. **Given** one server is skipped, **When** the turn runs, **Then** no tool from the skipped server is offered, and every other configured server and all native tools still work.
3. **Given** several configured servers, **When** discovery runs, **Then** it is performed concurrently and the run's total discovery-attributable delay is bounded by the fixed bound regardless of how many servers are configured.

**Functional Requirements (FR)**:

- **FR-008**: MCP discovery MUST be bounded by a fixed, small per-server deadline that is **independent of** (and much smaller than) the per-tool call timeout; a server that does not answer within it MUST be skipped with a warning.
- **FR-009**: A skipped or failed server MUST NOT abort the run, MUST NOT block discovery of other servers, and MUST NOT prevent the turn from proceeding with the tools it does have.
- **FR-010**: The overall delay attributable to MCP discovery MUST NOT exceed the fixed bound, independent of the number of configured servers or their health.

**Non-Functional Requirements (NFR)**:

- **NFR-002**: The fixed discovery bound MUST be small enough that an unreachable server does not materially delay a turn (a low single-digit-second ceiling), and it MUST be observable/verifiable.

---

### User Story 3 - Disable a server declaratively instead of commenting it out (Priority: P3)

As an operator with a known-flaky server, I want to turn that server off with a configuration switch, so I never again have to comment the entry out (losing its definition) just to keep my commands fast.

**Why this priority**: A convenience/robustness control that complements US2. US2 bounds the *unexpected* case; this removes the reason to hand-edit the config for a *known* case. It is independently verifiable and small.

**Independent verification**: set `ENABLED: false` on a server, run a turn, and confirm tellme makes no connection attempt to it and offers none of its tools.

**Acceptance Scenarios**:

1. **Given** a server marked `ENABLED: false`, **When** tellme runs a turn, **Then** it makes no connection attempt to that server and offers none of its tools.
2. **Given** an absent `ENABLED` key, **When** tellme runs a turn, **Then** the server is treated as enabled (default true).

**Functional Requirements (FR)**:

- **FR-011**: A remote MCP server marked `ENABLED: false` MUST be skipped entirely — no connection attempt, no discovery, no registered tools.
- **FR-012**: `ENABLED` MUST default to true so existing configurations are unaffected.

**Non-Functional Requirements (NFR)**:

- **NFR-003**: The `ENABLED` switch is an **availability** control (whether tellme talks to the server at all), not a security/consent control.

---

### Edge cases

- **A configured server returns malformed or unexpected protocol data** → tellme MUST skip it (warn) without aborting the run, consistent with FR-009.
- **A server's tool name produces a namespaced name longer than the wire maximum** → tellme MUST produce a deterministic, still-unique name (FR-004).
- **Two servers expose a tool of the same name** → the namespaced names MUST remain distinct (FR-004).
- **No `MCP_SERVERS` configured (or all disabled)** → tellme MUST behave exactly as before the round; no connection attempt, no change to the offered native tool set.
- **An MCP tool call fails at the server (tool-level error)** → the failure MUST be surfaced as a recoverable tool result fed back into the turn (the loop continues), distinct from a transport failure.
- **An MCP tool call exceeds its deadline** → it MUST be surfaced through the existing tool timeout result path (no new failure class).

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain several stories or cannot be reasonably attributed to a single one.

### Global requirements

#### Functional Requirements

- **FR-013**: This round MUST support the **remote Streamable HTTP** transport only; the **local stdio** (`COMMAND`) transport is out of scope (a later round). A `COMMAND`-shaped entry MUST be **skipped with a warning** — deterministically, without contacting it and **without failing the run** — so a real-world config carrying a stdio server does not invert the round's purpose (replacing "wait 30 s" with "tellme won't start"). **Transport shape is classified first:** a `COMMAND` entry (no `URL`) is never judged an "invalid remote entry" under FR-002 — the remote-entry validation applies only to a genuine remote entry.
- **FR-014**: The round MUST NOT change the existing native tool surface, the sequential tool-execution model, or the tool resource contract; MCP tools MUST honour the same per-tool bounding/timeout as native tools.
- **FR-015**: The round MUST NOT change the frozen class-phrase vocabulary or the exit-code set; MCP integration failures MUST reuse the existing classed failure surfaces.
- **FR-016**: MCP discovery and any server contact MUST occur only on the prompt-bearing turn path; the offline paths (`--version`, `-d`, `-l`, prompt-less `--new`, boot) MUST make no network contact.
- **FR-017**: Any credential used for an MCP server (and any derived Authorization header) MUST NOT be logged or written to disk.
- **FR-018**: An MCP **tool-call failure MUST NOT abort the run**: **every** call-time failure — a tool-level error reported by the server, a **transport/connection failure**, or a call against an **already-closed client** — MUST be surfaced as a **recoverable tool result** (a nil-error result carrying the error text) fed back into the loop, so the turn continues to a final answer — consistent with FR-009. **No call-time failure produces a terminal run abort.**
- **FR-019**: Before a server's tool is offered to the model, its argument schema MUST be **normalized and verified** to be well-formed — a JSON object declaring its offered arguments (`properties`) and its mandatory arguments (`required`) with `required ⊆ properties` (the round-031 / issue #64 invariant). A non-object, unparseable, or otherwise unsafe schema MUST be **skipped with a warning** (the tool is not offered and cannot 400 the turn); tellme MUST NOT offer a third-party schema verbatim when it can fail a strict provider.
- **FR-020**: Resolving a server's credentials via the external `gh` token source MUST be **bounded by the same fixed fast-fail deadline** as discovery; on timeout/failure it MUST warn and fall back to anonymous (never fail the run). The token source MUST be an **injectable seam**, so tests never spawn `gh`.
- **FR-021**: An MCP tool's call timeout MUST follow the existing tool resource contract — a resolved default, overridable by the server's `TIMEOUT`, and **clamped by the contract's timeout ceiling — a fixed 7200 s** (distinct from the contract's **token-bound** ceiling of `effectiveBudget ÷ 2`); a server-set value MUST NOT defeat the ceiling. The MCP tool's default timeout is **300 s** (the networked-tool class, matching `execute_command` and the reference's MCP default), **not** the 30 s local-reader default.

#### Non-Functional Requirements

- **NFR-004**: tellme MUST remain POSIX-only (no Windows variant), consistent with its standing scope.
- **NFR-005**: The MCP client implementation MUST be isolated behind a domain boundary so the transport/externals remain swappable and the rest of the codebase does not depend on the protocol library.

### Key entities

- **MCP server (configuration)**: an entry under `MCP_SERVERS` — a stable server key, a remote `URL`, optional credentials (`TOKEN`/`USERNAME`) and `AUTH` mode, an optional `TIMEOUT`, and an `ENABLED` switch. Not persisted by tellme (read from configuration).
- **MCP tool**: a tool a remote server advertises — its name, description, and argument schema — surfaced to the model under a deterministic namespaced name.
- **MCP client (domain port)**: the seam through which tellme lists a server's tools and calls one, abstracting the transport/library.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: With a reachable fake MCP server configured, a prompt whose model answer requests one of its tools completes end-to-end: the tool is executed on the server, its result is fed back, and the final answer is produced (hermetic E2E; covers US1).
- **SC-002**: With a server that never answers, a prompt-bearing turn completes within the fixed discovery bound, with the dead server skipped and all other tools still available — witnessed by a falsifiability check (covers US2 / NFR-002).
- **SC-003**: A server marked `ENABLED: false` produces **no** connection attempt and offers **no** tools; an absent key behaves as enabled (covers US3).
- **SC-004**: Each auth mode (`auto`/`gh`/`bearer`/`basic`/`none`) resolves credentials as specified, and no credential is logged (unit-pinned; covers FR-006, FR-017).
- **SC-005**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green; the offline paths remain network-free; the native tool set and class-phrase/exit-code vocabulary are unchanged (covers FR-013…FR-016).

## Assumptions

- The reference's startup-stall is caused by **unbounded discovery on the critical path**; a small fixed fast-fail bound plus an `ENABLED` switch removes it (this is the round's defining requirement, not a nice-to-have).
- The active `github` server in the operator's config resolves under `auto` via its **explicit** `${GITHUB_TOKEN}` — the `gh` fallback (which spawns the external `gh` CLI) runs only when no explicit token is present.
- Adopting the official MCP Go SDK (Q3 → 1) is a deliberate techstack decision recorded by the truth owner (`techstack.md`), not a spec requirement; the SDK is walled behind the domain port (NFR-005).
- Manual live confirmation against a **real** remote MCP endpoint is a closeout check, not part of the hermetic automated gate (re-using the round-013/031 precedent that `make verify` stays offline).
- Cross-invocation tool caching (Q2 Option 2), the stdio transport (Q1 deferral), and MEMORY/PLUR (Q5) are **out of scope** and recorded as forward items.
- This is a **CLI-interface** capability round: `/axb-system-analysis` is expected to record the CLI end carried to `/axb-dsl-refine`; `/axb-api-plan` is **NOOP**; `/axb-data-plan` is **NOOP** (no persisted state — caching deferred); `/axb-dsl-refine` **ADDs** the executable MCP interface truth; the truth change also touches `specs/truth/techstack.md` (the MCP rows).
- **Recorded divergences (round 032):** `ENABLED` is a tellme **addition** (the reference has no such key — the operator's stopgap was to comment the block out); and `REQUIRES_CONSENT` is **dropped** (tellme has no consent layer — a settled exclusion). Recorded rather than silent, per the round-028 ADR-0004 precedent.
- **MCP diagnostics (`-d`) are out of scope this round** (TD8): discovery is prompt-path-only, so `-d` neither reports an MCP server's reachability nor its tools; a non-dialing MCP diagnostic is a recorded forward item.
- **MCP tool schemas are an untrusted input** (B1): a third-party server's advertised schema is normalized/verified before offering (FR-019), because the round-031 recurrence gate only sees the **static** `agentTools()` assembler and cannot see dynamically registered tools.
