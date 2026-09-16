# Specification Quality Checklist: tool-schema well-formedness

**Created**: 2026-09-16

**Feature Directory**: `specs/plans/031-tool-schema-wellformedness`

**Spec Path**: `specs/plans/031-tool-schema-wellformedness/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria present (or explicitly noted N/A — this round corrects an advertised schema, adding no persisted data)

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
- [x] Low-risk undecided details disclosed as assumptions (the live-check is manual; `execute_command` left inline; broader schema-conformance out of scope)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Clarify Round 1 (2 questions, both answered "Option 1").** (Q1) *Verification surface* → **hermetic tool-schema well-formedness gate only** + a **manual** live Vertex/Gemini closeout check (round-013 style); keeps `make verify` offline. (Q2) *Recurrence guarantee* → **minimal fix + gate**: the shared schema construction declares the `reason` property; `execute_command` (inline, already compliant) is left unchanged; add the for-every-registered-tool `required ⊆ properties` check.
- **Grounding (live code).** `resourceSchema` (`internal/infrastructure/tools/filesystem.go:51`) emits `properties:{<extraProps>, max_output_tokens, timeout}` + `required:[<required>]`; 5 callers pass `"reason"` as *required* but never as a *property* → `required ⊄ properties`. `executeCommand.Parameters()` (`internal/infrastructure/tools/command.go`) builds inline and **does** declare `reason` (the lone compliant tool). The blind-spot test `TestToolSchemasRequireReason` asserts only that `required` *contains* `reason`, over the three readers only. Regression origin: round-024 fold `cfa005c`; carried forward by round-029 fold `eb0367c`.
- **Deliberately out of scope (recorded).** Teaching the hermetic E2E fake to schema-validate; a real-endpoint leg inside `make verify`; a structural "impossible-by-construction" schema builder; other finish reasons / provider-strictness modelling — all recorded forward items coordinated with `#60`.
- **No blocking gap remains.** The fix target (declare the `reason` property + gate), the verification surface (hermetic gate + manual live check), and the fail surface (unchanged class phrase / exit codes) are all locked. The exact check placement (which package/test file) and the gate's wording are ordinary RD decisions delegated to `/axb-technical-research` / `/axb-tasks`.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope derives from issue #64; both high-impact gaps were resolved in Clarify Round 1. Remaining open points (the check's exact location and naming) are low-risk RD details disclosed as assumptions. This plan may proceed to `/axb-spec-by-example` (expected skipped) and `/axb-technical-research`.
