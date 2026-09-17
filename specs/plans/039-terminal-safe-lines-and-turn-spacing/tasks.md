# Tasks — round 039 (terminal-safe `[Tool …]` lines + turn-output blank-line grouping)

Plan package: `specs/plans/039-terminal-safe-lines-and-turn-spacing`

Task markers: `[UNIT]` unit test · `[BDD-RED]` new/aligned stepdef that must fail before the product change · `[BDD-GREEN]`/`[BDD-REFACTOR]` feature green/refactor · `[CODE-REMOVE]` delete dead code · `[REGRESSION]` regression + falsifiability.

## Core Inputs (must read)

- `spec.md` (FR-001..FR-011), `research.md` (D1–D9), `plan.md`, `truth-delta.md`.
- Truth: `specs/truth/techstack.md` (Agent tool loop + `execute_command` rows — MODIFY this round), `specs/truth/features/cli/chat/watching-the-tool-loop.feature` (two new Rules), `specs/truth/features/cli/chat/dsl.md` (`## Given (round 039)` + `## Then (round 039)`).
- Governance: `docs/decisions/0008-terminal-safe-lines-and-blank-line-grouping.md` (ADD, supersedes ADR 0007).
- Code seams: `internal/ui/tooloutput.go` (holds `sanitizeControl` today), `internal/ui/toolcall.go` (the three sibling formatters), `internal/agent/agentloop.go` (`logAction`/`logEngine`/`reasonsOf`), `internal/cli/call_renderer.go` (`OnCallEnd` `emit`).

## Setup

_(none — stdlib-only; no new dependency, no build/SDL change.)_

## Foundational

- [X] **T001** — Move the sanitizer into a dedicated single-owned home: create `internal/ui/sanitize.go` and relocate `sanitizeControl` + its helpers (`isControlRune`, `escSequenceLen`, `csiLen`, `oscLen`, `genericEscLen`, the `csiScanLimit`/`oscScanLimit` constants) there **verbatim**, updating the `[Tool Output]` call site's import only. **只做** a byte-identical move; **不做** any behavior/class change. `[scaffold]`
- [X] **T002** — Add the round-039 E2E landing skeleton as **independent files** (Zero Shared Edits): `tests/e2e/steps/step_r039_control_free.go` (register the 3 control-free `Then`s + the escape-bearing-reason `Given`) and `tests/e2e/steps/step_r039_blank_lines.go` (register the 4 blank-line `Then`s), each with stub bodies returning `nil`. **只做** registration + stubs; **不做** the assertions. `[scaffold]`

## Phase 3 — Test alignment & implementation (Red first)

- [X] **T003** `[UNIT]` — `internal/ui/toolcall_sanitize_test.go`: hostile fixtures for the three sibling formatters (an ESC-bearing `reason`; an ESC-bearing result snippet; an escape-bearing argument value **and** argument key) asserting each rendered line is **control-free and valid UTF-8** with the visible text and the rune caps (`{189, 200, 200}`) preserved. **RED** before T005.
- [X] **T004** `[UNIT]` — (a) a `[Tool Output]` **byte-identical regression pin** (`internal/ui/tooloutput_sanitize_test.go` extend: the round-038 fixtures still produce identical output after the T001 move); (b) an **escape-only-reason** pin (`internal/agent/agentloop_blank_reason_test.go` extend: a `reason` that is only control data emits **no** `[Tool Reason]` row). **RED** before T005/T007.
- [X] **T005** `[UNIT]` — `internal/agent/agentloop_test.go` (extend): the call-begin **blank-line ordering** — a blank line immediately precedes the call block's first line (the `[Tool Reason]` line when the call states a reason; else the `[Tool Action]` line), **per call**. **RED** before T007.
- [X] **T006** `[BDD-RED]` — `tests/e2e/steps/step_r039_control_free.go`: implement the `Given` `…asks tellme to read "{path}" with a reason that carries terminal control data and then answers with "{answer}"` (a fixed escape-bearing `reason`) and the 3 `Then`s (`the run reported the reason/result/action for the tool call "{tool}" free of terminal control sequences`) + their control-free/UTF-8 helpers. **RED** before T005.
- [X] **T007** `[BDD-RED]` — `tests/e2e/steps/step_r039_blank_lines.go`: implement the 4 `Then`s (`each tool call's report begins after a blank line`; `the action of the call without a reason begins after a blank line`; `the trailing reason summary follows a blank line`; `the turn's closing status follows a blank line`) + their blank-adjacency helpers. **RED** before T007 (the blank writes) / already-RED for the reason-less call.
- [X] **T008** — Review gate: after T003–T007, every new step matches exactly one `DSLRow` and the suite reports **0 undefined** steps.

