# Feature Specification: tellme turn counter counts AI-endpoint calls (round 027)

**Feature Branch**: `027-ai-call-turn-counter`

**Created**: 2026-09-15

**Status**: Draft — semantics **operator-locked** (session 2026-09-15): tellme's turn chrome stays **as-is** (one `╭─⠿ Turn N - <mode>` header + one `╰─⠿ Ready` per prompt-bearing turn, plain text, no denominator), but the header's `N` changes **meaning** to the session's **running AI-endpoint-call index** — one more than the number of provider calls the session's prior turns made — so it matches `tell-me-go`'s counter.

**Input**: Operator request, verbatim: *"I want tellme as-is, but the counter show total calls to AI endpoint same as tell-me-go."* — i.e. keep tellme's presentation and cadence; change only what the header's `N` counts.

**Scope note**: a **presentation-accounting** round. It changes (a) how the round-017 turn-header number is **computed** (completed user turns → completed **AI-endpoint calls** + 1) and (b) the **persisted state** needed to reconstruct that count across invocations. It does **not** change tellme's chrome **cadence or format** (still exactly one header + one `╰─⠿ Ready` per prompt; the reference's per-call header/footer cadence is **not** adopted), the prompt path's `stdout` (byte-exact), the frozen class-phrase vocabulary, the pre-flight/post-turn payload lines, the round-018 post-turn metrics/`Ready` lines, the round-019 spinner, the round-022 tool-loop log line, or the tool set/semantics (round 024/021).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The turn header number counts AI-endpoint calls (Priority: P1)

As an operator, I want the `╭─⠿ Turn N - <mode>` header's `N` to count the session's **AI-endpoint calls** — not the number of user prompts — so the number advances by however many times the model was actually called and reads the same as `tell-me-go`.

**Why this priority**: it is the entire slice. Today `N` counts completed user turns (`len(prior)+1`), so a tool-using turn (which calls the model several times) advances it by only one — diverging from the reference. Fixing the counted unit is the deliverable.

**Independent verification**: run a turn hermetically on a fresh session and assert `Turn 1`; then resume a session whose prior turns, in total, made `C` provider calls and assert the next header is `Turn <C+1>`; and assert that a turn which runs a tool round (≥2 calls) advances the counter by its call count, while a tool-less turn advances it by one. In every case the chrome **cadence** is unchanged (exactly one header per prompt) and `stdout` is byte-exact.

**Acceptance Scenarios**:

1. **Given** a fresh session, **When** the operator runs a plain (tool-less) prompt, **Then** the header reads `Turn 1`.
2. **Given** a session whose prior turns made `C` total provider calls, **When** the operator runs the next prompt, **Then** the header reads `Turn <C+1>`.
3. **Given** a prompt that runs one tool round (two provider calls), **When** the operator then runs the following prompt, **Then** that prompt's header reads `Turn <C+3>` — the counter advanced by the two calls, not one.
4. **Given** any prompt-bearing turn, **When** it runs, **Then** exactly **one** `╭─⠿ Turn …` header and exactly **one** `╰─⠿ Ready` are emitted (no per-call header or footer), and the header format is unchanged (`╭─⠿ Turn <N> - <mode>`, plain text, no denominator).

**Functional Requirements (FR)**:

