# Phase 0 Research: tellme tool-usage accounting — per-tool outcomes + an offline roll-up (round 026)

<!--
  Decision-driven research for round 026. Each decision keeps:
  Decision / Rationale / Alternatives considered.
-->

**Context.** Round 024 shipped the tool resource contract and the bash-first `execute_command`; the agent
tool set is now the reader trio (`list_files` / `read_files` / `get_tree`) plus `execute_command`. Every
invocation already flows through `AgentLoop.Run` (`internal/agent/agentloop.go`), which holds both the
tool's returned **error** (`terr`) and the **per-call deadline** (`context.WithTimeout`) — but nothing
counts invocations, and nothing records whether a tool succeeded, failed, or ran out of time. Issue #53
asks for a **measurement** slice: count per-tool invocations + pass/fail so the operator can see which
tools the AI actually uses (prune candidates) and which fail (fix candidates) — evidence for the
surface-minimalism thesis.

**Operator-locked decisions (this session, clarify round 1).**
- **Q1 → 1**: classify each executed invocation as **`ok` / `error` / `timeout`**; a "failure" = **error or timeout**.
- **Q2 → Others**: persist to a **global** append-only JSONL log at **`~/.tellme/tools-count.jsonl`**, one record per invocation `{timestamp, tool, outcome}`, shared across repos/envs/modes and **never reset by `--new`**.
- **Q3 → 1**: surface via a **dedicated offline reporting flag** that prints the roll-up to `stdout` and exits.

## Decision 1: Three-way outcome from structural loop signals (never result-text sniffing)

