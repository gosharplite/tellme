# Feature Specification: spinner tail residue — the per-call tail must not share a row with a live frame (round 035)

**Feature Branch**: `035-spinner-tail-residue`

**Created**: 2026-09-17

**Status**: Draft — bug fix. Operator-locked decisions Q1–Q4 (see below). Anchor issue [#72](https://github.com/gosharplite/tellme/issues/72).

**Input**: Operator request: *"I think we should fix the bug issue https://github.com/gosharplite/tellme/issues/72."* Issue #72 reports that on a **tool-using turn** at a terminal, round 034's per-call **tail** (grouped `[Tool Reason]` lines + measured payload + metrics + `Ready`) is written **without yielding the round-019/025 spinner**, so the spinner's live frame shares a physical row with the tail's first grouped `[Tool Reason]`, and the frame **residue survives the finished turn**.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → (a).** Fix approach: **yield around the tail** — the per-call tail's non-final emission is written with the progress indicator yielded (the mechanism the tool-log lines and the `[Tool Output]` sink already use). `Spinner.OnCallEnd` stays a no-op (ADR 0005 D1 preserved).
- **Q2 → pure bug fix.** Round 035 fixes **only** the spinner/tail collision. The `internal/cli` composition-root refactor (#69, incl. the `BindToolOutput` constructor-injection debt and the `LoopObserver` interface segregation) stays its own future round.
- **Q3 → siblings out.** The two sibling round-034 recorded divergences — the failed-turn display-only `Ready` overstatement (G2) and the per-call numbering skew — remain recorded forward items, **not** in scope.
- **Q4 → E2E + unit.** The witness is an **E2E forced-gate tool-round scenario** (primary; pairs the round-019 `stderr` terminal seam with a scripted tool-calling provider) **plus** a **unit ordering pin** (secondary; the emit/yield order at the loop→renderer→spinner seam).

**Scope note**: an **implementation-vs-truth** defect (`truth-current`), **not** a spec gap. The violated rules already exist as executable truth (round-019/025 spinner residue rows). This round changes **no** CLI flag, exit code, line format, cadence, tool schema, tool result, provider transport, persisted record, or `stdout` byte. It adds the missing spinner **yield** around one diagnostic write and pins it with an executable witness.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A tool-using turn leaves no live-frame residue (Priority: P1)

As a developer watching tellme at a terminal, I want a tool-using turn's finished output to be clean — every grouped `[Tool Reason]` starting on its own line and **no** leftover braille-frame + phase-status (`⠋ Executing …`) anywhere — so the diagnostic stream reads as a coherent log rather than a torn frame glued to the closing status.

**Why this priority**: it is the operator's reported defect and the round's sole substance.

**Independent verification**: run a turn whose scripted provider requests ≥1 tool and then answers, with the diagnostics shown at a terminal (the round-019 `stderr` terminal seam gates the spinner on); read the captured `stderr` as a terminal (the `\r`-redraw simulation) and assert there is **no** braille frame followed by a phase status, and that the number of `[Tool Reason]` **lines** equals the number of occurrences.

**Acceptance Scenarios**:

1. **Given** a tool-using turn on the non-TUI prompt surfaces with the diagnostics shown at a terminal, **When** the turn completes, **Then** the finished diagnostic stream shows **no** progress-spinner residue (no braille frame followed by a phase status such as `Executing …`), and the answer has been written.
2. **Given** the same tool-using turn, **When** the call's grouped post-call tail is written, **Then** every grouped `[Tool Reason]` line begins on its **own** line (a tail write MUST NOT be appended to a live spinner frame row).
3. **Given** the spinner is gated **off** (diagnostics not at a terminal, or the raw flag), **When** the turn runs, **Then** the run is unchanged — no spinner is drawn, the tail renders as today, and the output is byte-identical to a gated-off run before this fix.

**Functional Requirements**:

- **FR-001**: On a **completed** tool-using turn with the spinner gated **on**, the finished diagnostic stream MUST carry **no** spinner residue — read as a terminal (the existing `\r`-redraw negative), it MUST show no braille frame followed by a phase status (`Thinking …` / `Executing …`). The existing rules `the progress spinner no longer appears once the answer is written` and `the run shows no progress spinner` MUST hold for the **tool-tail** path.
- **FR-002**: The per-call tail's writes MUST be yielded to the progress indicator: the indicator MUST be **synchronously cleared** before the tail's first line so the write starts at column 0 of a cleared line, and the tail's lines MUST NOT be appended to a live spinner frame row. As a consequence every grouped `[Tool Reason]` in the tail MUST begin on its own line (N occurrences ⇒ N lines).
- **FR-003**: The yield MUST cover the whole tail emission (the grouped reasons, the measured payload line, the metrics line, and the `Ready` line) — not only its first line.
- **FR-004**: The fix MUST be free of residue at the **whole-turn** level: no write that follows the tail (notably the next AI-endpoint call's status **frame**) may leave the indicator sharing a row. The exact re-activation point of the indicator (whether it resumes immediately after the tail or is re-activated by the next waiting phase) is a research/implementation detail pinned so the acceptance scenarios hold **end to end** for a multi-call turn.
- **FR-005**: The fix MUST be confined to the spinner/tail interaction: `stdout` stays byte-exact; the turn-chrome, payload, metrics, `Ready`, and tool-call **line formats** and the per-call **cadence** are unchanged; the **deferred final-call tail** (already emitted after `Stop()`) is unchanged; and a run with the spinner **gated off** is unchanged.

**Non-Functional Requirements**:

- **NFR-001**: Exercisable **hermetically** — no pty: the round-019 `stderr` terminal seam forces the spinner on, and the fake provider scripts the tool round; the negative reads the captured stream through the existing `\r`-redraw simulation.
- **NFR-002**: **stdlib-only** — no new module dependency (`go.mod` / `go.sum` unchanged).

---

### Edge cases

- **A multi-call tool-using turn** (≥1 tool round then an answer) → the non-final tail(s) and every frame are residue-free; the final (deferred) tail is unchanged.
- **A single-call tool-less turn** → unchanged (the tail is the deferred final tail, emitted after `Stop()`).
- **A bound-reached turn** → the non-final tail is residue-free; the run fails with `the tool request failed` (exit 7) with no frame residue.
- **The spinner gated off** (non-terminal `stderr`, or `-r`) → unchanged; the yield is a no-op.
- **A narrow terminal** (a soft-wrapped frame) → the round-025 row-aware clear still applies; the fix MUST NOT regress the existing narrow-terminal residue rule.

### Key entities

- **Per-call tail** — the grouped `[Tool Reason]` lines + measured payload line + metrics line + `╰─⠿ Ready` emitted at a call's end (round 034); deferred past the answer for the final call.
- **Progress indicator** — the round-019/025 spinner (braille frame + phase status + elapsed + resource), drawn on a row with no newline terminator; cleared synchronously by `Stop()` / `BeforeToolLog()`.

## Requirements *(mandatory)*

### Global requirements

- **FR-006**: The round MUST NOT change tellme's CLI flags, exit codes, frozen class-phrase vocabulary, tool surface/schemas/results, provider transport, persisted `history.jsonl` shape, or `stdout` bytes.
- **FR-007**: The offline paths MUST remain unchanged (no spinner, no tail, no new output).

### Out of scope (recorded)

- The `internal/cli` composition-root refactor and the `BindToolOutput` / `LoopObserver` debt ([#69](https://github.com/gosharplite/tellme/issues/69)).
- The failed-turn display-only `Ready` overstatement (G2) and the per-call numbering skew (round-034 recorded divergences).

## Success criteria *(mandatory)*

- **SC-001**: A tool-using turn at a terminal yields a finished diagnostic stream with **no** spinner residue (no braille frame + phase status) — the existing residue rows hold for the tool-tail path (US1).
- **SC-002**: Every grouped `[Tool Reason]` in a call's tail begins on its own line (occurrences == lines) (US1/FR-002).
- **SC-003**: A spinner-gated-off run is byte-identical to before the fix; `stdout` byte-exact; line formats/cadence unchanged (FR-005).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and the fix has a **falsifiability witness**: reverting the yield fails the new/updated scenario (a `⠋` + `Executing …` survives in the capture); restoring it is green.

## Assumptions

- **The observed defect is the round-034 per-call tail** (`call_renderer.OnCallEnd`) writing without a yield while the loop's `Spinner.OnToolsEnd()` is a no-op (so the indicator is still live). The `[Tool Output]` block already yields; the tool-log lines already yield via `AgentLoop.withToolLog`; the tail is the un-yielded write.
- **The fix is a spinner yield (Q1 a)**: clear before the tail's first line. The precisely correct re-activation point (resume-after-tail vs. re-activate-on-next-phase) is pinned in `/axb-technical-research` from the E2E witness, because a resume that lets the **next** call's frame write append to a live row would relocate the residue (FR-004); the acceptance scenarios are the arbiter.
- **The witness is E2E + unit (Q4)**: the E2E scenario pairs `the diagnostics are shown at a terminal` (the forced `stderr`-TTY seam) with a scripted tool-calling provider; the unit pin covers the emit/yield order at the seam.
- **Truth scope**: the fix makes the **existing** residue rules hold for the tool-tail path; `/axb-dsl-refine` re-anchors/extends the `chat` module's spinner/status carriers (a **MODIFY**) only where a carrier is needed for the gated tool-tail journey, and records it in `truth-delta.md`. `/axb-api-plan` is **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted state changes).
- This is a **CLI-interface** round: `/axb-spec-by-example` is **NOOP** (the violated acceptance rules already exist as executable truth; no new PM journey); `/axb-dsl-refine` is the CLI end contract owner.
