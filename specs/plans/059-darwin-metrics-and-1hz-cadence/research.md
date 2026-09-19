# Technical Research: round 059 — real macOS CPU/MEM + a 1 Hz resource-sample cadence

**Feature Branch**: `059-darwin-metrics-and-1hz-cadence`
**Created**: 2026-09-19
**Status**: Final — clarify CLOSED (Q1 CPU source · Q2 dependency · Q3 throttle scope · Q4 MEM definition · Q5 platform scope; all → option 1); the PR #125 review folds (B-059-1 + F-059-1…F-059-4) are recorded in §D7/§D8 and the `tasks.md` fold ledger.

## Terminology

- **resource segment** — the ` [CPU: <c>% | MEM: <m>%]` suffix on the tool-phase spinner line.
- **machine-wide CPU** — the host's busy fraction, `(1 − Δidle/Δtotal) × 100` (Linux `/proc/stat`; macOS Mach tick counters via cgo).
- **process CPU** — tellme's own CPU, from Go's `runtime/metrics` `/cpu/classes/total:cpu-seconds` ÷ Δwall ÷ `NumCPU`.
- **cgo leg / cgo-less leg** — the darwin build under `CGO_ENABLED=1` / `CGO_ENABLED=0` respectively.

## Reference findings (`~/tmp/github/gosharplite/tell-me-go`)

- The port is raw counters (`GetCPUStats() (total, idle)` + `GetMemoryPercent() float64`); the UI computes the %.
- `darwin && cgo` (`system_metrics_darwin_cgo.go`): Mach `host_statistics64`/`HOST_CPU_LOAD_INFO` for CPU; `HOST_VM_INFO64` → `(active+wired+compressor)/hw.memsize` for memory; `golang.org/x/sys/unix` for `sysctl`.
- `darwin && !cgo` (`system_metrics_darwin_nocgo.go`): `runtime/metrics` **agent** CPU; `sysctl` free+speculative+purgeable × `0.6` for memory. Never a flat `0`.
- `linux`: `/proc/stat` `cpu ` line + `/proc/meminfo`. Fallback: `runtime/metrics` agent CPU.
- Cadence: frame tick **200 ms**; `updateSystemMetrics` throttles the sample to **≥ 1 s** (`metrics_shouldSample`), reusing the last value.

## Decisions

### D1 — CPU source: machine-wide, reference-style per-build split (Q1 → 1)

`darwin && cgo` samples the machine-wide Mach tick deltas; `darwin && !cgo` samples the process's own CPU via `getrusage` (real, non-zero — see D7a). This mirrors the reference exactly and replaces the round-019 hardcoded `0`. cgo is allowed on darwin; the accurate machine-wide figure runs on a normal `CGO_ENABLED=1` build.

### D2 — `golang.org/x/sys/unix` for the macOS `sysctl` reads (Q2 → 1)

`unix.SysctlUint64`/`SysctlUint32`/`Getpagesize` replace the hand-rolled `syscall.Sysctl` + fixed-width decode, which is **deleted**. Root cause measured this session: `syscall.Sysctl("hw.memsize")` returns **NUL-trimmed** bytes (7 bytes for 16 GiB), so the old `len>=4` branch read the first four bytes = `0`; `vm.page_free_count` (3 bytes) errored out. `golang.org/x/sys` was already in the module graph (v0.41.0) → promoted to a **direct** dependency.

### D3 — The `Sample()` port stays percentage-valued

The reference's raw-counter port is **not** copied (A3); the domain port `Sample() (cpuPercent, memPercent float64)` is unchanged, so `internal/ui` (a pure formatter) and the CLI are untouched.

### D4 — macOS memory = reference-exact per platform (Q4 → 1)

cgo leg: `(active + wire + compressor) × pagesize / hw.memsize` (Mach VM stats). cgo-less leg: `(total − (free+spec+purgeable)×pagesize)/total × 0.6`. Both clamped `[0,100]`. The formulas are pure and live in the build-tag-free `system_metrics_math.go` so they pin on any host.

### D5 — A 1 Hz resource-sample throttle in the shared spinner (Q3/Q5 → 1)

