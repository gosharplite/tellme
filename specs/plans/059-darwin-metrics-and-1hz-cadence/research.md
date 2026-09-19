# Technical Research: round 059 — real macOS CPU/MEM + a 1 Hz resource-sample cadence

**Feature Branch**: `059-darwin-metrics-and-1hz-cadence`
**Created**: 2026-09-19
**Status**: Draft (clarify CLOSED — Q1 CPU source · Q2 dependency · Q3 throttle scope · Q4 MEM definition · Q5 platform scope; all → option 1)

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

`darwin && cgo` samples the machine-wide Mach tick deltas; `darwin && !cgo` samples the process's own CPU via `runtime/metrics` (non-zero under load). This mirrors the reference exactly and replaces the round-019 hardcoded `0`. cgo is allowed on darwin; the accurate machine-wide figure runs on a normal `CGO_ENABLED=1` build.

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

### D7 — Witness plan (falsifiable, reproduced then reverted)

- **(a)** restore the old darwin `memPercent` (the mis-decoded `sysctl`) ⇒ the darwin `Sample()` pin + the strengthened E2E resource row red (MEM `0.0%`).
- **(b)** hardcode the darwin CPU back to `0` ⇒ the darwin `Sample()` non-zero-memory pin stays green but a **cgo** noise check / the reference comparison reds; recorded as the CPU-path witness (a non-zero CPU under load is host-dependent, so the durable pin is the math table + the mach path compiling under cgo).
- **(c)** sample on every frame (drop the throttle) ⇒ `TestSpinnerResourceSampleThrottle` red (calls = 5 within a second).

### D8 — Truth impact & governance

- `specs/truth/techstack.md`: the **System metrics provider (telemetry)** row (rewritten: the darwin split, `x/sys`, the reference-exact MEM, the 1 Hz cadence, the recorded narrowing).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`): the resource Rule's comment + the resource row's `必查` (non-zero MEM; the cadence unit-pin narrowing).
- **ADR 0029** (this round), indexed.
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
