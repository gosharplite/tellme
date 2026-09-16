# Specification Quality Checklist: skills system — load & list pre-loaded skills

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/033-skills-system`

**Spec Path**: `specs/plans/033-skills-system/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement (loading/parsing/schema details live in research/truth, not as spec requirements)
- [x] Edge cases cover the main high-risk situations (nested dirs, malformed frontmatter, duplicates, absent dir, large catalog, offline paths)
- [x] Key entities and success criteria present

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order (US1 list → US2 on-demand use)
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 surface model, Q3 list delivery — both locked in-session)
- [x] This round's clarify question count held to a small set (Q1 + Q3; Q2 moot given Q1 → 1)
- [x] Low-risk undecided details disclosed as `NEEDS CLARIFICATION` / assumptions (discovery rule; list field detail)
- [x] Remaining `NEEDS CLARIFICATION` items marked explicitly non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Decisions locked in-session (clarify).** Q1 → 1 (**on-demand only** — list + reuse `read_files`; no automatic injection, no selector); Q2 → moot (no injected block); Q3 → 1 (**a model-callable read-only `list_skills` tool**; not an offline flag).
- **Motivating gap (grounded in the code).** tellme has no skill subsystem: no `.go` file mentions "skill", no `domain/skills`/`infrastructure/skills`/`agent/skills`, and no `list_skills` tool. The `docs/skills/` folder tellme.sh copies is currently inert. The operator wants a **minimal** system that loads only `<TELL_ME_HOME>/docs/skills` and drops the skills.sh `skillssh` toolkit.
- **Deliberately out of scope (recorded).** skills.sh `.skills/` source; `search_skills`/`install_skill`/`remove_skill`; cross-source merging; automatic relevance-based injection (the reference's `SkillSelector` + `ContextTransformer`).
- **No blocking gap remains.** The surface model and the list-delivery mechanism are locked; the remaining open points (exact discovery rule; list field detail; tool schema) are ordinary RD decisions delegated to `/axb-technical-research` / `/axb-tasks`.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope derives from the operator request plus the locked Q1/Q3. This plan may proceed to `/axb-technical-research` (the loading/parsing decision + `techstack.md`) and `/axb-system-analysis`. `/axb-spec-by-example` may be **skipped** (a listing tool is CLI-observable but the sole new journey is thin — the API/data planners are NOOP and the CLI end is carried to `/axb-dsl-refine`).