## Phase 4 — Feature green / refactor

- [X] **T009** `[BDD-GREEN]` — `watching-the-tool-loop.feature` Rule *Every tool-loop line is free of terminal control sequences* green: apply `sanitizeControl` inside `FormatToolReason`/`FormatToolResult`/`FormatToolAction` (values **and** keys), ordered **fold+trim → sanitize → rune cap**. Test Scope: the control-free Rule's 3 Examples.
- [X] **T010** `[BDD-REFACTOR]` — tidy the three formatters + `formatToolArgs` under green protection (single sanitize application point; no allocation regression on a control-free line).
- [X] **T011** `[BDD-GREEN]` — `watching-the-tool-loop.feature` Rule *The live turn output is grouped by blank lines* green: (a) `agentloop.logAction` writes a leading blank before the call block's first line; (b) `callRenderer.OnCallEnd` `emit` writes a blank before the non-empty grouped `[Tool Reason]` block and a blank before the post-status group (when `usage.Reported`); (c) widen the blank-reason suppression to read the **sanitized** value (D3) across `logAction`/`reasonsOf`/the defensive `OnCallEnd` guard. Test Scope: the grouping Rule's 3 Examples.
- [X] **T012** `[BDD-REFACTOR]` — tidy the blank-emit seams under green protection (one blank per site; no blank on a non-tool turn).

## Regression & falsifiability

- [X] **T013** `[REGRESSION]` — `make verify` + the full godog suite green; `stdout` byte-exact; every existing tool-log Example unchanged-green (the round-021/034/038 pins). Re-run the Gherkin/DSL topology audit (`--root specs/truth/features/cli`) and confirm PASSED.
- [X] **T014** — Falsifiability witnesses (each reproduced then reverted): (a) revert a sibling's sanitize → T003 + the corresponding control-free E2E Example fail; (b) remove a blank insertion → the corresponding blank-line E2E Example fails; (c) narrow the reason guard back to `TrimSpace(raw)` → the escape-only-reason pin fails.
- [X] **T015** — Update `STATUS.md` + the daily summary; open the PR; close **#80** on delivery.

## Boundary

- In scope: `internal/ui/sanitize.go` (new) + `internal/ui/{tooloutput,toolcall}.go`, `internal/agent/agentloop.go`, `internal/cli/call_renderer.go`, and the E2E step files.
- Out of scope (unchanged): the `[Tool Output]` header/separator literals, the block bound/stop/spinner-yield, `output_file`, the tool **result** fed to the model, the call's stored arguments, `stdout`, `-r`, flags, exit codes, the class-phrase vocabulary, the provider transport, every persisted record, and the offline paths (`spec.md` FR-009/FR-010/FR-011).

## Parallel Hint

- T003/T004 are independent unit files; T006/T007 are independent step files (Zero Shared Edits). T005, T009–T012 touch product files (`internal/ui`, `internal/agent`, `internal/cli`) and must serialize.

## Outcomes (implementation)

