# ADR 0014 — Yield-policy owner + `LoopObserver` hook split

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Supersedes:** —
- **Amends:** ADR 0005 **D1** (by reference — the renderer/composite partition; 0005 partitioned *rendering*, **not** *yield*)
- **Related:** round 045 (`specs/plans/045-yield-policy-owner`; issue [#105](https://github.com/gosharplite/tellme/issues/105) — **R3** of [#92](https://github.com/gosharplite/tellme/issues/92));
  round 019 (`internal/ui/spinner.go` mechanism; research D10); round 025 (row-aware clear); round 035 (`compositeObserver` per-call tail yield; issue [#72](https://github.com/gosharplite/tellme/issues/72));
  round 040 (the `[Tool Output]` block yield + `AdmitResume` + the lock-order invariant; ADR 0009); round 042 (layer-discipline gate; ADR 0011); round 044 (composition root; ADR 0013)

## Context

The turn path yields the progress indicator in three places, reached by four call sites over **three different routes**, and each place carried its own copy of the rule:

- **H1** — `internal/agent/agentloop.go` `withToolLog`: clear before, resume after, every `[Tool …]` diagnostic write (via the observer port).
- **H2** — `internal/cli/composite_observer.go` `yieldIndicatorBeforeTail`: on a **non-final** `OnCallEnd`, clear and **never** resume (a *phase-boundary* yield) — reached by calling `Spinner.BeforeToolLog()` **directly** on the composite's own field.
- **H3** — `internal/ui/coordinator.go` `ToolOutputCoordinator`: per-line clear inside a `[Tool Output]` block; idle-gap resume via the watcher; `End` = stop+join → clear → separator → resume — reached via `Spinner.Stop()` / `Spinner.AdmitResume()` / `Spinner.AfterToolLog()`.

The **mechanism** was never the problem: `Spinner.deactivate()` / `resume()` / `AdmitResume()` already give a synchronous, goroutine-joined clear, an epoch-preserving resume, and the round-025 row-aware clear, and the round-040 lock order (block-writer mutex → spinner mutex; mutual exclusion + join; no frame write inside the block critical section) is correct. What had **no owner** was the **policy**: *when a clear is permitted, that a clear never resumes, and that restoration is a phase-boundary act*.

**What this ADR amends (precise scope).** ADR 0005 **D1** decides two things: the *call-begin/call-end* hooks on `LoopObserver` (and the seam becoming a composite), and that *"the CLI stays the only renderer/accounting owner"*. **Both are unchanged** by this round — the pair renamed here is **round 019's** `Before/AfterToolLog`, not D1's creation, and rendering/accounting ownership is untouched. D1's *"the CLI stays the only renderer/accounting owner"* therefore stands for **rendering and accounting**; what D1 **left unnamed** — the **yield policy**, which the CLI composite had been deciding locally (round 035) — is now owned by `internal/ui` (`YieldController`), with the loop keeping the route. So 0005 **D1** is **narrowed/refined** (its unnamed yield axis is now named), not rewritten; 0005's body is not edited.

Separately, the `agentport.LoopObserver` port named its yield hooks after a tool-log write (`BeforeToolLog` / `AfterToolLog`), yet **H2's use is not a log write** — the names lied about one of their four uses, and H2 reached into the spinner instead of the port. The port conflated *a phase started*, *a line is about to be written*, and *the indicator must yield*.

## Decision

**D1 — The yield policy has exactly one owner: `internal/ui` `YieldController` (`internal/ui/yield.go`).** It exposes `Yield()` (clear-only), `Restore()` (resume, synchronous first frame), and `Admit()` (resume, goroutine-drawn first frame), over a `*Spinner` it may hold nil (the gated-off case → no-ops; `Enabled()` reports it). The policy is stated **once**, in the type's doc comment:

- **Yield clears** because a non-indicator line is about to be written to the diagnostic stream (so the line starts on its own cleared row). *A yield never resumes.*
- **Restore resumes** — a **phase-boundary** act: the next waiting phase, the end of a streaming block, or (via `Admit`) the idle-gap watcher — never mid-write.
- **Admit** is the one resume whose first frame is drawn on the redraw goroutine, for the caller that must not block inside another component's critical section.

All four call sites (H1/H2/H3) route through this owner: the loop and the composite via the `*Spinner` port adapter, the coordinator via a held `YieldController`.

**D2 — The `LoopObserver` yield pair is split: `YieldIndicator()` (clear-only) + `RestoreIndicator()` (resume), replacing `BeforeToolLog()` / `AfterToolLog()`.** The port states the loop's **route** to a yield and defers the **policy** to `YieldController`. No alias pair is kept — keeping four names with two meanings would re-introduce the overload the split removes. The composite observer reaches the spinner through the same `YieldIndicator()` its port forward uses, so H2's reach-through is gone.

**D3 — The mechanism is unchanged and stays in `internal/ui/spinner.go`.** `deactivate()` / `resume()` / `AdmitResume()`, the one I/O mutex, the epoch fields, the row-aware clear, and `Stop()` (the idempotent **turn-teardown** clear, distinct from a yield) are untouched.

## Alternatives considered

1. **Keep `BeforeToolLog`/`AfterToolLog` as intent aliases and *add* `YieldIndicator`/`RestoreIndicator`** (the #105 candidate C-R3-3 "alias" variant) — rejected: two names with one meaning per action re-creates the conflation.
2. **Move the wrapping to the presenter** (the loop stops yielding; the CLI brackets its own writes) — rejected: a larger rewrite that overlaps R4 (`internal/agent`'s `internal/ui` coupling) / [#101](https://github.com/gosharplite/tellme/issues/101); R3 stays behaviour-preserving by construction.
3. **Generalise `ToolOutputCoordinator` into a turn-scoped "presentation coordinator" owning all yields** — rejected: a rewrite of the block for no R3 benefit; a thin owner is enough, and the coordinator keeps its single-responsibility.
4. **A port-level owner in `internal/cli`** — rejected: the composite is a *client* of the policy, not its home; `internal/ui` already owns presentation and already holds H3.

## Consequences

- **Single ownership achieved**: the yield *rule* lives once (`YieldController`); the three former homes hold only *when* to call it, not *what* it does. A future mutation to the yield rule has one edit site.
- **Behaviour-preserving**: every existing ordering (Y1 clear+resume; Y2 clear-no-resume; Y3 synchronous joined clear; Y4 resume incl. the goroutine-drawn first frame; `End` = stop+join → clear → separator → resume) is byte-identical, pinned by **unit ordering pins** (the E2E suite cannot observe a clear/resume order — a green suite alone is false confidence, per [#92](https://github.com/gosharplite/tellme/issues/92)).
- **`internal/agent` stays presentation-free for the yield**: the loop holds the port, not the spinner; the `internal/agent -> internal/ui` baseline entry is driven by the **formatters + the `ToolReasonRenders` predicate**, **not** the yield port, so R3 does **not** lift it — that is **R4**'s DoD (1 → 0).
- **ADR 0005 D1 is amended by reference** (rendering vs yield); 0005's body is not edited. **No ADR is superseded.**
- **The two `#92` records** (the one-concurrent-block limit; the `End`-while-write-stalled residual) stay **records**.
- **Witness (c) — the lock-order relation (reproducible by a third party).** The falsifiability of "no self-locking owner" is the **lock order**: the idle watcher calls `YieldController.Admit()` **inside** the writer's `withLock` critical section, and the per-line clear calls `YieldController.Yield()` **inside** `WriteWith`'s critical section — so any owner implementation that acquired the spinner's own mutex would **self-deadlock** (a non-reentrant mutex). That is the same self-locking-accessor class round-040 review **R-11** caught (the locked idle query).
- Immutable once `Accepted`; a future change to the yield semantics, the port vocabulary, or the owner's shape supersedes this ADR rather than editing it.
