# Feature Specification: every `[Tool Output]` line grey (round 058)

**Feature Branch**: `058-grey-tool-output-content`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify round 1 OPEN** (one boundary question: the trailing partial line). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)** — a follow-up to round 057:

> *"I want all the `[Tool Output]` lines grey, not just the first line and the two horizontal lines."*

**Behaviour intent**: **MODIFY (CLI-visible diagnostic chrome, one element).** Round 057 greyed the `[Tool Output]` **header** line and the **two** horizontal separators, and deliberately left the streamed **content** lines plain (assumption A3). This round extends the grey to **every** line of the block — header, **each content line**, and both separators. It reverses round 057's A3 only; everything else round 057 delivered (the yellow `[Tool Action]`, the 500-rune cap, the payload increment) is untouched.

---

## Grounded in the current system *(measured this session, `dev` @ `50ace10`)*

| Site | Current shape |
| --- | --- |
| `internal/ui/tooloutput.go` → `ToolOutputWriter.Colour` | The round-057 flag. `Begin` writes `formatToolOutputHeaderColour(now, w.Colour)` (header) + `toolOutputSeparatorColour(w.Colour)` (opening separator); `EndWith` writes the reset + `toolOutputSeparatorColour(w.Colour)` (closing separator). **The content path does NOT consult `w.Colour`**: `WriteWith` writes `FormatToolOutputLine(now, line)`, whose body is `fmt.Sprintf("[%s] [Tool Output] %s", formatClock(t), sanitizeControl(line))` — **plain**. |
| `internal/ui/colour.go` | `colorGray = "\033[0;90m"`, `colorYellow`, `colorGreen`, `colorReset`; wrappers `grey`/`yellow`/`green` over one `wrap(s, code, enabled)` (empty string never wrapped; disabled ⇒ verbatim). |
| `internal/ui/coordinator.go` | The coordinator builds the writer with `Colour: colour` (round 057). The per-line clear hook (`WriteWith(p, beforeLine)`) runs **before** the line is written. |
| `tests/e2e/steps/step_r057_payload_and_colour.go` | `thenToolOutputGrey` asserts the grey **header** + ≥2 grey **separators**; and (round 057's own pinned fact) it asserts **the content line stays plain**: `if strings.Contains(out, gray+"[...] [Tool Output] hello")` — that assertion must be **inverted** by this round. |
| `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` | The round-057 Rule *"The tool output frame is grey and the action line is yellow at a terminal"*, whose comment records A3 (*"the streamed content lines … stay plain"*). |
| `specs/truth/techstack.md` | The **Agent command tool** row records the round-057 grey scope ("the header line and both horizontal separators") and the round-038 sanitize policy for the content. |

**Consequence (current behaviour):** on a terminal the block reads as a grey frame with **plain** content inside it — the operator wants the whole block grey.

---

## Settled design *(operator-stated this session)*

| # | Decision |
| --- | --- |
| **S-1** | On a colour-enabled stream (`stderr` is a terminal **and** `-r` is off), **every** `[Tool Output]` line is grey: the header, **each streamed content line**, and both separators. |
| **S-2** | The **content is sanitized first** (`sanitizeControl`), so the grey wrap encloses an already-control-free string — no command output can escape the wrapper (the round-038 policy is unchanged; the wrap is *outside* the browser-sanitized text). |
| **S-3** | The colour gate is the round-054 one, reused unchanged; off a terminal / under `-r` the block is **byte-identical** to the plain form. |
| **S-4** | `turns.log`/`-t` stays **plain** (the round-053/054 file-leg rule; the per-call `emit` renders the file leg with `NewLines(false)`). |
| **S-5** | The **neutral-close restore** (`\x1b[0m` before the closing separator, round 038) is unchanged, and the closing separator is still grey after it. |
| **S-6** | Nothing else changes: the yellow `[Tool Action]`, the 500-rune cap, the payload increment, the sanitizer semantics, the writer's locking/idle behaviour, `stdout` byte-exactness. |

---

## Clarify round 1 *(one question; ≤ 3 this round)* — **OPEN**

| # | Question | Status |
| --- | --- | --- |
| **Q1** | **Does "all lines" include the trailing partial line?** A content line is only ever written when it is **complete** (terminated); a trailing partial line is **dropped, never flushed** (round 034 FR-010). It is therefore never printed, so "all lines" most likely means *all printed lines* and the drop behaviour is unchanged. Options: **(A)** the drop is unchanged (nothing to grey — recommended; the partial line is never rendered); **(B)** the trailing partial line is now flushed (and greyed) — a behaviour change to the round-034 FR-010 drop, out of the request's spirit. | ⏳ **OPEN** |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — only the **colour of the content lines** changes; their **text** (sanitize, the `[HH:MM:SS] [Tool Output] ` framing, the drop-a-partial rule) is byte-identical.
- **I-2** — colour never enters `stdout` or `turns.log`; the plain path is byte-identical (the round-054 gate).
- **I-3** — the neutral-close restore and the round-038 sanitizer are unchanged; the wrap is applied to the sanitized value.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the whole output block reads as grey (Priority: P1)

As the **operator**, when a command streams output at a terminal I want **every** line of the `[Tool Output]` block — header, content, and the two rules — to be grey, so the block reads as one visually distinct region instead of a grey frame around plain text.

**Why this priority**: it is the entire request.

**Independent verification**: on a terminal run with `-r` off, the captured `stderr` carries the grey wrap on the header, on **each content line**, and on both separators; with `-r` on / off a terminal, none of those escapes appear.

**Acceptance Scenarios**:

1. **Given** a terminal run with `-r` off, **When** a command prints lines, **Then** the `[Tool Output]` header, **every** streamed content line, and both separators are wrapped grey (`\033[0;90m … \033[0m`).
2. **Given** the same run, **When** a content line is inspected, **Then** its text is exactly today's (sanitized, `[HH:MM:SS] [Tool Output] <line>`) and the grey encloses the whole line.
3. **Given** a command whose output carries a control sequence, **When** the line is rendered, **Then** the sequence is still removed (round 038) and the grey wrap is the **only** escape on the line.
4. **Given** a run whose diagnostic stream is **not** a terminal, **or** `-r` is on, **Then** the block (header, content, separators) is **byte-identical** to the plain form (no escapes).
5. **Given** a terminal run, **When** `turns.log` is inspected, **Then** it carries the **plain** block (no escapes) — the file leg stays plain.

**Functional Requirements**:

- **FR-001**: tellme MUST wrap **each streamed content line** of the `[Tool Output]` block in grey on a colour-enabled stream (S-1).
- **FR-002**: the content line's **text** MUST be unchanged — the same timestamp/prefix framing and the same `sanitizeControl` result (I-1); the grey wraps the sanitized line (S-2).
- **FR-003**: the colour MUST be emitted **only** when the diagnostic stream is a terminal **and** `-r` is off (the round-054 gate, S-3), the plain path byte-identical (I-2).
- **FR-004**: `turns.log`/`-t` MUST stay **plain** (S-4); colour MUST NOT reach `stdout` or `turns.log`.
- **FR-005**: the round-038 neutral-close restore MUST be preserved, and the closing separator MUST remain grey after it (S-5).
- **FR-006**: the header + both separators MUST remain grey (round 057's behaviour is a subset of this round's) — no regression.

---

## Edge Cases

- **A command that prints nothing** — no content lines; only the header + separators (all grey). Unchanged shape.
- **A trailing partial line** (no terminating newline) — **dropped, never flushed** (Q1 → A, the round-034 rule); nothing to grey.
- **A command that times out mid-output** — the drain path writes complete lines (grey); the drop + reset + grey closing separator are unchanged.
- **A non-terminal / `-r` run** — byte-identical plain block.
- **`turns.log`** — plain leg.
- **A very long content line** — the sanitizer/wrap order is unchanged; the grey adds an escape pair around the (already-sanitized) line; no new cap is introduced (the block's content is not rune-capped today).

## Key Entities

- **The `[Tool Output]` block** — the round-034 header + content lines + the two separators.
- **A content line** — one `[HH:MM:SS] [Tool Output] <line>` row (sanitized).
- **The grey accent** — the round-057 `colorGray` (`\033[0;90m`) wrap.

## Success Criteria

- **SC-001**: On a colour-enabled terminal run, the header, **every** content line, and both separators carry the grey wrap; off a terminal / under `-r`, all are byte-identical to the plain form.
- **SC-002**: A content line's text is byte-identical to today's (the round-038 sanitize result) — only the surrounding escapes differ.
- **SC-003**: `turns.log` and `stdout` never contain the grey escape (diff-level check).
- **SC-004**: The round's own gates hold: `gofmt`/`go vet` clean · `make verify` **OK** · `go test -count=1 ./...` green · `go.mod`/`go.sum` unchanged.
- **SC-005**: The changed behaviour has a reproduced falsifiability witness (drop the content wrap ⇒ the new Example/pin reds).

## Assumptions

- **A1**: "All the `[Tool Output]` lines" = the header, every **complete** content line, and both separators (the printed lines); the **trailing partial line** is never printed and stays dropped (Q1 → A unless the operator says otherwise).
- **A2**: The colour codes are the reference's (`colorGray`), the element set tellme's own (round-054/057 convention); the gate is the round-054 one.
- **A3**: Truth impact is expected in `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (+ `chat/dsl.md`) and `specs/truth/techstack.md` (**Agent command tool** row + the Turn-chrome row); `/axb-api-plan` + `/axb-data-plan` are **NOOP**; `/axb-spec-by-example` ✅; `/axb-dsl-refine` ✅ (one Rule text + rows).
- **A4**: An **ADR (0028)** records the block-wide grey (superseding round 057's A3), with a §Forward.
- **A5**: This round **reverses round 057's assumption A3** explicitly (recorded in ADR 0028 and the round-057 §Forward lineage) — the round-057 **package is frozen** and is **not** edited.

## Out of scope (recorded forward items)

- Any **other** colour change (the yellow `[Tool Action]`, the green accents, the payload increment) — round 057's, unchanged.
- Flushing the trailing partial line (Q1 → B) — a round-034 FR-010 behaviour change, not requested.
- A non-terminal colour mode (`--color=always`) — the gate stays terminal-only.
- The locked exclusions: no security/consent layer · no Windows · sequential tool calls ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
