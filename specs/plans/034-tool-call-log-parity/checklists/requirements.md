# Specification Quality Checklist: tool-call log parity (round 034)

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/034-tool-call-log-parity`

**Spec Path**: `specs/plans/034-tool-call-log-parity/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement (the emission site, stream, and separator literals live in research/DSL, not as spec requirements)
- [x] Edge cases cover the main high-risk situations (multi-call rounds, long arg/result truncation, binary result, unbounded output, empty reason, tool-less turn, surfaces, offline paths)
- [x] Key entities and success criteria present

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order (US1 decomposed log → US2 per-call cadence → US3 live `[Tool Output]`)
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1–Q7 — all locked in-session, one decision at a time)
- [x] This round's clarify question count held to a small set (7 locked decisions; no unresolved question)
- [x] Low-risk undecided details disclosed as assumptions (exact separator literal; emission site; clock seam)
- [x] Remaining `NEEDS CLARIFICATION` items marked explicitly non-blocking (none block)

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Decisions locked in-session (clarify, one at a time).** Q1 → 1 (full per-AI-call parity; `[Tool Reason]` twice — pre-action and grouped post-call); Q2 → 1 (`[Tool Output]` live stream); Q3 → 1 (verbatim reference templates/constants; all calls carry `reason`); Q4 → 1 (keep tellme's status-block line formats); Q5 → 1 (unbounded `[Tool Output]`); Q6 → 1 (keep spinner, inside the `[Tool Output]` block); Q7 → 1 (all prompt surfaces; `-r` does not suppress).
- **Motivating gap (grounded in the code + the reference).** tellme's tool log is a single line (`internal/ui/toollog.go` `FormatToolLog` → `[HH:MM:SS] [Tool] <name> - <reason>`, emitted unconditionally at `internal/agent/agentloop.go:239`). The reference (`tell-me-go` `internal/ui/renderer_metrics.go` + `internal/tools/workspace/shell.go`) instead decomposes each call (Engine/Reason/Action/Output/Result) and emits the status block per AI-endpoint call.
- **Superseded (recorded).** Round 022's single-line `[Tool]` log and its "one blank line before the answer" rule; round 022's "states no reason" Rule/DSL row (all calls carry `reason`); rounds 017/023/027's turn-chrome cadence becomes per-AI-call (the `Ready` footer + measured payload + metrics per call).
- **Out of scope (recorded).** The reference's `[Info] Starting chat...` line; the `[Bypassed]` security line (no security layer); the reference's own status-block line formats (Q4).
- **No blocking gap remains.** The rendering shape, the templates/constants, the cadence, the surfaces, and the streaming/bound behavior are locked; the remaining open points (exact separator literal; emission site; clock seam reuse) are ordinary RD decisions delegated to `/axb-technical-research` / `/axb-dsl-refine`.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope derives from the operator request plus the locked Q1–Q7. `/axb-spec-by-example` is expected to render the acceptance journeys (a tool-using run shows the decomposed log + per-call blocks + a live `[Tool Output]` stream). The plan proceeds to `/axb-technical-research`, `/axb-system-analysis` (CLI end carried to `/axb-dsl-refine`), `/axb-dsl-refine` (MODIFY `chat/watching-the-tool-loop.feature` + `chat/dsl.md`), and `/axb-tasks`.
