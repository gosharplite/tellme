# Tasks — Prompt suggestion parity: the newest-50 candidate pool (round 064)

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Mode**: one-shot (`/axb-implement`) — a single round, test-aligned first (RED), then GREEN.
**Pipeline position**: specify ✅ · clarify ✅ (Q1 → A) · spec-by-example ✅ · technical-research ✅ (ADR 0034) · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · implement ⏳

## Phase 0 — Setup

- [x] **T001** [SETUP] Confirm the branch `064-prompt-suggestion-parity` (off `dev` @ `4092ae7`) and the plan package present.

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research phase): `techstack.md` row MODIFY + ADR 0034 + index (done in the research phase).
- [x] **T003** [BDD-ALIGN] CLI truth: add the depth Rule/Example to `prompting-with-suggestions.feature` + the `dsl.md` row/note (done in the dsl-refine phase).
- [x] **T004** [SETUP] E2E fixture: the new `step_r064_suggestions.go` Given (`the shared prompt log holds {count} newer prompts about other topics`).

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T005** [RED][UNIT] Add `TestSuggest_RecentPromptPoolDepth` in `internal/app/suggestions/service_test.go`: a fake `PromptSource` records the requested `n`; assert the engine asks for the **deepened depth** (50) and never the shallow 10; and a many-match case still surfaces **≤10** (the cap). *Fails against today's code (asks for 10).*
- [x] **T006** [RED][E2E] The new depth Example (the shared log holds a match followed by 12 unrelated records; typing the term still offers that match). *Fails against today's code (the 10-prompt window excludes it).*

## Phase 3 — Feature

- [x] **T007** [GREEN][IMPL] `internal/app/suggestions/service.go`: split `maxSuggestions` into **`promptPoolDepth = 50`** (the history-source request) + **`maxSuggestions = 10`** (the surfaced cap); `addPrompts` requests `promptPoolDepth`. No other change (source set, ordering, subsequence, dedup, cap, empty-query behaviour, the `-i` surface).
- [x] **T008** [VERIFY] `go build ./...`, `go vet ./...`, `gofmt -l` clean; `go test -count=1 ./...` GREEN (incl. the E2E contract); `make verify` OK.

## Phase 4 — Regression & witnesses

- [x] **T009** [REGRESSION] The existing suggestion Examples stay GREEN (recent-prompt, workspace path, tool name, accept, >3-line drop) — the change is depth-only.
- [x] **T010** [WITNESS] Reproduce then revert: (a) restoring `promptPoolDepth` to **10** turns T005 + T006 **RED**; (b) the cap is non-vacuous (T005's many-match case asserts ≤10).
- [x] **T011** [RECORD] `truth-delta.md` dsl-refine rows + STATUS update; commit the round.

## Conventions

- **TD-6** — the fold ledger names fold heads + prior ledgers only (a ledger commit cannot name its own SHA).
- RED first: T005/T006 must be observed failing against the pre-change tree before T007 lands.
