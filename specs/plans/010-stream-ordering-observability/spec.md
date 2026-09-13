# Feature Specification: tellme Stream-Ordering Observability (round 010)

**Feature Branch**: `010-stream-ordering-observability`

**Created**: 2026-09-13

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "010 — stream-ordering observability" (anchor issue [#28](https://github.com/gosharplite/tellme/issues/28)). This round makes **cross-stream output ordering** — the interleave of the diagnostic stream (`stderr`) with the answer stream (`stdout`) — a **first-class, checkable contract**. It responds to round 009, which shipped an ordering defect (the post-turn payload line printed *before* the answer) that **every gate in the pipeline was structurally unable to see**: no layer could *represent* "a `stderr` line trails the `stdout` answer". The round (i) pins the **required** orderings, (ii) adds the **missing oracle** — a **merged-stream witness** in the E2E harness — and (iii) asserts the orderings at **both** the unit and E2E layers. It changes **no** observable output *content* and adds **no** dependency: it is a contract + verification round.

This round's behaviour intent is **MODIFY** — it strengthens existing round-008/009 behaviour (the payload-status line and the live tool-loop log already exist) into a **required, witnessed contract**, with no user-visible content change. It is not a re-litigation of round 009: the payload-status ordering was already corrected in `7bcb2d3`; round 010 makes the ordering an explicit, falsifiable guarantee.

**Clarify Round 1 (2026-09-13)** resolved three high-impact decisions:

- **Q1 → Option 2 (payload-status + tool-loop ordering required)**: the **required** cross-stream orderings are (a) the payload-status lines bracketing the answer (`pre-flight < answer < measured`), and (b) the live tool-loop log lines preceding the answer. Both are tellme-owned deliberate `stderr` diagnostics. The round-006 **degraded-render warning** is ruled **incidental** — documented as not guaranteed and **not** asserted (avoids over-asserting a fallback path).
- **Q2 → Option 1 (merged `2>&1` capture)**: the E2E harness observes ordering through a **single merged buffer** (`stderr` redirected into the same capture as `stdout`). The current separate `stdout`/`stderr` capture (`harness.RunResult{Stdout, Stderr}`) is structurally unable to express an interleave — that is the missing oracle. A merged buffer needs no product change and is deterministic under tellme's sequential writes.
- **Q3 → Option 1 (assert at both layers)**: each required ordering is asserted at the **unit** layer (the turn's emit order — extending `7bcb2d3`'s `TestRunTurn_PostTurnStatusFollowsAnswer`) **and** the **E2E** merged-stream layer (the new oracle).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The payload status brackets the answer in order (Priority: P1)

As an operator, I want the per-turn payload status to appear in the order that matches what it describes — the pre-flight estimate *before* the answer and the measured size *after* it — so the terminal narrative is coherent and I can trust which figure is the estimate and which is the measurement.

**Why this priority**: This is the exact bug class round 009 shipped and that no gate could observe; it is the smallest independently verifiable increment (one observable interleave) and the reason the round exists.

**Independent verification**: run a prompt against a provider that reports usage, capture the two streams **merged**, and confirm the pre-flight line appears before the answer bytes and the measured line after them.

**Acceptance Scenarios**:

1. **Given** a prompt run in which the provider reports its usage, **When** the run completes, **Then** in the merged output the pre-flight estimated payload line appears **before** the answer and the measured payload line appears **after** the answer.
2. **Given** a prompt run in which the provider reports no usage, **When** the run completes, **Then** in the merged output the pre-flight estimated payload line appears **before** the answer and **no** measured payload line appears.

**Functional Requirements**:

- **FR-001**: On a prompt-bearing run, the system MUST write the pre-flight estimated payload status line to `stderr` **before** it writes the answer to `stdout`.
- **FR-002**: When the provider reports usage for the turn, the system MUST write the measured payload status line to `stderr` **after** it writes the answer to `stdout`.
- **FR-003**: In the observable interleave of the two streams, the order MUST be `pre-flight < answer < measured`.
- **FR-004**: The ordering MUST NOT be achieved by relocating any diagnostic content onto `stdout`; the answer stream MUST remain byte-exact.

**Non-Functional Requirements**:

- **NFR-001**: The payload-status ordering MUST be deterministic and verifiable offline, against the in-process fake provider.

---

### User Story 2 - Tool-loop activity is visible before the answer (Priority: P2)

As an operator, I want the live tool-loop log lines to appear before the final answer, so I can follow the work that produced the answer in the order it happened instead of seeing a report that claims to describe work that has not yet "happened" on screen.

**Why this priority**: It is the same `stderr`-diagnostic ordering class as Story 1 and is required to close the bug class completely; it depends on Story 1's contract and witness being established.

**Independent verification**: run a prompt in which the model requests a tool, capture the two streams merged, and confirm the tool-loop log line naming the tool appears before the answer.

**Acceptance Scenarios**:

1. **Given** a prompt run in which the model requests a tool, **When** the run completes, **Then** in the merged output the tool-loop log line naming the tool appears **before** the answer.

**Functional Requirements**:

- **FR-005**: When a turn emits tool-loop diagnostics, the system MUST write each tool-loop log line to `stderr` **before** it writes the answer to `stdout`.
- **FR-006**: The tool-loop diagnostics MUST remain on `stderr`; the answer stream MUST remain unchanged.

**Non-Functional Requirements**:

- **NFR-002**: The tool-loop ordering MUST be deterministic and verifiable offline, against the in-process fake provider.

---

### Edge Cases

- When the provider reports **no usage**, the measured line MUST be omitted; the pre-flight line still precedes the answer (Story 1, scenario 2).
- When the prompt arrives via **piped stdin**, the required ordering MUST be identical to a positional-argument prompt.
- When **`-r/--raw`** is set, the payload-status lines MUST still be written to `stderr` in the required order, and `stdout` MUST stay raw and byte-exact.
- When a **non-prompt path** runs (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`), **no** payload-status line and **no** tool-loop line is emitted, so the ordering contract is vacuous there.
- When a turn ends **without a final answer** (the tool-loop bound is reached, or a tool request fails), the tool-loop lines MUST still precede any terminal outcome; the "measured after the answer" requirement is vacuous (no answer) and MUST NOT fabricate a measured line.
- The round-006 **degraded-render warning** relative to the answer is **incidental** — it MUST be documented as not guaranteed and MUST NOT be asserted (FR-011).

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-007**: The E2E suite MUST be able to observe the **interleaved** order of `stdout` and `stderr` (a merged-stream capture, e.g. both streams into one buffer), so that an inter-stream ordering regression fails the suite. The current separate `stdout`/`stderr` capture does not satisfy this requirement.
- **FR-008**: Each required ordering (FR-001–FR-006) MUST be asserted at **both** the unit layer (the turn's emit order) and the E2E layer (the merged-stream interleave).
- **FR-009**: Each required ordering MUST be **mechanically assertable** — expressible as an executed acceptance/interface rule and asserted by an automated test — not merely described in prose. (This closes the round-009 leak, where the ordering intent existed only as acceptance prose and was dropped from the executable interface truth.)
- **FR-010**: The round MUST NOT regress rounds 001–009; `stdout` MUST remain byte-exact (under piping and `-r`) and all prior acceptance scenarios MUST stay green.
- **FR-011**: Interleaves **not** declared required — specifically the degraded-render warning relative to the answer — MUST be treated as **incidental**: documented as not guaranteed and not asserted.

#### Non-Functional Requirements

- **NFR-003**: The round MUST remain **stdlib-first**: the merged-stream capture and any harness change introduce **no** new dependency.
- **NFR-004**: The E2E ordering assertion MUST be deterministic — no `time.Sleep`; a sequential-writer program yields a stable merged order under a single buffer.
- **NFR-005**: The interleave witness MUST be **hermetic** — independent of the developer shell / ambient environment, consistent with the round-009 E2E hermeticity fix.

### Key Entities *(include if feature involves data)*

- **Ordering contract**: the set of **required** cross-stream interleaves — the payload-status bracketing (`pre-flight < answer < measured`) and the tool-loop log preceding the answer. A guarantee, not persisted state.
- **Merged-stream witness**: the harness capture that records `stdout` and `stderr` into one **ordered** buffer, so the interleave is observable end-to-end. A test-only mechanism; no product change.
- **Incidental interleave**: an interleave tellme does **not** guarantee (the degraded-render warning) and MUST NOT assert.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of prompt-bearing runs with a usage-reporting provider show the order `pre-flight < answer < measured` in the merged output.
- **SC-002**: In the acceptance set, 100% of tool-requesting runs show the tool-loop log line before the answer in the merged output.
- **SC-003**: A deliberately inverted ordering (e.g. a payload-status line emitted before the answer) fails the suite at **both** the unit and E2E layers — the guarantee is falsifiable, not merely asserted.
- **SC-004**: `stdout` is byte-identical to the pre-010 behaviour and all round-001–009 acceptance scenarios remain green.
- **SC-005**: Every required ordering is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.
- **SC-006**: No new dependency is introduced (`go.mod` / `go.sum` unchanged).

## Assumptions

- **Observable interleave**: ordering is observed in the **merged** (`2>&1`) view; tellme writes sequentially within a turn, so the merged order is deterministic.
- **Required set (Q1)**: payload-status bracketing **and** tool-loop-precedes-answer are required; the degraded-render warning is incidental.
- **Witness (Q2)**: a merged single-buffer capture is the E2E oracle; no product surface change.
- **Layering (Q3)**: the ordering is asserted at both the unit and E2E layers.
- **Behaviour already correct**: after `7bcb2d3` the payload-status order is correct; the round adds the contract and the oracle, not a behaviour re-fix.
- **`stdout` byte-exactness** (rounds 005/006) is preserved; ordering is achieved by `stderr` write order only.
- **Stdlib-only** witness; no new dependency.
- **No persistence / API / provider-transport change** — the session model, the provider gateway `Response`, and any contract surface are unchanged.
- The round keeps the single-turn execution model (round 004), stdin piping (round 005), rendered/raw output (round 006), durable history (round 007), the agent tool loop (round 008), and the payload status line (round 009). It adds **no** streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, or TUI.
