# Tasks: round 059 — `059-darwin-metrics-and-1hz-cadence`

**Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/**`
**Tests gate**: `gofmt` + `go vet` + `make verify` + `go test -count=1 ./...` (the E2E is Strict).

## Phase 1 — Setup

- [X] **T001** Add `golang.org/x/sys` as a direct dependency (`go mod tidy`; it is already in the module graph at v0.41.0).

## Phase 2 — Foundational

- [X] **T002** Add the build-tag-free math helpers `internal/infrastructure/telemetry/system_metrics_math.go` (`agentCPUPercent`, `darwinMemPercentCGO`, `darwinMemPercentNoCGO`, `clampPercent`, `darwinMemoryUsageFactor`).

## Phase 3 — Test Alignment & Implementation

### US1 — real macOS CPU/MEM (P1)

- [X] **T003** [RED] `system_metrics_math_test.go` — pin the agent-CPU + both darwin memory formulas (exact values).
- [X] **T004** [RED] `system_metrics_darwin_sample_test.go` (`//go:build darwin`) — `Sample()` reports a non-zero memory figure (fails on round-058 code).
- [X] **T005** [GREEN] Replace `system_metrics_darwin.go`'s stub/decoder with the shared page-size helper; add `system_metrics_darwin_cgo.go` (Mach `host_statistics64` CPU + `HOST_VM_INFO64` memory) and `system_metrics_darwin_nocgo.go` (`runtime/metrics` CPU + `sysctl` memory); use `golang.org/x/sys/unix`.
- [X] **T006** [GREEN] Strengthen the E2E resource stepdef (`tests/e2e/steps/step_t010_chat_then_reports_resources.go`) to require a non-zero MEM figure.
- [X] **T007** [REFACTOR] Confirm the linux adapter is byte-unchanged; `gofmt`/`go vet` clean; witness (a) reproduced + reverted.

### US2 — the 1 Hz resource-sample cadence (P2)

- [X] **T008** [RED] `internal/ui/spinner_sample_throttle_test.go` — one `Sample` per second across 200 ms frames; nil provider → zeros.
- [X] **T009** [GREEN] Add `ResourceSampleInterval` + the `sampleResources` throttle in `internal/ui/spinner.go`; route `renderLocked` through it.
- [X] **T010** [REFACTOR] `gofmt`/`go vet` clean; witness (c) reproduced + reverted.

## Phase 4 — Truth, governance, verification

- [X] **T011** `specs/truth/techstack.md` **System metrics provider** row (the darwin split, `x/sys`, the reference-exact MEM, the 1 Hz cadence, the recorded narrowing).
- [X] **T012** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`): the resource Rule's round-059 comment + the resource row's `必查`.
- [X] **T013** **ADR 0029** (+ the decisions index row).
- [X] **T014** Truth-delta owner rows finalized (technical-research MODIFY; api/data NOOP; dsl-refine MODIFY; governance ADD).
- [X] **T015** Verify: `gofmt` clean · `go vet ./...` clean · `make verify` **OK** (incl. the `CGO_ENABLED=0` cross-compile matrix over the darwin cgo-less leg) · `go test -count=1 ./...` green (E2E Strict — **248 scenarios · 1836 steps**, 0 undefined) · diff secret scan clean · `go.mod` change = `golang.org/x/sys` indirect→direct only (`go.sum` unchanged).

## Fold ledger

| Fold | Commit | Note |
| --- | --- | --- |
| B-059-1 | (this fold) | cgo-less darwin CPU source → `unix.Getrusage` (the `runtime/metrics` `/cpu/classes/*` metric read `0` under load on the darwin cgo-less host); ADR 0029 D1/D1a/§Consequences + §Alternatives, the techstack **System metrics provider** row, the spinner-feature round-059 comment, the `chat/dsl.md` note, and the PR body corrected. |
| F-059-1 | (this fold) | Two seeded-delta CPU pins — `TestDarwinNoCGoCPUIsPopulated` (`darwin && !cgo`) + `TestDarwinCGoCPUIsPopulated` (`darwin && cgo`); both **RED** on their respective stub mutations, green on the fix; `research.md` D7 restated as a witness matrix. |
| F-059-2 | (this fold) | `tasks.md` T001–T015 `[X]` + this fold ledger; `research.md` status → Final. |
| F-059-3 | (this fold) | The **Turn progress spinner (operator)** techstack row carries the 1 Hz resource-figure cadence (the owning row, not just the sibling). |
| F-059-4 | (this fold) | Acceptance Examples rebuilt from **existing** sentences (no new stepdefs); Rule 2 recorded as a **documented narrowing** (the cadence unit pin); `R-059-c` PM follow-up recorded in `STATUS.md`. |
| TD-059-1 | (this fold) | ADR 0029 §Forward: nothing compiles the `darwin && cgo` leg on a non-darwin host; a darwin release build MUST be `CGO_ENABLED=1`. |
| TD-059-2 | (this fold) | The foreign `// architect-acceptance:` marker is removed; the darwin-only Mach acceptance is homed in ADR 0029 §Forward. |
| R-059-1 | (this fold) | ADR 0029 §Forward: the throttle cache is not re-armed at `OnToolsStart` (≤ 1 s staleness by design). |
| R-059-2 | (this fold) | ADR 0029 D1a/§Forward: the `÷ NumCPU` normalisation is explicit (machine-fraction scale, distinct from the machine-wide legs). |
| R-059-3 | (this fold) | `chat/dsl.md` resource row notes the strengthened E2E `Then` is host-dependent by construction. |
| TF-059-1 | (this fold) | The ADR 0029 **index** row (`docs/decisions/README.md`) corrected: cgo-less `runtime/metrics` → `getrusage` + the divergence label (the round-057 TF-057-1 / round-058 TF-058-1 class). |
| TD-059-3 | (this fold) | `system_metrics_math.go`'s `agentCPUPercent` doc provenance sentence corrected (fed by `getrusage`, not the reference's `runtime/metrics`). |
| R-059-4 | (this fold) | The cgo-less seeded pin's ≥ 1 ns-CPU design assumption documented in the test comment. |
| R-059-5 | (this fold) | The `techstack.md` System-metrics row carries the **recorded divergence** label (symmetry with `spec.md`). |
