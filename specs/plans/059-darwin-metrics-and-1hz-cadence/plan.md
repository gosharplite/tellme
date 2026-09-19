# Plan: round 059 — `059-darwin-metrics-and-1hz-cadence`

**Feature Branch**: `059-darwin-metrics-and-1hz-cadence`
**Created**: 2026-09-19
**Input**: `spec.md` (US1 real macOS CPU/MEM — P1; US2 1 Hz cadence — P2) · `research.md` · `truth-delta.md`

## 1. Interface inventory

| # | Interface | End | Module | Change |
| --- | --- | --- | --- | --- |
| I-1 | the tool-phase progress spinner's resource segment (`[CPU: x% | MEM: y%]` on `stderr`) | CLI | `chat` | **MODIFY** — the figures are populated on macOS and refresh once per second |

There is **one** CLI interface end (no HTTP/contract surface, no persisted data shape):

- `/axb-api-plan`: **NOOP** — no `specs/truth/contracts/**`.
- `/axb-data-plan`: **NOOP** — no `specs/truth/data/**` change (the sample cache is spinner-local presentation state, not domain state).
- `/axb-ui-plan`: **N/A** — a pure CLI round (no HTML/TUI mockup).
- `/axb-dsl-refine`: owns I-1 (`specs/truth/features/cli/chat/**`).

## 2. Analysis waves

- **Wave 1 — the metrics adapter (`internal/infrastructure/telemetry`)**: split the darwin arm into `darwin && cgo` (Mach CPU + Mach VM memory) and `darwin && !cgo` (`runtime/metrics` + `sysctl` memory); delete the hand-rolled decode; add the build-tag-free math helpers; adopt `golang.org/x/sys/unix`. Linux unchanged.
- **Wave 2 — the shared spinner throttle (`internal/ui`)**: sample the provider at most once per second, reuse the cached pair between frames, keep the 200 ms frame cadence.
- **Wave 3 — pins**: build-tag-free math tables; a darwin `Sample()` non-zero-memory pin; the spinner throttle pin; the strengthened E2E resource row (non-zero MEM).

## 3. Dependencies

- US1 and US2 are **independent**: US1 changes the darwin sampler, US2 changes the shared spinner cadence. Either can land alone.
- US2's throttle is a shared-spinner rule; US1's darwin sampler is a per-platform provider. No interface change in either.

## 4. Architecture notes

- The `internal/domain/metrics` port is **unchanged** (`Sample() (cpu, mem float64)`) — the infra adapters stay behind it (`ui` stays formatter-only; RULE-E baseline unaffected).
- The darwin split uses build tags (`darwin && cgo` / `darwin && !cgo`); the shared darwin helper file is `//go:build darwin`. The pure math is build-tag-free so it pins anywhere.
- The `CGO_ENABLED=0` cross-compile gate now builds the darwin **cgo-less** leg; the cgo leg is exercised by the host build (`CGO_ENABLED=1` darwin/arm64).

## 5. Truth & governance handoff

- `truth-delta.md` records: `/axb-technical-research` (techstack row) · `/axb-dsl-refine` (the spinner feature/`dsl.md`) · `/axb-api-plan` **NOOP** · `/axb-data-plan` **NOOP** · governance **ADR 0029**.
- Acceptance (PM-owned): `features/acceptance/reporting-real-machine-resources.feature` (carried by the `chat` interface rows; the cadence facet is a recorded unit-pin narrowing).
