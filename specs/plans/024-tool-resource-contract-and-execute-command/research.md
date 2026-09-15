# Phase 0 Research: tellme Tool Resource Contract, `execute_command` & Reader Retrofit (round 024)

**Topic**: how tellme adds a bash-first `execute_command` and applies a uniform, centrally-enforced **tool resource contract** (a token bound + a timeout) across the whole agent tool surface, retrofitting the reader trio (`list_files` / `read_files` / `get_tree`). Each decision supports `spec.md` (US1/US2 · FR-001..FR-017) and the operator-locked decisions **D1–D7** + clarify **Q1–Q3**.

## Decision 1: `execute_command` runs via `bash -c` (bash-first, POSIX-only)

- **Decision**: the command tool runs the model's `command` string through `bash -c "<command>"` (POSIX only). No argv-parsing shell-feature detection, no Windows branch, no `sh` abstraction.
- **Rationale**: the operator-locked direction (README → *Design Intent*: no Windows, bash-first); bash already provides pipes, wildcards, and redirection, so one `bash -c` invocation covers the reference's `execute_command` **and** `pipe_commands` surfaces at once.
- **Alternatives considered**:
  - Argv-direct execution with shell-feature detection + an `sh`-wrap fallback (the reference's route) — rejected: reintroduces the Windows translator/wrapper machinery and the `pipe_commands` split the operator rejected.
  - `sh -c` instead of `bash -c` — rejected: the target is explicitly bash.

## Decision 1a: Terminate the whole process group on timeout (review B2)

- **Decision**: each `execute_command` child runs in its **own process group** — `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` — and on the deadline the tool kills the **negative pgid** (`syscall.Kill(-pgid, SIGKILL)`) so bash **and its descendants** (a spawned compiler, `sleep`, a server) die together. `cmd.WaitDelay` (Go 1.20+) bounds the capture/`Wait` after cancellation, so an orphaned descendant holding the stdout pipe cannot wedge the wait. `syscall`/`os/exec` are stdlib (no dependency change).
- **Rationale**: `exec.CommandContext` + ctx-cancel SIGKILLs **only the direct child (`bash`)**; grandchildren survive **and keep the inherited stdout pipe open**, so the reader/`Wait` blocks past the deadline — the timeout does not actually fire and FR-003's "process tree" wording is unfulfilled. This is exactly the hazard the reference solved with `SysProcAttr`/`configureProcAttrs`.
- **Alternatives considered**:
  - ctx-cancel + `Wait` only (the existing loop seam) — rejected: leaves the descendant tree alive and can wedge on the open pipe.
  - Kill the direct pid only, with no `WaitDelay` — rejected: a descendant holding the pipe blocks `Wait` indefinitely.

## Decision 2: A non-zero exit is a successful tool result carrying the exit code (clarify Q1→1)

- **Decision**: a command that exits non-zero returns a **successful** tool result whose text carries the exit status (and bounded output); the loop continues. A tool **failure** is reserved for a **timeout** and for a tool/argument error (e.g. a missing `command`).
- **Rationale**: the reference returns `Exit Code: N` as a normal result; a non-zero exit is information the model needs (a failing `go test`, a `grep` no-match exits 1), and treating it as terminal would break the agent loop on ordinary outcomes. tellme's loop already feeds a returned tool *error* back as a non-terminal `"error: …"` result; a **contract-level** failure (the frozen `the tool request failed` phrase, exit 7) must **not** fire on a plain non-zero exit.
- **Alternatives considered**:
  - Non-zero exit = a terminal tool failure (exit 7) — rejected: breaks the loop on normal failures; contradicts reference parity.
  - Terminal only above an exit-code threshold — rejected: arbitrary and surprising.

## Decision 3: `output_file` / `append` — structured output capture (clarify Q2→1)

- **Decision**: `execute_command` accepts optional `output_file` (string) and `append` (boolean). When set, the tool binds the child's **stdout and stderr** directly to the opened file handle (`cmd.Stdout = f` / `cmd.Stderr = f`; `append` selects `O_APPEND|O_CREATE|O_WRONLY` vs `O_TRUNC|O_CREATE|O_WRONLY`) — the output **streams to disk and never enters the tool's memory**, so there is no re-materialisation. **Both streams are redirected** (a recorded behavioural choice): a failing command with `output_file` set therefore surfaces only the exit status + target to the model (no inline error text). The tool result reports the exit status and the capture target, with **no inline preview** (a preview is a subsequent bounded `read_files`). This is the sanctioned **large-output escape hatch**.
- **Rationale**: lets the model obtain output larger than any safe bound **without raising the bound**, and without the OOM path of capture-then-write; the result reports the target structurally. No path gate (D1).
- **Alternatives considered**:
  - Prepend a shell redirect into the `bash -c` string — rejected: quoting / `set -e` / pipe-merge hazards, and the target is not reported structurally.
  - Capture the bytes then write them — rejected: re-materialises the very output the escape hatch exists to avoid (OOM for `yes`, a runaway build).
  - Inline bounded preview alongside `output_file` — rejected (review D2): forces capture, defeating the escape hatch.
  - Defer — rejected (Q2→1).
- **Note (honest)**: this is close to a bash redirect — a convenience knob on an existing tool, not a new tool.

## Decision 4: The token bound and timeout are uniform, centrally-resolved tool parameters (three tiers)

- **Decision**: every agent tool accepts optional `max_output_tokens` (integer) and `timeout` (number, seconds). The **loop** is the single resolution/enforcement point: it resolves the effective value (**param → clamped to the ceiling → else the default**), applies `context.WithTimeout` (already present per call, round-008 FR-009), and **clamps the returned result string** to the effective token bound. **Every tool MUST bound its own output at the source** (review D1): `execute_command` streams the child's stdout through a bounded buffer and stops at the effective byte budget (draining/discarding the remainder), and the readers apply the aggregate limit **incrementally** while reading; the loop's clamp is the **backstop**, not the only defence.
- **Rationale**: one enforcement seam keeps a single behaviour and a single set of truth rows; the loop already owns the per-tool timeout and is the one place every call passes through.
- **Alternatives considered**:
  - Per-tool enforcement (each tool clamps + times itself) — rejected: N implementations that can drift.
  - No central clamp (trust the tools) — rejected: a new tool could ship unbounded.

## Decision 5: Default/ceiling derive from the model-aware effective budget (clarify Q3→1)

- **Decision**: the tool bound derives from the **effective budget** = `min(configured MAX_HISTORY_TOKENS, the active model's configured context window)`, where the window is an optional per-model `MODELS.<model>.CONTEXT_WINDOW` (absent → the configured budget). The **default** `max_output_tokens` = `effectiveBudget ÷ 4`; the **ceiling** = `effectiveBudget ÷ 2` (headroom reserved for the conversation + answer — review D3). The **timeout** default is per-tool: `execute_command` = **300 s** (aligning with `agent.DefaultToolTimeout`), readers = **30 s**; the **ceiling** = **7200 s**. A param above a ceiling is **clamped**, never rejected.
- **Rationale**: `MAX_HISTORY_TOKENS` alone is **model-independent** (env→file→default, `DefaultMaxHistoryTokens = 1000000`; it never consults the model — review B1), so a single tool result could exceed a 200 k-window model and fill it. Capping the effective budget by the model's configured window makes the bound **track the model**; the ÷4 default and ÷2 ceiling reserve headroom so one call cannot fill the window (with **no pruning**, a settled exclusion). At the shipped 1 M budget with no window configured the default is 250000 tokens ≈ 1 MiB (behaviour preserved); with a 200 k window configured the effective budget is 200000, so the default is 50000 and the ceiling 100000. **Opt-in / recorded limitation:** the window is **configured** — when `CONTEXT_WINDOW` is absent the effective budget is the operator's `MAX_HISTORY_TOKENS` (default 1000000), so the bound is model-blind in that case; #49's protection is therefore **config-gated** (the truth states the fallback honestly), and a one-time log when the active model has no configured window is a recorded implementation note.
- **Alternatives considered**:
  - A model-blind budget (`MAX_HISTORY_TOKENS` only) — rejected (review B1): cannot protect a small-window model; the "scales with the model" claim would be false.
  - A fixed token default (e.g. 250000) independent of the budget — rejected: same non-scaling defect.
  - Ceiling = the full budget — rejected (review D3): with no pruning, one call could self-inflict a `context_overflow`; hence the ÷2 reserve.
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

- **Decision**: no `SafePath`/consent/whitelist (D1); POSIX/bash only (D2); **stdlib only** — the command tool uses `os/exec` (`bash -c`) inside the tool adapter, the readers stay `os`/`path/filepath`/`io`; no new module. The E2E harness stays hermetic (fake provider + working-directory runner + scripted tool calls); a scripted `execute_command` drives the command tool; no pty; the child's `stdout`/`stderr` are bound to the tool's own buffers (`cmd.Stdout`/`cmd.Stderr`), **never** inherited from `os.Stdout`/`os.Stderr`, so the byte-exact `stdout` guarantee survives a command that writes heavily.
- **Rationale**: the operator direction and the round-008/021 precedent; keeps the surface stdlib-only and the tests offline. The reference's injected `ProcessRunner` domain port is **not** adopted this round — the command tool calls `os/exec` behind the tool seam; a port can be extracted when a second process consumer appears (recorded trigger: the first **write** tool, or any second managed-process consumer — so it is not re-litigated per round).
- **Alternatives considered**:
  - Introduce a `ProcessRunner` domain port now (the reference's ADR-074 route) — deferred: only one process consumer exists this round; adding the port would be speculative indirection.
  - Run via `sh`/a shell abstraction — rejected (D1/D2).
