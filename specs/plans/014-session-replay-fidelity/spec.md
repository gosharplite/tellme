# Feature Specification: tellme Session-Replay Fidelity — Persist the Tool-Call Signature (round 014)

**Feature Branch**: `014-session-replay-fidelity`

**Created**: 2026-09-14

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "tellme should resume a tool-using Vertex/Gemini session faithfully" — anchor issue [#34](https://github.com/gosharplite/tellme/issues/34). Today `tellme` drives Vertex Gemini end-to-end (round 013), but a **resumed** session whose earlier turns used tools replays the assistant `functionCall` parts **without** their `thoughtSignature`:

> `Function call is missing a thought_signature in functionCall parts. This is required for tools to work correctly…` → `the provider request failed` (exit 6)

`AgentLoop` records each executed step as `history.Step{Tool, Arguments, Result}` — the signature is not stored — and on resume `BuildMessages` re-synthesises the tool call with **no signature**. Round 013 fixed the **in-turn** loop (the signature rides through because it copies `resp.ToolCalls`); what remains is the **cross-process** case: replaying the tool steps read back from disk on resume. This round **MODIFIES** the persisted history model (a **data-model** change owned by `/axb-data-plan`), the domain `Step`, the loop's record/replay path, and the store round-trip. Behaviour intent: **MODIFY** the persisted tool-step record and its replay; **ADD** the resume-with-tools acceptance.

The reference is round-013 `research.md` **Decision 10** + its residual, and the PR [#33](https://github.com/gosharplite/tellme/pull/33) re-certification forward item (`#5653360657`).

**Clarify Round 1 (2026-09-14)** resolved two high-impact decisions:

- **Q1 → Option 1 (shape)**: a dedicated nullable **`signature`** field on the persisted tool step — a provider-agnostic name; only Vertex/Gemini populates it, the OpenAI-compatible family leaves it empty. No generic `metadata`/`extensions` container this round.
- **Q2 → Option 1 (legacy behaviour)**: when a resumed history's tool step has **no** stored signature, tellme keeps today's **unchanged best-effort** replay — the provider's rejection stays the existing `the provider request failed` (exit 6); **no** pre-flight detection is added.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Resuming a tool-using Gemini session replays its tool steps faithfully (Priority: P1)

As an operator, I want a session that used a tool on a Vertex Gemini provider — after I quit and later resume it — to reply to my next prompt by faithfully replaying the earlier tool step **with** its signature, so the resumed conversation completes instead of failing with `the provider request failed`.

**Why this priority**: it is the round's whole reason to exist — the cross-process resume case that still fails today; every other requirement only guards it.

**Independent verification**: with the in-process fake provider, run a prompt on a `gemini` provider that requests a tool, let the turn complete, then start a **fresh process** with the same history and re-prompt; confirm the replayed request carries the historical `functionCall` **with** its signature (asserted from the fake's recorded request) and the turn completes (exit 0).

**Acceptance Scenarios**:

1. **Given** a `gemini` provider whose turn uses a tool, **When** the turn completes, **Then** tellme persists that tool step **including** the provider signature it carried.
2. **Given** a history that holds a completed turn with a tool step carrying a signature, **When** the operator resumes and re-prompts, **Then** the request tellme sends replays the historical `functionCall` **with** the persisted signature (verbatim), and the turn completes with exit 0.
3. **Given** the same resumed session on the Gemini provider, **When** the replayed request is built, **Then** it does **not** trigger the provider's "missing thought_signature" rejection (no `the provider request failed`).
4. **Given** a turn whose provider supplied **no** signature for a tool step (e.g. the OpenAI-compatible family), **When** the turn completes and is later resumed, **Then** the step is persisted and replayed with an empty signature and the conversation is unaffected.

**Functional Requirements**:

- **FR-001**: When a completed turn that used tools is persisted, the system MUST also persist, per tool step, the provider signature the tool call carried (empty when the provider supplied none).
- **FR-002**: When a persisted conversation is replayed on resume, each replayed tool call MUST carry its persisted signature verbatim.
- **FR-003**: With the signature carried, a resumed session that previously used a tool on a provider that requires a signature (Vertex Gemini 3) MUST complete a re-prompt without the provider rejecting the replayed tool call.
- **FR-004**: A tool step for which the provider supplied no signature MUST be persisted and replayed with an empty signature and behave exactly as today (no new failure path).

**Non-Functional Requirements**:

- **NFR-001**: The signature MUST round-trip through the existing append-only JSON-Lines session store without changing the storage mechanism — one line per completed turn, the turn's ordered steps embedded as today.

---

### User Story 2 - Existing stored history is preserved (Priority: P2)

As an operator, I want the round to be **additive** to my stored history — a turn that used no tools, or a tool step that carried no signature, must serialize exactly as it does today — so upgrading `tellme` never rewrites or disturbs the `history.jsonl` I already have.

**Why this priority**: it protects the operator's existing data and keeps the change trustworthy; it depends on Story 1 (there is a new field only because of Story 1).

**Independent verification**: capture the bytes of a `history.jsonl` written for a non-tool turn and for a turn whose step had no signature before the change, and confirm they are byte-identical after the change; confirm `-l N` and resume still read them unchanged.

**Acceptance Scenarios**:

1. **Given** a completed turn that used no tools, **When** it is persisted, **Then** its `history.jsonl` line is byte-identical to the pre-round shape (no new field appears).
2. **Given** a completed turn whose tool step carried no signature, **When** it is persisted, **Then** the step's serialized shape is byte-identical to today (the new field is omitted when empty).
3. **Given** a `history.jsonl` written before this round, **When** the operator resumes or inspects it (`-l N`), **Then** it is read unchanged and the behaviour is unaffected.

**Functional Requirements**:

- **FR-005**: The change MUST be additive — a tool step whose signature is empty MUST serialize exactly as today (the new field omitted when empty), so every existing `history.jsonl` line stays byte-identical.
- **FR-006**: A turn that used no tools MUST remain byte-identical (no new field appears on a tool-less turn).
- **FR-007**: The persisted history model MUST stay provider-neutral — the field is populated only by providers that supply a signature (Vertex/Gemini); the OpenAI-compatible family leaves it empty and its behaviour is unchanged.

**Non-Functional Requirements**:

- **NFR-002**: The round MUST be stdlib-only — no new module (`go.mod`/`go.sum` unchanged).

---

### Edge Cases

- A `history.jsonl` written **before** this round (tool steps with no signature) resumed on a provider that requires one (Gemini 3): tellme keeps the **unchanged best-effort** replay — the provider rejects the replayed call and the turn fails with the existing `the provider request failed` (exit 6); **no** pre-flight detection (Q2 = Option 1).
- A turn with **multiple** tool steps: each step MUST carry its own signature and be replayed independently.
- A turn whose model requested a tool but the provider gave no signature for it: persisted/replayed with an empty signature; unaffected.
- The signature MUST NOT be surfaced by `-l N` (which still shows only prompt and answer).
- Resuming with a non-tool turn, or with the OpenAI-compatible family, MUST be unaffected.
- The signature is stored and replayed verbatim (opaque); tellme MUST NOT interpret, transform, or expire it.

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-008**: The round MUST NOT regress rounds 001–013 — every prior acceptance scenario MUST remain green, and the frozen class-phrase vocabulary MUST be unchanged (`the provider request failed` = exit 6; `the provider configuration is invalid` = exit 3). No new class phrase is introduced.

#### Non-Functional Requirements

- **NFR-003**: All new assertions MUST be deterministic (no `time.Sleep`); the fake provider's recorded request (the replayed conversation) is the verification surface for both layers.

### Key Entities *(include if feature involves data)*

- **Tool step (persisted)**: one tool execution performed during a completed turn, embedded in the turn's line as an element of its ordered steps array — `{tool, arguments, result, signature?}`. The signature is the provider's opaque token for that tool call; the field is absent when empty (round-007 determinism).
- **Provider signature**: the provider-specific opaque token (Gemini 3 `thoughtSignature`) that must be echoed verbatim on a replayed tool call. Empty for providers that do not use one (the OpenAI-compatible family). Provider-neutral in name and treatment; only Vertex/Gemini populates it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A resumed session that used a tool on a Gemini provider replays the historical tool step **with** its signature and completes — 0 provider "missing thought_signature" rejections (fake-recorded), exit 0.
- **SC-002**: 100% of completed turns persist each tool step's signature, and resume replays it verbatim.
- **SC-003**: `history.jsonl` lines for tool-less turns and for empty-signature steps are byte-identical to round 013.
- **SC-004**: The OpenAI-compatible family's request shape is unchanged, and every prior acceptance scenario remains green.
- **SC-005**: No new module is introduced (`go.mod`/`go.sum` unchanged).
- **SC-006**: The resume-with-tools behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (shape)**: a dedicated nullable **`signature`** field on the persisted tool step — a provider-agnostic name (Q1 = Option 1). No generic `metadata`/`extensions` container this round; a forward-looking container is a later refinement if another provider needs one.
- **A2 (legacy behaviour)**: an absent signature on resume keeps the **unchanged best-effort** replay; a required-but-missing signature surfaces as the provider's own rejection → `the provider request failed` (exit 6). No pre-flight detection (Q2 = Option 1).
- **A3 (provider-agnosticism)**: the persisted history model stays provider-neutral; the field is populated only by providers that supply a signature, and only Vertex/Gemini does so today.
- **A4 (determinism)**: the field is omitted when empty so existing `history.jsonl` lines stay byte-identical (the round-007 append-only determinism).
- **A5 (out of scope)**: the other 014 candidates listed in issue #34 are **not** in this round — the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials, and concurrent tool-call matching.
- **A6 (unchanged)**: rounds 001–013 semantics are unchanged except the persisted signature — single-turn execution, stdin piping, rendered/raw output, durable history, the agent tool loop, the payload status line, cross-stream ordering, persona-on-the-wire, the interactive multi-line prompt, and the Vertex/Gemini transport. No streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, or `SafePath`/consent.
