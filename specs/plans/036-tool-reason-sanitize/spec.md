# Feature Specification: tool reason sanitize — the model-authored reason must not break its own log row (round 036)

**Feature Branch**: `036-tool-reason-sanitize`

**Created**: 2026-09-17

**Status**: Draft — bug fix. Operator-locked decisions Q1–Q2 (see below). Anchor issue [#74](https://github.com/gosharplite/tellme/issues/74).

**Input**: Operator request: *"Fix bug issue https://github.com/gosharplite/tellme/issues/74."* Issue #74 reports that `internal/ui/toolcall.go`'s **`FormatToolReason`** is a bare `fmt.Sprintf("[%s] [Tool Reason] %s", formatClock(t), reason)` — **no `oneLine` fold, no `TrimSpace`, no rune cap** — while its two siblings sanitize **and** cap. The `reason` is the **only model-authored free-text field** in the tool log (`toolReason` returns it raw), so it is the one field a provider can make multi-line: a `\n` **splits** the `[Tool Reason]` row across lines and a `\r` **truncates** it (dropping the `[HH:MM:SS] [Tool Reason] ` prefix). This is a **pre-existing `dev`/`main` defect introduced by round 034** (PR #71 `806cede`) — round 034 replaced round 022's folding `FormatToolLog` (which folded+trimmed, pinned by `TestFormatToolLogFoldsNewlines`) with `FormatToolReason`, **demoting round-022 B1's guarantee without recording the divergence**.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → Option 1 (unit-pin-only witness).** The witness is a **unit pin with hostile fixtures** (`\n`, `\r`, whitespace-only, over-cap) in `internal/ui/toolcall_test.go`. No new E2E Example and **no new `DSLRow`** are minted. Precedents: the round-022 B1 guarantee was witnessed by a **unit pin only**, and round 035 **deliberately declined to mint a row for a display-only defect** (its SC-002 permanent narrowing). The unit pin deterministically covers all four inputs, whereas either E2E count formulation silently misses one (see below). Truth effect = a `chat/dsl.md` **reason-row prose MODIFY** + the cap constant landed; a **doc-level** truth change that adds no feature row, so no topology-audit trigger.
- **Q2 → Option 1 (suppress a blank reason).** A `reason` that is empty or whitespace-only (after folding + trimming) renders **no** `[Tool Reason]` line on either surface. The suppression is done by the **two callers** testing the **sanitized** value before calling the formatter; `FormatToolReason` stays **pure** (no empty-string sentinel). Rationale: round-022 B1's `TrimSpace` was documented *precisely* so "a whitespace-only reason takes the no-tail branch"; today it renders a dangling `[HH:MM:SS] [Tool Reason]    ` row (a trailing-whitespace artifact).

**Scope note**: a **presentation-contract** defect (`truth-current`), **not** a spec gap. The one-line promise exists (the sibling result row already says "on a single line"); this round makes the reason formatter honour it. This round changes **no** CLI flag, exit code, frozen class-phrase vocabulary, tool schema/result, provider transport, persisted record, or `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A model-authored reason never breaks its own log row (Priority: P1)

As a developer reading the tool log, I want every `[Tool Reason]` line to occupy exactly one line regardless of what the model wrote — so a `\n`/`\r`/over-long reason reads as a coherent single-line log rather than a split, truncated, or runaway row.

**Why this priority**: it is the operator's reported defect (issue #74) and the round's core substance; it restores a guarantee round 034 dropped.

**Independent verification**: unit-pin `FormatToolReason` with hostile fixtures — a `reason` containing `\n`, a `reason` containing `\r`, a `reason` of 201 runes — and assert the rendered value is a single line, carries no `\n`/`\r`, and (for the over-cap case) is exactly `reasonValueCap` runes ending in one U+2026.

**Acceptance Scenarios**:

1. **Given** a reason whose text contains a newline, **When** the reason line is rendered, **Then** the rendered value contains **no** `\n`/`\r` — the newline is folded to a space and the row stays on one line.
2. **Given** a reason whose text contains a carriage return, **When** the reason line is rendered, **Then** the rendered value contains **no** `\r` and retains the full `[HH:MM:SS] [Tool Reason] ` prefix followed by the whole reason (nothing is truncated by the `\r`).
3. **Given** a reason longer than the cap, **When** the reason line is rendered, **Then** the reason text is **at most `reasonValueCap` runes** and ends in exactly one U+2026, cut on a rune boundary (never splitting a multi-byte rune).

**Functional Requirements**:

- **FR-001**: `FormatToolReason` MUST fold embedded `\n` and `\r` to spaces and trim surrounding whitespace (the `oneLine` + `TrimSpace` discipline), so the reason can never occupy more than one line.
- **FR-002**: `FormatToolReason` MUST cap the folded reason at a **named rune constant** (`reasonValueCap`) with **one** U+2026 counted inside the cap and a cut on a **rune boundary** — exactly mirroring the sibling formatters (`argValueCap`/`resultValueCap`; ADR 0005 D5).
- **FR-003**: The fold + cap MUST live **inside the pure `internal/ui` formatter** (the round-022 B1 discipline). Both production callers — `agentloop.logAction` (the action line) and `cli/call_renderer.go` `OnCallEnd` (the grouped per-call tail) — MUST NOT each re-implement it; the single-site fix repairs both surfaces.
- **FR-004**: A **well-formed** reason (single-line, in-cap, non-blank) MUST render byte-identically to today, and the **sibling** formatters (`FormatToolAction` cap 189, `FormatToolResult` cap 200) MUST be unchanged.

**Non-Functional Requirements**:

- **NFR-001**: Witnessed by a **hostile-fixture unit pin** — the current `TestFormatToolReason` uses a clean single-line reason (`"checking the launch code"`) where a fold and a cap are both **no-ops**; the restored guarantee MUST be pinned with fixtures that actually exercise it (`\n`, `\r`, over-cap), per the round-022 B1 lesson (the original regression was *"untested — the fixtures used single-line reasons"*).
- **NFR-002**: **stdlib-only** — no new module dependency (`go.mod` / `go.sum` unchanged).

---

### User Story 2 - A blank reason renders no reason row (Priority: P2)

As a developer reading the tool log, I do not want a dangling `[Tool Reason]` line when the model supplied only whitespace, so the log has no empty-looking artifact row.

**Why this priority**: it is the second half of the same defect (the whitespace remnant); it removes a spurious line but does not change the one-line contract itself.

**Independent verification**: unit-pin the emission path with a whitespace-only reason (`"   "`) and a newline-only reason (`"\n"`) and assert **no** `[Tool Reason]` line is emitted on either surface.

**Acceptance Scenarios**:

1. **Given** a tool call whose reason is `"   "` (whitespace-only), **When** the call's log line is written, **Then** **no** `[HH:MM:SS] [Tool Reason]` line appears (neither the action line nor the grouped tail).
2. **Given** a tool call whose reason is `"\n"` (newline-only), **When** the call's log line is written, **Then** **no** `[Tool Reason]` line and **no** spurious blank line appear.

**Functional Requirements**:

- **FR-005**: A `reason` that is empty or whitespace-only **after** folding + trimming MUST emit **no** `[Tool Reason]` line on **either** surface (the loop action line and the per-call tail).
- **FR-006**: The suppression MUST be evaluated on the **sanitized** value at the **two call sites**; `FormatToolReason` MUST remain a **pure** formatter (it MUST NOT be relied on to return an empty-string sentinel, because the tail caller emits its return via `Fprintln` and would still print a bare newline).

---

### Edge cases

- **A reason containing both `\n` and `\r`** → folded to spaces, single line.
- **A reason with leading/trailing whitespace around real text** → trimmed; the text is preserved.
- **A reason longer than the cap whose cap boundary falls mid-multi-byte-rune** → cut on a rune boundary; exactly one U+2026 inside the cap.
- **A reason absent from the arguments** → no reason line (today's behaviour; unchanged).
- **A whitespace-only reason on the tail surface** → no reason line and no blank line (the `Fprintln` sentinel wrinkle is avoided by the caller-side trim check).
- **A tool-less turn** → unchanged (no tool log at all).

### Key entities

- **`[Tool Reason]` line** — the `[HH:MM:SS] [Tool Reason] <reason>` diagnostic row (round 034), emitted per call at its begin (`logAction`) and re-emitted grouped in the per-call tail (`callRenderer.OnCallEnd`).
- **`FormatToolReason`** — the pure `internal/ui` formatter that renders the row; the single site the fix touches for fold + cap.
- **`reasonValueCap`** — the named rune cap constant the formatter applies (recorded as an assumption below; ratified in `/axb-technical-research`).
- **Emission sites** — `agentloop.logAction` and `cli/call_renderer.go OnCallEnd`, the two callers that add the blank-reason suppression check.

## Requirements *(mandatory)*

### Global requirements

- **FR-007**: The round MUST NOT change tellme's CLI flags, exit codes, frozen class-phrase vocabulary, tool surface/schemas/results, provider transport, persisted `history.jsonl` shape, or `stdout` bytes; the **only** affected surface is the `stderr` tool-call log.
- **FR-008**: The offline paths (`--tool-usage`, `--version`, `-d`) MUST remain unchanged.

### Out of scope (recorded)

- **Not** round 035's spinner-yield invariant. Sanitizing the reason keeps a *row* intact; the round-035 phase-boundary yield keeps *the row the spinner shares* clear. They are independent and MUST NOT be conflated.
- The `internal/cli` composition-root refactor and the `BindToolOutput`/`LoopObserver` debt ([#69](https://github.com/gosharplite/tellme/issues/69)).
- Any change to the sibling caps (189/200) or to `FormatToolResult`'s "fold but do not trim-and-suppress" behaviour.

## Success criteria *(mandatory)*

- **SC-001**: With hostile fixtures (`\n`, `\r`, over-cap), `FormatToolReason` yields a **single line** with no `\n`/`\r`, and the over-cap reason is **exactly** `reasonValueCap` runes ending in one U+2026 (US1).
- **SC-002**: A whitespace-only and a newline-only reason emit **no** reason line on either surface — no dangling prefix, no blank line (US2).
- **SC-003**: A well-formed reason renders **byte-identically** to today; `stdout` is byte-exact; the sibling caps and formats are unchanged (FR-004/FR-007).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit (re-run if the reason-row note is landed) are green.
- **SC-005**: The fix has a **falsifiability witness**: removing the fold fails the `\n`/`\r` pins; removing the cap fails the over-cap pin; removing the trim-check re-emits the whitespace-only row; each reproduced then reverted.

## Assumptions

- **Cap value (demoted from clarify — a single numeric threshold).** `reasonValueCap` = **200**, mirroring the sibling `resultValueCap`. Ratified in `/axb-technical-research` and landed in the `chat/dsl.md` reason-row note (which already enumerates the siblings' `189`/`200` caps).
- **`oneLine`'s home (demoted — structure).** Optional tidy: move `oneLine` from `internal/ui/toollog.go` into `internal/ui/toolcall.go` (its sole/third consumer); if deliberately left, `toollog.go` must not be misread as a formatter module. A research/implementation detail.
- **Record the round-034 divergence.** The round-034 removal of round-022 B1's guarantee is currently **unrecorded**; it MUST be recorded as a divergence in **ADR 0005** (`docs/decisions/0005-tool-call-log-parity.md`). This is a project-discipline default, not an open choice.
- **Single-site fold.** The fold + cap go inside `FormatToolReason`; only the blank-reason suppression (Q2) touches the two callers.
- **Hostile fixtures are mandatory.** The unit pin MUST use `\n`, `\r`, whitespace-only, and over-cap fixtures (NFR-001), not the existing clean fixture.
- **Truth ownership.** `chat/dsl.md` is a `DSL` `TruthArtifact` owned by `/axb-dsl-refine` (`truth-single-owner`): that skill owns the reason-row prose edit + the cap-constant note, recorded in this round's `truth-delta.md`. `/axb-api-plan` is **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted state).
- **`/axb-spec-by-example` scope — CONFIRMED NOOP (2026-09-17).** This is an implementation-vs-truth fix: the one-line promise is a rendering *contract* and the violated guarantee is a code behaviour, not a new PM journey — no new plan-side acceptance feature is written. The round-034 acceptance carrier already states the tool-call log's shape (`specs/plans/034-tool-call-log-parity/features/acceptance/showing-the-tool-calls.feature`: "Long values are shortened predictably"), and the reason's one-line guarantee is carried at the **contract** level (the `chat/dsl.md` reason row), which `/axb-dsl-refine` reconciles. Writing a *new* plan-side acceptance rule here would create an `AcceptanceFeature` rule with **no** `InterfaceFeature` carrier (Q1 declined a new interface row/Example) — violating the `acceptance-coverage` invariant (every acceptance rule is carried by ≥1 `InterfaceFeature`). Hence NOOP, consistent with the round-035 display-only precedent.
- **Loop-tier pin update.** `internal/agent/agentloop_reason_test.go` asserts the exact literal `"[12:34:56] [Tool Reason] because"`; if the fold/trim change touches that fixture, it MUST be updated in the same change.
