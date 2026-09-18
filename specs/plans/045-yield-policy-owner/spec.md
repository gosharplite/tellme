# Feature Specification: yield-policy owner + `LoopObserver` hook split (round 045)

**Feature Branch**: `045-yield-policy-owner`

**Created**: 2026-09-18

**Status**: Draft — full round (specify → clarify → spec-by-example → technical-research → system-analysis → tasks → implement). Anchor issue [#105](https://github.com/gosharplite/tellme/issues/105) — **R3** of [#92](https://github.com/gosharplite/tellme/issues/92). Clarify round 1 locked **C-R3-1 … C-R3-6** (below).

**Input**: Issue [#105](https://github.com/gosharplite/tellme/issues/105) (R3 of the #92 gate-first split), resolved via the AIxBDD process. #105: *"the spinner-yield policy is scattered … three different packages each decide, locally, when to clear and whether to resume, and each carries its own rationale in a code comment"* and *"`agentport.LoopObserver` … conflates three concepts: phase started, line about to be written, and indicator must yield."* R1 ([#93](https://github.com/gosharplite/tellme/issues/93), round 042, ADR 0011) shipped the layer-discipline gate; R2 ([#100](https://github.com/gosharplite/tellme/issues/100), round 044, ADR 0013) moved the composition root and ratcheted the baseline **8 → 1**. R3 owns the **yield axis**; the **remaining** baseline entry (`internal/agent -> internal/ui`) is **R4**'s (the loop's `internal/ui` formatting + predicate coupling), not R3's.

**Behaviour intent**: **MODIFY (structural — behaviour-preserving)**. Give the spinner-yield policy **one named owner** and split the overloaded `LoopObserver` tool-log hooks into intent-named pairs, with **zero** change to any flag, exit code, stream contract, formatting, or the DSL vocabulary. The round's witness is **unit ordering pins + the new ADR** (the E2E suite cannot observe a clear/resume ordering — a green suite alone is false confidence, the round-009 trap).

---

## Locked decisions (clarify round 1 — C-R3-1 … C-R3-6)

> Locked one at a time from the six candidates enumerated on [#105](https://github.com/gosharplite/tellme/issues/105). The issue states the **problem**; the mechanism is chosen here.

| # | Decision |
| --- | --- |
| **C-R3-1 → owner (1)** | The yield policy gets **one named owner**: a new `internal/ui` type **`YieldController`** (`internal/ui/yield.go`) — a thin, single-purpose value over the spinner's yield mechanism with the policy stated **once** in its doc comment. It is the layer that owns presentation and where H3 already lives. Rejected: a port-level owner in `internal/cli` (the composite is a *client*, not a home) and generalising `ToolOutputCoordinator` into a turn-scoped "presentation coordinator" (a larger rewrite of the block, with no R3 benefit). |
| **C-R3-2 → (A)** | The **loop keeps calling the port** (`YieldIndicator()`/`RestoreIndicator()` around its diagnostic writes). Option B (the presenter owns the wrapping; the loop stops yielding) is a bigger rewrite and overlaps **R4** / [#101](https://github.com/gosharplite/tellme/issues/101). Option A is behaviour-preserving **by construction**. |
| **C-R3-3 → replace (not alias)** | The port's `BeforeToolLog()`/`AfterToolLog()` are **replaced** by `YieldIndicator()` (clear-only) / `RestoreIndicator()` (resume) — the split's whole point is to remove the *line-about-to-be-written* conflation, so **no alias pair is kept** (keeping four names with two meanings would re-introduce the overload the split removes). This is the #105 *Goal* wording verbatim: *"the overloaded `BeforeToolLog`/`AfterToolLog` pair is replaced by intent-named hooks"*. Doc comments carry the new intent; the yield **policy** lives with `YieldController` (C-R3-1), so the port states only the loop's *route*, never the rule. |
| **C-R3-4 → baseline stays 1** | R3 does **not** remove the `internal/agent -> internal/ui` baseline entry — that is **R4**'s DoD (1 → 0). The loop's `internal/ui` import is driven by the **four formatters** (`FormatToolEngine/Reason/Action/Result`) **and** the predicate (`ui.ToolReasonRenders`), **not** by the yield port. R3 removes the *yield* rationale from three homes; it does not touch the loop's formatting/predicate coupling. Recorded explicitly so the ratchet does not silently regress. |
| **C-R3-5 → unit pins only** | The witness is unit **ordering** pins (the `internal/cli/composite_observer_test.go` recorded-sequence harness, extended to `internal/ui`) — **no new E2E Example**. The E2E suite cannot observe a clear/resume ordering (a flat byte capture has no terminal grid); a new Example would be a **vacuous carrier**. |
| **C-R3-6 → records stay records** | The coordinator's `End`-while-write-stalled accepted residual and the one-concurrent-block record stay **recorded, not fixed** (both are `#92` records; neither is R3 work). |

---

## Grounded in the current system

Static read 2026-09-18 @ `dev` `8da0b88`.

### The 3 homes (each decides, locally, *when* to clear and *whether* to resume)

| # | Home | Decision | Rationale today lives in |
| --- | --- | --- | --- |
| **H1** | `internal/agent/agentloop.go` → `withToolLog` (the loop) | clear **before** and resume **after** every `[Tool …]` diagnostic write | the `withToolLog` doc comment |
| **H2** | `internal/cli/composite_observer.go` → `yieldIndicatorBeforeTail` (the composite observer) | on a **non-final** `OnCallEnd`, clear **and never resume** (a *phase-boundary* yield) | the `OnCallEnd` + `yieldIndicatorBeforeTail` doc comments (round 035) |
| **H3** | `internal/ui/coordinator.go` → `ToolOutputCoordinator` | per-line **clear** inside the block; **idle-gap resume** via the watcher; `End` = stop+join watcher → clear → separator → resume | the `coordinator.go` header + `End`/`watch`/`clearIndicator` doc comments (round 040) |

### The 4 call sites (three *different* routes to the mechanism)

| # | Call site | Shape |
| --- | --- | --- |
| **Y1** | `agentloop.go` `withToolLog` — `BeforeToolLog()` … write … `AfterToolLog()` | clear + resume (paired) — via the **port** |
| **Y2** | `composite_observer.go` `OnCallEnd` → `yieldIndicatorBeforeTail` → `c.spinner.BeforeToolLog()` | clear, **no** resume — a **direct** reach-through (the port was bypassed) |
| **Y3** | `coordinator.go` `clearIndicator` → `c.sp.Stop()`; per output line via `WriteWith(beforeLine)` | clear (synchronous, goroutine-joined) — a **third** route |
| **Y4** | `coordinator.go` `End` → `c.sp.AfterToolLog()`; watcher → `c.sp.AdmitResume()` | resume (two flavours: synchronous first frame; goroutine-drawn first frame) |

**The port conflates three concepts**: *a phase started*, *a line is about to be written*, *the indicator must yield*. H2 needed a yield that is **not** a tool-log write, found no vocabulary, and reached into the spinner directly. The hook names (`BeforeToolLog`/`AfterToolLog`) now lie about two of their four uses.

The mechanism — `Spinner.deactivate()`/`resume()`/`AdmitResume()`, one I/O mutex, a **synchronous, goroutine-joined** clear, epoch-preserving resume — is correct and **stays in `internal/ui/spinner.go`** (round-019 D10 / round-040 ADR 0009 D4). What has no owner is the **policy**.

### Invariants that must survive (round-040 review TD-3 / R-11)

- **Lock order** = block-writer mutex → spinner mutex; **mutual exclusion + join**; the block critical section **never spans a frame write** (the WS-A resume draws its first frame on the redraw goroutine).
- **No self-locking yield accessor** inside a block critical section (a self-locking `IdleSince` was exactly what round-040 R-11 caught).
- **Hermetic** — no pty, no `time.Sleep`, the ~200 ms cadence is an **injected ticker**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the yield policy has one named owner (Priority: P1)

As a maintainer, I want the spinner-yield **policy** to live in exactly one named place, so H1/H2/H3 stop each holding a private copy of the rule.

**Why this priority**: it is the round's reason to exist; the "3 homes → 1" is the ledger item.

**Independent verification**: `internal/ui/yield.go` defines `YieldController`, the sole place the policy is stated; every yield in the turn path routes through it (H1 via the port adapter on `*Spinner`, H2 via the same, H3 via the coordinator's `YieldController`). No yield *rationale* comment survives in more than one of H1/H2/H3.

**Acceptance Scenarios**:

1. **Given** the round head, **When** the yield routes are enumerated, **Then** every clear and every resume passes through `ui.YieldController` (`Yield()` / `Restore()` / `Admit()`), and H1/H2/H3 contain no yield decision beyond *when* to call it.
2. **Given** `Spinner.Stop()` (turn teardown) vs `YieldController.Yield()` (a yield), **When** the coordinator wants to clear mid-turn, **Then** it calls the owner (`Yield()`), not the teardown or a low-level primitive.

**Functional Requirements**:

- **FR-001**: A new `internal/ui/yield.go` MUST define the **single named owner** of the yield policy — `YieldController` — exposing `Yield()` (clear-only), `Restore()` (resume), and `Admit()` (goroutine-drawn resume), and MUST state the policy **once** in its doc comment (when a clear is permitted, that a clear never resumes, that restoration is a phase-boundary act).
- **FR-002**: `ToolOutputCoordinator` MUST route its clears/resumes/admits through the owner (holding a `YieldController`, not a bare `*Spinner`); its block logic (writer, watcher, lock order) is unchanged.
- **FR-003**: The `Spinner`'s port-facing yield methods MUST adapt to the owner (delegate to `YieldController`) rather than implementing a second policy; `Spinner.Stop()` MUST remain the **turn-teardown** clear (not a yield) and MUST NOT be used by the coordinator as a mid-turn clear.

**Non-Functional Requirements**:

- **NFR-001**: The consolidation MUST be **behaviour-preserving** — every yield ordering is byte-identical; the mechanism (synchronous joined clear, epoch-preserving resume, row-aware clear) is untouched.

### User Story 2 - the `LoopObserver` hook split (Priority: P2)

As a maintainer, I want the observer port to expose yield intent by name, so no caller uses a *tool-log* hook to mean a non-log yield.

**Why this priority**: it is the *mechanism* that removes H2's reach-through and the port's conflation — the second half of the coupled pair.

**Independent verification**: `agentport.LoopObserver` declares `YieldIndicator()` + `RestoreIndicator()` and **no** `BeforeToolLog`/`AfterToolLog`; `compositeObserver`'s tail yield calls `YieldIndicator()` (the same method the loop calls) instead of a tool-log-named hook.

**Acceptance Scenarios**:

1. **Given** the split port, **When** the loop writes a tool-log line, **Then** it calls `YieldIndicator()` before and `RestoreIndicator()` after (Y1 unchanged in *shape*).
2. **Given** the composite's non-final `OnCallEnd`, **When** it yields the tail, **Then** it calls `YieldIndicator()` and never `RestoreIndicator()` (Y2: clear, no resume).
3. **Given** any caller, **When** it yields, **Then** no tool-log-named hook exists to misuse (the pair is gone).

**Functional Requirements**:

- **FR-004**: `agentport.LoopObserver` MUST replace `BeforeToolLog()`/`AfterToolLog()` with `YieldIndicator()` (clear-only) and `RestoreIndicator()` (resume); its doc comment MUST describe the loop's **route** and defer the **policy** to `ui.YieldController`.
- **FR-005**: `compositeObserver` MUST implement `YieldIndicator()`/`RestoreIndicator()` and MUST route both the port forward **and** its own tail yield through the spinner's `YieldIndicator()` (one route, no reach-through to a tool-log name).
- **FR-006**: `agentloop.withToolLog` MUST call `Observer.YieldIndicator()` / `Observer.RestoreIndicator()`; every other observer hook is unchanged.

**Non-Functional Requirements**:

- **NFR-002**: The port stays **network-free and presentation-free** (`internal/agent` imports no `internal/ui` for the **yield**; it stays persona/pricing/store-free per ADR 0005 D1/D2).

### User Story 3 - zero behavioural change (Priority: P3)

As a maintainer, I want the round to be a **pure structural refactor**: under the same inputs every flag, exit code, stream byte, and formatting outcome is identical, and the full regression is green.

**Why this priority**: it bounds the round to structure and protects the #92 "no user-facing change" constraint.

**Independent verification**: `make verify` green (incl. the R1 layer-discipline gate: **0 new / 0 stale**, baseline still **1**); godog E2E green; `stdout` byte-exactness unchanged; DSL vocabulary **11**; topology audit unchanged.

**Acceptance Scenarios**:

1. **Given** a supported prompt path, **When** the round head runs it, **Then** `stdout` is byte-identical and every exit code is unchanged.
2. **Given** the full `make verify` + E2E, **When** it runs, **Then** it is green (no behavioural/stream-contract regression).
3. **Given** the R1 gate, **When** it evaluates the module graph, **Then** it reports **0 new, 0 stale** with the baseline still `internal/agent -> internal/ui` (C-R3-4).

**Functional Requirements**:

- **FR-007**: The round MUST NOT change any user-facing CLI behaviour: no flag, exit code, stream contract, formatting, DSL row (vocabulary stays **11**), or stored record.
- **FR-008**: The round MUST NOT remove the `internal/agent -> internal/ui` baseline entry (R4's DoD) and MUST NOT add a new violation; the R1 gate stays green with **0 new / 0 stale**.
- **FR-009**: The round's witness MUST be **unit ordering pins + the ADR**, NOT the E2E suite; it MUST include **falsifiability witnesses**: (a) drop the Y2 clear ⇒ the ordering pin fails; (b) resume after the Y2 tail ⇒ the residue pin fails (the *reproduced* round-035 result — clear+resume **relocates** the residue); (c) make the yield owner self-locking ⇒ the coordinator stress reds. Each reproduced **then reverted** (ADR 0010).

**Non-Functional Requirements**:

- **NFR-003**: `make verify` (incl. `verify-architecture`, `verify-mcp-sdk-confinement`, 4/4 cross-compile, lint, govulncheck), `go test -count=1 ./...`, `go test -race ./internal/ui/...`, and the Gherkin/DSL topology audit MUST be green; **no** new Gherkin/DSL row (`/axb-dsl-refine` **NOOP**).

---

## Requirements *(mandatory)*

### Global Requirements

#### Functional Requirements

- **FR-010**: The round MUST record the yield-policy ownership + the hook split in a new **ADR 0014** (`docs/decisions/0014-*.md` + the `docs/decisions/README.md` index row), **amending ADR 0005 D1** by reference (0005 stays `Accepted`; its index row may annotate the amendment, as round 040 did for **D7 → ADR 0009**). No ADR is superseded.
- **FR-011**: `specs/truth/techstack.md` MUST be updated for the renamed seam + the single-owned policy (`truth-current`) — **MODIFY**, not NOOP; the affected rows are the **Agent tool loop** row (observer-seam vocabulary) and the **Turn progress spinner** row (the yield policy now has one named owner).
- **FR-012**: Scope guard — the round MUST touch only: `internal/domain/agent/observer.go`, `internal/ui/{yield.go (new), spinner.go, coordinator.go}`, `internal/cli/composite_observer.go`, `internal/agent/agentloop.go`, the affected `_test.go` files, `docs/decisions/**`, `specs/truth/techstack.md`, and the plan package. It MUST NOT change adapter behaviour, domain business logic, flags, exit codes, or stream contracts.

#### Non-Functional Requirements

- **NFR-004**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.
- **NFR-005**: **POSIX-only** (the repo is bash/POSIX-only); no Windows variant.

### Key Entities

- **`ui.YieldController`**: the single named owner of the yield policy (`Yield`/`Restore`/`Admit`; nil-safe; `Enabled()`).
- **`agentport.LoopObserver`** (split): phase hooks + `YieldIndicator()`/`RestoreIndicator()`.
- **`compositeObserver`**: the loop's single observer; forwards phase hooks to the spinner and routes both the port forward **and** its own tail yield through the split yield method.
- **`ToolOutputCoordinator`**: the `[Tool Output]` block owner; now holds a `YieldController` for its clears/resumes/admits.

### Truth obligations (explicit — `truth-current`)

- **Agent tool loop** row — "Round 019 adds a `LoopObserver` hook seam … the loop invokes `OnInferenceStart/End`, `OnToolsStart/End`, and `Before/AfterToolLog`" → the yield pair is now `YieldIndicator`/`RestoreIndicator` (round 045).
- **Turn progress spinner** row — the yield behaviour paragraphs (rounds 034/035/040) now name the **single owner** (`ui.YieldController`) instead of three homes.

`/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted/runtime state change); `/axb-dsl-refine` = **NOOP** (no feature/DSL change).

## Success Criteria *(mandatory)*

- **SC-001**: `internal/ui/yield.go` defines the single named `YieldController`; H1/H2/H3 route every yield through it; no yield rationale survives in more than one home. (covers FR-001, FR-002, FR-003)
- **SC-002**: `LoopObserver` declares `YieldIndicator()`/`RestoreIndicator()` and **no** `BeforeToolLog`/`AfterToolLog`; the composite routes both the forward **and** the tail yield through `YieldIndicator()` — no tool-log-named hook is used for a non-log yield. (covers FR-004, FR-005, FR-006)
- **SC-003**: All four yield orderings (Y1..Y4) are pinned by **unit** witnesses and green, including the `[spinner.yield, call.OnCallEnd]` (Y2) ordering and the clear-before-line (Y3) ordering. (covers FR-009, NFR-001, C-R3-5)
- **SC-004**: Full regression green with **zero** behavioural / stream-contract change — `make verify`, godog E2E, exit codes, `stdout` byte-exactness, DSL vocabulary **11**, topology audit unchanged. (covers FR-007, NFR-003)
- **SC-005**: The R1 gate reports **0 new / 0 stale** with the baseline still **1** (`internal/agent -> internal/ui`); `go test -race ./internal/ui/...` green. (covers FR-008, C-R3-4, NFR-003)
- **SC-006**: `specs/truth/techstack.md` carries the two MODIFY rows; **ADR 0014** + its index row are present; `go.mod`/`go.sum` unchanged. (covers FR-010, FR-011, NFR-004)
- **SC-007**: Falsifiability witnesses (a)/(b)/(c) reproduced and reverted. (covers FR-009)

## Edge Cases

- **A gated-off spinner (nil)** → `YieldController{nil}` makes `Yield`/`Restore`/`Admit` no-ops; `Enabled()` is false so the coordinator starts no watcher and `EndWith`/`WriteWith` hooks are inert. The block still renders (round-040 FR-011 preserved).
- **The final call's deferred tail** → emitted after the renderer's `Stop()`; the composite must not touch the spinner on `final` (Y2 is non-final only) — preserved (round 035 / ADR 0005 D5).
- **`End` while a line write is in flight** → unchanged; the clear runs inside the writer's critical section via `EndWith`'s hook, now calling `YieldController.Yield()` (same `deactivate()`). The accepted residual stays recorded.
- **The one-concurrent-block limit** → unchanged and still recorded (C-R3-6).
- **`Spinner.Stop()` vs a yield** → `Stop()` is the turn teardown (idempotent, called at turn end); a mid-turn clear must go through the owner. Reviewing the two is an explicit Edge Case so a future refactor does not collapse them.
- **A future `internal/ui` yield primitive** → adding one **outside** `YieldController` would re-create the scattering R3 removes; the ADR records `YieldController` as the single entry point.

## Assumptions

- **A1 (weak E2E carrier)** — the round changes no `tellme` CLI behaviour, so the E2E suite is a **weak acceptance carrier**; the witness is the **unit ordering pins** + the ADR (FR-009). `/axb-spec-by-example` is therefore **NOOP** (no new user-facing business journey) — precedent: rounds 020/031/036/041/042/043/044.
- **A2 (mechanism is RD)** — the exact `YieldController` method set, the coordinator's field change, and the test-harness placement are **RD** decisions (`research.md` D-x), not FRs.
- **A3 (ADR required)** — **ADR 0014** records the single-owned yield policy + the split vocabulary (a project-level rule future rounds must cite). ADR **0005 D1 is amended by reference**, not edited.
- **A4 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-ui-plan` **skipped** (not user-facing).
- **A5 (`/axb-dsl-refine` NOOP)** — no feature Rule, Example, step, or `DSLRow` change (the topology audit stays unchanged); the acceptance carrier is the unit pins.
- **A6 (scope guard)** — see FR-012.
- **A7 (C-R3-4 rationale)** — R3 does not lift the baseline entry because that entry is **formatting + predicate**, not the yield port; R4 owns 1 → 0.

## Out of scope (recorded)

- The **blank-reason predicate** (3 sites, 1 dead) and the loop's control-flow dependence on `ui.ToolReasonRenders` — **R4** ([#92](https://github.com/gosharplite/tellme/issues/92) scope ledger #4).
- **Removing** the `internal/agent -> internal/ui` baseline entry (R4; C-R3-4).
- **Strict de-coupling** (R5 → [#101](https://github.com/gosharplite/tellme/issues/101)) and its PR #104 deferrals (F-4/F-6/F-7/F-8).
- The **ride-alongs** (suggestion-selection `set(items, cursor)`; `NewCommandTool(sink)` ctor injection) — parked.
- Changing user-facing CLI behaviour, flags, exit codes, or formatting.
- Re-architecting the loop, the coordinator, or the observer beyond the yield axis.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#92](https://github.com/gosharplite/tellme/issues/92) scope ledger **#2** (spinner-yield policy — single owner) + **#3** (observer port hook split).
- [#105](https://github.com/gosharplite/tellme/issues/105) — the round's anchor (grounded 3-home / 4-call-site inventory + the six clarify candidates).
- [#93](https://github.com/gosharplite/tellme/issues/93) → round 042 (ADR 0011) — the gate + baseline; [#100](https://github.com/gosharplite/tellme/issues/100) → round 044 (ADR 0013) — the composition root + the 8 → 1 ratchet.
- PR [#73](https://github.com/gosharplite/tellme/pull/73) (round 035) — the observer-hook overload + the second yield home; PR [#86](https://github.com/gosharplite/tellme/pull/86) (round 040, ADR 0009) — the `[Tool Output]` block yield + `AdmitResume` + the lock-order invariant.
- Round **045** clarify **C-R3-1 … C-R3-6** (locked above).

## References

- [`internal/domain/agent/observer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/domain/agent/observer.go) — the port to split.
- [`internal/cli/composite_observer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/composite_observer.go) — H2 + the tail yield.
- [`internal/agent/agentloop.go`](https://github.com/gosharplite/tellme/blob/dev/internal/agent/agentloop.go) — H1 (`withToolLog`).
- [`internal/ui/coordinator.go`](https://github.com/gosharplite/tellme/blob/dev/internal/ui/coordinator.go) · [`internal/ui/spinner.go`](https://github.com/gosharplite/tellme/blob/dev/internal/ui/spinner.go) — H3 + the mechanism.
- [`docs/decisions/0005-tool-call-log-parity.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0005-tool-call-log-parity.md) (D1 partitions *rendering*, **not** *yield*) · [`docs/decisions/0009-spinner-dual-timer-and-streaming-liveness.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0009-spinner-dual-timer-and-streaming-liveness.md) · [`docs/decisions/0010-test-deadline-decoupling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0010-test-deadline-decoupling.md) · [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md).
- [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the 1 remaining entry R4 owns.
