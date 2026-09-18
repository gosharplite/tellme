# Technical Research: blank-reason owner + the loop's presentation de-coupling (round 046)

**Plan Package**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`
**Created**: 2026-09-18
**Status**: Draft — RD decisions for the locked clarify answers (**C-R4-1 = A**, **C-R4-2 = delete the dead site**, **C-R4-3 = unit pins + the gate**, **C-R4-4 = ADR 0015**).
**Grounded**: static read @ `dev` `1d36509`.

> The `spec.md` §Locked decisions fix the *what*; this file fixes the *how* (RD, per spec A2). The mechanism is **behaviour-preserving by construction**: the loop keeps its write schedule and the round-045 yield bracket; only the *string source* changes.

---

## D1 — The mechanism: an injected `ToolLineRenderer` port (C-R4-1 = A)

The loop's `internal/ui` dependency is **pure string production** — four formatters plus the blank-reason predicate. Replace the direct import with an **injected port**, declared in `internal/domain/agent` (co-located with `LoopObserver`: the loop's second presentation-facing port), implemented in `internal/ui`, and wired at the composition site.

```go
// internal/domain/agent/presenter.go
type ToolLineRenderer interface {
    EngineLine(t time.Time, step, total int) string
    ActionLine(t time.Time, tool, arguments string) string
    ResultLine(t time.Time, tool, result string) string
    // ReasonLine renders the `[Tool Reason]` line and reports whether it renders
    // at all — ONE evaluation of the single reason transform decides both.
    ReasonLine(t time.Time, reason string) (line string, renders bool)
}
```

**Tier legality.** The edge `internal/agent` (tier 4) → `internal/domain/agent` (tier 0) is **downward** (legal); the edge `internal/ui` (tier 5) → `internal/domain/agent` (tier 0) is downward (legal). The `internal/agent` → `internal/ui` edge is **removed** → the RULE-A baseline entry disappears (**1 → 0**).

**Exact method set is RD's call (spec A2).** All four take `t time.Time` so the timestamp is stamped by the presenter from the loop's injected clock seam (`a.now()`), keeping the line bytes identical to today's.

## D2 — The port implementation lives in `internal/ui` (`ui.ToolLineRenderer`)

`internal/ui/toolcall.go` gains a thin adapter over the existing pure formatters:

```go
type ToolLineRenderer struct{}

func (ToolLineRenderer) EngineLine(t time.Time, step, total int) string { return FormatToolEngine(t, step, total) }
func (ToolLineRenderer) ActionLine(t time.Time, tool, args string) string { return FormatToolAction(t, tool, args) }
func (ToolLineRenderer) ResultLine(t time.Time, tool, result string) string { return FormatToolResult(t, tool, result) }