`Sample()` runs at most once per `ResourceSampleInterval` (1 s); the last pair is reused between samples; the **braille frame keeps its 200 ms cadence**. The throttle reads the injected clock (`s.now`) → deterministic under test. Applies on **every** platform (a shared-spinner rule, not a platform leg).

### D6 — Linux unchanged

`system_metrics_linux.go` is not touched; its machine-wide computation and bytes are identical.

### D7 — Witness matrix (falsifiable, reproduced then reverted)

| Witness | Mutation | Kill |
| --- | --- | --- |
| **(a)** darwin MEM | restore the round-019 darwin arm (or force `darwinCGOMemPercent()` → `0`) | the darwin `Sample()` MEM pin **and** the strengthened E2E resource row **RED** |
| **(b)** cgo-less CPU | hardcode the cgo-less CPU to `0` (`skip` the `procCPUNanoseconds` branch) | `TestDarwinNoCGoCPUIsPopulated` **RED** (seeded delta ⇒ deterministic) |
| **(b′)** cgo CPU | hardcode the cgo CPU to `0` (`return 0, darwinCGOMemPercent()`) | `TestDarwinCGoCPUIsPopulated` **RED** (seeded prior ticks) |
| **(c)** throttle | `ResourceSampleInterval := 0` (sample every frame) | `TestSpinnerResourceSampleThrottle` **RED** (calls = 5 within a second) |

Both CPU legs are therefore **killed by a seeded-delta pin** — the earlier "a non-zero CPU under load is host-dependent" admission was wrong (a seeded delta is deterministic), corrected here and by F-059-1.

### D7a — the cgo-less CPU source (B-059-1)

The first shape used `runtime/metrics` `/cpu/classes/total:cpu-seconds`. Review `B-059-1` showed the metric is the *available* CPU budget (`GOMAXPROCS` integrated over wall time), not consumed CPU, **and** is never updated on the darwin cgo-less host (measured `0` under a saturated core; `getrusage` read 3.003 s over the same window) — i.e. it re-shipped `CPU: 0.0%` on the gate's own build. The source is now `getrusage(RUSAGE_SELF)` (user+system CPU ÷ Δwall ÷ `NumCPU`) via `golang.org/x/sys/unix`, which is cgo-free and already the round's dependency; verified non-zero and monotonic on the darwin cgo-less build.

### D8 — Truth impact & governance

- `specs/truth/techstack.md`: the **System metrics provider (telemetry)** row (rewritten: the darwin split, `x/sys`, the reference-exact MEM, the 1 Hz cadence, the recorded narrowing) **and** the **Turn progress spinner (operator)** row (F-059-3: the 1 Hz resource-figure cadence).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`): the resource Rule's comment + the resource row's `必查` (non-zero MEM; the cadence unit-pin narrowing; the R-059-3 host-dependence note).
- **ADR 0029** (this round), indexed, incl. §Forward TD-059-1/TD-059-2/R-059-1…3.
- `specs/truth/data/data-model.dbml`: **NOOP** (the sample cache is spinner-local presentation state).
- New dependency: `golang.org/x/sys` (direct).

### D9 — Scope guard

In scope: the darwin cgo/cgo-less samplers, the decode fix, `x/sys`, the reference-exact MEM, the 1 Hz throttle, the unit pins, the strengthened E2E row, the truth/ADR. Out of scope: the linux sampler; the port signature; the resource-segment format; any other chrome element; a Windows sampler; a user-configurable interval; `--color=always`.

## Risks & residual items (RF-059-x)

- The Mach tick counters are 32-bit `natural_t` and wrap — no rollover handling (mirrors the reference).
- The cgo-less macOS leg reports the **process** CPU, not the machine's (D1/D7; the `CGO_ENABLED=0` cross-compile gate covers this leg only).
- The sample cache is spinner-local — a future non-spinner consumer of the port would re-introduce 5 Hz sampling.
- The **cadence** is carried by a unit pin (a flat capture cannot observe a refresh rate) — a deliberate narrowing, recorded in ADR 0029 / the `presenting-the-progress-spinner` DSL row.
- `x/sys` is now a direct dependency.
