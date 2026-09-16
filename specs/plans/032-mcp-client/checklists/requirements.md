# Specification Quality Checklist: remote MCP client

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/032-mcp-client`

**Spec Path**: `specs/plans/032-mcp-client/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement (the SDK/config shape live in research/truth, not as spec requirements)
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria present

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order (US1 capability → US2 non-stall robustness → US3 disable switch)
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated (resolved in-session as Q1–Q5; no `/axb-clarify` delegation needed — all decisions are operator-locked)
- [x] This round's decisions held to a small set (Q1–Q5, each one decision)
- [x] Low-risk undecided details disclosed as assumptions (exact discovery-bound value; check placement; live-check is manual)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Decisions locked in-session (Q1–Q5).** Q1 → 1 (remote Streamable HTTP only); Q2 → 1+3 (fixed small fast-fail discovery bound + per-server `ENABLED`; no cache); Q3 → 1 (official MCP Go SDK, confined behind a `tools.MCPClient` port); Q4 → 1 (full auth parity + reference tool naming); Q5 → 1 (MCP client only; stdio + MEMORY/PLUR + caching deferred).
- **Motivating defect (grounded in the reference).** `tell-me-go` builds MCP clients at the start of **every** CLI invocation, then the MCP plugin `ListTools` per server under a 30 s discovery cap and `wg.Wait()`s — so one offline server (`hf`) delayed **every** command (the operator had to comment it out). tellme bounds this with FR-008…FR-010 + NFR-002 (fixed fast-fail bound) and FR-011/FR-012 (`ENABLED`).
- **Deliberately out of scope (recorded forward items).** Local stdio transport; cross-invocation tool caching; MEMORY/PLUR automatic injection/learning; tool-call concurrency (still `#47` `not planned`).
- **No blocking gap remains.** The capability, the non-stall guarantee, the disable switch, the auth set, and the transport boundary are all locked. Remaining open points are ordinary RD decisions (exact bound value; discovery-context wiring; gate placement) delegated to `/axb-technical-research` / `/axb-tasks`.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope derives from the operator request plus the locked Q1–Q5. This plan may proceed to `/axb-spec-by-example` (an MCP journey may be added — the capability is CLI-observable) and `/axb-technical-research`.
