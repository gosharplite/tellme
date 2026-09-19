# Feature Specification: real macOS CPU/MEM + a 1 Hz sample cadence (round 059)

**Feature Branch**: `059-darwin-metrics-and-1hz-cadence`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify (5 decisions) CLOSED** (Q1→1, Q2→1, Q3→1, Q4→1, Q5→1, below). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)**:

> *"We are in mac os. When spinner shows up during [Tool Output], cpu/mem stay at zero. When cpu/mem do show numbers, the refresh rate is too fast, should be 1 refresh per second."*
> *"please reference tell-me-go, cpu/mem work on ubuntu/mac over there."*

**Behaviour intent**: **MODIFY (CLI-visible diagnostic chrome + one platform sampler).** Two operator-observed defects on the round-019 spinner resource segment (`[CPU: x% | MEM: y%]`), both divergences from `tell-me-go`:

1. **On macOS the segment reads `0.0%` for both figures.** Two independent causes — (a) the darwin CPU leg is a hardcoded stub (`return 0, …`) with **no** non-zero fallback, whereas the reference's darwin **nocgo** build reports a real (agent) CPU; (b) the darwin **memory** leg's hand-rolled `syscall.Sysctl` decoder assumes a fixed width, but darwin returns the value **NUL-trimmed**, so `hw.memsize` decodes to `0` and the whole percentage collapses to `0.0%`.
2. **The CPU/MEM figures refresh 5×/second** (recomputed on every 200 ms spinner frame) instead of the reference's **1×/second**.

The fix mirrors `tell-me-go`: a `darwin && cgo` Mach sampler + a `darwin && !cgo` `getrusage` (process CPU) leg — a **recorded divergence** from the reference's cgo-less `runtime/metrics` path, which reads `0` on this host (the metric is the *available* CPU budget, never updated here) — `golang.org/x/sys/unix` for the sysctl reads, the reference's per-platform memory definition, and a **1 Hz throttle on the sample/digits** (the braille wheel stays at 200 ms) on **all** platforms.

---

## Grounded in the current system *(measured this session, `dev` @ `43c5eda`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/telemetry/system_metrics_darwin.go` | `Sample()` returns `0, darwinMemPercent()` — **CPU is a hardcoded 0** ("pending a cgo mach sampler"). `darwinMemPercent` reads `hw.memsize`/`hw.pagesize`/`vm.page_*` via a hand-rolled `sysctlUint64` that decodes `binary.LittleEndian.Uint64(b[:8])` / `Uint32(b[:4])` on the **raw** `syscall.Sysctl` bytes. |
| *Measured on this macOS host* | `syscall.Sysctl("hw.memsize")` (16 GiB) returns **7 bytes** `00 00 00 00 04 00 00` → the `len>=4` branch reads **the first 4 bytes = 0** ⇒ `total == 0` ⇒ `0.0%`. `vm.page_free_count` returns **3 bytes** ⇒ neither branch ⇒ error ⇒ contributes 0. So MEM is `0.0%` regardless of load. |
| `internal/infrastructure/telemetry/system_metrics_linux.go` | `/proc/stat` `cpu ` line (total = Σcols, idle = idle col only) + `/proc/meminfo` → correct today; unchanged by this round. |
| `internal/domain/metrics/metrics.go` | The port is `Sample() (cpuPercent, memPercent float64)` — **percentages**, not the reference's raw `(total, idle)` headroom. Kept as-is (the spinner formats percentages). |
| `internal/ui/spinner.go` | `renderLocked()` (every 200 ms tick) calls `s.metrics.Sample()` **every frame** and formats `FormatResourceSegment(cpu, mem)` → ` [CPU: %.1f%% | MEM: %.1f%%]`. No sample throttle exists. `SpinnerInterval = 200 ms`. |
| `~/…/tell-me-go` (reference) | Providers expose raw counters: `GetCPUStats() (total, idle)` + `GetMemoryPercent()`. `darwin && cgo` = Mach `host_statistics64`/`HOST_CPU_LOAD_INFO` (CPU) + `HOST_VM_INFO64` `(active+wired+compressor)/hw.memsize` (MEM). `darwin && !cgo` = `runtime/metrics` `/cpu/classes/total:cpu-seconds` (agent CPU) + sysctl free+speculative+purgeable × **0.6** (MEM). Sysctl via `golang.org/x/sys/unix`. Cadence: frame tick 200 ms, **`updateSystemMetrics` throttled to ≥ 1 s** (`metrics_shouldSample`), reusing the last value in between. |
| `Makefile` (`verify-cross-compile`, line 212) | Pins `CGO_ENABLED=0 GOOS=… GOARCH=…` for all targets ⇒ the gate covers the **nocgo** darwin leg; the cgo leg is a host-native (`darwin/arm64`, `CGO_ENABLED=1`) build. |
| `specs/truth/techstack.md` (**System metrics provider (telemetry)** row) | Records the round-019 shape and the known gap *"the macOS CPU leg is pending a cgo mach sampler and reports `0.0%` until it lands"* — this round lands it. |

