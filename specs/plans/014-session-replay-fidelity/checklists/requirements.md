# Spec Quality Checklist: tellme Session-Replay Fidelity — Persist the Tool-Call Signature (round 014)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/014-session-replay-fidelity`

**Spec Path**: `specs/plans/014-session-replay-fidelity/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement (the storage mechanism is stated as an observable constraint; the exact field mapping is left to data-plan/DSL)
- [x] Edge cases cover the main high-risk scenarios (pre-round history with no signature, multiple steps, signature-less step, `-l` unaffected, OpenAI-family unaffected, opaque/verbatim treatment)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (the resume fix P1 before the data-preservation guarantee P2, which depends on it)
- [x] Each user story is independently verifiable
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries (no-regression + frozen vocabulary)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify` (persisted-field shape; legacy-signature behaviour)
- [x] This round's clarify questions were held to the budget (Round 1 = 2 questions)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A6) rather than asked
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain — both resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths (resume-with-tools replays the signature; empty-signature and tool-less turns unaffected)
- [x] Success criteria are measurable, verifiable, and technology-neutral (fake-recorded request; byte-identical `history.jsonl`)
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-14)

- **Q1 → Option 1 (shape)** — a dedicated nullable provider-agnostic `signature` field on the tool step; no generic metadata container (`FR-001`, `FR-002`, `A1`).
- **Q2 → Option 1 (legacy behaviour)** — an absent signature on resume keeps the unchanged best-effort replay; the provider's rejection is `the provider request failed` (exit 6); no pre-flight detection (`edge cases`, `A2`).

### Folded assumptions (accepted, not re-asked)

- **A3** — the history model stays provider-neutral; only Vertex/Gemini populates the field.
- **A4** — the field is omitted when empty, keeping existing lines byte-identical (round-007 determinism, `FR-005`).
- **A5** — the other 014 candidates (Gemini API family, ADC, concurrent tool-call matching) are out of scope.
- **A6** — rounds 001–013 semantics unchanged except the persisted signature.

### Still open (non-blocking — later-phase determinations)

- **Exact field placement in the record** — the DBML `history_step` column and the JSON key (omitted when empty) — a `/axb-data-plan` determination (the data owner adjudicates the provider-specific blob in the provider-agnostic model).
- **Resume witness shape** — how the two-process resume is exercised hermetically in the E2E suite (a second CLI run over the same history) — a `/axb-technical-research` / `/axb-dsl-refine` determination.
- **DSL step vocabulary** — the Given/Then rows for "a tool step that carried a signature" and "the replayed request carried the signature" — a `/axb-dsl-refine` determination.
- **`google`/other families** — whether any non-Vertex provider can supply a signature later; out of scope now (the field is simply left empty).

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (resume-with-tools P1 · existing-history preservation P2). Both clarify items are resolved; only data-plan / research / DSL-layer determinations remain. `/axb-spec-by-example` and `/axb-technical-research` can proceed in parallel.