- **T001** done — `internal/ui/sanitize.go` created; `sanitizeControl` + helpers moved **verbatim** out of `tooloutput.go` (byte-identical for `[Tool Output]`; the round-038 tests are the regression pin).
- **T002–T008** done — `internal/ui/toolcall_sanitize_test.go` (hostile fixtures for the three siblings + `ToolReasonRenders`); `internal/agent/agentloop_spacing_test.go` (per-call blank ordering + escape-only-reason suppression); `internal/cli/call_renderer_reason_test.go` (tail blanks); `tests/e2e/steps/step_r039_control_free.go` (1 Given + 3 Thens) + `tests/e2e/steps/step_r039_blank_lines.go` (4 Thens). Review gate: every new step matched exactly one `DSLRow`; suite **0 undefined**.
- **T009/T010** done — `FormatToolReason` / `FormatToolResult` / `FormatToolAction` (keys **and** values) sanitize **before** their rune caps.
- **T011/T012** done — a leading blank before each call's begin block (`agentloop.logAction`); a blank before the grouped tail block and before the post-status group (`callRenderer.OnCallEnd` `emit`); the blank-reason guard reads the **sanitized** value (`ui.ToolReasonRenders`) across `logAction` / `reasonsOf` / the tail guard.
- **T013** — `make verify` **OK** (lint 0 · govulncheck 0 reachable · cross-compile 4/4 · offline witness · no test-sleep · MCP-SDK confinement) · `go test -count=1 ./...` green — E2E **225 scenarios (225 passed) · 1672 steps** (was 218); `stdout` byte-exact; topology audit **PASSED** (44 features · 6 modules · 16 root + 323 module rows · 1648 steps).
- **T014** — **falsifiability witnesses** (each reproduced then reverted): (a) remove `FormatToolReason`'s sanitize → `TestFormatToolReason_StripsControlSequences` + the E2E `A reason that carries control data is shown as plain text` fail (the pre-fix line rendered `[Tool Reason] \x1b[31m…`); (b) remove the `logAction` leading blank → the E2E `A call that states no reason still starts a fresh block` fails; (c) narrow the reason guard back to `TrimSpace(raw)` → `TestToolReasonRenders`' escape-only case fails.
- **T015** — `STATUS.md` + daily summary updated; PR opened; #80 closes at merge (closeout Step 8).

## Round-039 review fold (PR #81)

- **B1** [blocker] — gated the post-status blank on a `renderedToolRound` marker set on a **non-final** `OnCallEnd`, so a **tool-less** usage-reporting turn gains no blank (FR-009 / the edge-case list / ADR 0008 D5's closing sentence hold in the **code**, not just prose). Witnesses: `TestCallTailNoBlankBeforePostStatusWithoutToolRound` (unit) + the `presenting-the-post-turn-status` Example `A tool-less turn's closing status is not blank-separated` (E2E). ADR 0008 D5 row 3 now states the tool-round condition.
- **TD-1** [pre-freeze] — corrected the three stale plan-side claims that said *"ADR 0007 is amended in place"* → **superseded by ADR 0008** (`spec.md` Q4 + Assumptions, `checklists/requirements.md`); **governance outcome: ADR 0007 superseded by 0008 (not amended in place) — research D9** (recorded here before the package freezes).
- **TD-2** — corrected the false *"the whole list capped at 189 rendered runes"* clause in `techstack.md:37` to the true per-**value** cap (keys sanitized but uncapped; the joined list uncapped).
- **TD-3** — homed on **#69**: the loop's control flow now depends on a `ui` predicate, and the three-site blank-reason predicate is **extended** (≈3 evaluations/call), not consolidated; plus the environment note (no layer-discipline `verify-architecture` analogue). Non-blocking, recorded on the live issue.
- **RF-1** — single-sourced the reason transform as `toolReasonText` (fold+trim → sanitize → cap); both `FormatToolReason` and `ToolReasonRenders` derive from it.
- **RF-2** — the `[Tool Action]` key path now **folds** (`oneLine`) and the key list is **sorted after sanitizing** (a hostile key cannot render out of ascending order); unit pins added.
- **RF-3** — narrowed the reason-step DSL row to *"the reason line(s) the run emitted"* (the `[Tool Reason]` line carries no tool name), dropping the unused `tool` capture; made the closing-status assertion **total** over all measured payload lines.
- **Observation** — recorded that the round-level `[Tool Engine] Step i/M` marker sits in the **header group** before the per-call blanks (ADR 0008 D5 row 1 + `techstack.md`).
