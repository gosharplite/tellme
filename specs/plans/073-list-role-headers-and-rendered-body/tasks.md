# Tasks — the `-l` listing presentation (round 073)

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [X] **T001** Confirm the surfaces: `internal/cli/cli.go` `renderHistoryList` (the flat `role: content` line), `toMessages`, `dispatchReporting`, `runtimeEnv`; `internal/ui/colour.go` (`wrap`, the palette); `internal/domain/render/ports.go` (the `Lines`/`Answer` family); `cmd/tellme/deps.go` (`NewAnswer`/`NewLines`).
- [X] **T002** Confirm glamour reuse: `internal/ui/renderer.go` `NewRenderer().Render(markdown, width)` (style/sanitizer/word-wrap/degraded fallback) is the one rendering policy.

## Phase 2 — Foundational

- [X] **T003** `internal/ui/colour.go`: `colorBlue` (`\033[1;34m`) / `colorMagenta` (`\033[1;35m`) + `blue(s, on)` / `magenta(s, on)` wrappers (via the existing `wrap`).

## Phase 3 — Test Alignment & Implementation

- [X] **T004 (RED)** `internal/ui/listing_test.go`: `TestListingReportsRoleHeadersAndSeparators` · `TestListingSeparatesEveryMessageOnce` · `TestListingRendersOnlyTheModelBody` · `TestListingColourGate` · `TestListingRawSuppressesColour`.
- [X] **T005 (RED)** `internal/cli/list_render_test.go`: `TestRenderHistoryListProjectsAndGates` (the role projection, the truncation, the stdout-gated colour, the raw flag, the warn sink) · `TestRenderHistoryListRawAndNoTerminal` · `TestRenderHistoryListWidthIsBestEffort` (an unreadable config ⇒ width 0, still Success; a readable `WRAP_WIDTH` honoured). `dispatch_test.go` updated to the new `-l` bytes + the factory argument; `history_projection_test.go` re-pointed at `listingMessages`.
- [X] **T006 (RED)** `internal/cli/list_render_test.go`: `TestStdoutTerminalSeam` (`TELL_ME_FORCE_STDOUT_TTY` forces the probe; nil ⇒ not-a-terminal; an injected probe is consulted).
- [X] **T007 (GREEN)** `internal/domain/render/ports.go`: `ListingRole{ListingOperator,ListingModel}`, `ListingMessage{Role,Body}`, `ListingSpec{Colour,Raw,Width,Warn}`, the `Listing` interface.
- [X] **T008 (GREEN)** `internal/ui/listing.go`: `NewListing()` — header (`[USER]`/`[MODEL]`, accented only when colour), the body (the model body via `NewRenderer` unless raw; the operator body verbatim), the blank separator.
- [X] **T009 (GREEN)** `internal/app/deps/deps.go` `NewListing func() render.Listing`; `cmd/tellme/deps.go` binds `ui.NewListing()`; `internal/cli/testdeps_test.go` gains `fakeListing`.
- [X] **T010 (GREEN)** `internal/cli/cli.go`: `runtimeEnv.stdoutTTY` + `stdoutIsTerminal()`; `stdoutTerminalDetector()` (the `TELL_ME_FORCE_STDOUT_TTY` seam, wired in `Run`); `renderHistoryList` rewired (the port, the stdout-gated colour, `-r`, the best-effort width); `listingMessages` (the successor of `toMessages`, which is deleted); `dispatchReporting` threaded; the `-r` help text names the listing.
- [X] **T011 (RED)** `tests/e2e/steps/step_r073_listing.go`: the Given `the output is shown at a terminal`, the When `… as raw output`, the seven Thens, and the `listingBlocks`/`stripANSI` helpers.
- [X] **T012 (RED)** The legacy listing Thens re-pointed: `step_t015_history_then_lists_last.go` (`tellme lists the last N messages`), `step_t023_history_then_operator_only.go`, `step_r053_history.go` (`tellme lists the assistant message "…"`).
- [X] **T013 (GREEN)** `go test -count=1 ./...` green (the round-073 Examples + the re-pointed `-l` Examples + the unit pins).

## Phase 4 — Feature (Green / Refactor)

- [X] **T014 (witness — reproduced, then reverted)** (a) drop the separator ⇒ `TestListingReportsRoleHeadersAndSeparators` RED **and** the E2E `Listing the last two messages` (`tellme lists the last 2 messages`) RED; (b) render the operator body ⇒ `TestListingRendersOnlyTheModelBody` RED **and** the E2E `A listed prompt is echoed verbatim, never reformatted` (`the listed operator prompt is shown verbatim`) RED; (c) un-gate the colour ⇒ `TestListingColourGate` RED **and** the E2E `… the listing carries no accents` RED.
- [X] **T015 (witness — reproduced, then reverted)** ignore `-r` in the listing ⇒ `TestListingReportsRoleHeadersAndSeparators`/`TestListingRawSuppressesColour` RED **and** the E2E `The raw listing shows the answer's source` (`the listed model answer is shown as its raw source`) RED.
- [X] **T016** Truth: `inspecting-the-session-history.feature` (+5 Rules) + `history/dsl.md` (+9 rows, 3 rewritten) at dsl-refine; `specs/truth/techstack.md` ×4 + the domain model at research.
- [X] **T017 (refactor)** `renderHistoryList` stays small (the mapping/width helpers extracted); one rendering policy (the answer path's renderer reused).

## Phase 5 — Delivery

- [X] **T018** `make verify` OK (layer 0 · modelith-check ×3 · lint 0 · govulncheck clean) · `go test -count=1 ./...` green · `go.mod`/`go.sum` unchanged · topology audit **the same 5 pre-existing errors, none new** (50 features · 16 root + 404 module rows · 2071 steps) · `turns.log` untouched by the listing (no chrome, no colour).

## Notes

- **Accent-gate scope (research D3 + the renderer-style nuance).** The gated accent is the **header** SGR pair only. The rendered model body carries glamour's own style output into a pipe — exactly as the answer path does (round 006: rendering is gated by `-r` alone) — so the "no accents" rows assert *no `\033[1;34m`/`\033[1;35m`*, and the `-r` row asserts *no `\033` at all*. Recorded in `research.md` D6/D9 + the truth rows.
- **The separator Example carries 2 messages.** `-l 1` legitimately lists a single message, so the separator Then requires ≥1 header (a separator is asserted *between* messages) — the `-l` default-1 Example is the pin.
