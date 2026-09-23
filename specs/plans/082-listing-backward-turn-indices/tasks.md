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
- [X] **T005** — `internal/cli/list_render_test.go`: pin `listingMessages` — same index on both messages of an entry; the **true** distance on a partial slice (the odd-`-l` leading message keeps `- 2`, not `- 1`). **Red** before T002.
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

_(reserved — review folds land here.)_
