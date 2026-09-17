# Specification Quality Checklist: tool-output sanitization + empty-submit no-op (round 038)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/038-tool-output-sanitize-and-empty-submit`

**Spec Path**: `specs/plans/038-tool-output-sanitize-and-empty-submit/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (two folded operator issues: #78 output control-sequence leak; #76 empty-submit quit)
- [x] No implementation technology, framework, or code detail is written as a requirement (the escape classes are stated as behaviour; the neutral close is a contract)
- [x] Edge cases cover the high-risk situations (dropped partial line, killed-mid-output, colour-with-reset, split-across-writes, repeated empty submits, whitespace-only)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (US1 control-sequence-free output P1; US2 empty-submit no-op P2; US3 neutral close P3)
- [x] Each user story is independently verifiable (US1: E2E + unit hostile fixture; US2: E2E empty-then-real submit; US3: writer unit pin)
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story (FR-001..FR-005)
- [x] Global requirements hold only cross-story / non-attributable items (FR-006/FR-007)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's clarify is one round, three decisions (Q1 strip class; Q2 neutral close; Q3 empty-submit parity) — all resolved with the operator (`1,1,1`)
- [x] Low-risk undecided details are disclosed as assumptions (sanitizer home; `/axb-ui-plan` skipped; folded issue) rather than asserted as decided
- [x] Any remaining `NEEDS CLARIFICATION` is marked for blocking/non-blocking status (none remain — Q1–Q3 resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Two folded issues disclosed**: #78 (the reported colour leak) and #76 (empty-submit quit) are **separate** operator-filed defects folded into one round at the operator's direction; each keeps its own acceptance rule and witness.
- **Recorded divergence**: the reference has **no** output sanitizer — this round intentionally diverges (it fixes a real terminal-state leak present in both), recorded in `research.md` + `truth-delta.md`.
- **Scope guard recorded**: FR-006/FR-007 pin the unchanged surfaces (flags, exit codes, vocabulary, tool surface, transport, records, tool result, `stdout`, offline paths).
- **Witness layering**: US1's E2E carrier is valid (the escape byte is directly observable in the captured `stderr`); US3's neutral close is a byte-level guarantee pinned at the unit layer.

## Ready verdict

- [x] Ready for downstream planning
- [ ] Still needs high-impact requirement gaps closed

**Note**: a presentation/parity change. `/axb-spec-by-example` is **NOT** a NOOP (both changes are user-visible). `/axb-ui-plan` is **skipped** (no chrome change). The load-bearing downstream steps are `/axb-technical-research` (two `techstack.md` row MODIFYs) → `/axb-system-analysis` (1 CLI end → `/axb-dsl-refine`; api/data NOOP) → `/axb-dsl-refine` (2 new truth Rules + `chat/dsl.md` rows) → `/axb-tasks` → `/axb-implement` (unit + E2E pins).
