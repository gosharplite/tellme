# Tasks — round 086 `086-tools-listing-line`

Task list for `/axb-implement`. One task per batch; mark `[X]` only after the task's `Test Scope`
passes. Round 086 has **no unobservable claim**, so there is **no Phase 4W `[WITNESS]` lane** — every
clause is carried by an E2E (`[BDD-GREEN]`) or a unit pin (see the Claim→Witness ledger).

## Phase 1 — Setup

- No new technology: the round reuses the existing glamour renderer, the `yellow()` colour helper,
  and the `render.Listing` port. **No Setup task** (nothing to install, no smoke test).

## Phase 2 — Foundational

- No shared component/entry/fixture is introduced beyond the E2E fixture field in T002. **No
  Foundational task.**

## Phase 3 — Test Alignment & Implementation

- [ ] T001 [BDD-ALIGN] Add the round-086 E2E step definitions.
  - Read: `specs/truth/features/cli/history/inspecting-the-session-history.feature` → the
    `## Then (round 086)` rows; `tests/e2e/steps/step_r073_listing.go` (the existing listing parse)
  - Do: add `tests/e2e/steps/step_r086_tools.go` — a turn-block parse helper (`[USER]`/`[TOOLS]`/
    `[MODEL]`) and the Thens `the listing reports each turn's tool activity` + `the listing accents
    the tool-activity line in yellow`.
  - Test Scope: `go test -count=1 -run TestE2E ./tests/e2e` — the new Examples resolve (no
    "undefined step") and **redden** at the new Thens (the product emits no `[TOOLS]` line yet).
- [ ] T002 [BDD-ALIGN] Record the arranged tool count in the E2E fixture.
  - Read: `tests/e2e/steps/step_t022_history_given_tool_using_exchange.go`,
    `tests/e2e/steps/scenario_context.go` (`exchange`/`recordExchange`)
  - Do: add `toolCalls int` to `exchange`; set it to `1` in the tool-using Given (0 elsewhere). The
    Then derives the expected count from the fixture, so it is independent of the observed output.
  - Test Scope: `go test -count=1 -run TestE2E ./tests/e2e`
- [ ] T003 [BDD-REMOVE] Retire the superseded "operator only" listing step.
  - Read: `tests/e2e/steps/step_t023_history_then_operator_only.go`;
    `specs/truth/features/cli/history/dsl.md` `Then` row `tellme lists only the operator's messages`
  - Do: replace the step with `tellme lists the tools' activity but not their contents` (assert the
    count line is present **and** no tool content — tool name, arguments, result — is surfaced).
  - Test Scope: `go test -count=1 -run TestE2E ./tests/e2e` (the reframed Example passes once T005
    lands; the tool-content negative holds immediately).

## Phase 4 — Feature phases

- [ ] T004 [BDD-GREEN] Emit the `[TOOLS] - M (N calls)` line from the listing.
  - Dependencies: T001, T002, T003
  - Read: `specs/truth/features/cli/history/dsl.md` → `## Then (round 086)`; `internal/domain/render/ports.go`
    (`Listing`/`ListingMessage`); `internal/ui/listing.go`; `internal/cli/cli.go` (`listingMessages`)
  - Do: add `ToolCount int` to `ListingMessage` (doc: the turn's tool-step count; the `TurnIndex`
    precedent); in `listing.go` emit — for each **model** message — the tool line immediately **before**
    the `[MODEL]` header, as its own block (a blank line before and after), yellow-wrapped under
    `spec.Colour && !spec.Raw`, else plain (`[TOOLS] (N calls)` when the turn index is ≤ 0);
    in `cli.go` set `ToolCount: len(e.Steps)` on the turn's model message.
  - Test Scope: `internal/ui/listing_test.go` · `internal/cli/list_render_test.go` ·
    `go test -count=1 -run TestE2E ./tests/e2e`
- [ ] T005 [BDD-GREEN] Update the unit byte pins + the CLI wiring pin.
  - Dependencies: T004
  - Read: `internal/ui/listing_test.go`; `internal/cli/list_render_test.go`
  - Do: update the exact-byte expectations for the new line; add pins for: the line placement
    (between `[USER]` and `[MODEL]`), `(0 calls)` vs `(N calls)`, the bare `[TOOLS] (N calls)`
    fallback (non-positive index), the whole-label yellow, and `-r`/non-terminal plainness; add the
    CLI wiring pin that `ToolCount = len(Steps)` and **not** `Calls` (a fixture whose `Calls ≠ len(Steps)`).
  - Test Scope: `go test ./internal/ui/ ./internal/cli/`
- [ ] T006 [BDD-REFACTOR] Tidy under green.
  - Dependencies: T005
  - Do: reconcile comments/docs (`colour.go` notes the listing `[TOOLS]` accent reuses `yellow`;
    the `Listing` port doc names the line); no behaviour change.
  - Test Scope: `go test ./internal/ui/ ./internal/cli/ && go test -count=1 -run TestE2E ./tests/e2e`
- [ ] T007 [BDD-GREEN] Truth + records.
  - Dependencies: T006
  - Do: `specs/truth/techstack.md` (*Output rendering* + *Session lifecycle flags* notes);
    `docs/domain-model/tellme.modelith.yaml` (the listing colour invariant) + `make modelith-render`;
    `docs/decisions/0057-tools-listing-line.md` + the index row (already drafted — verify).
  - Test Scope: `make modelith-check` · `make verify-adr-index`
