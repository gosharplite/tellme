# Tasks — round 092 `092-single-source-agent-tool-set`

Constraint-ordered execution.

## Phase 1 — Setup

- [X] **T001** Create the branch `092-single-source-agent-tool-set` off `dev` and the plan package
      `specs/plans/092-single-source-agent-tool-set/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational

- [X] **T002** Reproduce the baseline: the base-set composition is duplicated in `cmd/tellme/deps.go`
      (`assembleAgentTools`) and `tests/e2e/steps/tool_usage.go` (`registeredToolNames()` /
      `recordableToolNames()`); confirm no carrier binds the two.

## Phase 3 — Test Alignment & Implementation

- [X] **T003** Add `internal/infrastructure/tools/agentbase.go` — `NewAgentBaseTools(sink domaintools.OutputSink)`
      the canonical base-set owner (readers · `search_files` · write pair · `execute_command` · `list_skills`).
- [X] **T004** `cmd/tellme/deps.go` — `assembleAgentTools` delegates its base half to `NewAgentBaseTools(spec.Sink)`;
      the `vision` append + `resolveImageCeiling` unchanged.
- [X] **T005** `tests/e2e/steps/tool_usage.go` — `registeredToolNames()` / `recordableToolNames()` delegate to
      `NewAgentBaseTools(nil)` (+ the unchanged `read_image` union).
- [X] **T006 (carrier)** `cmd/tellme/deps_test.go` — `TestAgentToolsIsTheCanonicalBaseSet` (production base set
      == the canonical owner).
- [X] **T007 (carrier)** `tests/e2e/steps/tool_usage_test.go` (NEW) — `TestRegisteredToolNamesIsTheCanonicalBaseSet`
      (the e2e enumerator == the canonical owner).
- [X] **T008 (witness, re-run)** Reproduce W-1/W-3 (re-inline a divergent copy in either caller) ⇒ the carrier
      reddens; W-2 (drop a tool from the owner) ⇒ round-090's doc carrier reddens; revert all.

## Phase 4 — Green / verify

- [X] **T009 (Green)** `make check` green; E2E counts unchanged (330 · 2487); `verify-architecture` 0
      violations; `modelith-check` no drift; `gofmt`/`goimports` clean; `go.mod`/`go.sum` unchanged.

## Phase 5 — Records

- [X] **T010** **ADR 0062** (`docs/decisions/0062-*.md`) + `docs/decisions/README.md` index row;
      `specs/truth/techstack.md` placement note.
- [X] **T011** `STATUS.md` — round 092 in flight (→ delivered at closeout); the day summary; the PR.

---

## Claim → Witness ledger

| Claim | Witness | Mutant that reddens it |
| --- | --- | --- |
| (FR-001) one canonical owner; both callers derive from it | `infratools.NewAgentBaseTools`; both callers delegate (I-1) | re-inline the constructors in either caller ⇒ the matching carrier reddens |
| (FR-002) the e2e enumerator is bound to the owner | `TestRegisteredToolNamesIsTheCanonicalBaseSet` | `registeredToolNames()` re-inlined and diverging ⇒ red |
| (FR-003) the production base set is bound to the owner | `TestAgentToolsIsTheCanonicalBaseSet` | `assembleAgentTools()` re-inlined and diverging ⇒ red |
| (FR-004) the capability gate is unchanged | existing `TestCompositionResolvesTheFamilyAwareImageCeiling`; the unchanged union enumerator | drop the `vision` gate ⇒ the existing pin reds |
| (W-2) the canonical owner is load-bearing | round-090's `TestOfferedSetDocMatchesTheLiveRegistry` | remove a tool from `NewAgentBaseTools` ⇒ doc (8) vs live (7) red |
| (I-2 / W-4) behaviour identity | the full unit suite + the godog E2E, **no assertion changed**; E2E counts unchanged (330 · 2487) | a set/order change ⇒ a count/behaviour pin red |
| (I-3) layer baseline 0 | `make verify-architecture` | a new import edge ⇒ a RULE-A/B/E violation |
