# Tasks — Prompt suggestion parity: the newest-50 candidate pool (round 064)

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Mode**: one-shot (`/axb-implement`) — a single round, test-aligned first (RED), then GREEN.
**Pipeline position**: specify ✅ · clarify ✅ (Q1 → A) · spec-by-example ✅ · technical-research ✅ (ADR 0034) · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · implement ⏳

## Phase 0 — Setup

- [x] **T001** [SETUP] Confirm the branch `064-prompt-suggestion-parity` (off `dev` @ `4092ae7`) and the plan package present.

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research phase): `techstack.md` row MODIFY + ADR 0034 + index (done in the research phase).
- [x] **T003** [BDD-ALIGN] CLI truth: add the depth Rule/Example (and the cap Rule/Example) to `prompting-with-suggestions.feature` + the `dsl.md` rows/note. (Authored together with the implementation in the single round commit `6fc1ba9`; the round ran as one work session, so there is no separate `/axb-dsl-refine` commit.)
- [x] **T004** [SETUP] E2E fixtures: `step_r064_suggestions.go` (the two deep-log Givens) and `step_r064_then_no_recent_prompt.go` (the negative Then). (Same single round commit `6fc1ba9`.)

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T005** [RED][UNIT] Add `TestServiceSuggestAsksForDeepenedPoolAndCapsAtTen` in `internal/app/suggestions/service_test.go`: a fake `PromptSource` records the requested `n`; assert the engine asks for the **deepened depth** (50), that `promptPoolDepth > maxSuggestions`, and that a many-match case still surfaces **≤10** (the cap). *Fails against today's code (asks for 10 / the constants are equal).*
- [x] **T006** [RED][E2E] The depth Example (the shared log holds a match followed by 12 unrelated records; typing the term still offers that match) **and** the cap Example (20 matching prompts: the newest is offered, the eleventh is not). *The depth Example fails against today's code (the 10-prompt window excludes it).*

## Phase 3 — Feature

- [x] **T007** [GREEN][IMPL] `internal/app/suggestions/service.go`: split `maxSuggestions` into **`promptPoolDepth = 50`** (the history-source request) + **`maxSuggestions = 10`** (the surfaced cap); `addPrompts` requests `promptPoolDepth`. No other change (source set, ordering, subsequence, dedup, cap, empty-query behaviour, the `-i` surface).
- [x] **T008** [VERIFY] `go build ./...`, `go vet ./...`, `gofmt -l` clean; `go test -count=1 ./...` GREEN (incl. the E2E contract); `make verify` OK.

## Phase 4 — Regression & witnesses

- [x] **T009** [REGRESSION] The existing suggestion Examples stay GREEN (recent-prompt, workspace path, tool name, accept, >3-line drop) — the change is depth-only.
- [x] **T010** [WITNESS] Reproduce then revert: (a) restoring `promptPoolDepth` to **10** turned the unit pin T005 **RED** (`promptPoolDepth (10) must be strictly deeper than the surface cap (10)`) and the E2E depth Example **RED** (`267 scenarios (266 passed, 1 failed)`); (b) raising the **surface cap** (`maxSuggestions = 20`) turned the **E2E cap Example RED** (`The ten nearest matches are offered and the surplus is dropped`, `267 passed, 1 failed`) while the **unit pin stayed GREEN** — its `len(got) == maxSuggestions` clause is self-referential, so the E2E cap Example is the value-binding cap carrier (review R-4).
- [x] **T011** [RECORD] `truth-delta.md` dsl-refine rows + STATUS update; commit the round.

## Conventions

- **TD-6** — the fold ledger names fold heads + prior ledgers only (a ledger commit cannot name its own SHA).
- RED first: T005/T006 must be observed failing against the pre-change tree before T007 lands.
