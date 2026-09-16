# Specification Quality Checklist: provider-transport truncation guard

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/030-provider-truncation-guard`

**Spec Path**: `specs/plans/030-provider-truncation-guard/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria present (or explicitly noted N/A — this round adds behaviour, not data)

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated to `/axb-clarify`
- [x] This round's clarify held to 1–3 questions (2 asked — see below)
- [x] Low-risk undecided details disclosed as assumptions (the unset-`MAX_TOKENS` posture; the exact error detail wording; other finish reasons out of scope)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Clarify Round 1 (2 questions, both answered "Option 1").** (Q1) *Trigger scope* → **universal**: any `finish_reason == "length"` / `finishReason == "MAX_TOKENS"`, text **or** tool call, fails the turn (mirrors the reference; conservative). (Q2) *Failure class* → **reuse the existing provider class**: `the provider request failed` + exit code **6**; no new class phrase, no new exit code.
- **Grounding (reference evidence).** `tell-me-go`'s `gemini/response.go` `checkGeminiTruncation` (terminal; universal — text-only pinned) and `openai/metrics.go`'s `finish_reason == "length"` check (terminal; universal). tellme's own gap verified at PR #61 review: no `finishReason`/`finish_reason` read in either adapter.
- **Deliberately out of scope (recorded).** Other finish reasons (Vertex/Gemini `SAFETY`/`RECITATION`, OpenAI-compatible `content_filter`); the **request** side (no default output cap is added — the issue asks only for the documented "unset `MAX_TOKENS` → provider default" sentence).
- **No blocking gap remains.** The trigger set (`length`/`MAX_TOKENS`), the two families, the universal scope, and the failure class are all locked. The exact error-detail wording and the decode-site placement are ordinary RD decisions delegated to `/axb-technical-research` / `/axb-dsl-refine`.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope is operator-directed, and both high-impact gaps were resolved in Clarify Round 1. Remaining open points (the exact error detail wording, the guard's decode-site ordering vs the existing "no usable answer" error) are low-risk RD details disclosed as assumptions; none blocks planning. This plan may proceed to `/axb-spec-by-example` and `/axb-technical-research`.
