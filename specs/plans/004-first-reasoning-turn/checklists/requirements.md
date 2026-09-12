# Specification Quality Checklist: tellme First Reasoning Turn (round 004)

**Created**: 2026-09-12

**Feature Directory**: `specs/plans/004-first-reasoning-turn`

**Spec Path**: `specs/plans/004-first-reasoning-turn/spec.md`

## How to use

- Check each item against the current `spec.md`.
- When an item does not pass, record the concrete gap and correction in "Issues & Correction Log".
- If `NEEDS CLARIFICATION` items remain, state explicitly whether they block subsequent planning.

## Content Completeness

- [x] All mandatory sections are completed
- [x] Feature topic, scope, and main flows are clearly expressed
- [x] No implementation technology, framework, or code detail is written as a requirement (the HTTP transport and SDK choice are deferred to `/axb-technical-research`; the wire family is a domain concept carried from the round-003 provider schema)
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria are present, or explicitly marked not applicable

## User Stories & Requirement Attribution

- [x] User stories are ordered by business value and delivery order
- [x] Each user story can be verified independently
- [x] Each user story includes acceptance scenarios
- [x] FR/NFR attributable to a single story are attached directly under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement is duplicated between the story and global sections

## Gaps & Clarification Strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify`
- [x] This round's clarify budget stayed within 1–3 questions (exactly 3 questions addressed)
- [x] Low-risk undecided details are disclosed via assumptions or non-blocking edge cases
- [x] All high-impact decisions are ratified with zero remaining blocking gaps

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises and boundaries, with no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Correction Log

- **Clarify Round 1 Q1 (user-ratified)**: Scope = one in-memory, non-streaming turn per process; no persistence, no session loop. Streaming, `history.jsonl`, and the interactive loop are deferred.
- **Clarify Round 1 Q2 (user-ratified)**: The first adapter targets the OpenAI-compatible family (`openai`/`deepseek`/`kimi`), matching the `deepseek-flash` harness default; Gemini/Vertex and Anthropic are deferred.
- **Clarify Round 1 Q3 (user-ratified)**: A provider/transport failure uses the new frozen class phrase `the provider request failed` and the new distinct exit code `6` (extends the frozen `0/2/3/4/5` table to `0/2/3/4/5/6`).
- **Deferred to `/axb-technical-research` (not asked)**: transport = stdlib `net/http` with no provider SDK; the round-001 no-network capability guard is re-scoped to the offline boot/`--version`/`-d` paths (research owns the Decision-5 amendment). Recorded as assumptions in the spec.
- **Out of scope (deferred)**: tool calls / agentic loop, MCP, memory, `Turn`/`History` persistence, summarisation, cost/metrics, TUI/`-i`, `browse`/retry/edit, multi-provider failover, `cobra` subcommands.

## Ready Determination

- [x] Ready for subsequent planning
- [ ] Still requires high-impact requirement gaps to be filled

**Note**: All 3 high-impact questions were resolved in Clarify Round 1. The two remaining undecided items (transport choice, capability-guard re-scope) are technical decisions owned by `/axb-technical-research` and are recorded as explicit assumptions here, so they do not block `/axb-spec-by-example` or `/axb-technical-research`. The specification is complete, self-consistent, and ready for acceptance-criteria formalization.
