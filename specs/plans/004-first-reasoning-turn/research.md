# Phase 0 Research: tellme First Reasoning Turn (Round 004)

Topic: the round-004 reasoning slice — turn **one non-tool prompt** into **one provider request** and print the response, per Clarify Round 1 (Q1 single in-memory non-streaming turn; Q2 OpenAI-compatible family first; Q3 provider-failure class `the provider request failed` + exit code `6`). This research settles the provider gateway architecture, HTTP transport, request assembly, response normalization, the deterministic failure contract, the network-path test strategy, and the **amendment of round-001 Decision 5's no-network capability guard** (the issue explicitly routes that amendment here).

Scope note: the language (`Go 1.26`), module, CLI flag layer, testing harness (`godog` + stdlib `testing`), and base tooling were locked in rounds 001–003. The system still has **one CLI end** (the operator terminal); the provider is an **external dependency reached by the CLI**, not a new system end — so the three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end / `godog` / E2E black-box) and are not re-decided here.

---

## Decision 1: Provider gateway as a domain port with one OpenAI-compatible HTTP adapter

- **Decision**: Introduce a provider gateway **port** in a new `internal/domain/llm` package — an interface `Gateway` with a single method `Complete(ctx, Request) (Response, error)`, plus the network-free value types `Request`, `Response`, and the typed `ProviderError`. Implement exactly **one** adapter for the OpenAI-compatible family in a new `internal/infrastructure/llm/openai` package (stdlib `net/http`). The `internal/cli` dispatch wires the port to the adapter (the composition seam) and runs the turn.
- **Rationale**: Matches the reference's provider-agnostic gateway role (`Chatter` talking to the `Provider` gateway) and keeps the domain layer **network-free**, so (a) the transport is swappable (Gemini/Anthropic families later), and (b) the turn is testable against an in-memory fake gateway without any HTTP. It is the seam every later agentic slice (tools, failover, history) builds on, so establishing it now — rather than inlining HTTP — avoids a later rewrite.
- **Alternatives considered**:
  - **Call `net/http` inline from `internal/cli`** (no port): fewest files, but couples transport to the CLI, provides no fake seam (forcing HTTP into every unit test), and turns later families/failover into a rewrite.
  - **A single flat `internal/provider` package** (interface + adapter together): simpler, but mixes the domain abstraction with infrastructure and forfeits the layered (domain / infrastructure) posture the project mirrors from the reference.

## Decision 2: HTTP transport = Go stdlib `net/http` (no provider SDK)

- **Decision**: Perform the provider call with the standard library `net/http` client and `encoding/json`, over the OpenAI-compatible Chat Completions wire shape. No provider SDK and no generic HTTP-sugar library.
- **Rationale**: The OpenAI-compatible request/response is a small JSON POST; stdlib is dependency-free, deterministic, fully observable, and trivially pointed at an in-process `httptest.Server` for the E2E fake. It keeps the dependency surface minimal (the project already favours stdlib over frameworks) and honours the offline/CI constraints (no vendored SDK, no background transports).
- **Alternatives considered**:
  - **`sashabaranov/go-openai`** (OpenAI SDK): convenient, but adds a dependency and hides the wire, making per-vendor behaviour harder to control, and it covers only the OpenAI shape.
  - **Official per-vendor SDKs** (`google.golang.org/genai`, `anthropic-sdk-go`): one dependency per family — premature while only one family ships, and contrary to the "one simple transport first" decision.

## Decision 3: Request assembly from the resolved provider entry