- [ ] T008 [REGRESSION] Full gate + topology audit.
  - Dependencies: T007
  - Do: run the round gates; re-run the advisory topology audit; confirm `go.mod`/`go.sum` unchanged.
  - Test Scope: `make check` · `python3 …/audit_feature_dsl_topology.py --root specs/truth/features/cli`

## Pre-Delivery 盤點與覆蓋對照

### 1. Pre-Delivery Orphan Coverage Sweep (孤立產物盤點)

- `truth-delta.md` non-NOOP items → `specs/truth/features/cli/history/**` (T001/T003/T004),
  `specs/truth/techstack.md` (T007).
- `research.md` decided Decisions (D1–D8) → all carried by T004/T005/T007 or a truth row.
- `specs/truth/techstack.md` changed sections → T007.
- Isolated artifacts: **0**.

### 2. Claim→Witness 盤點對照表 (Claim→Witness Ledger)

| Claim ID | Source | Atomic Effect Claim | Witness type | Bound task / carrier | Discriminating falsifier | Status |
| --- | --- | --- | --- | --- | --- | --- |
| CLM-001 | FR-001 | a listed turn shows `[TOOLS] - M (N calls)` between its `[USER]` and `[MODEL]` | `[BDD-GREEN]` | T004 · E2E `the listing reports each turn's tool activity` | move the emit after the `[MODEL]` header (or drop it) ⇒ E2E red | pending |
| CLM-002 | FR-002 | the line is printed for every turn incl. `(0 calls)` | `[BDD-GREEN]` | T004 · E2E `A turn without tools reports zero` | emit only when `N > 0` ⇒ E2E red | pending |
| CLM-003 | FR-003 | the line is its own block (a blank line above and below) | `[BDD-GREEN]` | T005 · unit byte pin | drop a blank ⇒ byte pin red | pending |
| CLM-004 | FR-004 | the whole label is yellow on a terminal stdout | `[BDD-GREEN]` | T004/T005 · E2E `the listing accents the tool-activity line in yellow` + unit pin | un-gate the yellow ⇒ E2E/unit red | pending |
| CLM-005 | FR-005 | plain under `-r` / a redirected stdout | `[BDD-GREEN]` | T005 · unit pin `TestListingToolLineSuppressedUnderRawOnATerminal` (the `-r` × terminal case; fold F-086-2) + E2E `A raw listing on a terminal …` / `the listing carries no accents` | gate the tool line on `spec.Colour` only (drop `!spec.Raw`) ⇒ the unit pin + the E2E Example red | pending |
| CLM-006 | FR-006 | `-l N` still selects the last N messages (the line is a rider) | `[BDD-GREEN]` | T004 · the existing `tellme lists the last {count} messages` Then (unchanged) | count the line as a message ⇒ the Then red | pending |
| CLM-007 | FR-007 | only the count is surfaced — no tool content | `[BDD-GREEN]` | T003 · E2E `tellme lists the tools' activity but not their contents` | emit the tool name/args/result ⇒ red | pending |
| CLM-008 | FR-008 | the line is not emitted by `-t`/`turns.log` nor the chrome | `[BDD-GREEN]` | T004 · the existing `-t` Thens (`tellme prints exactly the turn log line …`) | write the line into `turns.log` ⇒ red | pending |
| CLM-009 | FR-009 | the count is `len(Entry.Steps)`, not `Entry.Calls` | `[BDD-GREEN]` | T005 · the CLI wiring pin (`Calls ≠ len(Steps)` fixture) + E2E (the arranged turn has `calls: 2`, one step) | report `Calls` ⇒ `(2 calls)` ≠ `(1 calls)` ⇒ red | pending |
| CLM-010 | NFR-001 | the listing stays offline | `[BDD-GREEN]` | T004 · the existing `tellme sends no request to any provider` Then | add a provider call ⇒ red | pending |
| CLM-011 | NFR-002 | no new dependency | `[BDD-GREEN]` | T008 · `git diff --exit-code go.mod go.sum` | add an import ⇒ `go.mod` diff ⇒ red | pending |
| CLM-012 | EC-001 | a non-positive `M` prints the bare `[TOOLS] (N calls)` | unit pin | T005 | print ` - 0` ⇒ unit pin red | pending |
| CLM-013 | EC-002 | a partial listing still shows the turn's tool line | `[BDD-GREEN]` | T004 · E2E `A partial listing still reports the answer's tool activity` | emit the tool line for the `[USER]` message instead of the `[MODEL]` message ⇒ measured **15** E2E scenarios red incl. the partial-listing Example, + the unit byte pins (TD-086-1) | pending |
| CLM-014 | EC-003 | a legacy line with no `steps` ⇒ `(0 calls)` | `[BDD-GREEN]` | T004 · E2E `A turn without tools reports zero` | error on a missing `steps` ⇒ red | pending |
| CLM-015 | SC-003 | gates green; `go.mod`/`go.sum` unchanged | `[BDD-GREEN]` | T008 · `make check` | — | pending |

**Unwitnessed and unapproved claims: 0.** (No `accepted-unwitnessed` record is required; no Phase 4W.)
