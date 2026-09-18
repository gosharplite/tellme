# Specification Quality Checklist: composition-root extraction (round 044)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/044-composition-root-extraction`

**Spec Path**: `specs/plans/044-composition-root-extraction/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (the exact `Dependencies` fields, the `MCPDiscoverer` signature, the `cmd/tellme` layout are explicit RD decisions — `research.md` D-x, not FRs)
- [x] Edge cases cover the main high-risk situations (tier-2 `app/deps` ceiling, untested `cmd/**`, assembler-gate relocation, offline `--tool-usage` path, `-i` TUI path, global-mutating tests, MCP confinement, the forced tier split of the seams)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (`internal/cli` stops being the composition root — the gate proves it) → US2 (the injected domain-typed seam; globals gone) → US3 (zero behavioural change)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-011, FR-013, NFR-005, NFR-006)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` — **seven decisions asked one at a time** (the standard cap was exceeded deliberately: this is the split's *mechanical big one* and each answer constrained the next)
- [x] No remaining `NEEDS CLARIFICATION` — round-1 answers **locked**:
  - **Q1 → (A) `cmd/tellme`** is the composition-root home (exempt from the R1 tier table).
  - **Q2 → (β)** a new **`internal/app/deps`** package holds the domain-typed `Dependencies` struct.
  - **Q3 → (1)** relocate `agentTools()`; inject `NewToolRegistry`/`BindToolOutput`/`BindSkillsCatalog`; move `ToolOutputSink` → `domaintools.OutputSink`.
  - **Q4 → (a)** move MCP discovery into `internal/infrastructure/mcp`; inject a func-typed `MCPDiscoverer`.
  - **Q5 →** R2's DoD is the **7 RULE-B edges → 0**; the strict form is follow-up **[#101](https://github.com/gosharplite/tellme/issues/101)** (R5).
  - **Q6 → (T1)** one `cli.Options` (`deps.Dependencies` + the two `ui` seams); delete every factory var.
  - **Q7 → (a)** the strict-scope follow-up lives as a child of [#92](https://github.com/gosharplite/tellme/issues/92) (**#101**).
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (mechanism is RD A2; ADR A3; `/axb-dsl-refine` NOOP A5)

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (gate green with the 7 entries removed; behaviour unchanged; falsifiability witness red→green)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-007)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Q2↔Q6 coupling recorded**: because `internal/app/deps` is tier 2 and **cannot** import `ui` (tier 5), the **one** `ui`-typed presentation seam (`RunTUIPrompt`, over unexported `resolution`/`runtimeEnv`) lives in a cli-local `cli.Options`, **not** in `Dependencies`; the `newRenderer` var is deleted (inlined). This is *forced by the R1 tier table*, not a preference — recorded in the spec Edge Cases.
- **Grill-round fold (PR #102; operator G1–G4 + fixes 1–8)**: the migration surface was under-recorded (TUI runner + 3-reader `NewTUIRegistry`, renderer, `turn_test.go`, `prompt_multiline_test.go`, `newTurnSpinner`/`augmentRegistryWithMCP`, `UserHomeDir`, `run` callers) and the pinned `OutputSink` literal omitted `Enabled()` — **the 7 → 0 DoD was formally unreachable until this fold**. The `cmd/tellme` test takes **no stream assertions**; no `internal/cli` test file may import `internal/infrastructure/*`. Recorded in `spec.md` *Grill-round fold* + the amended FRs, and `research.md` D2/D3/D5/D11/D12b.
- **Q5 scope boundary made explicit**: RULE-B is an *upward*-import rule; it does **not** constrain a downward import, so after R2 `internal/cli` still legally imports `agent`/`ui`/`config`/`home`/`app`. The stricter #92 AC2 clause is parked on **#101** (needs a new gate rule + ADR).
- **Assembler-gate relocation**: #100's constraint is the `agentTools()` *property* (parameterless, read-free, non-overridable), not its *file*; the round-031 `TestAgentToolSchemasAreWellFormed` moves with the assembler to `cmd/tellme`. Recorded as Assumption A7.
- **Truth-current is an obligation**: the techstack rows naming `agentTools()`, `verify-mcp-sdk-confinement`, and the MCP discovery/normalization/credential seams are real **MODIFY**s, not NOOPs (FR-011 / the Truth obligations table).
- **Ratchet discipline**: the baseline is regenerated **in the same commit** as each removed edge (FR-008), so `dev` is green at every commit.

## Ready determination

- [x] Ready to proceed to downstream planning
- [ ] A high-impact requirement gap must be closed first — **none remaining** (Q1–Q7 locked)

**Note**: plan half — this branch runs `/axb-specify` → `/axb-spec-by-example` (expected NOOP) → `/axb-technical-research` → `/axb-system-analysis`. `/axb-tasks` + `/axb-implement` are a later branch (the repo convention: the `/axb-tasks` half opens on a fresh branch off `dev` after the plan half merges).
