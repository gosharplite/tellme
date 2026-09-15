# Specification Quality Checklist: agent write tools

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/029-agent-write-tools`

**Spec Path**: `specs/plans/029-agent-write-tools/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria present (or explicitly noted N/A)

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated to `/axb-clarify`
- [x] This round's clarify held to 1–3 questions (0 asked — see below)
- [x] Low-risk undecided details disclosed as assumptions (atomic temp-file placement; empty-`content` handling; the `--tool-usage` re-evaluation intent)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Operator-directed design (this session).** The write surface is exactly two tools — `replace_text` (**strict-unique**: `0` → error, `>1` → error, exactly `1` → replace) at **P1**, and `write_file` (**create-only** + **atomic** temp+rename) at **P2**. Both carry no security gate; both require `reason`; both use a 30 s default timeout. Decision trail: *"Let's do (b)"* → build the pair; *"A"* → create-only; *"replace_text = A (strict-unique)"*.
- **Deliberate omissions.** `append_text`, `undo_file_change`, `delete_path`, `create_directory` are **not** built (the shell covers them — `>>`, `git checkout`, `rm`, `mkdir -p`; none clears the "beats bash" bar). The reference's `SecurityManager` gate and `backupManager` snapshots are dropped (tellme's settled no-security/no-undo posture), so the tools reduce to pure filesystem operations.
- **No clarify needed.** All design decisions that could change story boundaries or acceptance were locked in-session before opening the round (scope = the pair; `write_file` = create-only; `write_file` = atomic; `replace_text` = strict-unique; timeout = 30 s; no security/undo). No high-impact gap remains; the atomic-write mechanism and the exact error wording are ordinary RD decisions delegated to `/axb-technical-research` / `/axb-dsl-refine`.
- **Recorded intent (measure, don't assume).** `write_file`'s marginal value is lower than `replace_text`'s; whether it survives is to be **measured** via the round-026 `--tool-usage` report over real dogfooding usage and revisited in a later round (cross-linked to issue #60). Recorded here so it is not lost; explicitly **out of scope** for this round.
- **Dogfooding-track context.** This is the first round of the dogfooding-enablement track (umbrella issue #60); its purpose is to let tellme create/edit files so it can begin developing its own repo. Follow-on rounds (skills injection, context management) are tracked on #60, not here.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope is operator-directed and every design decision is locked. Remaining open points (the exact atomic temp-file naming/placement, the exact tool error wording) are low-risk RD details and are disclosed as assumptions; none blocks planning. This plan may proceed to `/axb-spec-by-example` and `/axb-technical-research`.
