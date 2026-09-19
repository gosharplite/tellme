# Tasks — round 058 (`058-grey-tool-output-content`)

**Inputs**: `spec.md` (US1) · `research.md` (D1–D9) · `plan.md` (1 CLI end; api/data NOOP) · `truth-delta.md`.
**Setup**: omitted (stdlib-only).

## Phase 1 — Truth alignment

- [X] **T001** — `specs/truth/techstack.md` (MODIFY ×2 rows) — owner `/axb-technical-research`.
- [X] **T002** — `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (the Rule text + comment) + `chat/dsl.md` (the `the tool output frame is shown in grey` row) — owner `/axb-dsl-refine`.
- [X] **T003** — **ADR 0028** + the `docs/decisions/README.md` index row.

## Phase 2 — Implementation (red-first where a carrier exists)

- [X] **T004** [BDD] — `internal/ui/colour_test.go`: invert the round-057 "content stays plain" pin to "content is grey"; assert the plain (colour-off) block is byte-identical.
- [X] **T005** [CODE] — `internal/ui/tooloutput.go`: `formatToolOutputLineColour`; `WriteWith` calls it with `w.Colour`.
- [X] **T006** [BDD] — `tests/e2e/steps/step_r057_payload_and_colour.go`: `thenToolOutputGrey` also requires ≥2 grey `[Tool Output] …` lines (header + ≥1 content).

## Phase 3 — Verification & regression

- [X] **T007** [REGRESSION] — `gofmt` · `go vet ./...` · `make verify` · `go test -count=1 ./...` (all Examples).
- [X] **T008** [WITNESS] — falsifiability witness: drop the content wrap ⇒ the unit pin + the grey Example red. Reproduce then revert.
- [X] **T009** — `truth-delta.md` owner rows + STATUS + the round branch/PR.

**Pre-Delivery Orphan Sweep**: 0 orphans (the modified sentence keeps exactly one `DSLRow`; no feature removed).

## Fold ledger - PR #124 review (5254692990, APPROVE WITH REQUIRED FOLDS - no blocker)

| # | Finding | Fold |
| --- | --- | --- |
| **F-058-1** (required) | The "block is plain" negative was **vacuous** - the only no-colour carrier used a tool-less provider (no block). | Added the **anti-vacuity plain-block Example** to `colouring-the-session-chrome.feature`'s plain Rule, built from two **existing** sentences - **no new `DSLRow`/stepdef**. Witness: leaking the colour gate now reds **two** scenarios (was one). |
| **F-058-2** (required) | ADR 0028 claimed a cross-link to ADR 0027 that did not exist. | ADR 0027's **`Status`** annotates the A3 supersession (0005-D7 precedent) + its **index row**; ADR 0028's claim corrected. |
| **TD-058-1** | The `Colour` field doc (+ `colour.go` package + `colorGray` docs) still said content stays plain. | All three corrected to the whole-block scope. |
| **TD-058-2** | `reGreyToolOutputPrefix` was byte-identical to `reGreyHeader`; `count >= 2` never proved a **content** line. | Replaced with `greyToolOutputContentLineCount` (excludes the header text; Go RE2 has no lookahead) + a `>= 1 grey CONTENT line` assertion; duplicate regex removed. |
| **R-058-2** | The round-058 assertion lived in a round-057-named test. | Added **`TestChromeColourRound058`**; restored the round-057 subtest to its scope. |
| *(review note)* | The round-038 sanitize predicate is now **scope-blind**. | Recorded as **RF-058-4** in ADR 0028 section Forward. |