- **FR-001**: The turn header `╭─⠿ Turn <N> - <mode>` MUST display `<N>` = **(the number of AI-endpoint calls made by the session's prior completed turns) + 1** — i.e. the running index of this prompt's first call.
- **FR-002**: An **AI-endpoint call** MUST be defined as **one inference round** of the agent loop — one provider completion request (`AgentLoop.Run`'s `Gateway.Complete` invocation). A tool-less turn makes **1** call; a turn that runs `k` tool-execution rounds makes **1 + k** calls. **Provider-internal retries** (backoff and empty-response retries, which stay within a single inference round) MUST **NOT** add to the count — matching `tell-me-go`, whose header fires only on the first attempt of a generation cycle.
- **FR-003**: The chrome **cadence and format** MUST be unchanged: exactly **one** header and **one** `╰─⠿ Ready` per prompt-bearing turn; plain text (no ANSI); no `/MAX_HISTORY_TURNS` denominator; the ` - <mode>` suffix; emitted on the diagnostic stream (`stderr`) inside the existing frame. The reference's per-call header/footer cadence is explicitly **not** adopted.
- **FR-004**: The displayed `<N>` MUST equal the number `tell-me-go` prints at that prompt's **first** call (parity target). This is a **value** requirement, not a cadence requirement: tellme shows only the first call's number per prompt; `tell-me-go` additionally shows the intermediate calls' numbers.

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The header value MUST be computable from persisted session state alone (no re-query of providers, no dependence on transient in-process state) — see US2.

---

### User Story 2 - The count is durable across invocations and session-scoped (Priority: P2)

As an operator, I want the AI-call count to survive across separate `tellme` invocations (each prompt is its own process) and to reset on `--new`, so the number stays correct on a resumed session and a fresh session correctly restarts at `Turn 1`.

**Why this priority**: US1 is only correct if the count can be reconstructed when a new process loads an existing session — without persistence the counter would reset every invocation. It is the durability half of the counter.

**Independent verification**: run two prompts as separate processes on one session and assert the second header continued the count; then run `--new` and assert the next header is `Turn 1`; then assert a resumed session (a later process reading the same history) continues the count.

**Acceptance Scenarios**:

1. **Given** a session created by an earlier process, **When** a new process runs the next prompt, **Then** the header continues the count from the persisted history (it does not restart at 1).
2. **Given** a session with prior turns, **When** the operator runs `tellme --new "<prompt>"`, **Then** the header reads `Turn 1` (the count resets with the archived session).
3. **Given** the persisted history, **When** the header number is computed, **Then** it is derived from the session's stored per-turn call counts (or an equivalent stored running total), independent of the provider.

**Functional Requirements (FR)**:

- **FR-005**: The session's AI-call total MUST be **persisted with the session history** so that a **separate invocation** (a new process resuming the session) reconstructs the same count. The persisted data MUST carry, per turn, the number of AI-endpoint calls that turn made (or an equivalent stored running total).
- **FR-006**: `--new` MUST reset the count to 0, so the first turn of a fresh session reads `Turn 1` — matching `tell-me-go` (whose session turn count restarts with the archived history).

---

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-007**: The round MUST NOT change any of: the prompt path's `stdout` (byte-exact); the frozen class-phrase vocabulary (`specs/truth/features/cli/dsl.md`); the pre-flight / measured payload lines (rounds 009/018); the post-turn metrics and `╰─⠿ Ready` lines (round 018); the spinner (rounds 019/025); the tool-loop log line (round 022); or the tool set and tool execution semantics (rounds 024/021).

#### Non-Functional Requirements

- **NFR-002**: The round MUST add **no new third-party module** (Go standard library only).
- **NFR-003**: Verification MUST be **hermetic** — injected streams and clock, the scripted fake provider, and a temp runtime home; no pty (the round-017/018 precedent).
- **NFR-004**: The round MUST be **POSIX-only** (tellme's standing scope).

### Edge cases

- A turn that **fails before completing** (a provider/transport error, or the tool-loop bound reached) persists **nothing**, so the counter is **unchanged** and the next prompt repeats the same number — matching `tell-me-go`, whose failed `Turn` does not advance `SessionTurns`.
- A prompt-less invocation, a `--version` / `-d` / `-l` / `--tool-usage` run, and an empty/cancelled interactive submission emit **no** turn header and do not affect the count.
- A **resumed** session (a later process) continues the count from persisted state.
- A turn whose header is emitted (showing `<prior calls + 1>`) and which then **fails** still does not advance the persisted counter — the shown number is the attempted first call's index.

### Key entities

- **Turn counter (`N`)** — the number displayed in `╭─⠿ Turn <N> - <mode>`: the session's running AI-endpoint-call index (prior calls + 1).
- **AI-endpoint call** — one inference round of the agent loop (one provider completion request); `1` for a tool-less turn, `1 + <tool rounds>` otherwise; provider-internal retries do not count.
- **Persisted turn record** — the per-turn session record persisted in `history.jsonl`; this round requires it to carry the turn's AI-call count (or an equivalent stored running total) so the counter is reconstructable across invocations.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: On a fresh session, the first prompt-bearing turn's header reads `Turn 1` (covers FR-001, FR-006).
- **SC-002**: After prior turns that made `C` total calls, the next header reads `Turn <C+1>`; a turn that runs a tool round advances the count by its call count (≥2), while a tool-less turn advances it by one (covers FR-001, FR-002, FR-005).
- **SC-003**: The chrome cadence is unchanged — exactly one header and one `Ready` per prompt; `stdout` is byte-exact; the class-phrase vocabulary and the round-018/019/022 surfaces are unchanged (covers FR-003, FR-007).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with a falsifiability witness for the call-based count (a tool-using turn must advance the counter by its call count, and the number must match the reference's value at the prompt's first call).

## Assumptions

- **"AI-endpoint call" = an inference round**, not an HTTP attempt — retries inside a single call do not count (operator-locked; matches the reference's first-attempt-only header). This is the unit the counter advances by.
- The header is emitted **before** the turn runs, so it can only show this prompt's **first** call index (`prior calls + 1`) — which is exactly the value `tell-me-go` prints at that prompt's first call.
- The per-turn call count is carried on the **persisted turn record** (the mechanism is an RD/`/axb-data-plan` decision; the requirement is only that the count survives across invocations).
- tellme's display stays literally "as-is": the word `Turn`, the ` - <mode>` suffix, and the plain-text/no-denominator format are unchanged; only the counted unit changes.
- This round **modifies in place** the round-017 turn-chrome truth (`specs/truth/features/cli/chat/presenting-the-turn.feature`, the `chat/dsl.md` row `the turn is headed "Turn {number}"…`, `specs/truth/techstack.md`) and **adds** a persisted per-turn call count to the data model; the actual truth changes are made by the truth-owner skills. No frozen plan package is touched (`fresh-package-per-round`).
