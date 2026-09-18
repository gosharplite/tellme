# ADR 0021 — Ride-alongs: the command tool's construction-time output sink + the suggester's single-owned selection policy; the durable home of the three `#92` records

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Related:** issues [#115](https://github.com/gosharplite/tellme/issues/115) (the two ride-alongs) and [#116](https://github.com/gosharplite/tellme/issues/116) (the three records);
  [ADR 0005](0005-tool-call-log-parity.md) (**D7** — the `[Tool Output]` pause/resume; superseded by 0009 for the pause only), [ADR 0009](0009-spinner-dual-timer-and-streaming-liveness.md) (the idle-gap resume), [ADR 0013](0013-composition-root-injection.md) (the injected `Dependencies` seam), [ADR 0020](0020-cli-ui-decoupling.md) (**F-8** made `domaintools.OutputSink` an interface; the `→ ui` de-coupling);
  round 034 (the `BindToolOutput` seam's origin, PR [#71](https://github.com/gosharplite/tellme/pull/71)), round 037 (the no-choice selection policy), round 040 (the coordinator's one-concurrent-block assumption, PR [#86](https://github.com/gosharplite/tellme/pull/86)), round 051 (the round-051 closeout split [#115](https://github.com/gosharplite/tellme/issues/115)/[#116](https://github.com/gosharplite/tellme/issues/116));
  round 052 (`specs/plans/052-ride-alongs-and-records` — this ADR's round)

## Context

At the round-051 closeout, [#92](https://github.com/gosharplite/tellme/issues/92) was split so it could close on its main theme (R1–R5 delivered). Two residuals moved to [#115](https://github.com/gosharplite/tellme/issues/115) — the **ride-alongs** — and three standing notes moved to [#116](https://github.com/gosharplite/tellme/issues/116) — the **records**. Round 052 is the round that closes both.

**R-2 (ledger #6).** The `execute_command` tool's `[Tool Output]` sink is bound **post-construction**: `NewCommandTool()` returns `executeCommand{output: &toolOutputBox{}}`, and `BindToolOutput(reg, sink)` mutates the registry's stored (value) tool through the shared pointer box (round 034). The dependency is therefore **invisible in the constructor** and set by a registry-rebind step. Since round 051 (**F-8**) `domaintools.OutputSink` is an **interface** the `ui.ToolOutputCoordinator` satisfies directly, so a constructor parameter can carry it.

**R-1 (ledger #5).** `suggester.set(items []string)` resets `s.cursor = noChoice` **internally** — a hidden side effect, repeated in three places (`newSuggester`, `set`, `cycle`) rather than owned by the caller.

**The records.** [#116](https://github.com/gosharplite/tellme/issues/116) holds three records: a permanent E2E narrowing (round 036 D4), the coordinator's one-concurrent-block assumption (round 040), and an `End`-while-write-stalled accepted residual. Its body states it is *"closed-by-intent only when each record is either resolved or relocated"* — and "resolve" is foreclosed for all three (permanent / declined-round-gated / bounded-then-accepted). Their current homes are **frozen** plan packages (`036-*`, `040-*`) or an un-indexed code comment, so none is a citable, non-frozen home.

## Terminology

- **Construction seam** — a dependency handed to an object through its constructor, not mutated in afterwards.
- **Registry-rebind step** — the round-034 pattern of setting a stored tool's sink *after* construction via a shared pointer box.
- **Record** — a settled standing fact with no code change expected (as opposed to *work*).

## Decision

**D1 — The command tool receives its sink at construction, through one sink-aware registry builder.** `NewCommandTool(sink domaintools.OutputSink) domaintools.Tool` is the constructor; `executeCommand` holds the sink directly (no pointer box). The agent registry is built by **one** parameterized builder:

```go
func agentTools() []domaintools.Tool { return assembleAgentTools(nil) }   // parameterless + read-free (round-033 FR-009)
func assembleAgentTools(sink domaintools.OutputSink) []domaintools.Tool { … NewCommandTool(sink) … }
func newToolRegistry(sink domaintools.OutputSink) domaintools.Registry  { return domaintools.NewRegistry(assembleAgentTools(sink)...) }
```

The `deps.Dependencies` field widens to `NewToolRegistry func(sink domaintools.OutputSink) domaintools.Registry`; the **prompt path** builds the progress object first and injects (`reg := dp.NewToolRegistry(prog.ToolOutput)`); the **offline** `--tool-usage` path passes **`nil`**. `agentTools()` stays **parameterless and read-free**, so the round-031 assembler well-formedness gate iterates it unchanged (its only edit is the sibling call `newToolRegistry(nil)`).

*Rejected:* (a) a **lazy sink seam** (mirror the round-033 catalog seam) — it keeps the indirection the ride-along removes; (b) a **second registry seam** — a duplicate builder and a drift hazard; (c) rebasing the assembler into `internal/cli` — forbidden (composition root, ADR 0013). *Precedent:* round 051 / ADR 0020 **F-7** reshaped the `MCPDiscoverer` `Dependencies` field inside a behaviour-preserving round.

**D2 — `BindToolOutput` and the pointer box are deleted.** Removed: the `toolOutputBox` type, `executeCommand`'s `sink()` accessor, `infratools.BindToolOutput`, the `deps.Dependencies.BindToolOutput` seam, and the `cli.go` rebind call. Kept: `sinkOpen` / `sinkClose` / `teeSink` (already nil-guarded), so a nil sink — the gate's and the offline assembler's — stays a no-op and the block renders unconditionally on the prompt path exactly as before (rounds 034/038/039/040 unchanged).

**D3 — The suggester's selection policy is owned by the caller.** `suggester.set(items []string, cursor int)` sets both fields and no longer resets the cursor; the two in-package callers pass `noChoice`, preserving the round-037 policy (*a refresh opens unselected*) as a **caller** decision. `newSuggester` keeps its construction default (`cursor: noChoice`); `cycle` / `selected` / `view` are untouched; `noChoice` stays package-private. The invariant `cursor ∈ {noChoice} ∪ [0, len(items))` is documented once, at `set`. *Rejected:* exporting `NoChoice` (no external caller); a `reset()` helper (redundant).

**D4 — The three `#116` records get a durable home in this ADR's §Records**, and [#116](https://github.com/gosharplite/tellme/issues/116) closes **completed**. The records are **copied** here (never moved); the frozen `036-*` / `040-*` packages are not edited; the in-code pins stay in place.

## Consequences

- The command tool's dependency is visible in its constructor; one registry builder exists; the assembler gate is unchanged in shape.
- One `Dependencies` field's **signature** changes (func-typed, so `Validate()` still covers it); the offline reporting path and the test fixtures pass `nil`.
- The suggester's reset is explicit at the call sites; the round-037 rendering is byte-identical.
- The three records are citable from a live, indexed artifact and survive the frozen packages.
- **Unchanged:** `stdout`/`stderr` byte-contracts, exit codes, flags, the `[Tool …]` literals (incl. `[Tool Output]`), the DSL vocabulary, the dependency set (`go.mod`/`go.sum`), the layer-discipline baseline (RULE-E stays **0**).

## Records (relocated from [#116](https://github.com/gosharplite/tellme/issues/116) at the round-052 close)

> **Status of these three records:** *standing records, not work.* Relocated here (from the frozen plan packages `specs/plans/036-*` / `040-*` and the `internal/ui/coordinator.go` doc comment) so they outlive the packages and are never re-raised as "we forgot". [#116](https://github.com/gosharplite/tellme/issues/116) closes on this relocation.

### Record 1 — Permanent E2E narrowing (round 036 D4)

The `\n`/`\r` **tool-reason row-collision** class has **no E2E carrier**. Its deterministic carrier is the **hostile-fixture unit pin** in `internal/ui/toolcall_reason_test.go`. This is a **permanent narrowing, not a deferral** — recorded so it is never re-raised as a missing E2E. *(Provenance: round 036 `research.md` D4; the round-036 `truth-delta.md` `chat/dsl.md` note; the live pin file.)*

### Record 2 — The coordinator models ONE concurrent block (round 040, PR [#86](https://github.com/gosharplite/tellme/pull/86) §5)

The `internal/ui` `ToolOutputCoordinator` assumes **one** open `[Tool Output]` block (one writer + one idle watcher). A future **concurrent-tools** round (currently **declined** — [#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`) would have to re-scope it to per-block state or serialize blocks; `compositeObserver.OnCallEnd` resumes unconditionally, which would violate the FR-002 idle-gap invariant with a second block open. Recorded so such a round reads it first. *(Provenance: round 040 / PR [#86](https://github.com/gosharplite/tellme/pull/86) §5; ADR 0009 D3/D4.)*

### Record 3 — `End`-while-write-stalled accepted residual

A wedged `stderr` write cannot be abandoned at `End` (documented in `internal/ui/coordinator.go`): `stopWatcher` joins the watcher, which may itself be parked inside `w.mu`, so nothing can abandon a blocked underlying write. **Bounded-then-accepted, not repaired** — the stream is assumed to make progress (a real terminal/pipe does). The unit stress `TestCoordinatorEndWhileLineWriteInFlight` exercises the **released** writer (the recoverable shape). *(Provenance: `internal/ui/coordinator.go`; the unit stress test.)*

## Forward

- **RF-52-1** — the widened `NewToolRegistry` seam takes a single sink; if a future round needs the registry built with a *different* per-turn dependency (e.g. a per-turn skills catalog), prefer adding a second **named** per-path seam over a growing positional parameter list (the round-051 **RF-51-6** warning).
- **RF-52-2** — Record 1's narrowing: a count formulation covering **both** `\n` and `\r` (neither existing formulation does) would let a future round add an E2E carrier; until then the unit pin is the deterministic carrier.
- **RF-52-3** — Record 2's re-scope remains gated on a concurrent-tools round ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