**Consequence (current behaviour):** at a terminal on macOS the tool-phase spinner reads `[CPU: 0.0% | MEM: 0.0%]`, and where figures do appear they change 5×/second.

---

## Settled design *(operator-clarified this session)*

| # | Decision (source) |
| --- | --- |
| **S-1** | **CPU source = machine-wide, reference-style split (Q1 → 1).** darwin reports **machine-wide** CPU%: under **cgo** via Mach `host_statistics64` (`HOST_CPU_LOAD_INFO`); under **!cgo** via `getrusage(RUSAGE_SELF)` (tellme's **own** CPU ÷ wall-time ÷ `NumCPU`) — a real, non-zero measurement (the reference's cgo-less `runtime/metrics` path was rejected: it read `0` here; ADR 0029 D1a). |
| **S-2** | **Sysctl reads via `golang.org/x/sys/unix` (Q2 → 1).** Add `golang.org/x/sys` as a direct dependency for `SysctlUint64`/`SysctlUint32`/`Getpagesize`; the hand-rolled trimmed-value decode is removed. Mach calls use C headers (unaffected by this dependency). |
| **S-3** | **Throttle = the CPU/MEM sample + digits only (Q3 → 1).** The braille frame keeps its 200 ms cadence; `Sample()` runs at most once per second, the last value repeated between samples. |
| **S-4** | **macOS MEM = reference-exact per platform (Q4 → 1).** cgo leg: `(active + wired + compressor) × pagesize / hw.memsize` (Mach `HOST_VM_INFO64`); nocgo leg: `(total − free − speculative − purgeable)/total`, **× 0.6** to approximate the cgo definition. |
| **S-5** | **Cadence applies on every platform (Q5 → 1).** One shared throttle in the spinner; Linux and macOS both sample once per second (reference parity). Providers stay ignorant of cadence. |
| **S-6** | Nothing else changes: the resource-segment **format** (` [CPU: %.1f%% | MEM: %.1f%%]`), the `Sample()` port signature, the linux sampler, the segment's absence outside the tool phase, the spinner's width/clear/yield behaviour, `turns.log` (the spinner is stderr-only), and the round-054 terminal colour gate. |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — the figures remain **machine-wide** percentages on the platforms where that is achievable (linux machine-wide = unchanged; darwin machine-wide under cgo, **process** CPU under nocgo — an explicit, recorded narrowing of "machine-wide" on the cgo-less darwin build, matching the reference).
- **I-2** — the resource segment is emitted **only** in the tool phase (`OnToolsStart…OnToolsEnd`), on a colour-gated terminal — unchanged.
- **I-3** — `Sample()` is still called **at most once per second**; the throttle never delays the first draw or blocks a frame.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - real macOS CPU/MEM in the spinner (Priority: P1)

As the **operator running tellme on macOS** at a terminal, while tools run I want the spinner's `[CPU: x% | MEM: y%]` to show **real machine-wide figures** (non-zero), so the resource segment is actually informative — matching what I see from `tell-me-go` on Ubuntu/macOS.

**Why this priority**: it is the operator's primary report ("cpu/mem stay at zero"); the zero is caused by two independent darwin defects.

