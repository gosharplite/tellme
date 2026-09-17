# Technical Research: tool reason sanitize (round 036)

**Plan package**: `specs/plans/036-tool-reason-sanitize`
**Anchor issue**: [#74](https://github.com/gosharplite/tellme/issues/74) — `FormatToolReason` is unsanitized and uncapped.
**Inputs**: `spec.md` (US1/US2, FR-001–008), the operator-locked clarify decisions Q1/Q2, the existing `techstack.md` tool-loop rows, and the live `internal/ui` / `internal/agent` / `internal/cli` code.

## AIxBDD must-ask questions (already settled — no clarify round needed)

Per `rules/AIxBDD必問問題與起始專案介面澄清判準.md`, the three must-asks are already answered by the **existing** `specs/truth/techstack.md` and this round does not re-judge them:

1. **BDD techstack** — `godog` (`godog.TestSuite` over `specs/truth/features/**`) for the E2E interface contract; Go `testing` for pure helpers (`techstack.md` E2E runner + Test strategy rows).
2. **Test strategy** — E2E (black-box) for the acceptance path; fast unit tests for pure helpers (`techstack.md` Test strategy row).
3. **System ends** — a single CLI end; no HTTP/API surface, no web frontend (`techstack.md`; `contracts/**` NOOP).

No new must-ask gap arises from a display-only bug fix, so `/axb-technical-research` proceeds without a clarify round.

---

## 決策 1：Sanitize single-site inside the pure `FormatToolReason` formatter

- **Decision**: The fold (`\n`/`\r` → space) + surrounding-whitespace trim is applied **inside** `internal/ui/toolcall.go`'s `FormatToolReason`, reusing the existing `oneLine` helper; neither production caller re-implements it.
- **Rationale**: `FormatToolReason` has **two** production callers — `internal/agent/agentloop.go` (`logAction`, the action line) and `internal/cli/call_renderer.go` (`callRenderer.OnCallEnd`, the grouped per-call tail) — both routed through the one formatter, so a single-site fix repairs both surfaces with no caller change. It also matches the round-022 **B1** discipline: the fold belongs in the pure formatter, never in the loop/CLI wiring, and reuses the round-034 helpers (`oneLine`) rather than introducing a parallel path.
- **Alternatives considered**:
  - Sanitize at each caller (loop + CLI): duplicates the rule in two places — exactly the drift a single pure formatter prevents, and a third caller would silently miss it.
  - A new dedicated sanitizer wrapper: adds indirection with no benefit over folding inside the existing formatter.

## 決策 2：The reason is capped at a named constant `reasonValueCap = 200`, mirroring `resultValueCap`

- **Decision**: `FormatToolReason` renders `capRunes(oneLine(reason), reasonValueCap)` with `const reasonValueCap = 200`, one U+2026 counted **inside** the cap, cut on a **rune boundary** — identical mechanics to `FormatToolResult` (`resultValueCap = 200`) and `FormatToolAction` (`argValueCap = 189`).
- **Rationale**: A free-text model field must be bounded like its siblings (the reason is, today, the only *uncapped* tool-log field). **200** reuses the sibling free-text scale (`resultValueCap`) rather than inventing a third number; the same `capRunes` helper keeps the rune-safe/U+2026 contract uniform. The chosen value is a *single numeric threshold* — the ensure rule (`axb-specify` 澄清升級門檻 Rule 1) explicitly does **not** escalate such a detail to clarify, so it is recorded here and the assumption in `spec.md` is thereby ratified.
- **Alternatives considered**:
  - `189` (the `argValueCap` value): that cap bounds *rendered argument values*; the reason is free-text prose, so the `resultValueCap` scale is the closer sibling.
  - A genuinely new value (e.g. 120): no evidence base; a third distinct cap adds comparison cost for no benefit.
  - No cap (fold + trim only): leaves the one unbounded field unprotected against a runaway reason — the sibling asymmetry the issue flags.

## 決策 3：Blank-reason suppression lives at the two callers; `FormatToolReason` stays pure

- **Decision**: A `reason` that is empty **or whitespace-only after folding+trimming** emits **no** `[Tool Reason]` line. The suppression is evaluated at **three** emission sites with a `strings.TrimSpace(reason) != ""`-style guard, replacing today's raw `reason != ""` guard: **(1)** `agentloop.logAction` (the action line) and **(2)** `agentloop.reasonsOf` (the grouped tail's **source list** — the production filter, so the tail already receives no blank) are both reachable from the real loop; **(3)** the `callRenderer.OnCallEnd` `emit` closure carries a fourth-home **defensive** guard that production cannot reach (kept as defence-in-depth; the single-ownership consolidation is recorded on [#69](https://github.com/gosharplite/tellme/issues/69), not refactored in-round — round-036 review TD-1). (Note: the real CLI symbol is the `emit` closure inside `OnCallEnd`; there is no `renderCallTail` symbol — round-036 review TD-2.) `FormatToolReason` is **not** asked to return an empty-string sentinel.
- **Rationale**: Today a `"   "` reason renders a dangling `[HH:MM:SS] [Tool Reason]    ` row (trailing whitespace) and a `"\n"` reason renders the prefix plus a blank line. Round-022 B1's `TrimSpace` was documented **precisely** so "a whitespace-only reason takes the no-tail branch", so suppression is the intended contract, not a new behaviour. The caller-side check is required because `callRenderer` emits the formatter's return via `fmt.Fprintln` — a `""` return would still print a bare newline, so the sentinel must be resolved by the caller, not the formatter. Keeping the formatter pure preserves its unit-testability (no output-shape coupling to the emitter).
- **Alternatives considered**:
  - Let `FormatToolReason` return `""` and have callers skip empty returns: the tail caller's `Fprintln` still writes a bare `\n` unless every caller adds a guard anyway — so it buys nothing and couples the formatter to an emitter detail.
  - Render a prefix-only row `[HH:MM:SS] [Tool Reason]`: keeps a useless row and diverges from B1's documented intent.
  - Do not suppress (match `FormatToolResult`'s fold-without-trim): leaves the trailing-whitespace artifact the issue reports.

## 決策 4：Witness = hostile-fixture unit pin + falsifiability witnesses; no new E2E Example

- **Decision**: The guarantee is witnessed by a **unit pin** in `internal/ui/toolcall_test.go` using hostile fixtures — a reason containing `\n`, one containing `\r`, a whitespace-only reason, and an over-cap (201-rune) reason — plus the emission-path pins for suppression (loop + tail). **No** new E2E Example and **no** new `DSLRow` are minted (operator Q1 → Option 1). Falsifiability witnesses are reproduced-then-reverted: remove the fold → the `\n`/`\r` pins fail; remove the cap → the over-cap pin fails; remove the trim guard → the whitespace-only row re-emits.
- **Rationale**: The E2E suite is **structurally blind** to this defect class — the round-035 residue rows look for a braille frame + phase status, which neither `\n` nor `\r` produces, so they pass on both (simulated: clean / `\n` / `\r` → residue row PASS in all three). A count-based E2E carrier has two formulations and **each silently misses one input** (a marker-bearing `[Tool Reason]` line count catches `\r` but not `\n`; a total tail-block line count catches `\n` but not `\r`). The unit pin deterministically covers **all four** inputs. This mirrors both precedents: round-022 B1 was witnessed by a unit pin only (`TestFormatToolLogFoldsNewlines`), and round 035 deliberately declined to mint a row for a display-only defect (its SC-002 permanent narrowing).
- **Residual risk (disclosed)**: the "no new E2E row" choice is a **permanent narrowing**, not a deferral — the collision class stays outside the E2E surface. The unit hostile-fixture pin is the deterministic carrier; a future round may revisit if a count formulation that covers **both** inputs is ever designed.
- **Alternatives considered**:
  - E2E marker-bearing count: catches `\r` only; mints a feature row + audit for half coverage.
  - E2E total tail-block count: catches `\n` only; same cost, half coverage.
  - Both E2E carriers: two new rows for a display-only defect, against the round-035 "no row for a display-only defect" precedent.

## 決策 5：Record the dropped guarantee + this round's reason-cap decision in a NEW ADR 0006 (not by editing ADR 0005)

- **Decision**: Add `docs/decisions/0006-tool-reason-fold-and-cap.md` recording (a) the **round-034 unrecorded divergence** — round 034 replaced round 022's folding `FormatToolLog` with `FormatToolReason`, dropping B1's fold/trim guarantee **without** recording it — and (b) this round's decision that the reason joins its siblings (fold + `TrimSpace` + `reasonValueCap = 200`) and that a blank reason is suppressed. Add the ADR to `docs/decisions/README.md`'s index.
- **Rationale**: Issue #74 suggests recording the divergence in ADR 0005, but **ADR 0005 is `Accepted` and immutable** (`docs/decisions/README.md`: "Immutable once `Accepted` … supersede with a new ADR rather than rewriting history"; ADR 0005's own Consequences: "a future change to the cadence, the caps, or the estimator seam supersedes this ADR rather than editing it"). Adding a **reason cap** is exactly "a future change to the caps". So the disciplined vehicle is a **new, successor ADR** that extends D5's cap family to the reason axis and records the historical drop — history is added, never rewritten.
- **Alternatives considered**:
  - Edit ADR 0005's D5/D7 in place (the issue's literal suggestion): violates the repo's ADR immutability rule and ADR 0005's own supersession clause.
  - Leave it unrecorded (status quo): reproduces the exact discipline violation the issue was filed to expose.

## 決策 6：Tidy — move `oneLine` into `toolcall.go` (its sole consumer family)

- **Decision**: Relocate the `oneLine` helper from `internal/ui/toollog.go` into `internal/ui/toolcall.go` and delete the now-empty `toollog.go`. The fix adds a **third** consumer (`FormatToolReason`), so the file named `toollog` would hold no tool-log formatter at all.
- **Rationale**: `internal/ui/toollog.go` is a 13-line orphan holding exactly one function whose only consumers are the round-034 formatters in `toolcall.go`; after this round it has three consumers, all in `toolcall.go`. Moving it keeps formatter + helper co-located and removes a misleadingly-named file.
- **Alternatives considered**:
  - Leave `toollog.go` as-is: the orphan the issue flags; a reader would misread it as a formatter module. If deliberately left, the file comment must say so.
  - Merge into a new shared file: no benefit over the natural home (`toolcall.go`).

## 決策 7：Scope guard — display-only; no flag / exit code / transport / persisted-state change; loop-tier pin updated in the same change

- **Decision**: The fix changes only the `stderr` tool-call log rendering. It touches no CLI flag, exit code, frozen class-phrase vocabulary, tool schema/result, provider transport, or persisted record; `stdout` stays byte-exact. The loop-tier pin `internal/agent/agentloop_reason_test.go` (which asserts the literal `[12:34:56] [Tool Reason] because`) is updated **in the same change** if the trim/fold touches its fixture; the clean-reason fixture continues to render byte-identically (FR-004), so no change is expected there but the pin is re-verified.
- **Rationale**: Issue #74 is a presentation-contract fix; the sibling caps and the `FormatToolResult` behaviour must not change (FR-004). The loop-tier pin is the second consumer of the formatter's output and must be re-verified to keep the guarantee single-sourced.
- **Alternatives considered**:
  - Also reconcile `FormatToolResult`'s fold-without-trim asymmetry: out of scope (a settled sibling behaviour; changing it is a separate decision).
  - Touch the `[Tool Action]`/`[Tool Result]` caps: out of scope — only the reason joins the sanitize+cap family.

---

## Truth impact summary (for `/axb-truth-delta` and the owner skills)

- **`specs/truth/techstack.md`** — **MODIFY** the **Agent tool loop** row: record that the per-call `[Tool Reason]` is folded (`\n`/`\r` → space) + trimmed and capped at **200** rendered runes (one U+2026, rune-safe) like its siblings, and that a blank (empty/whitespace-only) reason emits **no** reason line on either surface (the fix restores the round-022 B1 fold guarantee; the cap set grows from `{189, 200}` to `{189, 200, 200}`).
- **`specs/truth/features/cli/chat/dsl.md`** — **MODIFY** (owned by `/axb-dsl-refine`): reconcile the **reason row** prose to state the single-line guarantee explicitly (as the result row already does) and land the `reasonValueCap` constant next to the siblings' documented `189`/`200`; plus a round-036 note recording the fold+trim+cap and the blank-reason suppression. **No new `DSLRow`; no feature Example (spec-by-example NOOP).**
- **`specs/truth/contracts/**`** — **NOOP** (no HTTP surface).
- **`specs/truth/data/data-model.dbml`** — **NOOP** (no persisted-state change).
- **`docs/decisions/0006-*.md`** — **ADD** (governed-artifact decision record; the ADR index gains a row).

## Residual risks

- **Permanent E2E narrowing (D4)** — the collision stays outside the E2E surface; the unit hostile-fixture pin (`internal/ui/toolcall_reason_test.go`) is the deterministic carrier. This record also lives on the **live** issue [#69](https://github.com/gosharplite/tellme/issues/69) (and was commented on [#74](https://github.com/gosharplite/tellme/issues/74) before it closes), because this plan package **freezes on merge** — the round-035 G3 lesson: *durable surface = a live issue, not a frozen plan package* (round-036 review TD-3). Recorded so a future reader does not read the E2E coverage as complete.
- **Retro-documentation (D5)** — the round-034 divergence is documented after the fact; ADR 0006 records it, but the original round-034 package stays frozen (rule `plan-package-frozen`).
- **Cap-value ratification (D2)** — 200 is chosen by sibling-parity, not measurement; recorded as the assumption `spec.md` carried.
