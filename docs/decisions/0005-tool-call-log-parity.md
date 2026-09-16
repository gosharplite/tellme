# ADR 0005 — Tool-call log parity: per-call frame cadence, injected estimator seam, and rune-safe rendering

- **Status:** Accepted
- **Date:** 2026-09-16
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** round 034 (`specs/plans/034-tool-call-log-parity`); grill round on PR #70 (`architect` ⚔ `griller`, both bootstrapped via `SESSION-BOOTSTRAP.md`);
  round 017 (`presenting-the-turn` chrome); round 019/025 (the spinner + its row-aware clear); round 022 (the single-line `[Tool]` log this round reverses);
  round 027 (the AI-call `<N>` counter); round 028 (`docs/decisions/0004`); round 030 (the failed-turn usage discard this round interacts with);
  round 031 (the `agentTools()` well-formedness gate the seam must not disturb)

## Context

The operator asked tellme to *"show tool calls similar to tell-me-go"*. tellme's round-022 tool log is a single line per call (`[HH:MM:SS] [Tool] <name> - <reason>`), deliberately omitting the call's arguments and result. The reference instead decomposes each call into `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` lines and frames the turn **per AI-endpoint call**.

Round 034 (this ADR's round) re-specifies the rendering. A **grill round** (10 verified questions, subject `architect`, interrogator `griller`) established that the round's own draft spec contained several falsifiable defects, and the corrections below are the load-bearing, cite-worthy decisions that future rounds depend on.

## Problem

Four decisions from round 034 are project-level (other artifacts / future rounds must be able to cite them):

1. **Where the per-call status block is produced.** tellme's loop (`internal/agent`) has no persona/pricing/session-store knowledge, and the CLI can only reproduce call 1's inputs; a per-call frame + tail therefore needs explicit seams.
2. **How the per-call pre-flight payload estimate is computed** for every call after the first.
3. **The per-call cadence's interaction with the trailing answer and with usage persistence.**
4. **Rendering divergences from the reference** in truncation, argument ordering/number formatting, and the `[Tool Output]` bound.

## Decision

**D1 — Two loop observer seams; the CLI stays the only renderer/accounting owner.** The existing `agentport.LoopObserver` seam is extended (and becomes a **composite**, since the spinner is currently the sole observer) with a **call-begin** hook carrying `(callIndex, estimatedPayload)` and a **call-end** hook carrying `(callIndex, usage, roundReasons, measuredPayload, final)`. All rendering, pricing, the session roll-up, and the persistence flush stay in `internal/cli.runTurn`; `internal/agent` gains no persona/pricing/store knowledge.

**D2 — An injected estimator seam.** `AgentLoop` gains `PayloadEstimate func(messages []llm.Message) int` (nil ⇒ no estimate), set in `runTurn` with a closure capturing `res.Person` + `agent.ToolDefs(reg)` (the augmented registry). The loop passes its **fused** `base+turn` slice (not `llm.Request`, so the adapter's prompt-vs-messages rule is not duplicated). Call 1's value is byte-identical to the previous once-per-prompt estimate; calls 2..k grow monotonically.

**D3 — Per-call frame; a per-call tail with the final call deferred past the answer.** The status **frame** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload) is emitted per AI-endpoint call; `N` advances within a prompt (`prior-calls + k`). Each call emits a **tail** (grouped `[Tool Reason]` + measured payload + metrics + `╰─⠿ Ready`) at its end; the **final** call's tail is emitted **after** `writeAnswer` so the closing status still trails the answer.

**D4 — Recorded display-only divergence vs persistence.** Persistence stays **one `AppendBatch` per turn**, gated on the **final** call's `Reported` flag (the one-append-per-call alternative is **rejected** — it would persist calls of a turn that later fails, changing round-030/018 durability). Consequence: on a **failed** turn, or a **completed** turn whose final call reports no usage, the already-shown per-call `Ready` **session** field names totals that are never persisted. This is recorded, not engineered away.

**D5 — Rune-safe truncation caps.** Caps are **maximum total rendered rune lengths** (argument value **189**, result snippet **200**), with exactly one **U+2026** counted **inside** the cap, cut on a **rune boundary**, evaluated on the **folded** value (fold first). Argument values are also folded (the reference does not fold them). This is a **recorded divergence** from the reference's byte slicing + **three ASCII dots** (`[:186]+"..."` = 189 *bytes*; `[:197]+"..."` = 200 *bytes*).

**D6 — Sorted argument keys + `json.Number` rendering.** The `[Tool Action]` argument list removes `reason` first, sorts the remaining keys ascending, and renders values from a `UseNumber()` decode (raw numeric literal — `1000000`, never `1e+06`), strings unquoted, arrays/objects compact. **Recorded divergence** from the reference's Go-map iteration + `%v`-on-`float64`. Ordering is a **presentation** rule only — it never mutates `tc.Arguments` (round-014 replay fidelity).

**D7 — `[Tool Output]` is bounded-and-stopped, not "unbounded".** The shell stream carries every complete line up to the round-024 byte budget, after which the process group is stopped (the existing `trimmed` outcome); `output_file` calls emit **no** block; a trailing partial line is dropped. The spinner is yielded **once per call** (single writer). **Recorded divergence** from the reference's un-terminated `warnWriter`.

## Alternatives considered

1. **Post-hoc re-emission of per-call blocks over `result.Calls` after `loop.Run`** (no live seams) — rejected: `result.Calls` carries no ordering relative to the tool lines, so the frames/tails would file after the answer and break the "each block surrounds its tool activity" contract.
2. **`AgentLoop` owns `EstimatePayload` and gains a persona** — rejected: a third site of the leading-`system`-message rule (the drift the `ToolDefs` reuse exists to prevent).
3. **One tail per prompt** (A5's proposal) — rejected on the evidence: the reference's status middleware wraps every phase processor and `Engine.Run` allocates a fresh Turn per call ⇒ N calls = N tails.
4. **Per-line spinner yield during the output block** — rejected: the sink is invoked from two tool-owned `io.Copy` goroutines while the loop is blocked in `Execute`, and the spinner's clear is positional.
5. **`M = MAX_TOOL_LOOP + 1`** (a call-index denominator) — rejected: it prints a step header for a round the accounting never records; the executed-round unit keeps `1 ≤ i ≤ M` by construction.
6. **Reference byte-slice/`%v` rendering verbatim** — rejected: not rune-safe, and `%v` on `float64` loses numeric fidelity (`1e+06`).

## Consequences

- The rendering is reference-faithful in shape (per-call frames, decomposed lines, live output) while diverging deliberately in **D5/D6/D7** (rune-safe, sorted/`json.Number`, bounded-stopped) — each recorded in `spec.md` and `techstack.md`.
- `internal/agent` stays a use-case (no persona/pricing/store) via D1/D2; `agentTools()` stays parameterless and read-free, so round-031's assembler gate is untouched.
- The recorded **display/persistence divergence** (D4) and the failed-turn **numbering skew** (per-call frames reach `prior + k`, then the next prompt restarts at `prior + 1`) are operator-visible; both are documented, not silently accepted.
- Immutable once `Accepted`; a future change to the cadence, the caps, or the estimator seam supersedes this ADR rather than editing it.
