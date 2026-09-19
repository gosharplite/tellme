# Tasks: round 059 — `059-darwin-metrics-and-1hz-cadence`

**Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/**`
**Tests gate**: `gofmt` + `go vet` + `make verify` + `go test -count=1 ./...` (the E2E is Strict).

## Phase 1 — Setup

- [ ] **T001** Add `golang.org/x/sys` as a direct dependency (`go mod tidy`; it is already in the module graph at v0.41.0).

## Phase 2 — Foundational

- [ ] **T002** Add the build-tag-free math helpers `internal/infrastructure/telemetry/system_metrics_math.go` (`agentCPUPercent`, `darwinMemPercentCGO`, `darwinMemPercentNoCGO`, `clampPercent`, `darwinMemoryUsageFactor`).

## Phase 3 — Test Alignment & Implementation

### US1 — real macOS CPU/MEM (P1)

- [ ] **T003** [RED] `system_metrics_math_test.go` — pin the agent-CPU + both darwin memory formulas (exact values).
- [ ] **T004** [RED] `system_metrics_darwin_sample_test.go` (`//go:build darwin`) — `Sample()` reports a non-zero memory figure (fails on round-058 code).
- [ ] **T005** [GREEN] Replace `system_metrics_darwin.go`'s stub/decoder with the shared page-size helper; add `system_metrics_darwin_cgo.go` (Mach `host_statistics64` CPU + `HOST_VM_INFO64` memory) and `system_metrics_darwin_nocgo.go` (`runtime/metrics` CPU + `sysctl` memory); use `golang.org/x/sys/unix`.
- [ ] **T006** [GREEN] Strengthen the E2E resource stepdef (`tests/e2e/steps/step_t010_chat_then_reports_resources.go`) to require a non-zero MEM figure.
- [ ] **T007** [REFACTOR] Confirm the linux adapter is byte-unchanged; `gofmt`/`go vet` clean; witness (a) reproduced + reverted.

### US2 — the 1 Hz resource-sample cadence (P2)

- [ ] **T008** [RED] `internal/ui/spinner_sample_throttle_test.go` — one `Sample` per second across 200 ms frames; nil provider → zeros.
- [ ] **T009** [GREEN] Add `ResourceSampleInterval` + the `sampleResources` throttle in `internal/ui/spinner.go`; route `renderLocked` through it.
- [ ] **T010** [REFACTOR] `gofmt`/`go vet` clean; witness (c) reproduced + reverted.

## Phase 4 — Truth, governance, verification

- [ ] **T011** `specs/truth/techstack.md` **System metrics provider** row (the darwin split, `x/sys`, the reference-exact MEM, the 1 Hz cadence, the recorded narrowing).
- [ ] **T012** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`): the resource Rule's round-059 comment + the resource row's `必查`.
- [ ] **T013** **ADR 0029** (+ the decisions index row).
- [ ] **T014** Truth-delta owner rows finalized (technical-research MODIFY; api/data NOOP; dsl-refine MODIFY; governance ADD).
- [ ] **T015** Verify: `gofmt` clean · `go vet ./...` clean · `make verify` **OK** (incl. the `CGO_ENABLED=0` cross-compile matrix over the darwin cgo-less leg) · `go test -count=1 ./...` green (E2E Strict) · `go test -count=1 ./internal/... ` · diff secret scan clean · `go.mod`/`go.sum` change is `x/sys` become-direct only.

## Fold ledger

| Fold | Commit | Note |
| --- | --- | --- |
| — | — | (filled during review) |
