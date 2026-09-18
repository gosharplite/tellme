# Specification Quality Checklist: `-l` defaults to `1` + green chrome accents (round 054)

**Created**: 2026-09-19
**Feature Directory**: `specs/plans/054-l-default-and-chrome-colour`
**Spec Path**: `specs/plans/054-l-default-and-chrome-colour/spec.md`

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (US1 bare `-l` ⇒ 1; US2 four green accents)
- [x] No implementation/framework detail written as a *requirement* beyond the necessary contract (the green **code** `\033[0;32m` is an explicit requirement — the operator pinned it)
- [x] Edge cases cover the main high-risk situations (`-l` + prompt; `-l` + flags; attached value; colour gate precedence; sanitization order; `turns.log`)
- [x] Key entities and success criteria present

## User stories & requirement attribution

- [x] Ordered by value: US1 (`-l` default) → US2 (colour)
- [x] Both independently verifiable
- [x] Both carry acceptance scenarios
- [x] Story-specific FR under each (FR-001/002; FR-003…006)
- [x] No global requirements needed (no cross-story FR)

## Gaps & clarify strategy

- [x] **Q1 (colour gate) LOCKED** by direction + precedent (`stderr` TTY && `!raw`); operator may veto
- [x] **Q2 (the four elements + whole-line `[Tool Reason]`) LOCKED** by the operator, divergence accepted
- [x] **Q3 (`turns.log` stays plain) LOCKED** (round-053 RF-53-1)
- [x] No remaining `NEEDS CLARIFICATION`

## Verifiability & success criteria

- [x] Acceptance covers the success path + the gates (terminal colour; piped/`-r` plain; `turns.log` plain)
- [x] SC measurable (byte-level on `stdout`; colour presence by `TELL_ME_FORCE_STDERR_TTY`; gates)
- [x] Assumptions are premises only

## Issues & corrections

- The chrome is **plain by design** today (round-017 D3) → this round **supersedes** that clause for the four elements (recorded in ADR 0023).
- The E2E reads **non-terminal** stderr with regexes → the colour must be gate-off there (the existing assertions stay valid); the colour is asserted through the `TELL_ME_FORCE_STDERR_TTY` seam.
- `turns.log` is "control-free" (ADR 0022) → the tee keeps the plain chrome (FR-006).
- The reference's own green layout differs on 3 of 4 elements → recorded divergence (Q2).

## Ready determination

- [x] Ready to proceed to `/axb-spec-by-example` (a user-visible change)
- [x] No high-impact requirement gap remains open
