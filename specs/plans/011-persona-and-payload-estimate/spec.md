# Feature Specification: tellme Persona-on-the-Wire & Wire-Faithful Payload Estimate (round 011)

**Feature Branch**: `011-persona-and-payload-estimate`

**Created**: 2026-09-13

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "tellme needs to have `System/persona message`" and "tellme needs to estimate what will be send on the wire, which will be compare to the post-status token." This round makes `tellme`'s **outbound request faithful to its configuration and its pre-flight estimate faithful to the wire**. Two coupled defects are closed: (i) the configured `PERSON` persona is **parsed but never sent**, so the model never receives the operator's instruction; and (ii) the pre-flight payload estimate counts **only conversation message text**, so a trivial prompt reports `~5` while the provider measures `387` — the estimate ignores the tool declarations and (once (i) lands) the persona message that dominate the real request.

This round's behaviour intent is **MODIFY** — it changes the request **content** (adding the persona message) and the **inputs** the estimator consumes (persona + tool declarations + messages). It introduces no new capability surface, no new dependency, and no change to the answer stream or the payload-status **line format**. It strengthens the request contract round 004/008 established and closes the observability gap round 009's estimate shipped.

**Clarify Round 1 (2026-09-13)** resolved two high-impact decisions:

- **Q1 → Option 1 (estimate scope)**: the pre-flight estimate MUST count the **persona/system message + the tool declarations + the conversation messages** — the three input components the provider's `prompt_tokens` covers.
- **Q2 → Option 1 (guarantee strictness)**: the round pins the estimate's **inputs** (persona + tool declarations + messages) **and determinism**; the estimate is **not** required to equal the provider's reported `prompt_tokens` (no numeric-tolerance assertion). The exact heuristic (chars/token, per-tool allowance, any base term) is a `/axb-technical-research` determination, not a spec-level contract.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The configured persona reaches the model (Priority: P1)

As an operator, I want the persona/system instruction I configured (`PERSON`) to be sent to the provider on every reasoning request, so the model's answers follow the instruction I set — for example "respond concisely and accurately in English" — instead of the provider's default assistant behaviour.

**Why this priority**: The persona is the operator's primary control over model behaviour, and today it is silently dropped: `tellme` parses `PERSON` but never puts it on the wire, so every run behaves as a generic assistant regardless of configuration. It is the root cause of part of the `~5`/`387` gap and the smallest independently verifiable increment (one observable request-content change).

**Independent verification**: configure a non-empty `PERSON`, run a prompt against the in-process fake provider, and confirm the recorded request's `messages` array begins with a `system` message whose content equals the configured `PERSON` value, ahead of the conversation.

**Acceptance Scenarios**:

1. **Given** a resolvable configuration whose `PERSON` is a non-empty instruction and whose selected provider is reachable, **When** the operator starts tellme with a prompt, **Then** the request recorded by the provider carries a **leading `system` message** whose content is that `PERSON` value, before the conversation messages.
2. **Given** a configuration whose `PERSON` is empty, **When** the operator starts tellme with a prompt, **Then** the recorded request carries **no** `system` message (the conversation is unchanged from the no-persona shape).

**Functional Requirements**:

- **FR-001**: On a prompt-bearing run, the system MUST include the configured `PERSON` value as the **first** (`system`-role) message of the request sent to the provider.
- **FR-002**: The persona message MUST precede the conversation messages (prior turns followed by the current prompt) and MUST NOT be merged into, or prepended to, the user prompt text.
- **FR-003**: When `PERSON` is empty, the system MUST send **no** persona message.
- **FR-004**: The persona MUST be carried on **every** provider request made during the run — the main completion and any tool-driven completion (e.g. the session-summarisation tool).
- **FR-005**: The persona is request-only: it MUST NOT appear in `stdout` and MUST NOT alter the answer byte contract.

**Non-Functional Requirements**:

- **NFR-001**: The persona message MUST be deterministic and verifiable **offline** against the in-process fake provider.
- **NFR-002**: The persona content MUST be sent **verbatim** (byte-for-byte the configured value), with no truncation or decoration.

---

### User Story 2 - The pre-flight estimate reflects the wire payload (Priority: P2)

As an operator, I want the per-turn pre-flight `~` estimate to measure the same input the provider will tokenize — the persona message, the tool declarations, and the conversation — so the number I see before the request is comparable to the measured post-turn count and I can trust it as a payload-budget signal.

**Why this priority**: It depends on Story 1 (the persona must be on the wire before the estimate can count it) and on the tool declarations already sent since round 008. It closes the observability defect where the estimate reported `~5` for a request the provider measured at `387`, making the pre-flight line structurally misleading.

**Independent verification**: run a prompt against the fake provider and confirm the pre-flight estimate reflects the persona + tool declarations + messages — e.g. a run whose wired payload is strictly larger (a longer persona, or more declared tools) reports a strictly larger estimate — and that identical inputs always yield the same estimate.

**Acceptance Scenarios**:

1. **Given** a prompt run with a non-empty persona and one or more declared tools, **When** the run completes, **Then** the pre-flight estimated payload is computed over the persona message, the tool declarations, and the conversation messages.
2. **Given** two runs that differ only by a strictly larger wired payload (a longer persona, or an additional declared tool), **When** both run, **Then** the run with the larger payload reports a strictly larger pre-flight estimate.
3. **Given** identical inputs, **When** the run is repeated, **Then** the pre-flight estimate is identical.

