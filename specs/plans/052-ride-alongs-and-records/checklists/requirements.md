# Specification Quality Checklist: close [#115](https://github.com/gosharplite/tellme/issues/115) + [#116](https://github.com/gosharplite/tellme/issues/116) — ride-alongs (code) + records (disposition) (round 052)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/052-ride-alongs-and-records`

**Spec Path**: `specs/plans/052-ride-alongs-and-records/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (close #115: the two ride-alongs; close #116: the records' durable disposition)
- [x] No implementation/framework detail written as a *requirement* (the R-2 mechanism is an explicit RD decision behind FR-003; R-1's signature is the issue's own scalable shape)
- [x] Edge cases cover the main high-risk situations (gate set-drift; nil sink; seam-signature ripple; suggester invariant; frozen-record immutability; func-typed `Validate()`)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (the command tool is born with its sink) → US2 (the selection-policy owner) → US3 (the records' durable home)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR attached under the story
- [x] Global requirements hold only cross-story items (FR-007…FR-010)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] **Q1 → one round LOCKED** — the operator's goal names both issues for 052; #115 is small, #116 is a documentation close
- [x] **Q2 → (i) relocate + close completed LOCKED** — the three records move into ADR 0021 §Records (a non-frozen, indexed home); #116 closes completed
- [x] Questions settled by the operator's goal + the round-052 framing exchange; **no** open question (per the operator's direction to proceed)
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (ADR number A4; NOOP set A5/A6; R-2 mechanism A3)
- [x] No remaining `NEEDS CLARIFICATION` — clarify round 1 **CLOSED**

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (ctor injection; set signature; records in ADR 0021; byte-unchanged streams; gate green)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **`agentTools()` must stay parameterless + read-free** (round-033 FR-009) while the sink is injected at construction — the one real design tension; resolved as a `research.md` decision (D-series).
- **Seam ripple** — widening `deps.NewToolRegistry` to carry the sink touches `renderToolUsage` (offline, pass `nil`) and the test fixtures (`internal/cli/{cli,testdeps}_test.go`, `cmd/tellme/deps_test.go`) — enumerated at implementation.
- **Frozen-record immutability** — the relocation **copies** the three records; `specs/plans/036-*` / `040-*` are **not** edited.
- **No release valve** at RULE-E baseline 0 (ADR 0011/0016) — this round must not add a governed import.
- **Witness discipline**: unit pins + the gate (FR-009), never the E2E suite alone.
- **Atomicity** — per round: the code + ADR/truth + the records' relocation land as one PR. **No new Makefile target.**

## Ready determination

- [x] Ready to proceed to downstream planning — **clarify round 1 CLOSED** (Q1 one round · Q2 (i) relocate+close)
- [x] No high-impact requirement gap remains open

**Note**: plan package is the `/axb-specify` skeleton. Next pipeline step: `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
