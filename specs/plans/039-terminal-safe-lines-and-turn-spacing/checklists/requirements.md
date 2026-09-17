# Specification Quality Checklist: terminal-safe `[Tool …]` lines + turn-output blank-line grouping (round 039)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/039-terminal-safe-lines-and-turn-spacing`

**Spec Path**: `specs/plans/039-terminal-safe-lines-and-turn-spacing/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (two folded workstreams: issue #80 sibling sanitization; the operator's blank-line spacing)
- [x] No implementation technology, framework, or code detail is written as a requirement (the escape class and the blank positions are stated as behaviour; the single-owned home is stated as a policy, not a file layout requirement)
- [x] Edge cases cover the high-risk situations (escape-only reason, split-across-writes, folded newlines, escape-bearing key, multi-call/reason-less call, no-usage, non-tool turn, raw/-i/non-terminal)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (US1 control-free siblings P1 — it fixes a real terminal-affecting leak; US2 blank-line grouping P2 — additive readability)
- [x] Each user story is independently verifiable (US1: unit hostile fixture + E2E escape; US2: E2E blank-line positions)
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story (FR-001..FR-008)
- [x] Global requirements hold only cross-story / non-attributable items (FR-009/FR-010/FR-011)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's decisions were **all** resolved in-session with the operator (Q1–Q8, asked one at a time) — **no** separate `/axb-clarify` round is needed (the round-029/037 operator-locked pattern)
- [x] Low-risk undecided details are disclosed as assumptions (single-owned home is an enabler; `/axb-ui-plan` skipped; reference check deferred to research) rather than asserted as decided
- [x] No `NEEDS CLARIFICATION` remains (Q1–Q8 locked; the remaining items are research-phase determinations, not requirement gaps)

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Anchored + folded**: issue [#80](https://github.com/gosharplite/tellme/issues/80) is the anchor; the blank-line workstream is an operator request folded into the same round (the round-038 #78+#76 precedent).
- **Recorded decision**: the neutral-close restore stays `[Tool Output]`-scoped (Q4) — a deliberate **non**-change recorded so it is not silently re-litigated.
- **Byte-identical regression guard**: moving `sanitizeControl` to a single home must not change `[Tool Output]`'s existing output (a round-038 regression pin).
- **Interaction with #69**: this round `extends` the round-036 blank-reason predicate's input (FR-005) but does **not** consolidate its three sites (that single-ownership refactor stays on #69).
- **Witness layering**: US1's carriers are unit hostile fixtures (the escape byte is observable but the sibling E2E path is awkward) plus a direct E2E escape fixture; US2's carriers are E2E blank-position assertions (the blank is directly observable in the captured `stderr`).

## Ready verdict

- [x] Ready for downstream planning
- [ ] Still needs high-impact requirement gaps closed

**Note**: two `stderr`-presentation workstreams. `/axb-spec-by-example` is **NOT** a NOOP (both are user-visible). `/axb-ui-plan` is **skipped** (no chrome change). The load-bearing downstream steps are `/axb-technical-research` (a `techstack.md` MODIFY + a new **ADR 0008** that **supersedes ADR 0007** — research D9, not an in-place amendment — + the reference divergence check) → `/axb-system-analysis` (1 CLI end → `/axb-dsl-refine`; api/data NOOP) → `/axb-dsl-refine` (a sanitized-siblings Rule + a blank-line Rule + `chat/dsl.md` rows) → `/axb-tasks` → `/axb-implement` (unit + E2E pins).