- **Decision**: classify every **executed** invocation as exactly one of `ok` / `error` / `timeout`:
  - `error` — the tool returned a **non-nil error** (the loop's `terr != nil` path; the loop substitutes `error: <msg>`).
  - `timeout` — the tool returned a **nil-error result** but the **per-call context deadline was exceeded** (`errors.Is(tctx.Err(), context.DeadlineExceeded)`) — the round-024 FR-018 "stopped at the time limit" path shared by the readers and `execute_command`.
  - `ok` — otherwise, **including** a bounded/truncated result (a trim is a bounded success, not a failure).
- **Rationale**: both failure kinds are **structural** at the loop (`err`, and the loop-owned deadline); no tool-result text is ever sniffed, so the generic `internal/agent` loop stays decoupled from tool-specific markers (`capMarker`, `timeoutMarker`, …). This is the only classification the operator-locked Q1→1 requires, and it captures the two real "did not deliver" modes (a Go error; a stop at the time limit).
- **Stated loop↔tool invariant (required fold, review finding 3)**: the classification rests on an invariant the truth now states explicitly — *every tool derives its timeout from the loop-owned per-call `ctx`* (the readers `get_tree`/`list_files`/`read_files` gate on `timedOut(ctx)`, and `execute_command` uses `exec.CommandContext(ctx, …)` + `timedOut(ctx)`), so **a tool produces a timeout result only when that deadline has expired; a tool MUST NOT fabricate a timeout result outside that deadline.** Without it, a future tool that returns a timeout-like result off an internal timer would be silently misclassified as `ok`.
- **Tie-break (required fold, review finding 3)**: in `execute_command`'s `runCaptured` the byte budget ("trimmed") and the deadline ("stopped") can both fire; the tool then returns a **truncation** result while the loop's deadline signal is true — FR-002 and the structural signal can disagree. **DECISION: the accounting follows the loop's deadline signal** (records `timeout`), and the tool's trim-vs-stop guard remains a *result-shape* concern (the result text stays the bounded success). The boundary is **unit-pinned** in T016.
- **Alternatives considered**:
  - **Binary pass/fail where fail = non-nil error only** — rejected (Q1 not chosen): a timeout is a nil-error *result* by the FR-018 contract, so it would be miscounted as a pass and the "stopped at the time limit" signal lost.
  - **Four-way incl. `truncated`** — rejected (Q1 not chosen): truncation is only visible by matching a tool's result markers, which couples the generic loop to tool internals; and a bounded result is a success by design.
  - **Widen the `Tool` port to return a typed outcome** — rejected this round: a larger contract change than a measurement slice warrants; the loop already has the two signals it needs (`err` + deadline).

## Decision 2: A global append-only JSONL log at `~/.tellme/tools-count.jsonl`

- **Decision**: persist one JSON object per line — `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"<ok|error|timeout>"}` — appended to `~/.tellme/tools-count.jsonl` (the user home resolved via `os.UserHomeDir()`), opened `O_APPEND|O_CREATE|O_WRONLY` (0o644). The file is **global** (shared across repos, envs, and personas) and **never reset, archived, or truncated by `--new`**. There is **no** summary companion — the report reads the log once and sums. **Streaming read (required fold, review finding 4):** the report MUST aggregate in a **single streaming fold** (`scan(file, func(rec){ counts[rec.Tool][rec.Outcome]++ })`, O(tools) memory) and MUST NOT materialise every record (`Load()` whole-file allocation is reserved for a round-trip unit test only) — the file is unbounded and machine-global. **Reader resilience (required fold, review finding 5):** because the log is appended with **no `flock`** (a settled exclusion), a concurrent writer's final line can be observed half-written; a **malformed/torn line is skipped** best-effort and the report never fails or corrupts the roll-up on log content (the read-side complement of the write-side best-effort rule). **Lazy creation (required fold, review finding 8):** the `~/.tellme/` directory and the log file are created **on the first `Record`**, never at construction, so a tool-less turn leaves `~/.tellme` untouched. **Timestamp (required fold, review finding 8):** `timestamp` comes from the adapter's `time.Now()` — explicitly **unasserted** (no injected-clock determinism is claimed).
- **Rationale**: the thesis wants **long-run** evidence ("which tools does the AI never use / often fail"), so a machine-wide cumulative log beats a per-session one. Append-only JSONL matches tellme's state convention (`history.jsonl`, `tokens.log`, `global_prompts.jsonl`) and makes concurrent writers safe **without `flock`** (a single small `O_APPEND` write is atomic across processes — a settled no-`flock` exclusion). A summary file is unnecessary because the log is read only on the reporting path (once per invocation, O(N) streaming), unlike round 018's per-turn summary.
- **Rationale (placement)**: `~/.tellme/` (a **user-global** root) is the operator-locked choice (Q2 → Others). It is a **recorded divergence** from the domain model's `TellMeHome`-is-the-namespace convention (the home is otherwise the stable namespace for `configs/`, `output/`, etc.); tellme already records divergences (no security, no Windows, the spinner). It stays **local state** and is modelled by `/axb-data-plan`.
- **Alternatives considered**:
  - **Round-018-style per-mode `output/<mode>/tools.log` + summary** — rejected (Q2 not chosen): per-session (reset on `--new`) contradicts the global intent, and it duplicates files per mode.
  - **`~/.tellme/tools-count.yaml`** — rejected: YAML is tellme's *config* format, not its *state* format; and a single mutable counter file needs read-modify-write, which races across concurrent invocations without `flock`.
  - **Summary-only file** (cumulative counts, no per-invocation log) — rejected: loses the per-invocation audit and per-run slice, and still needs read-modify-write (races).

## Decision 3: The counting seam — an injected `ToolUsageSink` domain port in the loop

- **Decision**: the loop records each executed invocation through an injected sink — a small domain port (e.g. `internal/domain/history.ToolUsageSink` with `Record(ToolUsageRecord)`), **nil = no-op**. The CLI constructs the file-backed adapter (`~/.tellme/tools-count.jsonl`) and injects it, mirroring `gatewayFactory` / `historyStoreFactory` / `newUsageStore`. A sink error is **swallowed** (best-effort); `internal/agent` stays free of filesystem/`os.UserHomeDir` coupling. The classification (Decision 1) is computed in the loop; the sink only writes.
- **Rationale**: the loop is the only place holding `err` + the deadline (Decision 1), and injection keeps the loop pure/testable and the adapter swappable (mirroring the round-018 usage-store seam). Nil-safe keeps construction and unit tests trivial. Best-effort honours the spec's FR-004 (a log failure must never break a turn).
- **Port home (required fold, review finding 6)**: the port lives in **`internal/domain/history`** (a **new** file `tool_usage.go`) — chosen for **adapter co-location consistency with the round-018 `UsageStore`** (both are per-run accounting stores whose adapters live in `internal/infrastructure/history`). `internal/domain/tools` (home of `Tool`/`Registry`/`ByteBudget`) is the equally-defensible cohesion alternative; the round-018 precedent + adapter co-location wins here, and this is recorded so the choice is not re-litigated. (The concrete file is **new**, leaving `usage.go` untouched — see `plan.md` Structure Decision.)
- **Alternatives considered**:
  - **Record inside each tool** — rejected: couples every tool to accounting (tools are pure) and duplicates the classification N times.
  - **Persist on the `history.Step`** (add an outcome field) — rejected: `history.jsonl` is **per-session** (not global) and its shape is frozen truth (rounds 008/014); widening it for a global measurement is the wrong home.
  - **Emit a per-turn line to `stderr`** — rejected: breaks byte-exactness expectations on the diagnostic stream and adds noise (Decision 4).

## Decision 4: The reporting path — a new offline flag; deterministic; every registered tool

- **Decision**: add a new boolean flag — provisional **`--tool-usage`** — that, when set, prints the per-tool roll-up to **`stdout`** and exits, mirroring `-l` / `-d`. It is **strictly offline** (no provider request, no stdin read) and, like `-l` / `-d`, is evaluated in the dispatch precedence **before** any prompt/stdin access. The report lists **every registered agent tool** in the **registry's offer order** (`list_files`, `read_files`, `get_tree`, `execute_command`) with, per tool, total invocations and the `ok` / `error` / `timeout` breakdown, aggregated over the **whole** global log (a streaming pass — NFR-001). The exact flag name and the report's line wording are pinned by the CLI contract owner (`/axb-dsl-refine`).
- **Rationale**: mirrors the established offline reporting commands (`-l`, `-d`) so the prompt path's `stdout` stays byte-exact; a global file naturally yields a machine-wide view; listing every registered tool (from the **live registry**) makes **zero-use** tools visible (the prune signal); a fixed registry order makes the output **deterministic** independent of the log's line order (FR-009).
- **Footprint (required fold, review finding 1)**: the report is **`--version`-class**, *not* `-l`-class. `-l` runs `resolveWorkspace(homeDir)`, which **requires `TELL_ME_HOME`** (exit 4 if unset) **and** creates the session workspace via `home.EnsureWorkspace` — but the round-026 log lives in `~/.tellme` (resolved from `os.UserHomeDir()`), deliberately outside `TellMeHome`. So the report is pinned to require **neither `-c`, nor `TELL_ME_HOME`, nor a workspace**: it reads only `os.UserHomeDir()` + the live registry, writes **plain text to `stdout`**, is **not TTY-gated**, and is **not affected by `-r`**. An E2E assertion with `TELL_ME_HOME` **unset** distinguishes the two designs (a bare `tellme --tool-usage` in a shell without `TELL_ME_HOME` must succeed, not exit 4).
- **Alternatives considered**:
  - **Fold into `-d`** — rejected (Q3 not chosen): mixes two concerns (resolution vs usage) and shows only when `-d` runs.
  - **A per-turn `stderr` line** — rejected (Q3 not chosen): recurring noise, and a *global cumulative* number printed each turn is odd.
  - **A subcommand** (`tellme tool-usage`) — rejected: tellme has no subcommands and `spf13/cobra` is not introduced (Not Introduced Yet).

## Decision 5: Hermetic verification — HOME pointed at a temp dir; E2E + unit split

- **Decision**: because the write resolves the user home, the E2E harness and unit tests point **`HOME`** at a temporary directory (mirroring how the tests already control `TELL_ME_HOME`), so no test ever touches the operator's real `~/.tellme`. When the home cannot be resolved, the write is a **silent no-op**. The **record + report** are witnessed **end-to-end** (the E2E reads the log file and asserts the report on `stdout`), and the **outcome classification** — especially the `timeout` leg (which would need a scripted over-deadline tool) — is additionally **unit-pinned** on the loop with a fake tool that returns an error / blocks past a short injected deadline.
- **Rationale**: keeps the suite hermetic and deterministic, and covers the timeout branch where a Gherkin `Then` cannot practically script it (the round-024 FR-018 unit-pin precedent).
- **Alternatives considered**:
  - **A dedicated `TELL_ME_TOOLS_LOG` env override** — rejected: redundant with `HOME`, and an extra seam to justify; the existing `HOME` control is standard and sufficient.
  - **A real `~/.tellme` write in tests** — rejected: not hermetic; pollutes the operator's machine.

## Decision 6: No new dependency; POSIX-only; best-effort; `stdout` untouched; recorded divergence

- **Decision**: the log uses the Go standard library (`encoding/json` over `os`) — **no new module** (NFR-002). The behaviour is **POSIX-only** (`os.UserHomeDir()` → `$HOME`; a single `O_APPEND` write; no `flock` — NFR-003). The write is **best-effort** and emits **nothing** to `stdout`/`stderr` (FR-004/FR-005); the round-018 token store, the tool semantics/set, the post-turn status lines, and the spinner are **unchanged** (FR-011). The write happens **before** the tool result is folded back into the conversation, so a record exists for every executed call even if a later call aborts the run.
- **Recorded divergences / forward items**:
  - The `~/.tellme` global root is a **deliberate divergence** from `TellMeHome`-is-the-namespace (Decision 2).
  - The global append-only log grows **unbounded**; compaction/rotation is a **forward item** (out of scope) — noted in `spec.md` Assumptions.
  - Only **executed** (registry-resolved) invocations are recorded; a request for an unregistered tool name aborts the loop with the existing `the tool request failed` condition and records **nothing** (an unavailable name is not a registered tool).
- **Rationale**: keeps the round dependency-free and POSIX-consistent with the whole project, and makes the measurement non-invasive to the operator-facing surface.
- **Alternatives considered**:
  - A third-party logging/rotation library — rejected: violates the stdlib-only, minimal-surface direction.
  - Recording a hallucinated/unavailable tool **name** as an `error` — deferred: it is not a registered tool, so it would break the "every registered tool" roll-up; a forward candidate.
