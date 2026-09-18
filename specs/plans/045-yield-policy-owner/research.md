# Technical Research: yield-policy owner + `LoopObserver` hook split (round 045)

**Plan Package**: `specs/plans/045-yield-policy-owner`
**Anchor**: [#105](https://github.com/gosharplite/tellme/issues/105) — R3 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> RD-side technical decisions. Every decision below is a **technology/structure choice**; the *what* lives in `spec.md`. Decisions are grounded in a static read of `dev` @ `8da0b88`.

## Context

The spinner-yield **policy** is scattered across three homes (H1 the loop's `withToolLog`; H2 the composite observer's `yieldIndicatorBeforeTail`; H3 the `internal/ui` `ToolOutputCoordinator`) reached by four call sites over **three different routes**, and the `agentport.LoopObserver` port names its yield hooks after a tool-log write (`BeforeToolLog`/`AfterToolLog`) even though one of the four uses is not a log write. The **mechanism** (`Spinner.deactivate()/resume()/AdmitResume()`; one I/O mutex; a synchronous goroutine-joined clear; epoch-preserving resume; the round-025 row-aware clear) is correct and stays. The round gives the **policy** one named owner and splits the port's yield pair — **behaviour-preserving**. The witness is **unit ordering pins + ADR 0014** (the E2E suite cannot observe a clear/resume ordering).

### D1 — The mechanism stays; only the policy moves

`internal/ui/spinner.go` keeps `deactivate()` / `resume()` / `AdmitResume()` exactly as they are (round-019 D10, round-040 ADR 0009 D4). R3 does **not** touch: the one I/O mutex, the synchronous joined clear, the epoch fields (`epoch`/`callEpoch`), the row-aware clear (`lastRows`/`eraseRows`/`rowsForLine`), the dual-timer formatting, or `Stop()`'s idempotent teardown. A move of any of these would be a behaviour change, which `spec.md` US3 forbids. **R3 is a policy consolidation, not a mechanism change.**

### D2 — The owner: `internal/ui/yield.go` → `YieldController` (C-R3-1)

A new file `internal/ui/yield.go` defines the **single named owner**:

```go
type YieldController struct{ sp *Spinner }

func NewYieldController(sp *Spinner) YieldController
func (y YieldController) Enabled() bool   // false when the spinner is gated off (nil)
func (y YieldController) Yield()          // clear-only (deactivate)
func (y YieldController) Restore()        // resume (synchronous first frame)
func (y YieldController) Admit()          // resume with a goroutine-drawn first frame (WS-A)
```

- The **policy** is stated **once** in the type's doc comment: *a clear is clear-only and never resumes; restoration is a phase-boundary act (the next waiting phase, the block's `End`, or the idle-gap `Admit`) — never mid-write*.
- A **nil** spinner is allowed and makes every method a no-op (`Enabled()` false) — the gated-off case.
- The type is **thin by design**: it owns the *rule*, not the *mechanism*. The mechanism methods on `Spinner` become **unexported details** that only `YieldController` (and the round-040 `AdmitResume` unit pin) call.

