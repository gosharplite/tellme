# Phase 0 Research: tellme Tool Resource Contract, `execute_command` & Reader Retrofit (round 024)

**Topic**: how tellme adds a bash-first `execute_command` and applies a uniform, centrally-enforced **tool resource contract** (a token bound + a timeout) across the whole agent tool surface, retrofitting the reader trio (`list_files` / `read_files` / `get_tree`). Each decision supports `spec.md` (US1/US2 · FR-001..FR-017) and the operator-locked decisions **D1–D7** + clarify **Q1–Q3**.

## Decision 1: `execute_command` runs via `bash -c` (bash-first, POSIX-only)

- **Decision**: the command tool runs the model's `command` string through `bash -c "<command>"` (POSIX only). No argv-parsing shell-feature detection, no Windows branch, no `sh` abstraction.
- **Rationale**: the operator-locked direction (README → *Design Intent*: no Windows, bash-first); bash already provides pipes, wildcards, and redirection, so one `bash -c` invocation covers the reference's `execute_command` **and** `pipe_commands` surfaces at once.
- **Alternatives considered**:
  - Argv-direct execution with shell-feature detection + an `sh`-wrap fallback (the reference's route) — rejected: reintroduces the Windows translator/wrapper machinery and the `pipe_commands` split the operator rejected.
  - `sh -c` instead of `bash -c` — rejected: the target is explicitly bash.

## Decision 2: A non-zero exit is a successful tool result carrying the exit code (clarify Q1→1)

- **Decision**: a command that exits non-zero returns a **successful** tool result whose text carries the exit status (and bounded output); the loop continues. A tool **failure** is reserved for a **timeout** and for a tool/argument error (e.g. a missing `command`).
- **Rationale**: the reference returns `Exit Code: N` as a normal result; a non-zero exit is information the model needs (a failing `go test`, a `grep` no-match exits 1), and treating it as terminal would break the agent loop on ordinary outcomes. tellme's loop already feeds a returned tool *error* back as a non-terminal `"error: …"` result; a **contract-level** failure (the frozen `the tool request failed` phrase, exit 7) must **not** fire on a plain non-zero exit.
- **Alternatives considered**:
  - Non-zero exit = a terminal tool failure (exit 7) — rejected: breaks the loop on normal failures; contradicts reference parity.
  - Terminal only above an exit-code threshold — rejected: arbitrary and surprising.

## Decision 3: `output_file` / `append` — structured output capture (clarify Q2→1)

- **Decision**: `execute_command` accepts optional `output_file` (string) and `append` (boolean). When set, the command's output is written to that file (`append` selects `>>` vs `>`); the tool result reports the exit status and the capture target (and may include a bounded preview). This is the sanctioned **large-output escape hatch** — capture to disk, then read it back bounded.
- **Rationale**: lets the model obtain output larger than any safe bound **without raising the bound**; the result reports the target structurally. No path gate (D1).
- **Alternatives considered**:
  - Rely on the model writing `> file` inside the `command` string — sufficient under bash-first, but the target is not reported structurally and quoting is the model's problem; the explicit param is reference parity + a clean result.
  - Defer — rejected (Q2→1).
- **Note (honest)**: this is close to a bash redirect — a convenience knob on an existing tool, not a new tool.

## Decision 4: The token bound and timeout are uniform, centrally-resolved tool parameters (three tiers)

- **Decision**: every agent tool accepts optional `max_output_tokens` (integer) and `timeout` (number, seconds). The **loop** is the single resolution/enforcement point: it resolves the effective value (**param → clamped to the ceiling → else the default**), applies `context.WithTimeout` (already present per call, round-008 FR-009), and **clamps the returned result string** to the effective token bound. A tool MAY additionally early-stop its own output using the resolved bound (passed via the context) to avoid materialising a huge string, but the loop's clamp is the contract guarantee.
- **Rationale**: one enforcement seam keeps a single behaviour and a single set of truth rows; the loop already owns the per-tool timeout and is the one place every call passes through.
- **Alternatives considered**:
  - Per-tool enforcement (each tool clamps + times itself) — rejected: N implementations that can drift.
  - No central clamp (trust the tools) — rejected: a new tool could ship unbounded.

## Decision 5: Default/ceiling derive from the resolved context budget (clarify Q3→1)

- **Decision**: the **default** `max_output_tokens` = the resolved context budget / **4**; the **ceiling** = the resolved context budget. Both come from the round-009 budget resolver (`MAX_HISTORY_TOKENS`, default **1000000**). At the default budget the tool bound is **250000 tokens ≈ 1 MiB** (at the len/4 estimator) — preserving the round-021 "cannot exhaust the context" bound; at a smaller budget (e.g. a 200 k-token model) the default scales down to **50000 tokens**, **resolving issue #49**. The **timeout** default is per-tool: `execute_command` = **300 s** (aligning with `agent.DefaultToolTimeout`, the existing per-tool bound) and the readers (`list_files`/`read_files`/`get_tree`) = **30 s**; the **ceiling** = **7200 s** (2 h; reference parity). A param above a ceiling is **clamped**, never rejected.
- **Rationale**: the default must protect the context **and scale with the model** (the operator's core requirement); tying it to the same resolved budget the round-009 payload line already measures against gives one source of truth and retires the magic 1 MiB constant. The chosen numbers preserve current behaviour at the default budget.
- **Alternatives considered**:
  - Keep a fixed 1 MiB default — rejected: cannot scale with a model window smaller than ~1 MiB → issue #49.
  - A fixed token default (e.g. 250000) independent of the budget — rejected: same non-scaling defect.
  - Derive the timeout from the bound — rejected: they protect different resources (time vs size); independent knobs.

## Decision 6: The reader trio is retrofitted to one aggregate bound (no per-file cap, no paging)

- **Decision**: retire the fixed `readMaxPerFile` (100000) and `readAggregateCap` (1 MiB) constants. `read_files` returns each requested file **whole**, in request order, **stopping at the aggregate `max_output_tokens` bound**; on overflow it emits a marker naming what was read and what was skipped. A single file larger than the bound is truncated at the bound with a marker pointing to the shell for slices. `list_files`/`get_tree` are bounded by the same bound. The **≤50 files/call** cap is kept (degenerate-input guard).
- **Rationale**: D6 — no per-file math, no paging (a multi-file paging schema is AI-hostile, per the operator); the shell is the paging layer; whole-file reads fix the ">100 KB unreadable" defect.
- **Alternatives considered**:
  - Keep a per-file cap — rejected: blocks large files (the operator's objection).
  - Fair-share `aggregate ÷ len(filepaths)` — rejected: invisible arithmetic the AI gets wrong (the operator's objection).
  - Paging (`offset`/`limit`) in `read_files` — rejected: paging is single-file; the shell (`sed`/`head`/`tail`) is the paging layer.

## Decision 7: Uniform truncation / skip markers

- **Decision**: every reader's over-budget result ends with a single truncation marker (e.g. `... (truncated: result budget reached)`), and `read_files` additionally emits a **skip** marker naming the files it could not return (e.g. `... (not read: result budget reached): <paths>`). The exact strings are a `/axb-dsl-refine` detail; the approach is **one marker per cut kind**.
- **Rationale**: a clear, uniform signal the model can act on (shrink the request, or read a slice via the shell); replaces the round-021 `... (truncated)` / `... (truncated at the read budget)` pair.
- **Alternatives considered**:
  - Silent truncation — rejected: the model cannot tell it lost data.
  - Per-tool bespoke markers — rejected: drift; one vocabulary is clearer.

## Decision 8: No security, no Windows, no new dependency; hermetic harness

- **Decision**: no `SafePath`/consent/whitelist (D1); POSIX/bash only (D2); **stdlib only** — the command tool uses `os/exec` (`bash -c`) inside the tool adapter, the readers stay `os`/`path/filepath`/`io`; no new module. The E2E harness stays hermetic (fake provider + working-directory runner + scripted tool calls); a scripted `execute_command` drives the command tool; no pty; `stdout` byte-exact.
- **Rationale**: the operator direction and the round-008/021 precedent; keeps the surface stdlib-only and the tests offline. The reference's injected `ProcessRunner` domain port is **not** adopted this round — the command tool calls `os/exec` behind the tool seam; a port can be extracted when a second process consumer appears.
- **Alternatives considered**:
  - Introduce a `ProcessRunner` domain port now (the reference's ADR-074 route) — deferred: only one process consumer exists this round; adding the port would be speculative indirection.
  - Run via `sh`/a shell abstraction — rejected (D1/D2).