- **Decision**: Map the resolved `config.Provider` (carried on `resolution.Provider`, round-003 review finding #3) to the OpenAI-compatible request: `POST <URL>/chat/completions`; headers `Authorization: Bearer <API_KEY>` (when non-empty), `Content-Type: application/json`, with the configured `HEADERS` merged in; body `{"model": <MODEL>, "messages": [{"role": "user", "content": <prompt>}], "max_tokens": <MAX_TOKENS>}` (`max_tokens` omitted when `MAX_TOKENS <= 0`); `THINKING_LEVEL` carried as `reasoning_effort` when non-empty.
- **Rationale**: Satisfies spec FR-002 using exactly the fields round 003 landed. The endpoint convention (append `/chat/completions` to the base `URL`) matches the OpenAI-compatible family and the harness default (`https://api.deepseek.com` → `https://api.deepseek.com/chat/completions`; `https://api.openai.com/v1` → `…/v1/chat/completions`). Sending no body field the family does not understand keeps the request portable across `openai`/`deepseek`/`kimi`.
- **Alternatives considered**:
  - **Treat `URL` as the full request URL** (append no path): rejected — the round-003 `URL` is a base endpoint, so the family path must be appended.
  - **Map thinking only through `HEADERS`**: rejected as the primary mechanism — per-vendor thinking mapping varies; the round carries `reasoning_effort` and records the residual (below).

## Decision 4: Response normalization to a minimal answer

- **Decision**: Decode the OpenAI-compatible response and extract `choices[0].message.content` as the answer text. Treat an absent/empty `choices`, missing content, or unparseable JSON as a **provider failure** (Decision 5), never a spurious success. Tool-call parts are out of scope this round and ignored.
- **Rationale**: Satisfies spec FR-004 ("normalize to a minimal textual answer") — the operator must see the answer, not raw JSON. Keeping the normalized shape minimal defers the full provider-agnostic `Thought` model and tool-call shapes to a later slice (spec assumption).
- **Alternatives considered**:
  - **Introduce the full `Thought` model now**: deferred — the reasoning/tool-call `Thought` shape belongs to a later slice and is not needed to prove one turn.
  - **Print the raw response body**: rejected — exposes wire detail and fails the "minimal answer" requirement.

## Decision 5: Deterministic provider-failure contract mapping

- **Decision**: Collapse every turn failure into one typed `llm.ProviderError` that the CLI renders as a single stderr line `tellme: the provider request failed: <actionable detail>` with exit code `6`. Covered failures: transport errors (connection, TLS, timeout), non-2xx status (the status is included in the detail), and an uninterpretable or empty body (Decision 4). This class is distinct from configuration (`3`).
- **Rationale**: Satisfies spec FR-006/FR-007/FR-008 and Clarify Q3 (Option 1); it extends the frozen round-002 one-line contract and widens the pinned table `0/2/3/4/5` to include `6`, keeping provider/network failures programmatically separable from boot and configuration errors.
- **Alternatives considered**:
  - **Reuse the configuration class** (`3` + `the provider configuration is invalid`): rejected (Clarify Q3, Option 2) — conflates a network/provider failure with a configuration error.
  - **Per-status classes** (auth vs server vs timeout): rejected — over-engineering for one turn; the single class phrase plus the actionable trailing detail is sufficient and matches the existing contract.

## Decision 6: Network-path test strategy (local fake provider) + amended no-network guard

- **Decision**:
  - **E2E acceptance**: the `tests/e2e` suite starts an in-process `httptest.Server` acting as an OpenAI-compatible provider; each scenario's configuration points the provider `URL` at it. Assertions cover exit code, stdout (the answer) / stderr (the frozen phrase on failure), and the fake's **recorded request** (exactly one request; expected model/auth/body).
  - **Unit tests**: table-driven tests for request assembly and response normalization (pure helpers, no network).
  - **Guard amendment (amends round-001 Decision 5)**: **retire** the whole-binary capability guard ("no `net/http` in the `./cmd/tellme` closure" + "no dialing symbol") because the chat path now legitimately links `net/http`. Re-prove the **offline paths** (`--version`, `-d`, and no-prompt boot) with two witnesses: (i) a **no-dial canary** — point the configured provider `URL` at a recording sink and assert the offline paths leave it with **zero** connections; (ii) the retained **differential no-egress sandbox** for `-d`. Rewire the Makefile `verify-no-network` target and `TestDependencyGraphHasNoNetworkCapability` to the canary-based offline-path assertion and rename it to reflect the narrowed scope.
- **Rationale**: Satisfies spec NFR-001/NFR-003 and the issue's "amend the guard" requirement. The offline guarantee stays a genuine, falsifiable witness (a canary that must stay silent), while the chat path is allowed to dial. The whole-binary absence guard is simply no longer true by construction once a network-bearing path exists.
- **Alternatives considered**:
  - **Split the diagnostic into a separate network-free binary**: rejected — an extra build artifact and a second `main`, disproportionate to one witness.
  - **Keep a static import-graph guard scoped to the diagnostic path**: rejected — `internal/cli` legitimately imports the adapter on the chat path, so a package-closure assertion cannot exclude the diagnostic dispatch.
  - **Drop the guard entirely**: rejected — loses the offline witness; the canary + sandbox keep it cheaply.

## Decision 7: Prompt argument handling & dispatch precedence

- **Decision**: Read the prompt from `pflag`'s positional arguments (`fs.Args()`, first positional), matching the reference's `tell-me-go "<prompt>"`. Preserve the existing dispatch order — `--version` → `-d` → **prompt turn** → boot — so an explicit mode flag takes precedence over a prompt, and a run with no positional prompt keeps the existing offline boot behavior.
- **Rationale**: Matches the operator-facing interface the spec and acceptance fix (FR-001; edge cases "no prompt" and "flag precedence") and keeps the existing flag modes authoritative. It also explains why the current `tm`/`tellme.sh` aliases' prompts are silently ignored today (positional args are parsed but unused).
- **Alternatives considered**:
  - **A `--prompt` flag**: adds surface and diverges from the reference and the aliases.
  - **A `chat` subcommand**: deferred with the rest of `cobra` (subcommands land with `browse`/`retry`).

## Residual risks / forward links

- **Thinking-option wire mapping (Decision 3)**: `THINKING_LEVEL` → `reasoning_effort` is a best-effort mapping for the OpenAI-compatible family; per-vendor thinking semantics (e.g. Gemini `THINKING_BUDGET`, Anthropic thinking blocks) are deferred to the slice that adds those families. `THINKING_BUDGET` is configuration-modeled but not mapped onto the OpenAI-compatible wire this round.
- **Guard amendment (Decision 6)**: the offline-path witnesses are a **canary (zero-connection)** plus the **differential sandbox** — necessary-condition witnesses, not an absence *proof* at the binary level (the retired guard supplied that). Stated honestly: with a network-bearing chat path, absence-of-capability is no longer a whole-binary claim; it becomes an offline-path behaviour claim.
- **Provider families (Decision 2)**: only the OpenAI-compatible family ships; Gemini/Vertex and Anthropic require additional adapters behind the same `Gateway` port (deferred).
- **`Thought` model (Decision 4)**: the response is normalized to a minimal answer; the provider-agnostic `Thought` (text / tool-call / reasoning segments) and tool-call shapes are deferred.
- **Slice 003 dependency**: this round consumes `resolution.Provider` (resolved, variable-expanded) exactly as round 003 carried it (review finding #3) — no re-loading or re-parsing.
