# Tasks — 082-listing-backward-turn-indices

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`
**Spec**: [`spec.md`](spec.md) · **Plan**: [`plan.md`](plan.md) · **Research**: [`research.md`](research.md)

> One task per unit of work; a task is `[X]` only after its verification passes. The test-alignment layer precedes the feature work; refactor happens under green.

## Phase 1 — Setup & Foundational

- [X] **T001** — Add `TurnIndex int` to `render.ListingMessage` (`internal/domain/render/ports.go`) with a doc comment naming the backward-index contract (round 082; ADR 0054) — the filed shape the CLI fills and the adapter reads.
- [X] **T002** — Stamp the true turn distance in `internal/cli/cli.go` `listingMessages`: compute `len(entries) - i` per entry (i = 0-based entry index) and set it on **both** messages of that entry, **before** the message-count truncation.
- [X] **T003** — Update `internal/ui/listing.go` `header` to format `[USER] - N` / `[MODEL] - N` from `m.TurnIndex` (ASCII hyphen, single spaces), wrap the **whole** label in the existing `blue`/`magenta`, and fall back to the bare `[USER]` / `[MODEL]` label when `TurnIndex <= 0`.

## Phase 2 — Test Alignment & Implementation (Red → Green)

- [X] **T004** — `internal/ui/listing_test.go`: pin the suffixed header text (`[USER] - 2` / `[MODEL] - 2`), the whole-label colour unit (`\033[1;34m[USER] - 2\033[0m`), the plain (colour-off) form, and the `TurnIndex <= 0` bare-label fallback. **Red** before T003.
- [X] **T005** — `internal/cli/history_projection_test.go`: pin `listingMessages` — same index on both messages of an entry; the **true** distance on a partial slice (the odd-`-l` leading message keeps `- 2`, not `- 1`). **Red** before T002.
- [X] **T006** — `tests/e2e/steps/step_r073_listing.go`: update `listingBlocks` to recognise the suffixed header (`[USER] - N` / `[MODEL] - N`, role = prefix) and keep the existing stepdefs green; update `thenListingAccentsRoles` to assert the whole-label wrap (`\x1b[1;34m[USER] - ` / `\x1b[1;35m[MODEL] - `) and `thenListingCarriesNoAccents` to assert the plain suffixed label; update `thenMessagesSeparatedByBlankLine` header matching.
- [X] **T007** — `tests/e2e/steps/step_r073_listing.go`: implement the new Then `the listing heads each message with its backward turn index` — recompute the expected `[USER] - K` / `[MODEL] - K` sequence from the arranged exchanges + the `-l N` request (last N messages) with `K = len(entries) - i`, and compare it to the listed header sequence (an anti-vacuity check that the expected length comes from the request).
- [X] **T008** — Run the round-082 E2E Examples for the new Rule (the full listing, the partial listing, the tool-using turn, the terminal accent) — **Green**.

## Phase 3 — Truth & Records

- [X] **T009** — `specs/truth/features/cli/history/inspecting-the-session-history.feature`: the round-082 Rule + 4 Examples (done at `/axb-dsl-refine`).
- [X] **T010** — `specs/truth/features/cli/history/dsl.md`: the round-082 note + the new `backward turn index` Then row + the qualified accent / role-header rows (done at `/axb-dsl-refine`).
- [X] **T011** — `specs/truth/techstack.md` *Session lifecycle flags* MODIFY (done at `/axb-technical-research`).
- [X] **T012** — **ADR 0054** (`docs/decisions/0054-list-backward-turn-indices.md` + index) **amends ADR 0045** (done at `/axb-technical-research`).
- [X] **T013** — `docs/domain-model/**` **NOT modelled** (recorded in `plan.md` §5; `modelith-check` stays green).

## Phase 4 — Gates

- [X] **T014** — `gofmt -l .` + `goimports -l .` clean; `go vet ./...` clean.
- [X] **T015** — `make verify` **OK** (layer gate 0 · `modelith-check` ×3 no drift · `verify-fmt` · `verify-adr-index` · lint 0 · govulncheck clean · cross-compile).
- [X] **T016** — `go test -count=1 ./...` **green** (E2E incl. the new Examples); `go.mod`/`go.sum` unchanged.

## Fold ledger

**PR #166 architectural review (`architect` peer, head `76474d9`) — `APPROVE WITH REQUIRED FOLDS`** (no `[ARCHITECTURAL BLOCKER]`; the design was right, the folds were record/evidence quality). Fold commit: `FOLD-082`.

| # | Finding | Fold |
| --- | --- | --- |
| **F-082-1** | ADR 0045 has no amendment back-pointer (the amended ADR's index row read plain `Accepted`) | `docs/decisions/README.md` 0045 row → `Accepted (**amended by 0054** …)`; `0045-*.md` `Status` line → `Accepted (round 073) — amended by ADR 0054 …` (body untouched, immutable once Accepted) |
| **F-082-2** | `spec.md` SC-002 (`tellme -l 1` ⇒ only `[MODEL] - 1`) had no carrier | appended `Then the listing heads each message with its backward turn index` to the *Listing with no count* Example (the SC-002 carrier; `arrangedListCount` already yields 1) |
| **F-082-3a** | `tasks.md` T005 named the wrong test file | T005 → `internal/cli/history_projection_test.go` |
| **F-082-3b** | the stated W1 understated the measured effect | recorded below |

**Measured witnesses** (independently reproduced by the reviewer on a scratch copy, and by the round):

| # | Mutant | Unit pins red | E2E Examples red |
| --- | --- | --- | --- |
| **W1** | drop the `TurnIndex` stamp | 2 | **8** — the 4 new Examples **+** the 4 round-073 Examples the round re-pointed to the suffixed label (the round's original "3" understated) |
| **W2** | forward numbering (`i+1`) | 2 | 2 (incl. the partial-listing Example) |
| **W3** (reviewer) | renumber from the **truncated slice** (the alternative ADR 0054 D2 rejects) | **exactly 1** | **exactly 1** (the partial listing) — the true-distance claim is precisely falsifiable |
| **C** (reviewer) | accent only the role **word** (invariant I-4) | 1 | 2 (both terminal-accent Examples) |

**Nits / TD (recorded, not required):** **N-082-1** `internal/ui/colour.go` round-073 comment names the `[USER]` header line (now a whole-label unit) — folded (`custom` comment now names the whole label incl. the turn index); **N-082-2** the `chrome-colour-terminal-gated` invariant names the `-l` role-header accent — folded (a parenthetical naming the backward turn index; `docs/domain-model` re-rendered, `modelith-check` green); **TD-082-1** `len(entries)-i` re-derived in 3 places (optional helper); **TD-082-2** `fakeListing` prints the bare role (fidelity); **TD-082-3** the `TurnIndex <= 0` fallback is unreachable in shipped wiring (defensive; RF-082-2). **O-082-1 (pre-existing, unrelated)** — the reviewer observed `chat/presenting-the-interactive-prompt.feature:53` fail once in a full parallel run, then pass on re-run and 5/5 in isolation; recorded here, and raised as a live issue if it recurs (not a round-082 defect).
