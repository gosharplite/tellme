# Feature Specification: tellme First Reasoning Turn (round 004)

**Feature Branch**: `004-first-reasoning-turn`

**Created**: 2026-09-12

**Status**: Draft

**Input**: User request and GitHub Issue #10: "004 — First reasoning turn: one prompt → provider → response" — refined by Clarify Round 1 (2026-09-12):
- **Q1 -> Option 1**: A single, in-memory, non-streaming turn per process; no persistence (`history.jsonl`), no session loop.
- **Q2 -> Option 1**: The first concrete provider adapter targets the OpenAI-compatible family (`openai` / `deepseek` / `kimi`).
- **Q3 -> Option 1**: A provider/transport failure emits a new frozen class phrase `the provider request failed` and a new distinct exit code `6`.

*(Not asked this round — deferred to `/axb-technical-research` as assumptions: transport = stdlib `net/http` with no provider SDK; the round-001 no-network capability guard is re-scoped to the offline boot/`--version`/`-d` paths while the prompt-bearing chat path may dial the provider.)*

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run a single reasoning turn from the terminal (Priority: P1)

As an operator, I want to pass a prompt to tellme and receive a real answer from my configured provider in the terminal, so that tellme becomes a usable reasoning assistant instead of inert scaffolding.

**Why this priority**: This is the core value of the slice and the first reasoning capability of the whole project. Everything later (streaming, history durability, tools, MCP, memory) is built on a single working turn, so this is the first value to prove.

**Independent verification**: Run `tellme "<prompt>"` against a local fake OpenAI-compatible provider that returns a known message; confirm exactly one request is sent to the resolved endpoint with the resolved model and credential, the returned text is printed, and the process exits successfully.

**Acceptance Scenarios**:

1. **Given** a resolved configuration that selects an OpenAI-compatible provider (endpoint, model, and an expanded credential), **When** I run `tellme "<prompt>"`, **Then** tellme sends exactly one provider request carrying the prompt with the resolved model and credential, prints the provider's answer text, and exits with the success code.
2. **Given** the provider returns a valid text answer, **When** the turn completes, **Then** the printed output is the provider's answer text alone (no tool calls, no multi-turn loop, no session state), and the process exits with the success code.

**Functional Requirements**:

- **FR-001**: The system MUST accept the prompt as a positional command-line argument; when such an argument is present, the system MUST run exactly one reasoning turn.
- **FR-002**: The system MUST assemble the provider request from the resolved configuration for the effective selected provider — its endpoint (`URL`), model (`MODEL`), expanded credential (`API_KEY`), `MAX_TOKENS`, `HEADERS`, and thinking options.
- **FR-003**: The system MUST send exactly one provider request per invocation and consume the provider response in a non-streaming manner.
- **FR-004**: The system MUST normalize the provider response to a minimal textual answer and print that answer to standard output.
- **FR-005**: The system MUST exit with the success code after printing the response.

---

### User Story 2 - Deterministic provider-failure reporting (Priority: P2)

As an operator or automation-script author, I want a failed reasoning turn to fail with a fixed class phrase and a distinct exit code, so that I can tell a provider/network failure apart from boot and configuration errors programmatically.

**Why this priority**: The reasoning path is the first tellme code to contact the network, so its failure surface is new and must slot into the frozen round-002 failure contract; without a distinct class, scripts cannot distinguish a provider/network failure from a configuration error.

**Independent verification**: Point tellme at a fake provider that refuses the connection, returns an error status, times out, or returns an uninterpretable body; confirm for each that stderr begins with the frozen class phrase and the process exits with code `6`.

**Acceptance Scenarios**:

1. **Given** a prompt is supplied and the provider or transport fails (connection refused, timeout, or an HTTP error status), **When** I run tellme, **Then** stderr emits `tellme: the provider request failed` and the process exits with code `6`.
2. **Given** the provider returns a response that cannot be interpreted as a valid answer, **When** I run tellme, **Then** the run fails with the same class phrase and code `6` instead of printing a spurious success.

**Functional Requirements**:

- **FR-006**: When the provider request fails at the transport level (connection, TLS, or timeout) or the provider returns an error status, the system MUST emit exactly one stderr line beginning with the frozen class phrase `tellme: the provider request failed` and MUST exit with numeric code `6`.
- **FR-007**: When the provider response cannot be interpreted as a valid answer (e.g. an empty or structurally unusable body), the system MUST fail with the same class phrase and code rather than report success.
- **FR-008**: The trailing detail following the class phrase MUST be actionable — naming the provider and the underlying failure reason — while the class phrase and exit code remain the frozen contract.