**Why `internal/ui` and not `internal/cli`**: `internal/ui` is the layer that owns presentation and is where H3 already lives; a `strict de-coupling` round ([#101](https://github.com/gosharplite/tellme/issues/101)) would eventually move the port itself. Rejected: a port-level owner in `internal/cli` (the composite is a *client*) and generalising `ToolOutputCoordinator` into a turn-scoped "presentation coordinator" (a larger rewrite of the block with no R3 benefit).

### D3 — The hook split: replace `BeforeToolLog`/`AfterToolLog` (C-R3-2/C-R3-3)

`agentport.LoopObserver`:

```go
YieldIndicator()   // clear-only: the loop is about to write to the diagnostic stream
RestoreIndicator() // resume after a yield (typically at a phase boundary)
```

- The **loop** keeps calling the port (C-R3-2 option A): `withToolLog` calls `Observer.YieldIndicator()` / `Observer.RestoreIndicator()` — behaviour-preserving **by construction** (same clear-before/resume-after).
- The **composite observer** forwards both new hooks to its spinner half **and** routes its own tail yield through `YieldIndicator()` — so H1 (via the composite's forward) and H2 (via its own call) now reach the spinner by **one method**, removing H2's reach-through to a tool-log name.
- **No alias pair is kept** (C-R3-3): the split's purpose is to remove the *line-about-to-be-written* conflation; keeping four names with two meanings re-introduces the overload. The port's doc comment describes the loop's **route** and **defers the policy** to `ui.YieldController`.

**Rejected (C-R3-2 option B)**: the loop stops yielding and the presenter owns the wrapping. It is a bigger rewrite (the loop would emit a completion event and the CLI would bracket its writes) and overlaps **R4** (`internal/agent`'s `internal/ui` coupling) / [#101](https://github.com/gosharplite/tellme/issues/101).

### D4 — The coordinator holds the owner (H3), not the bare spinner

`ToolOutputCoordinator` changes `sp *Spinner` → `yc YieldController` (built from the spinner in `NewToolOutputCoordinator`). Its three yield calls become `c.yc.Yield()` (`Begin` + `clearIndicator` + the `EndWith`/`WriteWith` hooks), `c.yc.Restore()` (`End`), and `c.yc.Admit()` (the watcher). Its `sp != nil` checks become `c.yc.Enabled()`. **The block's writer, watcher, lock order, and mutual-exclusion+join invariant are untouched** — only the yield *route* changes and `Stop()` (a teardown) is replaced by `Yield()` (a yield) where the intent is a mid-turn clear (behaviour-identical: both call `deactivate()`).

### D5 — Governance: ADR 0014 (the yield-policy owner + the split)

A project-level rule future rounds (R4, #101) must cite: the yield policy's single owner, the split vocabulary, and the `Stop()` (teardown) vs `Yield()` (yield) distinction. **ADR 0005 D1 is amended by reference** — 0005 partitioned *rendering*, not *yield*, so its body is **not edited**; only its `Status`/index row may annotate the amendment (the round-040 **D7 → ADR 0009** precedent). No ADR is superseded. Follows the ADR-0011/0013 lineage.

### D6 — Interface posture: api/data/dsl NOOP

- `/axb-api-plan` = **NOOP** — no HTTP/OpenAPI surface.
- `/axb-data-plan` = **NOOP** — no persisted/runtime state change.
- `/axb-dsl-refine` = **NOOP** — no user-facing CLI contract change; the vocabulary stays **11**; the topology audit is unchanged. **Evidence for the NOOP:** the truth tree was grepped for the old hook names (`BeforeToolLog`/`AfterToolLog`) — they survive only in *explanatory notes* (the round-035/036 notes in `chat/dsl.md`), never in a `DSLRow` or a step, so no truth row goes stale (the `#92` audit blind spot is checked, not trusted).
- `/axb-ui-plan` **skipped** (not user-facing).

### D7 — Verification strategy (the witness is the unit pins, not the suite)

- `make verify` (incl. `verify-architecture` **0 new / 0 stale**, baseline still **1**; `verify-mcp-sdk-confinement`; 4/4 cross-compile; lint 0; govulncheck) · `go test -count=1 ./...` green · **`go test -race ./internal/ui/...`** green (the coordinator stress is the anti-vacuity guard for H3) · topology audit unchanged · `stdout` byte-exactness unchanged.
- **Unit ordering pins** (SC-003): extend the `internal/cli/composite_observer_test.go` recorded-sequence harness — `[spinner.YieldIndicator, call.OnCallEnd]` for the non-final tail (Y2); `[call.OnCallEnd]` for the final (Y2 negative); and the nil-safe no-op. `internal/ui/coordinator_test.go` (Y3/Y4) stays green unchanged; `internal/ui/spinner_test.go` pins the clear/resume continuation (Y1 arm).
- **Falsifiability witnesses** (SC-007, reproduced then reverted, ADR 0010): (a) drop the Y2 clear ⇒ the ordering pin fails; (b) resume after the Y2 tail ⇒ the residue pin fails — the *reproduced* round-035 result, where clear+resume **relocates** the residue rather than removing it (so "no resume" is a measured outcome, not a preference); (c) make the owner self-locking ⇒ the coordinator stress reds.

### D8 — Scope guard

Touch only: `internal/domain/agent/observer.go`, `internal/ui/{yield.go (new), spinner.go, coordinator.go}`, `internal/cli/composite_observer.go`, `internal/agent/agentloop.go`, the affected `_test.go` files, `docs/decisions/**`, `specs/truth/techstack.md`, and the plan package. **No** adapter behaviour change, **no** domain business-logic change, **no** flag/exit-code/stream change. The strict de-coupling stays out → [#101](https://github.com/gosharplite/tellme/issues/101).

### D9 — Risks & mitigations

| Risk | Mitigation |
| --- | --- |
| A behaviour drift in the yield ordering | the unit ordering pins (D7) + full E2E + `stdout` byte-exactness |
| The coordinator's resume/clear invariant regresses (lock order / join) | `go test -race ./internal/ui/...`; the coordinator body is unchanged except the *route* (D4) |
| The port rename leaves a stale truth row | grep the truth tree for the old names (D6) and record the evidence in `truth-delta.md` |
| A future yield primitive re-scatters the policy | ADR 0014 names `YieldController` as the single entry point; recorded as an Edge Case |
| Collapsing `Stop()` (teardown) into a yield | `Stop()` stays the turn teardown; the coordinator uses `Yield()`; recorded as an Edge Case |
| The baseline accidentally regresses to 0 (lifting R4's DoD) or gains a violation | the gate must report **0 new / 0 stale** with the baseline still `1` (C-R3-4 / FR-008) |

## Truth impact (summary)

`specs/truth/techstack.md` **MODIFY** — the **Agent tool loop** row (the yield hook vocabulary) and the **Turn progress spinner** row (the single-owned yield policy) — recorded in `truth-delta.md`. `contracts/**` + `data/**` + `features/**` **NOOP**. New **ADR 0014** + index row.