**Functional Requirements**:

- **FR-006**: The pre-flight estimated payload MUST be computed over the same input components sent on the wire — the **persona/system message**, the **tool declarations** sent in the request, and the **conversation messages** (prior turns + current prompt).
- **FR-007**: The estimate MUST be **deterministic** for identical inputs.
- **FR-008**: The estimate MUST be **responsive** to the wired payload — it MUST increase when the counted payload grows (e.g. a longer persona, or additional tool declarations).
- **FR-009**: The estimate is **not** required to equal the provider's reported `prompt_tokens`; the round pins the estimate's inputs and determinism, not numeric equality (no tolerance assertion).

**Non-Functional Requirements**:

- **NFR-003**: The estimator MUST remain **offline and dependency-free** (no BPE tokenizer); the existing hand-written heuristic is extended in **inputs**, not replaced by a tokenizer dependency.

---

### Edge Cases

- When `PERSON` is **empty**, no persona message is sent, and the estimate counts the tool declarations + conversation messages only.
- When the prompt arrives via **piped stdin**, the persona and estimate behaviour MUST be identical to a positional-argument prompt.
- When **`-r/--raw`** is set, the payload-status lines MUST still be written to `stderr` and `stdout` MUST remain raw and byte-exact.
- When a **non-prompt path** runs (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`), **no** provider request is made, so no persona is sent and no payload-status line is emitted (the contract is vacuous there).
- When a turn makes **tool calls**, the request carries the tool declarations, so the estimate MUST count them on that turn too.
- When the provider reports **no usage**, the measured line is omitted; the pre-flight estimate still counts the wired payload.
- When the persona value is **large**, the estimate MUST reflect it (FR-008) and the verbatim content MUST be preserved (NFR-002).

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-010**: The `stdout` answer stream MUST remain byte-exact — unchanged by the persona message and by the estimate change (rounds 005/006 fidelity preserved).
- **FR-011**: The change MUST be reflected in the **executable interface truth**: the request contract — now carrying a leading `system` message — MUST be assertable, and the existing assertion that a first-turn request "carries no earlier exchange / messages are exactly the current user prompt" MUST be updated to account for the leading persona message.
- **FR-012**: The **measured** (post-turn) payload status line MUST continue to source the provider's reported `usage.prompt_tokens`; only the **pre-flight estimate** changes. The status-line **format** is unchanged.
- **FR-013**: The round MUST NOT regress rounds 001–010; prior acceptance scenarios MUST stay green except where this round's request-content contract is deliberately strengthened.

#### Non-Functional Requirements

- **NFR-004**: The round MUST introduce **no new dependency** (`go.mod` / `go.sum` unchanged).
- **NFR-005**: All new assertions MUST be **deterministic** — no `time.Sleep`; the offline fake provider is the verification surface.

### Key Entities *(include if feature involves data)*

- **Persona message**: the leading `system`-role message carrying the configured `PERSON` value, sent as the first element of the request `messages` array. Request-only — never printed, never persisted as an answer.
- **Wire payload**: the input the provider tokenizes for a request — the persona message, the tool declarations, and the conversation messages (prior turns + current prompt).
- **Pre-flight estimate**: the offline, deterministic heuristic count over the wire payload, used by the pre-flight payload-status line; **not** required to equal the provider's measured count.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of prompt runs with a non-empty `PERSON` send a leading `system` message whose content equals the configured `PERSON` value.
- **SC-002**: In the acceptance set, 0 prompt runs with an empty `PERSON` send a `system` message.
- **SC-003**: The pre-flight estimate reflects the persona + tool declarations + messages: a run with a strictly larger wired payload (longer persona, or more declared tools) reports a strictly larger estimate, and identical inputs yield an identical estimate.
- **SC-004**: `stdout` is byte-identical to pre-011 behaviour, and all round-001–010 acceptance scenarios remain green (except where the request-content contract is deliberately strengthened).
- **SC-005**: Both the persona message and the estimate-inputs behaviour are carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.
- **SC-006**: No new dependency is introduced (`go.mod` / `go.sum` unchanged).

## Assumptions

- **Persona role (A1)**: the persona is a **separate leading `system` message** — not a `developer` message (tellme ships only the OpenAI-compatible family and performs no OpenAI-reasoner classification), and not merged into the user prompt.
- **Empty persona (A2)**: when `PERSON` is empty/absent, no system message is sent.
- **Every request (A3)**: the persona is carried on every provider request made during the run, including tool-driven completions — matching the reference's client-level injection.
- **Blast radius (A4)**: only the outbound request **content** and the **pre-flight estimate** change; the payload-status line format, the measured line's source (`usage.prompt_tokens`), and `stdout` byte-exactness are unchanged.
- **Heuristic is research (A5)**: the exact estimation heuristic (chars/token ratio, per-tool allowance, any base term) is a `/axb-technical-research` determination, not a spec-level contract.
- **No numeric equality**: the measured count comes from the provider's own tokenizer over the exact request bytes; it is not required to equal tellme's heuristic estimate.
- **Execution model preserved**: the single-turn model (round 004), stdin piping (round 005), rendered/raw output (round 006), durable history (round 007), the agent tool loop (round 008), the payload status line (round 009), and the cross-stream ordering contract (round 010) are unchanged. The round adds **no** streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, or TUI.
- **Stdlib-only**: no persistence, API, or provider-transport widening beyond the request content.
