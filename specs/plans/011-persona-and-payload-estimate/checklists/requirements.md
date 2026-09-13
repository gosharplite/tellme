# Spec Quality Checklist: tellme Persona-on-the-Wire & Wire-Faithful Payload Estimate (round 011)

**Created**: 2026-09-13

**Feature Directory**: `specs/plans/011-persona-and-payload-estimate`

**Spec Path**: `specs/plans/011-persona-and-payload-estimate/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement (the estimation *heuristic* is explicitly deferred to `/axb-technical-research`; the spec pins only behaviour/inputs)
- [x] Edge cases cover the main high-risk scenarios (empty persona, piped stdin, `-r`, non-prompt paths, tool turns, no-usage)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (persona-on-the-wire P1 before the wire-faithful estimate P2, which depends on it)
- [x] Each user story is independently verifiable
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps are escalated to `/axb-clarify`
- [x] This round's clarify questions are held to the budget (Round 1 = 2 questions)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-13)

- **Q1 → Option 1 (estimate scope)** — the pre-flight estimate counts the **persona message + tool declarations + conversation messages** (`FR-006`); the three input components the provider's `prompt_tokens` covers.
- **Q2 → Option 1 (guarantee strictness)** — the round pins the estimate's **inputs + determinism** (`FR-007`, `FR-009`); no numeric-equality/tolerance assertion against the measured count.

### Folded assumptions (accepted, not re-asked)

- **A1** — persona is a separate leading `system` message (not `developer`, not merged into the prompt).
- **A2** — empty `PERSON` → no system message.
- **A3** — the persona rides every provider request during the run (incl. tool-driven completions).
- **A4** — only request content + the pre-flight estimate change; line format, measured source, and `stdout` byte-exactness unchanged.
- **A5** — the exact estimation heuristic is a `/axb-technical-research` determination.

### Still open (non-blocking — later-phase determinations)

- **Estimation heuristic** — the precise chars/token ratio, per-tool-declaration allowance, and whether any base term is included; a `/axb-technical-research` determination. Offline/dependency-free is fixed (`NFR-003`).
- **Persona plumbing to tool-driven completions (`FR-004`)** — how the persona reaches the session-summarisation tool's own request (client-level injection vs an explicitly threaded persona); a `/axb-technical-research` / `/axb-system-analysis` determination.
- **Existing-truth update wording (`FR-011`)** — how the `the request carried no earlier exchange` assertion is reworded to account for the leading `system` message; a `/axb-dsl-refine` determination.
- **Persona-assertion granularity (US1)** — whether the persona message is asserted via a new Given/Then row or an extension of the existing fake-provider request-recording rows; a `/axb-dsl-refine` determination.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (persona-on-the-wire · wire-faithful estimate); the estimate-inputs assertion and the persona-message assertion are the round's executable deliverables. All clarify items are resolved; only research/DSL-level determinations remain. `/axb-spec-by-example` and `/axb-technical-research` can proceed in parallel.
