# Feature Specification: tellme spinner width safety — a bounded tool label and a residue-free clear (round 025)

**Feature Branch**: `025-spinner-width-safety`

**Created**: 2026-09-15

**Status**: Draft — scope and mechanism **operator-locked** (session 2026-09-15): **Q1 → 1** (fix **both** defects) · **Q2 → 1** (make the clear width-safe by **tracking the last frame's occupied rows and erasing them all**)

**Input**: Bug report [issue #55](https://github.com/gosharplite/tellme/issues/55), verbatim: *"The turn spinner's **tool-phase label enumerates every tool name** in the batch … With round 024's larger tool set (`execute_command`) and multi-tool batches, that line exceeds the terminal width … 1. **The trailing resource segment is clipped** … 2. **It can break the round-019 teardown contract** … residue can survive the turn."* Operator decision (this session): fix **both** defects; make the clear width-safe by **tracking the last frame's rendered-row count and erasing every row**.

**Scope note**: a **defect-fix** round on the round-019 spinner surface. It bounds the several-tool status label and makes the spinner clear width-safe. It does **not** change the model-phase label, the single-tool label, the no-names label, the resource segment, the `stderr`-terminal gate, the phase lifecycle, `stdout`, or the frozen class-phrase vocabulary.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The several-tool label stays bounded (Priority: P1)

As an operator watching a multi-tool turn, I want the spinner's tool-phase label to name the first tool and count the rest, so the line stays short no matter how many tools the model batches — the trailing `[CPU: … | MEM: …]` segment stays readable and the frame stays (in practice) a single row.

**Why this priority**: it is the headline defect — round 024's larger tool set and multi-tool batches make the enumerated label over-wide, clipping the resource segment on every such turn.

**Independent verification**: run a turn whose fake provider batches several tools (e.g. four) and assert the tool-phase spinner line's status is the bounded form `Executing tools [<first> and <N-1> more]...` — not the full enumeration; run a single-tool turn and assert the single-tool form is unchanged.

**Acceptance Scenarios**:

1. **Given** a turn that runs several tools at once, **When** the tool-phase spinner is drawn, **Then** its status reads ` Executing tools [<first> and <N-1> more]...` — bounded, independent of the batch size.
2. **Given** a turn that runs exactly one tool, **When** the tool-phase spinner is drawn, **Then** its status reads ` Executing [<name>]...` (unchanged).

**Functional Requirements (FR)**:

- **FR-001**: For a several-tool phase (**two or more** names), the tool-phase status label MUST be **bounded** — ` Executing tools [<first> and <N-1> more]...`, where `<first>` is the first name and `<N-1>` the count of the remaining names — and MUST NOT enumerate every name.
- **FR-002**: For a single-tool phase (**one** name), the label MUST remain ` Executing [<name>]...` (**unchanged**).
- **FR-003**: For a tool phase with **no** names, the label MUST remain ` Executing tools...` (**unchanged**).

### User Story 2 - The spinner leaves no residue on a wrapping terminal (Priority: P2)

As an operator on a terminal narrower than a spinner frame, I want the spinner's clear to erase **every** row the previous frame occupied, so no spinner residue survives into the answer even when a frame soft-wraps.

**Why this priority**: it repairs the round-019 teardown contract (`the progress spinner no longer appears once the answer is written`), which a **single-line** clear cannot honour once a frame soft-wraps across two or more rows.

**Independent verification**: with an injected narrow terminal width, render an over-wide frame and then clear; assert the emitted clear sequence accounts for **every** row the frame occupied (a **unit** pin — a real terminal grid cannot be reproduced hermetically); the round-019 merged-capture teardown assertion is retained.

**Acceptance Scenarios**:

1. **Given** a frame wider than the terminal width (so it soft-wraps across several rows), **When** the spinner clears it, **Then** every row the frame occupied is erased — no residue survives the turn.

**Functional Requirements (FR)**:

- **FR-004**: Before drawing a new frame, the spinner MUST erase **every terminal row the previous frame occupied** (not only the last row), so an over-wide (soft-wrapped) frame leaves no residue.
- **FR-005**: On every clear — the yield-to-interleaved-output, the phase end, a turn failure, and the final teardown — the spinner MUST likewise erase every row of the last frame, preserving the round-019 contract `the progress spinner no longer appears once the answer is written`.

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The occupied-row count MUST be derived from the frame's rendered width and the terminal width; when the terminal width is unknown, the spinner MUST degrade safely (a single-row best-effort clear — no crash). *Known bound (accepted):* a mid-frame terminal resize leaves the row count stale, so the clear may over-erase one row of prior output; the residue defect is eliminated and over-erase-under-reflow is a recorded limitation (the reference has no resize handling at all).

---

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-006**: The model-phase label (` Thinking [<model>]...`), the resource segment (` [CPU: … | MEM: …]`), the `stderr`-terminal gate, the phase lifecycle (stop/resume), and the turn-scoped elapsed counter MUST be **unchanged**.
- **FR-007**: The round MUST NOT change `stdout` (byte-exact), MUST NOT introduce a new frozen class phrase (the vocabulary stays unchanged), and MUST keep the spinner on `stderr` only.

#### Non-Functional Requirements

- **NFR-002**: The round MUST add **no new third-party module** — a terminal-width source already present in the module graph (or an injected seam) is used.
- **NFR-003**: The behaviour MUST be POSIX-only; verification is hermetic (an injected width seam + clock + streams; no pty).

### Edge cases

- **Two** tools → ` Executing tools [<first> and 1 more]...`.
- **One** tool → ` Executing [<name>]...` (unchanged).
- **Zero** names → ` Executing tools...` (unchanged).
- A frame exactly the terminal width → one row; one character over → two rows.
- Terminal width unknown → single-row best effort (no crash, no over-erase).
- The `-i` submit surface (round 023) draws the same spinner and **inherits** the fix.

### Key entities

- **Tool-phase status label** — the bounded several-tool label (FR-001) and the unchanged single-tool / no-names labels (FR-002/FR-003).
- **Rendered-row count** — the number of terminal rows the last frame occupied (derived from the frame width and the terminal width), tracked across redraws and consumed by the clear (FR-004/FR-005).

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: A several-tool turn's spinner status reads ` Executing tools [<first> and <N-1> more]...`; a single-tool turn's reads ` Executing [<name>]...` (covers FR-001–FR-003).
- **SC-002**: Given a narrow terminal width, clearing an over-wide frame erases every row it occupied — no residue (covers FR-004–FR-005; **unit-pinned**).
- **SC-003**: `stdout` is byte-exact, the class-phrase vocabulary is unchanged, and the model-phase label / resource segment / gate are unchanged (covers FR-006–FR-007).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with falsifiability witnesses for the bounded label and the row-aware clear.

## Assumptions

- The bounded several-tool form is the operator's chosen shape (issue #55): ` Executing tools [<first> and <N-1> more]...`.
- The width-safe clear is a **recorded divergence from the reference** — `tell-me-go` likewise enumerates every tool name and clears a **single** row (`internal/domain/events/types.go`, `internal/ui/renderer_spinner.go`); tellme deliberately bounds the label and hardens the clear where the reference does not (see `research.md`).
- The `-i` submit surface inherits the fix (the same spinner draws there).
- The row-aware clear needs a **terminal-width source**; the concrete seam (injected for tests; `golang.org/x/term` is already in the module graph) is a `/axb-technical-research` decision. The E2E carries the **label** bound; the width-safe **clear** is pinned at the **unit** layer (a flat byte capture cannot reproduce a terminal grid).
- This round **modifies in place** the round-019 truth (`specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md`); the round-019 plan package stays frozen (`fresh-package-per-round`).