**Independent verification**: on a `darwin/arm64` `CGO_ENABLED=1` build, a tool-running turn at a terminal (or a direct `Sample()` call under load) reports a **non-zero** MEM (and a non-zero CPU after the first second); a `darwin && !cgo` build compiles and reports a non-zero **process** CPU (`getrusage`) + non-zero MEM; `GOOS=linux` output is unchanged.

**Acceptance Scenarios**:

1. **Given** a cgo darwin build, **When** `Sample()` is called with memory in use, **Then** MEM reflects `(active+wired+compressor)/hw.memsize` as a percentage in `(0,100]` (the reference's cgo `GetMemoryPercent`).
2. **Given** a cgo darwin build, **When** `Sample()` is called twice ~1 s apart under load, **Then** CPU is the machine-wide `(1 − Δidle/Δtotal) × 100` from Mach tick deltas (a non-zero value under load).
3. **Given** a `darwin && !cgo` build, **When** `Sample()` is called twice ~1 s apart, **Then** CPU is the process CPU (`getrusage`; non-zero while the process burns CPU) and MEM is the reference's nocgo approximation (`(total−free−speculative−purgeable)/total × 0.6`).
4. **Given** a `linux` build, **When** `Sample()` is called, **Then** the CPU/MEM values and computation are byte-for-byte as today (no regression).
5. **Given** any platform, **When** a required sysctl/API read fails, **Then** the segment degrades to `0.0%` for that figure (never panics, never blocks the frame).

**Functional Requirements**:

- **FR-001**: On darwin under **cgo**, `Sample()` MUST report machine-wide CPU% from the Mach `host_statistics64` `HOST_CPU_LOAD_INFO` tick deltas (`(1 − Δidle/Δtotal) × 100`; `0.0` on the first call, no prior delta) (S-1).
- **FR-002**: On darwin under **!cgo**, `Sample()` MUST report a **non-zero** CPU% from `getrusage(RUSAGE_SELF)` (the process's own CPU ÷ wall-time ÷ `NumCPU` × 100) — never a hardcoded `0` (S-1; ADR 0029 D1a — the reference's cgo-less `runtime/metrics` metric is the *available* CPU budget and read `0` on this host).
- **FR-003**: On darwin, MEM% MUST be decoded correctly (NUL-trim-safe) via `golang.org/x/sys/unix` and MUST match the reference: **cgo** `(active+wired+compressor)/hw.memsize`; **!cgo** `(total−free−speculative−purgeable)/total × 0.6` (clamped to `[0,100]`) (S-2/S-4).
- **FR-004**: The linux sampler MUST be unchanged (machine-wide `/proc/stat` CPU + `/proc/meminfo` MEM) (S-6).
- **FR-005**: The `Sample() (cpuPercent, memPercent float64)` port signature and the resource-segment format ` [CPU: %.1f%% | MEM: %.1f%%]` MUST be unchanged (S-6).

### User Story 2 - the CPU/MEM figures refresh once per second (Priority: P2)

As the **operator**, while tools run I want the `[CPU: x% | MEM: y%]` figures to **refresh once per second** (not 5×/second), so the numbers are readable — while the spinner wheel keeps animating smoothly.

**Why this priority**: the operator's second report; independent of US1 (it is a cadence change in the shared spinner).

**Independent verification**: with a controllable clock/metrics double, `Sample()` is invoked at most once per second across a run of many 200 ms frames; the frame index still advances every 200 ms; the displayed figure repeats the last sampled value between samples.

**Acceptance Scenarios**:

1. **Given** a tool-phase spinner on any platform, **When** ≥ 5 frames elapse within one second, **Then** the resource provider's `Sample()` is called **once**, and the same figures render on the intervening frames.
2. **Given** the same spinner, **When** ≥ 1 s has elapsed, **Then** the next frame re-samples and the figures update.
3. **Given** any tool-phase run, **When** frames are drawn, **Then** the braille frame still advances every 200 ms (unchanged animation).

**Functional Requirements**:

- **FR-006**: The spinner MUST re-sample CPU/MEM at most **once per second** on **every** platform; between samples it MUST reuse the last sampled figures (S-3/S-5).
- **FR-007**: The braille frame cadence MUST remain **200 ms** (the throttle MUST NOT slow the animation) (S-3).
- **FR-008**: The throttle MUST NOT delay the first draw nor block a frame; a missing/zero sample renders as `0.0%` (S-6/I-3).

---

## Edge Cases

- **First second of a tool phase** — CPU shows `0.0%` on the first call (no tick/CPU delta yet), matching the reference's `initCPUTracking` (it seeds the baseline and reports `0.0`). MEM shows the real value immediately.
- **A provider read failure** (missing sysctl, Mach call non-`KERN_SUCCESS`) — that figure renders `0.0%`; no panic, no frame stall (FR-005 / I-3).
- **macOS process idles in the tool phase** — the cgo machine-wide CPU still reflects host load (not the process); the nocgo leg reflects the process's own CPU (recorded narrowing, I-1).
- **`GOOS=linux CGO_ENABLED=0` cross-build** — must still build + vet green; the darwin cgo file is excluded by build tag, the nocgo leg compiles on linux? **No** — the darwin files are `//go:build darwin`; the linux leg is compiled. Unchanged.
- **No `[CPU/MEM]` outside the tool phase** — the segment is tool-phase only; unchanged.
- **Clock/cadence seam in tests** — the throttle reads an injectable `now`, so a test drives it deterministically.

## Key Entities

- **The resource segment** — ` [CPU: <c>% | MEM: <m>%]` appended to the tool-phase spinner line.
- **The system-metrics provider** (`metrics.SystemMetricsProvider`) — the port the spinner samples; one platform adapter behind it.
- **The sample throttle** — the spinner-internal ≥ 1 s gate around `Sample()` + a cached figure pair.

## Success Criteria

- **SC-001**: On a cgo macOS build, a tool-running turn at a terminal reports a **non-zero** MEM and (after the first second) a **non-zero** CPU; on a nocgo darwin build, both figures are non-zero under load.
- **SC-002**: On Linux the CPU/MEM computation and rendered figures are byte-identical to today's.
- **SC-003**: Across a ≥ 5-frame, < 1 s window the provider's `Sample()` is invoked exactly once (a seam-based unit check); the braille frame advances every 200 ms.
- **SC-004**: `go build ./...` + `make verify` hold, including the `CGO_ENABLED=0` cross-compile matrix (which now exercises the darwin **nocgo** leg) and the lint/govulncheck gates; `go.mod`/`go.sum` gain `golang.org/x/sys` (only).
- **SC-005**: Each changed behaviour (darwin CPU, darwin MEM, the throttle) has a reproduced falsifiability witness.

## Assumptions

- **A1**: The macOS machine-wide CPU requires cgo (Mach) — accepted by the operator (Q1 → 1); the `CGO_ENABLED=0` cross-compile gate therefore covers only the nocgo leg, which now reports a non-zero **process** CPU (`getrusage`) (a deliberate, reference-matching narrowing of "machine-wide").
- **A2**: `golang.org/x/sys` is adopted as a direct dependency (Q2 → 1), matching the reference's usage; the techstack **System metrics provider (telemetry)** row is updated (dependency-free posture qualified).
- **A3**: The port stays **percentage-valued** (`Sample() (cpu, mem float64)`) — the reference's raw-counter port is *not* copied; only the sampling sources/definitions are mirrored.
- **A4**: Truth impact is expected in `specs/truth/techstack.md` (**System metrics provider** row) and possibly `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`) if the 1 Hz cadence is expressed as an interface Rule; `/axb-api-plan` + `/axb-data-plan` are **NOOP**.
- **A5**: A new **ADR (0029)** records the darwin sampler split + the 1 Hz cadence, with a §Forward.
- **A6**: `linux` behaviour is explicitly preserved (no machine-wide→agent change on linux).

## Out of scope (recorded forward items)

- A **Windows** sampler (locked exclusion).
- A colour change or any element beyond the resource figures (rounds 054/057/058 chrome is untouched).
- A user-configurable refresh interval / `--color=always`.
- Migrating the port to raw counters (A3) or a functional-options spinner constructor (recorded R-2 nit).
- The locked exclusions: no security/consent layer · no Windows · sequential tool calls ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
