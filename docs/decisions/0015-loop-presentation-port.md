# ADR 0015 — Loop presentation port: the loop owns the schedule, `internal/ui` owns the tool-line rendering

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Supersedes:** —
- **Amends:** — (ADR 0005 **D1** is reaffirmed, not amended: the loop loses no *rendering* ownership it had; it gains an injected seam)
- **Related:** round 046 (`specs/plans/046-blank-reason-owner-and-presentation-decoupling`; issue [#108](https://github.com/gosharplite/tellme/issues/108) — **R4** of [#92](https://github.com/gosharplite/tellme/issues/92));
  round 034 (decomposed tool-call rendering; ADR 0005); round 036 (the blank-reason fold/trim/suppress predicate — issue [#74](https://github.com/gosharplite/tellme/issues/74); ADR 0006);
  round 039 (the single-owned `internal/ui` sanitize policy applied by every `[Tool …]` formatter — issue [#80](https://github.com/gosharplite/tellme/issues/80); ADR 0008);
  round 042 (layer-discipline gate + baseline; ADR 0011) · round 044 (composition root; ADR 0013) · round 045 (yield-policy owner; ADR 0014)

## Context

`internal/agent/agentloop.go` — the bounded think→act→observe loop — imported `internal/ui` for two things, and nothing else:

1. the **four pure tool-line formatters** (`FormatToolEngine` / `FormatToolReason` / `FormatToolAction` / `FormatToolResult`), called from `logEngine` / `logAction` / `logResult` inside `withToolLog`; and
2. the **blank-reason predicate** `ui.ToolReasonRenders` — *"does this model-authored `reason` render a `[Tool Reason]` row?"* — derived from the single transform `ui.toolReasonText` (fold → trim → sanitize → cap).

The R1 layer gate (ADR 0011) ranks `internal/agent` = tier **4** and `internal/ui` = tier **5**, so this import is an **upward** edge → a **RULE-A** violation, and it was the **only** entry left in `tools/arch/baseline.txt` (R2 had taken the 7 `cli → infrastructure` edges; R3's `C-R3-4` deliberately left this one for R4).

The predicate was also **re-stated per site**: three evaluations of the same rule — `reasonsOf` (the tail filter), `logAction` (the begin-line guard), and `cli/call_renderer.go` `OnCallEnd` (a **dead** defensive re-check, because `reasonsOf` had already filtered) — with the `toolReasonText` transform re-run **≥3× per call** (round-039 review evidence). The rule itself was single-owned in `internal/ui`, but its *use* was scattered and one site could never fire.

## Decision

**D1 — The loop renders its tool-line diagnostics through an injected port, `agentport.ToolLineRenderer`.** The port is declared in `internal/domain/agent` (co-located with `LoopObserver` — the loop's second presentation-facing port) and takes the four line renderers, all stamped from the loop's injected clock:

```go
type ToolLineRenderer interface {
    EngineLine(t time.Time, step, total int) string
    ActionLine(t time.Time, tool, arguments string) string
    ResultLine(t time.Time, tool, result string) string
    ReasonLine(t time.Time, reason string) (line string, renders bool)
}
```

`ReasonLine` returns the rendered `[Tool Reason]` line **and** whether it renders at all, from **one** evaluation of the single reason transform. Implemented in `internal/ui` as `ui.ToolLineRenderer` (a thin adapter over the existing pure formatters); the loop holds the port nil-safely (like `Stderr` and `Observer`) and the composition site (`internal/cli` `runTurn`, per ADR 0013) injects it.

**D2 — The loop owns the SCHEDULE; `internal/ui` owns the RENDERING and the blank-reason predicate.** The loop decides *which* lines, *when*, in *what order*, and writes them to its own `Stderr` — including the round-039 per-call leading blank. The presenter decides *what the bytes are* and *whether a blank reason renders*. So:

- the loop keeps its write sites and the round-045 yield bracket (`withToolLog` → `Observer.YieldIndicator()` / `RestoreIndicator()`), so **ADR 0014 is untouched**;
- `internal/agent` no longer imports `internal/ui` (the RULE-A edge is gone → baseline **1 → 0**);
- the predicate's **single owner on the real path** is `ui.ToolLineRenderer.ReasonLine`: the loop's two live sites (`logAction`, the tail-filter) consume it, and the **dead** third site (`cli/call_renderer.go` `OnCallEnd`) is **deleted** — single ownership, not defence-in-depth (round-036 review TD-1 paid down).

**D3 — The tail keeps stamping its own timestamp.** The loop passes the **raw** (already-filtered) reasons to the call-end hook, and the tail formats them at emit time — so no clock reading moves and the emitted bytes are unchanged. Pre-rendering the tail line in the loop was rejected for this reason.

## Alternatives considered

1. **Relocate the writes into the presenter** (the loop emits *semantic* tool-line events through `LoopObserver`; the CLI `call` half formats + writes) — rejected: purer ownership, but it moves **observable write sites**, and this round's acceptance carrier is **weak** (#92: a green E2E suite is false confidence) — the round-035/039 spacing class would be in the blast radius for no falsifiable gain.
2. **As (1) plus the presenter owning the round-045 yield wrapping** — rejected: it would **amend ADR 0014** and churn the `YieldIndicator`/`RestoreIndicator` pair R3 had just split, for a benefit R4's DoD does not require.
3. **Move the pure formatters to a lower tier** (a new `internal/present` package or into `internal/domain/**`) — rejected: it forces either a **new tier-table entry** (a normative ADR-0011 change) or **presentation types in the domain** (violating AC5 purity); and it would split `internal/ui`'s single sanitize policy across packages.
4. **Keep the direct import and widen the baseline** — rejected: it defeats the ratchet (#92 AC3: **0** violations) and R3's deliberate deferral.
5. **Keep the dead tail guard as defence-in-depth** — rejected: #92 AC6 asks for **one** owner on the **real path**; a guard that can never fire is the "dead site" the ledger names.

## Consequences

- **The ratchet reaches its terminal state** — `tools/arch/baseline.txt` is **header-only (0)**; `verify-architecture` is green (0 new / 0 stale / 0 cycles) and a re-introduced `internal/agent → internal/ui` import **fails** the gate (the anti-bypass rule holds at 0).
- **Single ownership of the blank-reason predicate** — `ui.ToolLineRenderer.ReasonLine` is the owner, evaluated once per site on the real path; the dead tail guard is removed; the ≈3×-per-call re-run collapses.
- **Behaviour-preserving** — the write sites, their order, the per-call blanks, the bytes, `stdout`, the exit codes, and the DSL vocabulary are unchanged; the migration is a **string-source** change only.
- **ADR 0014 is untouched** (the yield mechanism/policy/port pair stand); **ADR 0005 D1 is reaffirmed** (the CLI remains the renderer/accounting owner — the loop merely stops *formatting*); **no ADR is superseded**.
- **Public surface preserved**: `ui.FormatTool*` and `ui.ToolReasonRenders` stay exported (the formatters + the public predicate); `ui.ToolLineRenderer` is the production entry point.
- **Recorded residual**: the loop's own `_test.go` files **may not** import `internal/ui` (the gate governs test imports), so they inject an in-package **fake renderer** — the loop's tests now assert the *schedule*, and the *formatting* is asserted at the `ui` tier + E2E.
- **The predicate's `ReasonLine` query costs a format at the tail-filter site** (round-046 review TD-1): `reasonsOf` calls `ReasonLine` and discards the line, so each reason costs a clock read + a full format + an allocation purely to obtain the boolean. Pre-formatting the tail in the loop is forbidden (it would move a clock reading — D3), so the trade is accepted; the alternative (splitting the port into `ReasonRenders(reason) bool` + the renderer) is recorded as available but not taken.
- **The domain-declared port is presentation-SHAPED** (round-046 review TD-4): `ToolLineRenderer`'s method set returns terminal-ready strings and owns the clock stamping, so a `internal/domain/agent` port's doc comments specify line formats. Moving the *formatters* into the domain was rejected (Alternatives #3); this smaller *shape leak* is an **accepted** trade, recorded so a future round does not treat it as an oversight.
- **The baseline at 0 has no release valve** (round-046 review TD-7): with `tools/arch/baseline.txt` header-only, a future *legitimate* downward violation can no longer be absorbed by appending a baseline line — it can only be fixed by refactor (or by a deliberate ADR-0011 amendment). This is the intended terminal state of the R1→R4 ratchet, recorded so the next round that reaches the gate does not re-litigate a settled decision.
- Immutable once `Accepted`; a future change to the port's shape, the schedule/presenter split, or the predicate's ownership supersedes this ADR rather than editing it.
