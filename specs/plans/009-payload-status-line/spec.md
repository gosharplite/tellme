# Feature Specification: tellme Payload Status Line (round 009)

**Feature Branch**: `009-payload-status-line`

**Created**: 2026-09-13

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "009 — payload status line + token accounting". This round gives tellme **payload-budget visibility**: a prompt-bearing run reports, per turn, the size of the payload it is about to send (a pre-flight **estimate**) and the size the provider actually reported (**actual**), each measured against a configurable **payload budget** — reproducing the reference's `[HH:MM:SS] Payload: <tokens>/<max> tokens - <mode> - <model>` status line. It is **observe-only**: it introduces **no** token-budget pruning or automatic summarisation (pruning stays a settled exclusion).

This round also introduces tellme's first **token accounting**: a deterministic, stdlib-only token estimator, and the provider-usage field the actual figure is read from (currently discarded). No new dependency; the session-history record is unchanged (tokens are recomputed, never persisted).

**Clarify Round 1 (2026-09-13)** resolved three high-impact decisions:

- **Q1 -> Option 1 (stderr)**: the status line is written to the **diagnostic stream (`stderr`)**, so `stdout` stays the byte-exact answer stream (piping/`-r` unaffected); it matches the reference's interactive path and tellme's existing round-008 tool-loop diagnostics.
- **Q2 -> Option 1 (estimate + actual)**: the line shows **both** the pre-flight **estimate** (`~<est>/<max>`) and the post-turn **actual** (`<actual>/<max>`). The actual figure requires widening the provider gateway `Response` to carry the provider's reported usage and parsing it in the OpenAI-compatible adapter (currently discarded).
- **Q3 -> Option 1 (always-on)**: the line is emitted on **every** prompt-bearing run — **not** gated by whether the stream is a terminal and **not** suppressed by `-r/--raw`. (This keeps the open PR #16 Obs 1 stdout-TTY probe untouched.)

**User-locked default (2026-09-13)**: the payload budget `MAX_HISTORY_TOKENS` defaults to **1000000**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See the estimated payload before each turn (Priority: P1)

As an operator, I want to see, before tellme sends a request, how large the payload is going to be relative to my budget, so that I can judge how close the conversation is to its limit instead of being surprised later.

**Why this priority**: This is the core value of the round — budget visibility — and it is the smallest independently verifiable increment (it needs no provider usage, only a deterministic estimator).

**Independent verification**: with a known history and a given budget, run a prompt and confirm a single status line appears on **`stderr`** **before** the answer, showing the estimated payload tokens against the budget and the active mode/model; confirm `stdout` is unchanged versus a run without the line.

**Acceptance Scenarios**:

1. **Given** a session with some prior history and a configured budget, **When** the operator runs a prompt, **Then** tellme writes **one** status line to `stderr` **before** the provider request, of the form `[HH:MM:SS] Payload: ~<estimate>/<budget> tokens - <mode> - <model>`.
2. **Given** any prompt-bearing run, **When** it produces the status line, **Then** the bytes written to `stdout` (the answer) are identical to a run in which no status line is produced.

**Functional Requirements**:

- **FR-001**: On a prompt-bearing run, the system MUST write exactly **one** pre-flight status line to the **diagnostic stream (`stderr`)** before the provider request: `[HH:MM:SS] Payload: ~<estimate>/<budget> tokens - <mode> - <model>`.
- **FR-002**: `<estimate>` MUST be a **deterministic, offline** estimate of the assembled payload's size (the resumed conversation plus the current prompt), produced by a stdlib-only estimator (no provider call, no network).
- **FR-003**: `<budget>` MUST be the resolved payload budget (Story 3; default **1000000**).
- **FR-004**: `<mode>` MUST be the effective mode and `<model>` the active provider's configured **model** (its `MODEL` attribute, matching the reference's status line) — **not** the registry key.
- **FR-005**: The pre-flight line MUST NOT perturb `stdout`: the answer stream stays byte-exact (piping and `-r` are unaffected).

