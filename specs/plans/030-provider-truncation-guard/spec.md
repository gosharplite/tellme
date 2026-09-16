# Feature Specification: provider-transport truncation guard (round 030)

**Feature Branch**: `030-provider-truncation-guard`

**Created**: 2026-09-16
**Status**: Draft — scope **operator-directed** (resolves issue [#62](https://github.com/gosharplite/tellme/issues/62)): make tellme's two provider transports **read the finish reason** and turn an **output-cap truncation** (`finish_reason == "length"` / `finishReason == "MAX_TOKENS"`) into a **loud, non-retryable provider failure** rather than returning a possibly-corrupted response.

**Input**: Operator request, verbatim: *"Let's resolve issue 62."* → clarify locks (both answered **"Option 1"**): the truncation trigger is **universal** (any output-cap truncation, text or tool call), and the failure **reuses the existing provider class** (`the provider request failed` + exit code 6).

**Scope note**: a **transport-layer** round. It reads the finish reason in **both** adapters (`internal/infrastructure/llm/openai` + `internal/infrastructure/llm/gemini`) and fails the turn on an output-cap truncation. It does **not** change the request side (no output cap is added or changed), the agent tool loop, the write tools, the CLI dispatch, the retry posture (there is none to change), or any non-transport behaviour. It does **not** add a security/consent gate.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A truncated tool call is refused, never silently written (Priority: P1)

As an operator driving tellme to edit its own repository, I want a provider response that was cut off at its output-token cap **mid-tool-call** to fail the turn loudly — not to hand an incomplete `write_file`/`replace_text` call to the tool — so the agent can never silently write a truncated or empty file.

**Why this priority**: This is the data-corruption path the round exists for. Round 029's write tools carry multi-KB arguments (`content`, `new_text`) that are the largest, last-emitted keys; a mid-string truncation can still decode while the args object loses the most important key, so a silent write corrupts a file (or writes an empty one). It is the round's core value; US2 can ship without it, not the reverse.

**Independent verification**: point a fake provider at tellme, script a response whose OpenAI `finish_reason` is `"length"` (or whose Vertex `finishReason` is `"MAX_TOKENS"`) at a `functionCall`; confirm the run fails with `tellme: the provider request failed: …`, exit code 6, the tool is **not** dispatched, and no file is created/edited.

**Acceptance Scenarios**:

1. **Given** the OpenAI-compatible provider truncates at the output cap (`finish_reason: "length"`) in a response that requests a tool call, **When** the turn runs, **Then** the turn fails with the provider class phrase and exit code 6, and the tool is **not** run.
2. **Given** the Vertex/Gemini provider truncates at the output cap (`finishReason: "MAX_TOKENS"`) on a `functionCall` part, **When** the turn runs, **Then** the turn fails the same way (provider class phrase + exit 6), and no file is written.

**Functional Requirements (FR)**:

- **FR-001**: The OpenAI-compatible adapter MUST read `choices[0].finish_reason`; the Vertex/Gemini adapter MUST read `candidates[0].finishReason`.
- **FR-002**: When the finish reason denotes an output-cap truncation (`"length"` for OpenAI-compatible; `"MAX_TOKENS"` for Vertex/Gemini), the adapter MUST fail the request with a provider/transport error **instead of** returning the response.
- **FR-003**: When the truncated response carries a tool call, the adapter MUST NOT return the call for execution — the tool MUST NOT be dispatched and MUST NOT write any file. The error detail MUST name the truncation; the **Vertex/Gemini** detail MUST additionally name the tool when the truncation site is a `functionCall` part (the OpenAI-compatible detail is family-agnostic — a deliberate reference-parity asymmetry, see `research.md` D6).

---

### User Story 2 - A truncated answer is surfaced, not passed off as complete (Priority: P2)

As an operator, I want any provider response cut off at its output-token cap — even one that carries only text and no tool call — to fail the turn loudly, so an incomplete answer is never presented as if it were the model's finished output.

**Why this priority**: this is the conservative, reference-mirroring extension of US1 (clarify Q1 → universal). It is P2 because the tool-call case (US1) carries the data-corruption risk; the text-only case is a correctness/consistency guard that the reference also pins.

**Independent verification**: script a truncated response with **no** tool call (`finish_reason: "length"` / `finishReason: "MAX_TOKENS"`); confirm the turn fails with the provider phrase + exit 6 and no answer is printed to `stdout`.

**Acceptance Scenarios**:

1. **Given** the provider truncates at the output cap in a text-only response, **When** the turn runs, **Then** the turn fails with the provider class phrase and exit code 6 (no partial answer is printed as the answer).

**Functional Requirements (FR)**:

- **FR-004**: The truncation guard MUST fire on **any** response whose finish reason denotes an output-cap truncation, **whether or not** it carries a tool call (the trigger is the finish reason, not the presence of a tool call).
- **FR-005**: A non-truncating finish reason MUST NOT trigger the guard — `"stop"` / `"tool_calls"` (OpenAI-compatible), `"STOP"` (Vertex/Gemini), and an **absent/empty** finish reason are all healthy (no false positives).

---

## Requirements *(mandatory)*

> The per-story FR are attached under each story above; this section holds only requirements that constrain both stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-006**: The truncation failure MUST surface through the **existing** provider-failure contract — the frozen class phrase `the provider request failed` and exit code **6** — carrying a detail that names the truncation. It MUST NOT add a new class phrase or a new exit code (clarify Q2 → reuse the provider class).
- **FR-007**: The truncation failure MUST NOT be silently retried or masked. tellme has **no** retry/failover/classification layer today; this round MUST NOT introduce one, and the guard MUST fail the turn rather than attempting recovery.
- **FR-008**: The guard MUST be implemented in **both** transport families — the OpenAI-compatible adapter and the Vertex/Gemini adapter.
- **FR-009**: The adapters MUST NOT change the **request** side. When `MAX_TOKENS` is unset tellme sends **no** output cap and the provider's own default governs; the guard still fires on whatever truncation the provider reports. (No default cap is added this round.)
- **FR-010**: Only output-cap truncation (`"length"` / `"MAX_TOKENS"`) is in scope. Other finish reasons are **out of scope** — their behaviour is unchanged this round — **explicitly including** Vertex/Gemini `SAFETY`, `RECITATION`, and `MALFORMED_FUNCTION_CALL` (a function-call integrity failure adjacent to this round's class, recorded as a tracked forward item), and OpenAI-compatible `content_filter`.

#### Non-Functional Requirements

- **NFR-001**: The round MUST add **no new third-party module** (Go standard library only) and MUST stay **POSIX-only** (tellme's standing scope).
- **NFR-002**: Verification MUST be **hermetic** — the in-process fake provider scripts the truncated wire shapes for **both** families (OpenAI `choices[0].finish_reason` and Vertex `candidates[0].finishReason`); no pty, no network. The guard is a pure decode-side check; it introduces no timing or nondeterminism.

## Key entities

- **None.** This round adds **behaviour** (a decode-side guard in the two transports), not persisted data; it introduces no new file format, table, or record.

## Edge cases

- **Vertex/Gemini success often omits `finishReason`**: an absent/empty finish reason MUST NOT trigger the guard (FR-005).
- **OpenAI-compatible normal tool-call completion** uses `finish_reason == "tool_calls"` — MUST NOT trigger the guard.
- **The truncated response's args JSON is unparseable**: the OpenAI-compatible adapter stores the raw arguments string **without** parsing it, so today this would NOT error on its own — the finish-reason guard is the residual-class catcher and MUST fire.
- **Truncation with empty content**: the guard SHOULD produce a truncation-specific error ahead of the generic "no usable answer" error, so the cause is unambiguous.
- **Truncation with both text and a tool call**: truncation dominates — the turn fails (no tool dispatched, no answer).
- **An unknown / new finish-reason value**: MUST NOT trigger the guard (only the two truncation values do).

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: A fake provider that returns `finish_reason:"length"` (OpenAI-compatible) or `finishReason:"MAX_TOKENS"` (Vertex/Gemini) makes the run fail with `the provider request failed` + exit code 6; the requested tool is **not** dispatched and no file is written (covers FR-001, FR-002, FR-003, FR-006, FR-008).
- **SC-002**: The same failure fires for a truncated **text-only** response (no tool call) — exit 6, no answer printed (covers FR-004).
- **SC-003**: Healthy responses (`stop`, `tool_calls`, Vertex/Gemini `STOP`, absent finish reason) do **not** error (covers FR-005) — the negative controls at both the unit and E2E layers.
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with **falsifiability witnesses**: removing the guard → the truncated-tool-call scenario silently returns/writes; removing a negative-control branch → a healthy response falsely fails.

## Assumptions

- tellme has **no** retry/failover/classification layer: every transport/provider failure becomes a `*llm.ProviderError`, which the CLI maps to the frozen phrase `tellme: the provider request failed: <detail>` and exit code 6. So the issue's "terminal (never auto-retried)" reduces to "a plain provider error"; the round must not add a retry.
- `MAX_TOKENS` is **unset by default**, so tellme sends no output cap and the provider's own default governs; the guard fires on the provider's reported truncation regardless. The issue asks only for this documented sentence — **no default cap is added** this round.
- A truncation failure is a **failed** turn: the agent loop records usage only for a **completed** call (round 018), so a refused truncation accounts **no** usage (token/cost) for that call. The reference returns `(content, metrics, err)` and still accounts the call; folding usage onto the error path would require a loop/`Gateway` change out of this round's scope, so the loss is **recorded** (`research.md` D1) — a conscious decision, not an unknown.
- The reference (`tell-me-go`) is the benchmark: its `checkGeminiTruncation` and its OpenAI `finish_reason == "length"` check are both **terminal** and **universal** (they pin a text-only case too), recorded as the evidence that motivated this round.
- The actual truth changes (`specs/truth/techstack.md` transport rows; a `features/cli/**` interface feature + `dsl.md` rows; `contracts/**` and `data/**` expected NOOP) are made by the truth-owner skills; this plan package records the intent only (`fresh-package-per-round`).