---

### Edge Cases

- When no prompt argument is supplied, tellme MUST keep its existing behavior (boot, `--version`, `-d`) and MUST NOT contact any provider.
- When an explicit mode flag (`--version` or `-d`) is present together with a prompt, the explicit mode MUST take precedence and no provider request is made.
- When `MAX_TOKENS`, `HEADERS`, or thinking options are omitted from the provider entry, the request MUST still be assembled with usable defaults so a valid provider call can be made.
- When the provider endpoint is unreachable (connection refused), the same provider-failure contract (class phrase + code `6`) applies.
- When the provider returns an empty answer, the run MUST fail with the provider-failure class rather than succeed with no output.
- When the provider returns a successful status whose body cannot be parsed into an answer, the run MUST fail with the provider-failure class rather than print a partial or empty success.

## Requirements *(mandatory)*

> Story-specific FR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-009**: The system MUST extend its deterministic exit-code table with code `6` for provider/transport failure, distinct from success (`0`), usage (`2`), configuration (`3`), environment (`4`), and diagnostic-unresolved (`5`); the round-001, round-002, and round-003 codes and class phrases MUST remain unchanged.

#### Non-Functional Requirements

- **NFR-001**: The boot, `--version`, and `-d` diagnostic paths MUST remain strictly offline and deterministic — zero network calls; only the prompt-bearing chat path performs network egress.
- **NFR-002**: The round MUST NOT regress the existing round-001/002/003 CLI contracts, exit codes, or class phrases; all pre-existing end-to-end scenarios MUST remain green.
- **NFR-003**: The reasoning path MUST be verifiable end-to-end against a **local fake provider** with no external network access, and the request-assembly and response-normalization helpers MUST be covered by fast, isolated unit tests.
- **NFR-004**: The reasoning path MUST be deterministic — identical provider responses MUST produce identical printed output and exit codes across repeated runs.

### Key Entities *(include if feature involves data)*

- **Prompt**: The operator's single input string passed as a positional command-line argument.
- **ProviderRequest**: The assembled outbound request for one turn — endpoint, authentication derived from the expanded `API_KEY`, model, `MAX_TOKENS`, `HEADERS`, thinking options, and the prompt.
- **TurnResponse**: The normalized minimal answer text extracted from the provider's response and printed to stdout.
- **ProviderFailureContract**: The operator-facing failure specification binding the frozen class phrase `the provider request failed` to exit code `6`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of valid single-turn runs send exactly one provider request and print the provider's answer text, exiting with code `0`.
- **SC-002**: In the acceptance set, 100% of provider/transport failure runs emit exactly one `tellme: the provider request failed` stderr line and exit with code `6`.
- **SC-003**: All pre-existing round-001/002/003 acceptance scenarios remain green, and the boot, `--version`, and `-d` paths perform zero network calls.
- **SC-004**: The single-turn acceptance path runs entirely against a local fake provider (no external network), and the request-assembly and response-normalization helpers achieve full unit-test coverage.

## Assumptions

- Round 004 delivers exactly one in-memory, non-streaming turn per process; a multi-turn session loop, streaming output, `Turn`/`History` persistence, summarisation, cost/metrics, tools, MCP, and memory are deferred to later slices (Clarify Q1).
- The first concrete adapter targets the OpenAI-compatible family (`openai`, `deepseek`, `kimi`); the Gemini/Vertex and Anthropic families are deferred (Clarify Q2).
- A provider/transport failure is reported with the new frozen class phrase `the provider request failed` and the new distinct exit code `6` (Clarify Q3).
- Transport is the Go standard library HTTP client with no provider SDK added; this is ratified by `/axb-technical-research` (not fixed here).
- The no-network capability guard from round-001 research Decision 5 is re-scoped: it asserts the boot, `--version`, and `-d` paths are network-free, while the prompt-bearing chat path is permitted to dial the provider; the amendment is owned by `/axb-technical-research`.
- The prompt is supplied as a positional argument; explicit modes (`--version`, `-d`) take precedence over it.
- Round 004 persists no session state; the request carries the single-turn prompt only (no prior-history window).
- Provider context-window and pricing mapping remain deferred to later runtime slices.