**Non-Functional Requirements**:

- **NFR-001**: The estimator MUST be deterministic and offline — the same assembled payload yields the same estimate.

---

### User Story 2 - See the provider's actual payload size after the turn (Priority: P2)

As an operator, I want to see the payload size the provider actually reported for the turn, so that I can trust the measurement and calibrate how far my estimate drifts from reality.

**Why this priority**: It completes the reference-parity picture (the `~` estimate vs the measured actual), and it depends on Story 1's line existing; it is a larger change because it needs the provider's usage data (currently discarded).

**Independent verification**: run a prompt against a provider that reports usage; confirm a second status line appears on `stderr` **after** the answer with the actual figure; confirm a provider that reports no usage omits the line.

**Acceptance Scenarios**:

1. **Given** a provider response that reports usage, **When** the turn completes, **Then** tellme writes **one** status line to `stderr` of the form `[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model>`, where `<actual>` is the provider's reported prompt-token count.
2. **Given** a provider response that carries no usage data, **When** the turn completes, **Then** tellme writes **no** post-turn status line (the Story 1 pre-flight line is unaffected).

**Functional Requirements**:

- **FR-006**: After a turn completes, the system MUST write exactly **one** post-turn status line to `stderr`: `[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model>`.
- **FR-007**: `<actual>` MUST come from the **provider's reported usage** for that turn. The provider gateway response MUST be widened to carry the reported prompt-token count, and the OpenAI-compatible adapter MUST parse it from the response's usage block.
- **FR-008**: If the provider response carries no usage data, the system MUST omit the post-turn line.
- **FR-009**: The status line MUST NOT be gated by whether the stream is a terminal, and MUST NOT be suppressed by `-r/--raw`.

**Non-Functional Requirements**:

- **NFR-002**: `<actual>` MUST reflect the provider's report for the turn (not an estimate).

---

### User Story 3 - Configure the payload budget (Priority: P3)

As an operator, I want to set the payload budget that the status line measures against, so that the denominator reflects my chosen limit rather than a hard-coded constant.

**Why this priority**: It decouples the displayed budget from a constant; Stories 1–2 work with the default before it lands.

**Independent verification**: set the budget via environment and via configuration, run a prompt, and confirm the denominator of both status lines reflects the chosen value; set a negative value and confirm a configuration error.

**Acceptance Scenarios**:

1. **Given** the budget is overridden by the environment, **When** the operator runs a prompt, **Then** the status line's `<budget>` shows the overridden value.
2. **Given** the budget is invalid (negative), **When** the operator runs, **Then** tellme reports a configuration error and exits with the existing configuration-error code/phrase.

**Functional Requirements**:

- **FR-010**: The system MUST resolve the payload budget from `MAX_HISTORY_TOKENS` (configuration) with an environment override (`MAX_HISTORY_TOKENS`), defaulting to **1000000**, mirroring the existing env-over-file precedence used for the tool-loop bound.
- **FR-011**: A negative `MAX_HISTORY_TOKENS` MUST be a **configuration error** reported with the existing general configuration-invalid class phrase (no new class phrase).

**Non-Functional Requirements**:

- **NFR-003**: Budget resolution MUST be pure and unit-testable (in-memory), like the existing effective-value resolvers.

---

### Edge Cases