func (ToolLineRenderer) ReasonLine(t time.Time, reason string) (string, bool) {
    text := toolReasonText(reason)              // ONE evaluation
    if strings.TrimSpace(text) == "" {
        return "", false
    }
    return fmt.Sprintf("[%s] [Tool Reason] %s", formatClock(t), text), true
}
```

- `ReasonLine` is the **single owning definition** in the production path: it evaluates `toolReasonText` **once** and derives both the rendered line and the `renders` decision — collapsing the old double-run inside `logAction` (P2) and re-deriving nothing at a second site (SC-001).
- The rendered line is **byte-identical** to `FormatToolReason(t, reason)` for every non-blank reason (same transform, same `formatClock`, same cap), and `ReasonLine(t, r).renders` ⇔ `ToolReasonRenders(r)` — so both consumers (the begin line and the tail filter) keep today's outcome.
- `ToolReasonRenders` and `FormatToolReason` **stay exported** (the predicate's public helper + the counter-formatter), so the ui test surface is unchanged; `ReasonLine` is the production owner.

## D3 — The loop: the port replaces the import; the schedule is untouched

`internal/agent/agentloop.go`:

- Drop `import "…/internal/ui"`; add a field `Lines agentport.ToolLineRenderer` (nil-safe, like `Stderr`/`Observer`).
- `logEngine` → `a.Lines.EngineLine(a.now(), step, total)`; `logAction` → `a.Lines.ReasonLine(...)` (print iff `renders`) then `a.Lines.ActionLine(...)`; `logResult` → `a.Lines.ResultLine(...)`.
- `withToolLog` is **unchanged** (the round-045 `YieldIndicator()`/`RestoreIndicator()` bracket stays; ADR 0014 untouched — D5).
- Nil `Lines` ⇒ the three log funcs return early (no lines, no blank, no yield) — the documented nil-safe default; the composition site always injects (D4), and the E2E suite asserts the real bytes.
- **The schedule and the blanks are byte-identical**: `logAction` still emits its leading blank, then the reason (when it renders) then the action line.

## D4 — Wiring: the composition site injects the adapter

`internal/cli/cli.go` `runTurn` builds the loop (the composition site, ADR 0013); it already imports `internal/ui`, so it injects `Lines: ui.ToolLineRenderer{}` alongside the existing `Now`/`ToolUsage` seams. No new dependency is threaded through `deps.Dependencies` (the loop is built in `internal/cli`, not in `cmd/tellme`), so the change stays inside the round's scope guard (FR-011).

## D5 — The yield route is R3's, and R4 does not touch it

`withToolLog`'s `Observer.YieldIndicator()`/`RestoreIndicator()` call sites are untouched, so **ADR 0014 is unchanged** (no amendment). `Spinner.Stop()`, `YieldController`, and the composite's tail yield are all out of scope. Recorded so the relation is explicit (SC-003).

## D6 — The predicate's single owner + the dead site P3 (C-R4-2)

Today the predicate `ui.ToolReasonRenders` is evaluated at P1 (`reasonsOf`), P2 (`logAction`), and P3 (`cli/call_renderer.go` `OnCallEnd` — **dead**: P1 already filtered). Under R4:

- **P1** (`reasonsOf`) becomes a **method** on `AgentLoop` that consults the port (`_, renders := a.Lines.ReasonLine(a.now(), r)`), appending the **raw** reason when it renders. (The raw value is appended — not the line — so the tail keeps stamping its timestamp at emit time, D7.)
- **P2** (`logAction`) consults the same port method (one `ReasonLine` call gives both the line and the decision).
- **P3** is **deleted**: the tail (`call_renderer.emit`) prints `roundReasons` directly (they are already filtered upstream by P1), so it no longer re-derives the predicate. The round-036 review **TD-1** "single-ownership consolidation" is thus paid down, not deferred.

One owning definition (the port → `ui.toolReasonText`), evaluated on the **real path** (P1 at call-end, P2 at call-begin) — satisfying #92 **AC6**.

## D7 — Why the loop passes the raw reason (not the rendered line)

The tail's `[Tool Reason]` line is stamped with `r.env.now()` **at emit time** in the presenter (round-009 clock-sharing). Pre-rendering the tail line in the loop (at `notifyCallEnd`, milliseconds earlier) would move a clock reading and risk a byte difference at a second boundary. So the loop appends the **raw** reason and the tail keeps formatting — the observable bytes are unchanged.

## D8 — Truth obligations (`truth-current`) — MODIFY, not NOOP

Two `specs/truth/techstack.md` rows are real MODIFYs:

1. **Agent tool loop** — the row says the loop "is produced by the CLI via two loop observer seams" and names the `internal/ui` formatters; R4 adds that the loop renders through the injected **`agentport.ToolLineRenderer`** port and the blank-reason predicate is **single-owned** (`ui.ToolLineRenderer.ReasonLine`), evaluated once per site with the dead tail guard removed.
2. **Layer-discipline gate** — the row states "At delivery the baseline holds the **8** … reaches 0 across R2–R4"; R4 delivers the terminal state — the baseline is **header-only (0)**.

## D9 — Non-interfaces (NOOP) + governance

- `/axb-api-plan` = **NOOP** (no `contracts/**` exists; no HTTP surface — `contract-authoritative` holds vacuously).
- `/axb-data-plan` = **NOOP** (no persisted/runtime state change).
- `/axb-dsl-refine` = **NOOP** (no feature/DSL change). **Stale-row guard (measured)**: `grep -rn 'ToolReasonRenders\|FormatTool' specs/truth/` — a renamed/relocated seam can leave a *stale* truth row passing green (the audit checks feature → row only); record the measured hits.
- **ADR 0015** (C-R4-4): `docs/decisions/0015-loop-presentation-port.md` + the index row; amends nothing; states its relation to ADRs 0005/0013/0014 (FR-009).

## D10 — Witness: unit pins + the gate (C-R4-3)

1. **The loop's schedule** — retarget the three `internal/agent` log tests to inject a recording **fake renderer** (in-package; the loop's own tests may **not** import `internal/ui` — the gate governs test imports): assert the four line kinds, their order, the per-call blank, and the blank-reason suppression (via the fake's contract).
2. **The adapter's contract** — a `ui` pin: `ToolLineRenderer.ReasonLine` (the conforming adapter) returns `("", false)` for blank/whitespace/escape-only reasons and a line byte-equal to `FormatToolReason` for the rest (incl. an over-cap reason). The **port** postcondition is the weaker, caller-facing invariant — when `renders` is false the `line` value is *unspecified* (fold-review N-3); the loop honours `renders`, never the line's content.
3. **The gate** — `tools/arch/baseline.txt` regenerated to **header-only**; `verify-architecture` green **0 new / 0 stale / 0 cycles**.
4. **Falsifiability (reproduced then reverted, ADR 0010)**: (a) re-add an `internal/agent → internal/ui` import ⇒ the gate reds; (b) make `ReasonLine`'s decision diverge from the format (e.g. render a line for a blank reason) ⇒ the port pin + a loop pin red; (c) add a stale baseline line at 0 ⇒ the gate reds.
5. **No new E2E Example** — the real formatting stays E2E-asserted through the production wiring (`watching-the-tool-loop.feature` + the round-039 spacing Examples).

## Residual risks (recorded)

- **A future second consumer of the predicate** must consume the owner, not re-derive (the ADR records `ui.ToolLineRenderer.ReasonLine` as the owner).
- **A nil `Lines`** silently omits the tool lines; the composition site always injects and the E2E suite guards the wiring (a miswire reds E2E, not a unit).
- **`ToolReasonRenders`** remains exported and used by the ui tests; a future refactor may fold it into `ReasonLine` (recorded).
