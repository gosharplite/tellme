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