- When the provider response carries **no usage data**, the post-turn line MUST be omitted (Story 2, scenario 2). The pre-flight line still appears.
- When the conversation is **empty** (first turn), the pre-flight estimate MUST cover the current prompt alone.
- When the prompt arrives via **piped stdin**, behaviour MUST be identical to a positional-argument prompt.
- When **`-r/--raw`** is set, the status line MUST still be emitted on `stderr`; the answer on `stdout` stays raw.
- When a non-prompt path runs (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`), **no** status line MUST be emitted.
- When the budget is **`0`**, resolution semantics MUST mirror the tool-loop bound resolver (a non-positive value falls back to the default) — a `/axb-technical-research` determination; if it diverges, it is recorded as a disclosed default, not a spec change.
- The status line MUST NOT be a frozen class phrase (no `tellme: ` prefix); the closed class-phrase vocabulary is **unchanged**.

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-012**: The status line(s) MUST be emitted on **every prompt-bearing run** (including `--new` prompt turns) and MUST NOT be emitted by any non-prompt path.
- **FR-013**: The system MUST NOT regress the flag surface or the acceptance behaviour of rounds 001–008; in particular, `stdout` MUST remain byte-exact under piping and under `-r` (the status line is diagnostic, on `stderr`).
- **FR-014**: The status line MUST NOT use the reserved `tellme: ` class-phrase prefix; the frozen class-phrase vocabulary is unchanged.

#### Non-Functional Requirements

- **NFR-004**: The round MUST remain **stdlib-first**: the estimator introduces **no** new dependency (a heuristic estimate, e.g. bytes/4, is sufficient; the exact constant is a research determination).
- **NFR-005**: The status-line format MUST be stable and machine-checkable, deterministic modulo the timestamp; the timestamp MUST come from an injectable clock seam so tests are deterministic.

### Key Entities *(include if feature involves data)*

- **Payload status**: the per-turn status record rendered into the line — estimated tokens, actual tokens (when reported), budget, mode, model, timestamp. A **display-only** value; it is not persisted.
- **Payload budget**: the resolved `MAX_HISTORY_TOKENS` value (default 1000000) that the status line measures against.
- **Provider usage**: the prompt-token count reported by the provider on a response (the source of the actual figure); widened onto the gateway `Response`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of prompt-bearing runs emit exactly **one** pre-flight status line on `stderr` before the request, and — when the provider reports usage — exactly one post-turn line after it.
- **SC-002**: In the acceptance set, 100% of prompt-bearing runs produce `stdout` **byte-identical** to the pre-009 behaviour (piping and `-r` unaffected).
- **SC-003**: For a fixed assembled payload, the pre-flight estimate is **deterministic** across runs.
- **SC-004**: When the provider reports usage, the post-turn `<actual>` equals the provider's reported prompt-token count; when it does not, no post-turn line is emitted.
- **SC-005**: The default budget is **1000000**, is overridable by env and by configuration, and a negative value is reported as a configuration error.
- **SC-006**: 100% of non-prompt paths (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`) emit no status line, and all round-001–008 acceptance scenarios remain green.

## Assumptions

- **Stream (Q1)**: the status line goes to `stderr`; `stdout` stays the byte-exact answer stream.
- **Figures (Q2)**: both a pre-flight **estimate** and a post-turn **actual** are shown; the actual requires widening the gateway `Response` and parsing the provider's usage block.
- **Gating (Q3)**: always-on — not TTY-gated, not suppressed by `-r`. (PR #16 Obs 1 — the stdout-TTY probe — is untouched.)
- **Budget default**: `MAX_HISTORY_TOKENS` = **1000000** (user-locked); env-over-file precedence; negative is a configuration error.
- **Persistence**: **NOOP** — per-turn token counts are computed for display and recomputed from history text on resume; `history_entry` is **not** widened.
- **Observe-only**: the budget is **displayed**, not enforced — no token-budget pruning or automatic summarisation is introduced (a settled exclusion).
- **Estimator**: a stdlib-only heuristic (e.g. bytes/4); the exact constant is a `/axb-technical-research` determination. No new dependency.
- **Metrics line out of scope**: the reference's `M:/H:/C:` and `$cost` metrics line is **not** reproduced (tellme has no pricing table).
- **Line format**: reference-parity `[HH:MM:SS] Payload: <tokens>/<max> tokens - <mode> - <model>`, with `~` prefixing the pre-flight estimate only; the timestamp comes from an injectable clock seam.
- The round keeps the single-turn execution model (round 004), stdin piping (round 005), rendered/raw output (round 006), durable history (round 007), and the agent tool loop (round 008). It does **not** add streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, or a TUI.
