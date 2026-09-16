# Tasks — round 034 (tool-call log parity)

Plan package: `specs/plans/034-tool-call-log-parity`
Core inputs: `spec.md` (FR-001–FR-017, folds G1–G10), `research.md` (D1–D10), `plan.md`, `truth-delta.md`, `docs/decisions/0005-tool-call-log-parity.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/chat/{watching-the-tool-loop,presenting-the-turn,presenting-the-post-turn-status,presenting-the-progress-spinner,reporting-the-payload-status,estimating-the-wire-payload,failing-the-tool-loop}.feature` and `chat/dsl.md` (**15** new round-034 Then rows).

## Phase 1 — Setup

*(Omitted — stdlib-only; no new dependency, no new module. `go.mod`/`go.sum` unchanged.)*

## Phase 2 — Foundational

- [ ] T001 [FOUNDATIONAL] Landing skeletons (Zero Shared Edits). **只做**：add `internal/ui/toolcall.go` (empty `FormatToolEngine`/`FormatToolReason`/`FormatToolAction`/`FormatToolResult`), `internal/ui/tooloutput.go` (empty header/separator + a per-line writer), a `compositeObserver` skeleton in `internal/cli`, the `CallBegin`/`CallEnd` observer types in `internal/domain/agent`, a `toolOutputSink` struct field on the command tool, and the stepdef landing files for the 15 new rows. **不做**：no behaviour, no wiring.
- [ ] T002 [FOUNDATIONAL] Extend `agentport.LoopObserver` with the **decoupled** interface and make the CLI's presenter a `compositeObserver` (block renderer + spinner). **只做**：the seam + a no-op default; **不做**：rendering. **Signature (approved — PR #70 re-review):**
  ```go
  type LoopObserver interface {
      OnCallBegin(callIndex int, messages []llm.Message)                          // CLI computes llm.EstimatePayload(person, toolDefs, messages)
      OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) // usage.PromptTokens IS the measured payload
      // ... existing waiting-phase methods (OnInferenceStart/End, OnToolsStart/End, Before/AfterToolLog) preserved
  }
  ```
  `AgentLoop` **gains no `PayloadEstimate` field** and no persona; no `measuredPayload` parameter.

## Phase 3 — Test Alignment & Implementation

- [ ] T003 [BDD-REMOVE] Retire round-022 assertions: the no-reason Rule + its stepdef; the round-022 blank-line rows and their stepdefs; the `[Tool] <name> - <reason>` shape assertions.
- [ ] T004 [BDD-ALIGN] Retarget the existing tool-log stepdefs (Reason row) to the decomposed `[Tool Reason]` line; align the `offering` / single-frame payload fixtures (per G5/G10).
- [ ] T005–T019 [P] [BDD-RED] One stepdef per **new round-034 DSL row** — **15 rows**, one row per task:
  - T005 `the run reported the tool step marker`
  - T006 `the run reported the action for the tool call "{tool}" without the reason`
  - T007 `the run reported the result for the tool call "{tool}"`
  - T008 `the run reported the reasons again before the measured payload status`
  - T009 `the run reported the action for the tool call "{tool_a}" before the action for the tool call "{tool_b}"`
  - T010 `the run streamed the command's output on its diagnostic output`
  - T011 `the run streamed no command output block for the writing command`
  - T012 `the turn is framed once per model request`
  - T013 `the turn number advances from one frame to the next`
  - T014 `the run reports the post-turn status once per model request`
  - T015 `the last post-turn status trails the answer`
  - T016 `the progress spinner does not appear while the command's output streams`
  - T017 `each model request reports an estimated payload status`
  - T018 `the run reported a tool step marker for each of the {count} executed rounds`
  - T019 `the final model request reports no tool step marker`
- [ ] T020 [P] [UNIT] `internal/ui.toolcall` formatters: `FormatToolEngine`/`Reason`/`Action`/`Result` — the rune-safe caps (190-rune arg → 189 runes; 201-rune result → 200 runes), the single U+2026 inside the cap, sorted-keys + `json.Number` (`1000000`, never `1e+06`), the `reason`-exclusion, and the unparseable-args → `<tool>()` branch.
- [ ] T021 [P] [UNIT] the `[Tool Output]` writer (header/separator/line) + the sink's line-assembly across two writers + the trailing-partial-line drop.
- [ ] T022 [P] [UNIT] the CLI estimator: the k-th frame's estimate equals `llm.EstimatePayload(res.Person, ToolDefs(reg), wire_k)` for a scripted ≥2-round turn; call 1 byte-identical; the observer's `OnCallBegin` carries messages (no loop field).
- [ ] T023 [P] [UNIT] the persistence invariant: the persisted record set is the `Reported` subset of `result.Calls` on completion and empty on every error exit.
- [ ] T024 review gate (subagent) — **all 15 rows' stepdefs landed**; no undefined steps; the assembler gate (`TestAgentToolSchemasAreWellFormed`) untouched.

## Phase 4 — Feature implementation

- [ ] T025 [BDD-GREEN] 4A — `chat/watching-the-tool-loop.feature`: the decomposed log (`Engine`/`Reason`/`Action`/`Result`) + the grouped post-call reasons + the in-order (distinct-tool) Example. **Test Scope**: `watching-the-tool-loop.feature`. **Read**: `spec.md` FR-001–FR-007; `research.md` D1/D3/D4/D9; `chat/dsl.md` (round-034 rows).
- [ ] T026 [BDD-REFACTOR] 4A tidy under green.
- [ ] T027 [BDD-GREEN] 4B — the per-AI-call frame cadence + the final-call tail deferral + the CLI-computed per-call estimate (observer messages). **Test Scope**: `presenting-the-turn.feature`, `presenting-the-post-turn-status.feature`, `reporting-the-payload-status.feature`, `estimating-the-wire-payload.feature`. **Read**: FR-008–FR-010b; D2/D5/D6.
- [ ] T028 [BDD-REFACTOR] 4B tidy under green.
- [ ] T029 [BDD-GREEN] 4C — the live `[Tool Output]` block (shell-class, non-`output_file`; bounded-and-stopped; single-writer yield). **Test Scope**: `watching-the-tool-loop.feature`, `presenting-the-progress-spinner.feature`. **Read**: FR-010–FR-012; D7.
- [ ] T030 [BDD-REFACTOR] 4C tidy under green.
- [ ] T031 [CODE-REMOVE] 4D — delete the round-022 `FormatToolLog` + its call site + the blank-line emission; retire the round-022 `dsl.md` rows.
- [ ] T032 [REGRESSION] 4E — the bound-reached witness in `failing-the-tool-loop.feature` (N frames, `M` engine lines, one engine-less final frame, failure, exit 7, nothing persisted); `make verify`; the topology audit; and the **falsifiability witnesses** for SC-006 (190-rune arg → 189; rune-boundary cut; sorted vs source order; `json.Number` vs `%v`; no engine line on the bound-reached call; frozen estimator seam; `output_file` block suppression; stop-vs-cap wording; per-call frame count; tail-before-answer rejected).
- [ ] T033 [REGRESSION] `stdout` byte-exact; offline paths unchanged; no new dependency; the tool surface / class-phrase vocabulary / exit codes unchanged.

## Pre-Delivery Orphan Coverage Sweep

All non-NOOP `truth-delta.md` rows and every `research.md` decision (D1–D10) are referenced by a task `Read` or delivered above; the grill folds G1–G10 map to FR-001–FR-012 → T025/T027/T029/T031/T032; the **15** new DSL rows map 1:1 to T005–T019. **Target: 0 orphans.**
