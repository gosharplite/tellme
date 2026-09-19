# ADR 0029 — Real macOS CPU/MEM in the spinner + a 1 Hz resource-sample cadence

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0010](0010-test-deadline-decoupling.md) (the test-deadline discipline — no forced wait), round 019 (`specs/plans/019-*` — the spinner + the system-metrics port), round 040 (**ADR 0009** — the idle-gap liveness / dual timer), [`tell-me-go`](https://github.com/gosharplite/tell-me-go) (the reference whose `system_metrics_*` + `metrics_shouldSample` this round mirrors)

## Context

Two operator-reported defects on the round-019 spinner's tool-phase resource segment (` [CPU: x% | MEM: y%]`), both divergences from `tell-me-go`:

1. **On macOS the segment reads `0.0%` for every host.** Two independent causes in `internal/infrastructure/telemetry/system_metrics_darwin.go`:
   - the **CPU** leg was a hardcoded stub — `Sample()` returned `0, darwinMemPercent()` — with **no** non-zero fallback (the round-019 note recorded "pending a cgo mach sampler");
   - the **memory** leg mis-decoded the `sysctl` value: `syscall.Sysctl` returns the value **NUL-trimmed**, but the hand-rolled decoder assumed a fixed width. Measured on a 16 GiB host: `hw.memsize` returns **7 bytes** `00 00 00 00 04 00 00`, so the `len>=4` branch read the first four bytes = **0** ⇒ `total == 0` ⇒ `0.0%`; `vm.page_free_count` returns 3 bytes ⇒ an error ⇒ contributes `0`.
2. **The figures refresh 5×/second.** `internal/ui/spinner.go`'s `renderLocked` called `s.metrics.Sample()` on **every** 200 ms frame, rather than the reference's throttled once-per-second sample.

The reference (`~/tmp/github/gosharplite/tell-me-go`) works on ubuntu/mac: a `darwin && cgo` Mach sampler, a `darwin && !cgo` `runtime/metrics` fallback, `golang.org/x/sys/unix` for the `sysctl` reads, and a `metrics_shouldSample` 1 s throttle around a 200 ms frame tick.

## Decision

**D1 — CPU source = machine-wide, reference-style per-build split.** On macOS under **cgo**, `Sample()` reports machine-wide CPU from the Mach `host_statistics64` `HOST_CPU_LOAD_INFO` tick deltas (`(1 − Δidle/Δtotal) × 100`; `0.0` on the first call — no prior delta). Under **!cgo**, it reports the **process's own** CPU from Go's `runtime/metrics` (`/cpu/classes/total:cpu-seconds` ÷ Δwall ÷ `runtime.NumCPU()`, non-zero under load) — never a hardcoded `0`, matching the reference's cgo-less build.

**D2 — `sysctl` reads via `golang.org/x/sys/unix`.** `unix.SysctlUint64`/`SysctlUint32`/`Getpagesize` replace the hand-rolled `syscall.Sysctl`+fixed-width decode (which is deleted). `golang.org/x/sys` becomes a **direct** dependency (it was already in the module graph); the Mach calls use C headers as before. This removes the decode bug class entirely.

**D3 — macOS memory = reference-exact per platform.** The **cgo** leg uses the Mach VM statistics — `(active + wire + compressor) × pagesize / hw.memsize`; the **!cgo** leg uses the `sysctl` free + speculative + purgeable page counts over total, scaled by the reference's `0.6` factor — both clamped to `[0, 100]`. The pure formulas live in the build-tag-free `system_metrics_math.go` so they pin on any host.

**D4 — A 1 Hz resource-sample throttle, on every platform.** The spinner samples the provider at most once per `ResourceSampleInterval` (1 s), reusing the last pair between samples; the **braille frame keeps its 200 ms cadence**. The throttle reads the spinner's injected clock, so it is deterministic under test. This matches the reference's `metrics_shouldSample` and applies to Linux too (the defect lives in the shared spinner, not a platform leg).

**D5 — Linux is unchanged.** The Linux sampler (machine-wide `/proc/stat` + `/proc/meminfo`) is not touched; only the darwin arm and the shared cadence change.

**D6 — The `Sample() (cpuPercent, memPercent float64)` port stays percentage-valued.** The reference's raw `(total, idle)` counter port is **not** copied; only the sampling sources/definitions are mirrored.

**D7 — The cgo-less macOS CPU is a recorded narrowing of "machine-wide".** On the `CGO_ENABLED=0` build the CPU figure is the process's own CPU (matching the reference's cgo-less path). The `make verify-cross-compile` gate pins `CGO_ENABLED=0`, so it exercises the cgo-less leg only; the accurate machine-wide CPU runs on a normal `CGO_ENABLED=1` darwin build.

## Consequences

- **Positive**: on macOS the resource segment is now informative (`[CPU: x% | MEM: y%]`), matching what the operator sees from the reference on ubuntu/mac; the figures stop changing 5×/s.
- **New dependency**: `golang.org/x/sys` (direct); recorded in `specs/truth/techstack.md`.
- **Cgo on darwin**: the accurate machine-wide CPU needs a C toolchain (as the reference does); the cgo-less leg always builds.
- **Unchanged guarantees**: the port signature, the resource-segment format, the linux sampler, the tool-phase-only emission, the round-054 colour gate, and `stdout`/`turns.log` byte-exactness.
- **Forward** (RF-059-x): (1) the Mach tick counters are 32-bit `natural_t` and wrap — no rollover handling (mirrors the reference); (2) the cgo-less leg reports process CPU, not the machine's (D7); (3) the sample cache is spinner-local — a future non-spinner consumer of the port would re-introduce 5 Hz sampling; (4) `x/sys` is now a direct dependency (keep it in the vendored/module budget); (5) a Windows sampler remains a locked exclusion.

## Alternatives considered

- **cgo-free everywhere (process CPU on all platforms)**: rejected by the operator — diverges from the reference's default build on Linux/macOS.
- **Hand-rolling a corrected stdlib decoder** instead of `x/sys`: rejected — `x/sys/unix` is the reference's canonical, trim-safe API and removes the bug class.
- **`(total − free − spec − purgeable)/total` on both macOS legs (no `0.6`)**: rejected — reads lower than the reference's cgo number; the operator chose reference-exact per-platform memory.
- **Slowing the whole spinner line to 1 Hz**: rejected — the wheel would visibly stutter; only the sample/digits are throttled.
