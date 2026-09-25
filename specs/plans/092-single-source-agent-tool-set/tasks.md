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
- [X] **T006 (carrier)** `cmd/tellme/deps_agentbase_test.go` — `TestAgentToolsIsTheCanonicalBaseSet` (production base set
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
| (FR-004) the capability gate is unchanged | `TestCompositionResolvesTheFamilyAwareImageCeiling` (the gate pin — **has teeth**: drop the `vision` gate ⇒ red, measured) | drop the `vision` gate ⇒ the pin reds. The **union half** (`recordableToolNames()` keeps `read_image`) is **unwitnessed** — TD-092-1 / ADR 0032 **RF-062-12** |
| (W-2) the canonical owner is load-bearing | round-090's `TestOfferedSetDocMatchesTheLiveRegistry`; `TestNewToolRegistryOffersAgentTools`; the four `search_files` E2E Examples | remove a tool from `NewAgentBaseTools` ⇒ doc (8) vs live (7) red **and** `TestNewToolRegistryOffersAgentTools` red **and** the 4 `search_files` Examples red (fold N-092-1, measured; the two 092 carriers stay green — they pin the delegation, not the owner's content) |
| (I-2 / W-4) behaviour identity | the full unit suite + the godog E2E, **no assertion changed**; E2E counts unchanged (330 · 2487) | a set/order change ⇒ a count/behaviour pin red |
| (I-3) layer baseline 0 | `make verify-architecture` | a new import edge ⇒ a RULE-A/B/E violation |

---

## Fold ledger (architect review — PR #192)

Review verdict: **`APPROVE WITH REQUIRED FOLDS`** (no `[ARCHITECTURAL BLOCKER]`; all folds **record /
live-state accuracy**, no design or behaviour change) — [`pull/192#issuecomment-5831066497`](https://github.com/gosharplite/tellme/pull/192#issuecomment-5831066497).

| Finding | Class | Resolution |
| --- | --- | --- |
| **F-092-1** the "single production site / two call sites only" claim (spec I-1 + ADR §Verification) is falsified by `newTUIRegistry()` + the round-031 literal-eight fixture | record accuracy | Scoped both claims to the **agent base** set and named the **two non-delegating listers** (`newTUIRegistry` at `cmd/tellme/deps.go:148`; `TestNewToolRegistryOffersAgentTools`'s literal eight) — `spec.md` I-1 · ADR 0062 §Verification · (RF-092-3 cross-referenced). |
| **F-092-2** the `cmd/tellme` carrier mis-cited as `deps_test.go` | record accuracy | Corrected to **`cmd/tellme/deps_agentbase_test.go`** in `plan.md` · `tasks.md` T006 · `truth-delta.md` · `checklists/requirements.md`. |
| **F-092-3** stale owner name left in `step_r021_t026_chat_then_offered_tools.go:37` ("the same constructors `cli.newToolRegistry` uses") | record accuracy | Rewrote the comment to cite the canonical owner `infratools.NewAgentBaseTools` (round 092 / ADR 0062), mirroring the reworded `tool_usage.go`. |
| **F-092-4** stale count/enumeration on live surfaces (`techstack.md:127` "all seven tools"; `deps.go`'s `agentTools()` comment omitted `search_files`) | record accuracy | Dropped the literal "seven" from the techstack row; reconciled the `agentTools()` doc comment (names `search_files` + the canonical owner). |
| **F-092-5** the recorded round base was `dev` `e0dcea2`; measured base is `65ad377` | live-state accuracy | Corrected `STATUS.md:4-5` + day-log §12 (intro `:651`, at-a-glance `:657`). |
| **F-092-6** `STATUS.md` roadmap listed #189 as a live seed while it is the in-flight anchor | live-state accuracy | Dropped #189 from the candidates line (leaving #191). |
| **TD-092-1** FR-004's carrier cell named the union enumerator as a witness (no teeth, measured) | technical debt | Annotated the cell honestly (the gate pin has teeth; the union half is **unwitnessed** — ADR 0032 RF-062-12) in `checklists/requirements.md` + this ledger; no cheap pin added (recorded as debt). |
| **N-092-1** W-2's witness enumeration was understated | nit | Restated the full enumeration (doc carrier + `TestNewToolRegistryOffersAgentTools` + the 4 `search_files` Examples) in the checklist, the ledger, and ADR §D5/§Verification. |
| **N-092-2** the PR body's `**Closes #189.**` bolds the closing keyword | nit | PR body edited to a plain `Closes #189.` |
| **N-092-3** RF-092-2 re-records ADR 0032's RF-062-12 without citing it | nit | Added the cross-reference (round 092 *halves* RF-062-12; the union half stays unwitnessed — TD-092-1). |

