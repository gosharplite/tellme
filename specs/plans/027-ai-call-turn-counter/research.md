# Phase 0 Research: tellme turn counter counts AI-endpoint calls (round 027)

<!--
  Decision-driven research for round 027. Each decision keeps:
  Decision / Rationale / Alternatives considered.
-->

**Context.** Round 017 gave tellme's non-TUI prompt turn the reference's operator chrome, including the header
`╭─⠿ Turn <N> - <mode>`, where `<N>` = the session's **completed-turn count + 1** (one `history_entry` line
per completed turn; `len(prior)+1` in `internal/cli/cli.go`). The reference (`tell-me-go`) computes its header
from `SessionTurns+1`, where `SessionTurns = History.GetTotalEntries()/2`
(`internal/agent/orchestrator/engine_phases.go`), and — crucially — it makes **each LLM generation cycle its own
`Turn`**: `engine.go`'s `Run` allocates a fresh `Turn` after any tool-calling cycle (`shouldStopRunning =
!HasToolCalls`), and `middleware.go`'s `withStatusReporter` publishes the header at the **start of every**
`Inference` phase (`RetryCount == 0`) and the metrics + `╰─⠿ Ready` at `Persisting`. So the reference shows a
`Turn`/`Ready` pair **per provider call**, and its number counts **model requests**, not user prompts.

tellme keeps a **single** header + `Ready` per prompt (operator-locked: "tellme as-is"), so this round changes
only what `<N>` **counts**: the running **AI-endpoint-call index**. A tool-less prompt advances `N` by one; a
prompt that consults a tool advances it by its number of inference rounds.

**Operator-locked decisions (this session).**
- tellme's chrome **cadence and format stay as-is** (one `╭─⠿ Turn <N> - <mode>` header + one `╰─⠿ Ready` per
  prompt; plain text; no denominator); only the counted **unit** changes.
- `N` = **(prior AI-endpoint calls) + 1** — the index of this prompt's first call, which equals the reference's
  value at that prompt's first call. No `/axb-clarify` was needed: the three decisions that could move a story
  boundary (what the number counts; what a "call" is; keeping the cadence) were settled directly with the operator.

## Decision 1: The counted unit is an inference round (generation cycle), not an HTTP attempt

- **Decision**: "one AI-endpoint call" = one **inference round** of the agent loop — one `Gateway.Complete`
  invocation in `AgentLoop.Run`. A tool-less turn makes **1** call; a turn that runs `k` tool-execution rounds
  makes **1 + k** calls. **Provider-internal retries** (backoff and empty-response retry, which stay inside a
  single round) do **NOT** add to the count.
- **Rationale**: this matches the reference exactly — `withStatusReporter` fires the header only when
  `RetryCount == 0`, so a retry produces no new header and does not advance its number. The loop already makes
  exactly one `Complete` per round (`internal/agent/agentloop.go`), so the unit is available directly (Decision 4).
- **Alternatives considered**:
  - **Count every HTTP attempt** (incl. retries) — rejected: a backoff/empty-response retry would inflate the
    number and diverge from the reference's first-attempt-only counting.
  - **Count user prompts** (the status quo) — rejected: it is the defect this round fixes (it hides the extra
    model calls a tool-using turn makes).

## Decision 2: Persist the per-turn call count on the session-history entry

- **Decision**: add the turn's AI-endpoint-call count to the persisted `history_entry` (a new integer field,
  e.g. `calls`), and compute the header number as **Σ (prior entries' call counts) + 1**. The value written is
  the loop's `len(AgentResult.Calls)` for the completed turn (Decision 4).
- **Rationale**: each prompt is a **separate process**, so the count must be **reconstructable from persisted
  state**. Today `N` derives from `len(prior)` (a *turn* count, which cannot see calls), and neither the entry
  shape `{prompt, answer, steps:[…]}` nor the step shape encodes a call count — the number of tool **steps** is
  not the number of tool **rounds** (a round may batch several tools), so the count cannot be derived. The
  append-only session history is the session's durable record, so the count belongs there. The new field is
  serialized alongside the existing frozen fields (reference `omitempty` discipline keeps any field-less line
  readable — Decision 5).
- **Alternatives considered**:
  - **Derive from the round-018 `tokens.log`** — rejected: it records one line per **reported** call only (a call
    that reports no usage is absent), it is best-effort, and it is a token-usage store, not a call index.
  - **A separate running-total file** — rejected: a second state file needing read-modify-write (races across
    concurrent invocations without `flock`), and `--new` would have to archive it in lockstep; the history entry
    already carries the per-turn granularity.
  - **No persistence (recompute each process)** — rejected: the count would reset to `Turn 1` every invocation.

## Decision 3: The header is emitted before the turn, so it shows this prompt's first-call index

- **Decision**: `N` = prior calls + 1, computed and printed **before** `loop.Run` (the existing chrome order).
  The running total advances by this turn's calls only when the turn **completes and is persisted**.
- **Rationale**: the header is a start-of-turn marker (round 017); it cannot know the turn's own call count
  before the loop runs. `prior + 1` is exactly the value the reference prints at that prompt's first call — the
  parity target (FR-004) — while preserving tellme's single-header cadence (a tool-using prompt still shows one
  header, not one per call).
- **Alternatives considered**:
  - **Print the header after the turn** — rejected: it would reorder the chrome the operator asked to keep
    (the header must open the turn, before the payload line and the answer).
  - **Show the count of completed calls (no `+1`)** — rejected: off by one against the reference's value.

## Decision 4: The loop already exposes the call count — no new loop seam

- **Decision**: reuse the existing `AgentResult.Calls` (`[]llm.Usage`, appended once per `Gateway.Complete` in
  `Run`) — `len(result.Calls)` is the turn's AI-endpoint-call count. No new field, port, or observer callback.
- **Rationale**: `Calls` was added in round 018 ("EVERY provider call's usage for the turn, in call order") and is
  appended immediately after each successful `Complete`, so it already equals the generation-cycle count. A
  completed turn is persisted only on success, and on a successful run `Calls` holds every round — so `len(Calls)`
  is the count to store.
- **Alternatives considered**:
  - **A dedicated `CallCount int` on `AgentResult`** — rejected: redundant with `Calls`.
  - **An observer callback per round** — rejected: a heavier seam (like the round-019 `LoopObserver`) for one integer.

## Decision 5: Session-scoped, reset by `--new`; legacy entries count as one

- **Decision**: the number = Σ over the **active session's** entries; `--new` archives the active history, so the
  sum restarts at 0 (the fresh first turn reads `Turn 1`). A persisted entry **without** a stored call count
  (written before this round) counts as **1**.
- **Rationale**: matches the reference, whose session turn count restarts with the archived history. A legacy
  entry's calls are unrecoverable from the old line; `1` is the floor (a tool-less turn) and keeps a resumed
  pre-027 session from failing. It is a recorded limitation, not a hidden assumption.
- **Alternatives considered**:
  - **A lifetime counter** — rejected: diverges from the reference (its number is session-scoped).
  - **Treat a legacy entry with steps as 2** — rejected: a guess; steps ≠ rounds, so it would still undercount a
    multi-round turn, and it would mis-price a legacy tool-less turn.

## Decision 6: No new dependency; POSIX-only; surface unchanged; recorded divergences

- **Decision**: the round uses only the Go standard library; it is **POSIX-only**; and it changes only the header
  **number** — `stdout` stays byte-exact, and the frozen class-phrase vocabulary, the pre-flight / measured
  payload lines, the post-turn metrics + `╰─⠿ Ready`, the spinner, the tool-loop log line, and the tool
  set/semantics are **unchanged**.
- **Rationale**: keeps the round dependency-free, POSIX-consistent, and non-invasive to every other operator
  surface — the operator asked to keep tellme as-is except for the number's unit.
- **Recorded divergences / forward items**:
  - tellme shows **one** header/`Ready` per prompt (the reference shows one per provider call) — an **accepted**
    operator decision; only the number's unit aligns with the reference.
  - **Legacy entries** without a stored call count count as 1 (Decision 5) — a resumed pre-027 session may
    undercount its tool-using turns.
  - A **failed** turn (provider/transport error, or the tool-loop bound reached) persists **nothing**, so the
    counter is unchanged and the next prompt repeats the number — matching the reference, whose failed `Turn`
    does not advance `SessionTurns`.
  - A future **summarisation/archive** path that drops entries must preserve the counter (the round-017 review
    finding 3 forward note); today no mid-session archive shrinks the active history, so `Σ entries` is stable.

- **Alternatives considered**:
  - **Adopt the reference's per-call chrome cadence** (a header/`Ready` per call) — rejected: the operator
    explicitly asked to keep tellme's "as-is" presentation; only the counter's unit was to change.
  - **Show a `/MAX_HISTORY_TURNS` denominator** (reference parity) — rejected/out of scope: tellme's header has no
    denominator (round 017), and the operator's request was limited to the counter's value.

## Residual risks / forward items

- **Legacy-entry floor.** A pre-027 resumed session may undercount its tool-using turns (each legacy entry counts
  as 1). Acceptable for a dev-stage tool; a one-time migration is not warranted.
- **Failed-turn repeat.** A turn that fails after its header is printed leaves the number unchanged, so the next
  prompt repeats it. Correct by contract (nothing persisted), and it matches the reference.
- **Future archive path.** If a later round introduces mid-session summarisation/archival that removes entries,
  the `Σ entries` computation must be revisited so the counter does not shrink (round-017 review finding 3).
