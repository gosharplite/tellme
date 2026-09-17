# Specification Quality Checklist: a hermetic `make` Go-toolchain environment (round 043)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/043-hermetic-make-go-env`

**Spec Path**: `specs/plans/043-hermetic-make-go-env/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (the GNU-make syntax is an explicit RD decision A3, not an FR)
- [x] Edge cases cover the main high-risk situations (per-recipe inline override, nested `go build`, `$(shell)` probes, direct-invocation bypass, cold cache/proxy, escape hatch)
- [x] Key entities and success criteria are present (the neutralise set / preserve set / hermetic boundary)

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (hostile env neutralised) → US2 (recorded + citable) → US3 (no clean-env behaviour change)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-013, NFR-004)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 approach · Q2 mechanism · Q3 sets · Q4 scope · Q5 governance · Q6 ownership vs round-042 `childEnv` · Q7 residuals · Q8 round ordering)
- [x] This round's clarify questions **asked one at a time; all eight answered**
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (exact GNU-make syntax A3; ADR A4; `/axb-dsl-refine` NOOP A5)
- [x] No remaining `NEEDS CLARIFICATION` — round-1 answers **locked**:
  - **Q1 → Option A** (neutralise at the invocation boundary).
  - **Q2 → A1** (one top-of-`Makefile` `export`/`unexport` block).
  - **Q3 → V1** (neutralise `GOENV`/`GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOWORK` + ambient `GOOS`/`GOARCH`/`GOARM`; preserve warm-cache + network/checksum sets; host-default `CGO_ENABLED`).
  - **Q4 → S1** (all targets).
  - **Q5 → G1** (ADR 0012 + index + `techstack.md`).
  - **Q6 → D1** (keep `tools/arch` `childEnv` as defence-in-depth; `Makefile` is the primary owner).
  - **Q7 → R1 + R2 + R3** (record the bare-`go` residual; fold the `fmt`/`tidy` positive control into verification; record the `GOENV=off` escape-hatch consequence).
  - **Q8 → O1** (round `043-*` now; R2 of #92 → `044-*`).

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (hostile envs neutralised; clean env unchanged)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-005)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **`GOENV=off` is load-bearing** — recorded explicitly (FR-002, research D4): unsetting/emptying `GOFLAGS` does **not** neutralise a persisted `go env -w GOFLAGS=…` (Go falls back to the env file for unset **and** empty values — round-042 F-2). This is the issue's exact failing case.
- **Not V2 (global `CGO_ENABLED=0`)** — would be a behaviour change (excludes future cgo-tagged code) and violate the "no behaviour change for a clean env" criterion (research D2/Q3).
- **The round-042 `childEnv` is retained as defence-in-depth** (research D5), so `tools/arch/**` behaviour is unchanged (frozen round-042 history); only a cross-reference **comment** may be added.
- **Residuals recorded** (research D8): bare-`go`-outside-`make`, and the `GOENV=off` env-file escape hatch.
- **Ordering**: the `STATUS.md` roadmap line (currently naming 043 = R2 of #92) MUST be updated at closeout to reflect 043 = this round, R2 → 044 (Q8/O1).

## Ready determination

- [x] Ready to proceed to downstream planning
- [ ] A high-impact requirement gap must be closed first — **none remaining** (round-1 answers locked)

**Note**: plan half only — stop before `/axb-tasks` and `/axb-implement` (the `Makefile` change and the witnesses are a later phase). Everything up to and including `/axb-system-analysis` is in scope for this branch. Next pipeline step after this half: `/axb-tasks` (operator's call).
