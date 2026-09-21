# Tasks — skill frontmatter block scalars (round 075)

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [X] **T001** Confirm the surfaces: `internal/infrastructure/skills/loader.go` `parseFrontmatter`; `renderSkills` (`internal/infrastructure/tools/skills.go`, unchanged); the `chat` truth feature + DSL (round 033 + round 075); the E2E helpers (`wire_skills.go`, `scenarioFrom`/`onlyFake`/`lastToolResult`); the reference's `parseSkill` (line-based — divergence grounding).

## Phase 2 — Foundational

- [X] **T002** `internal/infrastructure/skills/loader.go`: add block-scalar helpers (`isBlockIndicator`, `resolveBlockScalar`) and route `name`/`description` through a value resolver. *Only* the reader; no tool/rendering/domain-type change.

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (RED)** `internal/infrastructure/skills/loader_test.go`: `TestLoadFoldedBlockScalarDescription` · `TestLoadLiteralAndChompingBlockScalars` · `TestLoadInlineGreaterThanIsNotABlock` · `TestLoadBlockScalarWithNoBodyIsSkipped` · `TestLoadBlockScalarName`.
- [X] **T004 (RED)** `tests/e2e/steps/step_r075_skills_given_folded.go` + `step_r075_skills_then_described.go`: the folded-authoring Given and the described-text Then.
- [X] **T005 (RED)** Truth feature `specs/truth/features/cli/chat/listing-the-available-skills.feature` (+1 Rule/Example) and the `chat/dsl.md` round-075 Given/Then rows.
- [X] **T006 (GREEN)** `parseFrontmatter` resolves a block-scalar value: indicator `[>|][+-]?`; collect the blank/indented block; fold (`>`) / literal (`|`); chomping clip/strip/keep; `TrimSpace`. Inline values unchanged; no-body indicator ⇒ empty ⇒ skip.

## Phase 4 — Feature (Green / Refactor)

- [X] **T007 (witness — reproduced, then reverted)** Drop the block-scalar resolution (restore the single-line read): the unit pins redden **and** the E2E `A description written as a folded block scalar` reddens at `the listing describes the skill "golang-patterns" as "Idiomatic Go patterns for robust code"` (the result shows `- golang-patterns: > (`).
- [X] **T008** Truth: `specs/truth/techstack.md` §Skills *Skills catalog (load)* row MODIFY (block-scalar support + divergence note); **ADR 0047** + index.
- [X] **T009** Domain model: **not modelled** — the round changes frontmatter *parsing*, not a modelled entity/invariant; recorded in `plan.md` §5 (ADR 0041's escape hatch).

## Phase 5 — Delivery

- [X] **T010** `make verify` + `go test -count=1 ./...` green; `go.mod`/`go.sum` unchanged; the topology audit adds no new error.
