# Feature Specification: tellme Session History Persistence (round 007)

**Feature Branch**: `007-session-history-persistence`

**Created**: 2026-09-12

**Status**: Draft

**Input**: User request "007 — session history persistence". Framed through the round-007 scope discussion; the five **permanent exclusions** (`-b`/`--retry`, pinning, streaming responses, token-budget pruning, `SafePath`) and the **deferred** capability (summarisation — lands later, with the agent-tools round) are treated as settled and are **out of scope**. Clarify Round 1 (2026-09-12) resolved three high-impact decisions:

- **Q1 -> Option 1 (Auto-resume, always)**: a plain `tellme "<prompt>"` run loads the persisted session history from the session workspace and includes prior turns as context before the current prompt — matching `tell-me-go`'s persistent `Session`.
- **Q2 -> Option 1 (Both `--new` and `-l N`)**: the round delivers the session lifecycle surfaces — `--new` (archive the current history, start a fresh session) and `-l N` (print the last N persisted messages) — plus persistence and resume.
- **Q3 -> Option 1 (Reuse the environment class phrase)**: a history read/write failure is reported with the existing environment class phrase `the runtime home is not usable` and the environment error code; the frozen class-phrase vocabulary stays at **ten**.

*(Deferred to `/axb-technical-research` as parity/format details to pin: the history file name and location under the session workspace; the append-only line format; the archive mechanism for `--new`; the default message count for a bare `-l`; and crash/partial-write handling.)*

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A durable session across runs (Priority: P1)

As a developer using tellme from the terminal, I want each completed exchange to be remembered across invocations, so that a follow-up prompt is answered with the earlier conversation as context instead of starting from nothing — matching the reference's persistent session.

**Why this priority**: This is the core value of the round — without durable memory there is no session, and `--new`/`-l` are meaningless. It is the first value to prove.

**Independent verification**: run a first prompt whose answer the provider echoes; run a second prompt in the same session workspace and confirm the provider received the **prior** turn as context (observable through the request the provider records and the answer). Repeat on a **fresh** workspace and confirm the first run carries no prior context.

**Acceptance Scenarios**:

1. **Given** a successful prompt turn, **When** the turn completes, **Then** the exchange is persisted to the session history under the session workspace, and the process exits with the success code.
2. **Given** a session workspace that already holds a persisted history, **When** the operator runs a further prompt, **Then** tellme includes the prior turns as context before the current prompt, and the answer reflects the remembered conversation.
3. **Given** a session workspace with **no** persisted history, **When** the operator runs a prompt, **Then** tellme sends the current prompt with no prior context.

**Functional Requirements**:

- **FR-001**: After a turn completes, the system MUST persist that turn (the user prompt and the provider's answer) to the session history under the session workspace.
- **FR-002**: On a prompt run, the system MUST load the persisted session history and include the prior turns as context before the current prompt (auto-resume).
- **FR-003**: The session history MUST be **append-only** — a completed turn's stored record MUST NOT be rewritten; an interrupted turn MUST NOT be written.
- **FR-004**: When no session history exists yet, the system MUST proceed with no prior context and MUST NOT fail.

### User Story 2 - Start a fresh session (`--new`) (Priority: P2)

As a user who has finished one line of work, I want a `--new` flag that starts a fresh session with no prior context while retaining the old history, so that I can begin a new conversation without losing the previous one.

**Why this priority**: It is the reset surface that keeps a durable session usable (history would otherwise only ever grow); it depends on Story 1's persistence existing.

**Independent verification**: run a prompt to build history; run `--new`; confirm the next prompt carries **no** prior context and that the prior history is retained (not destroyed).

**Acceptance Scenarios**:

1. **Given** a session with persisted history, **When** the operator runs tellme with `--new`, **Then** a fresh session begins with no prior context, the prior history is retained (archived, not destroyed), and the process exits with the success code.
2. **Given** a session with **no** persisted history, **When** the operator runs tellme with `--new`, **Then** a fresh session begins and the run succeeds (no error).

**Functional Requirements**:

- **FR-005**: The system MUST provide a `--new` flag that begins a fresh session with no prior context.
- **FR-006**: `--new` MUST retain (archive) the prior session history rather than destroy it.

### User Story 3 - Inspect recent history (`-l N`) (Priority: P3)

As a user reviewing a session, I want an `-l N` flag that prints the last N exchanges, so that I can see what tellme remembers without sending a new prompt.

**Why this priority**: A read/inspection surface that makes the persisted history verifiable and useful; lower priority than establishing and resetting the session itself.

**Independent verification**: build a history with several turns; run `-l 2`; confirm exactly the last two messages are printed and no provider request is made; run on an empty session and confirm a clean, empty result.

**Acceptance Scenarios**:

1. **Given** a session with persisted history, **When** the operator runs tellme with `-l N`, **Then** tellme prints the last N messages to standard output, performs no provider request, and exits with the success code.
2. **Given** a session with no persisted history, **When** the operator runs tellme with `-l N`, **Then** tellme prints nothing and exits with the success code.

**Functional Requirements**:

- **FR-007**: The system MUST provide an `-l N` flag that prints the last N persisted messages to standard output.
- **FR-008**: An `-l` run MUST NOT contact a provider.
- **FR-009**: When the session holds fewer than N messages, `-l N` MUST print all available messages without error.

### Edge Cases

- When a completed turn is interrupted before the provider answers, the system MUST NOT write a partial history record.
- When the history file exists but cannot be read or written, the system MUST report the environment class phrase `the runtime home is not usable` and exit with the environment error code (FR-011).
- When `--new` is combined with a prompt, the fresh (empty) session MUST be used for that prompt's context.
- When `-l N` is given a non-positive or malformed N, the system MUST reject the usage with the existing usage-error contract and exit code.
- When the history contains entries from a prior, incompatible version, the system MUST fail deterministically (environment class phrase) rather than silently misreading it.

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-010**: The system MUST keep its existing flag surface (`-c/--config`, `-d/--diagnostics`, `--version`, `-r/--raw`) and MUST add **only** `--new` and `-l N` in this round.
- **FR-011**: A session-history read or write failure MUST be reported on standard error with the existing environment class phrase `the runtime home is not usable`, and the process MUST exit with the environment error code.
- **FR-012**: The system MUST NOT regress the existing acceptance behaviour of rounds 001–006, **except** where this round deliberately extends the chat turn to also persist and resume (a recorded truth MODIFY under `/axb-dsl-refine`).

#### Non-Functional Requirements

- **NFR-001**: The new surfaces MUST NOT introduce network access beyond the existing prompt turn; `--version`, `-d`, a prompt-less boot, `--new`, and `-l` MUST remain strictly offline and deterministic.
- **NFR-002**: For a given turn, the bytes the system writes to the session history MUST be deterministic across repeated runs.
- **NFR-003**: History path resolution, append, reload, and the `--new` / `-l` mode selection MUST be covered by fast, isolated unit tests, complementing the black-box end-to-end path.

### Key Entities *(include if feature involves data)*

- **Session history**: the persisted, append-only record of a session's turns, stored under the session workspace — created on first write, read on resume.
- **History entry**: one completed turn — the user prompt and the provider's answer.
- **Session workspace**: the existing per-mode directory under the runtime home (round-001 truth), which now also hosts the session history.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of prompt runs persist the completed turn, and a subsequent run in the same session workspace includes the prior turns as context — while a run on a fresh workspace carries no prior context.
- **SC-002**: In the acceptance set, 100% of `--new` runs begin a session with no prior context and leave the prior history retained.
- **SC-003**: In the acceptance set, 100% of `-l N` runs print the last N messages (or all available) and make no provider request.
- **SC-004**: In the acceptance set, 100% of history read/write failures emit the class phrase `the runtime home is not usable` and exit with the environment error code.
- **SC-005**: All pre-existing round-001–round-006 acceptance scenarios remain green, except where a recorded truth MODIFY updates them for the now-persistent turn.
- **SC-006**: History path resolution, append, reload, and the `--new` / `-l` mode selection are covered by isolated unit tests.

## Assumptions

- The session history is stored as an **append-only** file under the session workspace (the reference's JSON-Lines `history.jsonl`); the exact file name, location, and line format are **technical parity details owned by `/axb-technical-research`**.
- `--new` archives the prior history (moving it aside) rather than deleting it; the exact archive mechanism and location are owned by `/axb-technical-research`.
- A bare `-l` (no count) has a sensible default message count; the value is owned by `/axb-technical-research`.
- A **turn** remains one user prompt + one provider answer (round 004); tool calls are not in scope, so a history entry carries only prompt/answer content.
- Resume is scoped to the **effective mode's** session workspace (round-001 truth); there is no cross-mode session.
- **Settled exclusions (out of scope, not deferred)**: `-b`/`--retry`, history pinning, streaming responses, token-budget pruning, and `SafePath`. **Deferred (not excluded)**: summarisation — it lands later, with the agent-tools round.
- The round keeps the single-turn execution model (round 004), the stdin piping (round 005), and the rendered/raw output (round 006); it does **not** add a TUI (`-i`), thoughts/tools display, or the `--callback` worker.
