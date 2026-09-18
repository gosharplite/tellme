# Specification Quality Checklist: close [#103](https://github.com/gosharplite/tellme/issues/103) — offline session commands honour `-c` + a `-t` flag (round 053)

**Created**: 2026-09-19

**Feature Directory**: `specs/plans/053-offline-session-config-and-turns-flag`

**Spec Path**: `specs/plans/053-offline-session-config-and-turns-flag/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (US1 = `-c`-honouring for `-l`/`--new`; US2 = `-t`)
- [x] No implementation/framework detail written as a *requirement* beyond the fix's necessary constraint (the mode-only/offline clause is [#103](https://github.com/gosharplite/tellme/issues/103) AC3, a boundary not a design)
- [x] Edge cases cover the main high-risk situations (`-t`×`-l` precedence; `-t`+`--new`; env+`-c`; absent/invalid `-c`; absent default; `-c` path form)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (`-c` selects the session) → US2 (`-t` flag)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR attached under the story (FR-001…FR-005)
- [x] Global requirements hold only cross-story items (FR-006…FR-011)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] **Q1 → (C1) LOCKED** — `tellme` writes its **own** `turns.log` (its rendered turn chrome — no new format); `-t` prints it. Rejected (A) print `tokens.log`, (B) project `history.jsonl`, (C2) reproduce `tell-me-go`'s exact bytes.
- [x] **Q2 → (A) LOCKED** — an **explicit** `-c` that cannot be honoured **fails** (existing config-error phrase); an absent default stays tolerant.
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (A1–A6; the `-l`-default-1 parity is explicitly **out of scope**)
- [x] No remaining `NEEDS CLARIFICATION` — **clarify round 1 CLOSED** (Q1 (C1) · Q2 (A))

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (differential `-l -c`; env-wins; offline; `--new -c`; `-t`)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-005)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **US2 grew (Q1 → C1)** — the round now produces a **persisted `turns.log`** (a new writer on the turn path) **and** the `-t` reader; the writer's seam is an RD decision (`research.md`), the content = tellme's existing chrome renderers (no new format).
- **The single seam** — `resolveWorkspace`/`historyMode` is shared by `-l`, prompt-less `--new`, and (newly) `-t`; widening it once covers all three (FR-003/FR-006).
- **Offline boundary** — the `-c` read must be **mode-only** (`config.Load` is already a pure parse); routing through `resolve()` would violate [#103](https://github.com/gosharplite/tellme/issues/103) AC3.
- **`dispatchReporting` precedence** — `-t` must be given a defined slot in the offline reporting order (`-d` → `-l` → `--tool-usage`) (`cli.go:876-891`).
- **Persisted-artifact truth** — `turns.log` touches `specs/truth/data/data-model.dbml` + `--new` archiving (data-plan = MODIFY, not NOOP).
- **Closed phrase vocabulary** — any new error path must reuse an existing `tellme: {phrase}` (`specs/truth/features/cli/**/dsl.md`).
- **No new dependency** — stdlib-only (`go.mod`/`go.sum` unchanged).

## Ready determination

- [x] Ready to proceed to downstream planning — **clarify round 1 CLOSED** (Q1 → (C1) · Q2 → (A))
- [x] No high-impact requirement gap remains open

**Note**: plan package is the `/axb-specify` skeleton. Next pipeline step: **`/axb-spec-by-example`** (A5: invoked — a user-facing new flag + a corrected session selection) → then `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement`.
