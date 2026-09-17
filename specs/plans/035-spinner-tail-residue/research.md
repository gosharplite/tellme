# Technical Research — round 035 (spinner tail residue)

Plan package: `specs/plans/035-spinner-tail-residue`
Truth touched: `specs/truth/techstack.md` (row: **Turn progress spinner (operator)**).
Anchor: issue [#72](https://github.com/gosharplite/tellme/issues/72) — the per-call tail is not yielded to the spinner, so a frame row collides with the tail's first grouped `[Tool Reason]` and the residue survives the finished turn.

> Decision-driven research. Operator-locked inputs: **Q1 (a)** yield around the tail · **Q2** pure bug fix · **Q3** siblings out · **Q4** E2E + unit witness. `spec.md` FR-001..FR-007 / SC-001..SC-004 are the acceptance authority.

## Context / constraints

- **Rendering-only, implementation-vs-truth defect**: the violated rules already exist as executable truth (round-019/025 spinner residue rows). No flag, exit code, line format, cadence, tool schema/result, transport, persisted record, or `stdout` byte changes (FR-005/FR-006).
- **stdlib-only, POSIX-only, hermetic (no pty)**: the round-019 `stderr` terminal seam (`the diagnostics are shown at a terminal` → `TELL_ME_FORCE_STDERR_TTY`) forces the spinner on; the existing fake provider scripts the tool round (NFR-001/002).
- The failing write is `callRenderer.OnCallEnd`'s **non-final** `emit()` (`internal/cli/call_renderer.go:78-94`); the loop fires it while the spinner is still running because `Spinner.OnToolsEnd()` is a no-op (`internal/ui/spinner.go:207-208`) and `Spinner.OnCallEnd` is an intentional no-op (ADR 0005 D1). The `[Tool Output]` sink and the tool-log lines **already** yield (`cli.go:753-766`, `agentloop.go:withToolLog`); the per-call tail is the one un-yielded write.

## Decisions

### D1 — Phase-boundary yield: clear the spinner before the tail, and do NOT resume

The fix is a **synchronous clear of the progress indicator immediately before the per-call tail's first write**, with **no immediate resume**. The indicator is re-activated by the next **waiting-phase** hook (the next call's `OnInferenceStart`) or, for the final call, by the turn's existing `Stop()`.

Rationale — the tail sits at a **phase boundary**, unlike every existing yield site:
- The **tool-log lines** and the **`[Tool Output]` block** interleave **inside** a waiting phase, so they `clear → write → resume` (`BeforeToolLog`/`AfterToolLog`).
- The **per-call tail** is emitted at the call's **end** (`callEnd`), and the next activity is the next call's status **frame** (`OnCallBegin`), then that call's `OnInferenceStart`. There is nothing to resume *for* — the correct move is to leave the line cleared and let the next waiting phase re-activate.

Clearing is synchronous (`Spinner.deactivate` → `clearLocked` → `eraseRows`, round-019 D10 / round-025 row-aware), so the tail's first `fmt.Fprintln` starts at column 0 of a cleared line and the tail's lines begin on their own rows (FR-002). The turn-scoped elapsed epoch is preserved across deactivate/activate, so re-activation keeps counting (round-019 D4).

**Alternatives considered**:
- **Clear + resume after the tail** (the naïve reading of "yield"): **rejected** — a byte-level trace shows it *relocates* the residue. After the tail resumes, the spinner redraws on a fresh row; the next call's `OnCallBegin` then writes `ui.FormatTurnOpening` = `"\n" + rule + "\n" + header + "\n"` (`internal/ui/turn.go:51`), whose leading `"\n"` moves the cursor down and leaves the resumed frame on the row above. The spinner still believes "my row = current row", so its later redraws/`Stop()` erase the **wrong** row and the frame text survives — the same defect, moved one call later. (Removing the `"\n"` would require touching the pristine round-017 chrome writer.)
- **Make `Spinner.OnCallEnd` clear** (issue #72's option (b)): **rejected by Q1** — it moves the yield responsibility into the indicator and edges toward the ADR-0005-G1 `LoopObserver` segregation (a superseding decision, out of scope per Q2). Recorded as the alternative the operator declined.
- **Yield at every status/frame write too** (`OnCallBegin`): **not needed** — call 0's frame precedes any activation, and every later frame is written after the prior tail already cleared the line (D2 trace). Adding a second yield site would be redundant surface.

### D2 — The yield lives on the composite seam, keyed on `!final`

The clear is issued in **`compositeObserver.OnCallEnd`** (`internal/cli/composite_observer.go`), which already holds both halves (`call`, `spinner`): clear via the spinner **before** delegating to `call.OnCallEnd`, and **only when `!final`**. The final call's tail is deferred (`EmitFinalTail` after `Stop()`), so no write happens at `final` time — the clear is skipped to keep the final path byte-identical.

Coherence trace (2-call tool turn), after the fix:
1. `callBegin(0)` → frame 0 (clean; spinner not yet active).
2. `inferenceStart` → activate (`Thinking`). `toolsStart` → activate (`Executing tools`).
3. tool-log lines (`withToolLog`, `clear→write→resume`) → last resume leaves a live frame.
4. `toolsEnd` (no-op) → `callEnd(0, final=false)` → **clear (stop) → tail 0 lines on cleared rows**.
5. `callBegin(1)` → frame 1 on a clean row (spinner stopped).
6. `inferenceStart` → activate → frame drawn on a fresh row; final call → `Stop()` before the answer; deferred final tail after → clean.

`Rationale`: the composite is the single seam that owns both behaviours (ADR-0005 D1), so the fix is one method, leaves `callRenderer` and `internal/ui` untouched, and is unit-testable with a recording double. **Alternatives**: give `callRenderer` a spinner handle (couples the renderer to the indicator); a dedicated `Spinner` method (widens the indicator's API for one caller) — neither earns its surface.

**Yield-site inventory (PR #73 review RF-1).** After this round the turn's spinner-yield policy has **three** homes, none of them a named owner: (a) the composite's per-call tail (this round — clear-only, no resume, `!final`); (b) the `[Tool Output]` sink closure (`cli.go:753-766` — clear before the child, resume after the closing separator); (c) `runTurn`'s **two** `sp.Stop()` call sites (the deferred panic-safe guard, `cli.go:740`, and the inline stop before the answer, `cli.go:774`) plus the post-answer `EmitFinalTail()` (`cli.go:795`) — i.e. four call sites across three policy homes. ADR 0005 D1 partitions *rendering* ownership (loop → tool lines, CLI → frames/tails) but not the *yield* policy; single ownership of it is recorded on [#69](https://github.com/gosharplite/tellme/issues/69) (RF-1) so the next rendering round does not add a fourth site by path of least resistance.

### D3 — The E2E witness: force the gate + a tool round + the whole-stream residue row

The primary witness is an E2E Example that **pairs the spinner gate with a tool round** and asserts the **whole-stream residue row** `the run shows no progress spinner` (which reads the capture through the `\r`-redraw simulation, `tests/e2e/steps/spinner_helpers.go:spinnerVisible`):

```gherkin
Given the operator has a runnable tellme installation
And the runtime home is "ait-tmg"
And the diagnostics are shown at a terminal
And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good"
When the operator starts tellme with the prompt "read the notes"
Then the run shows no progress spinner
And tellme exits successfully
```

Rationale — this is exactly the scenario shape the audit found **missing** (D4): under the defect the collision row survives `spinnerVisible` as `⠋ Executing tools [...]...` → `hasSpinnerFrame` matches → the step **fails today**; after the fix the tail is written on a cleared line → the step **passes**. It reuses the existing step and the existing row (no new DSL row; `dsl-exact-one-match` preserved), and exercises the real CLI+loop+spinner wiring (the bug's actual home). Falsifiability: revert the clear → the step fails with a `⠋ …` residue in `stderr`.

**Alternatives**: the post-answer row `the progress spinner no longer appears once the answer is written` (t011) — **insufficient**: it inspects only the bytes **after** the answer, and this defect's residue is **pre**-answer (in the tail); it is why the existing terminal+tool-round Example stayed green. A pty-driven terminal capture — rejected (NFR-001, no pty; a flat capture plus the `\r` simulation suffices).

### D4 — Why the existing suite missed it (verification-coverage gap), and the go-forward

Not a missed red gate — a **renderer-layer coverage gap** (the round-009 pattern). Of the two residue rows:
- `the progress spinner no longer appears once the answer is written` (t011) checks the merged capture **after** the answer → blind to a pre-answer tail collision.
- `the run shows no progress spinner` (t012) checks the **whole** stream via `spinnerVisible` → *would* catch it, but its scenarios only exercise non-terminal surfaces / the failed-turn carrier, never a **terminal + tool round**. The one terminal tool-round Example asserted only t011.

Fix (truth-half): `/axb-dsl-refine` adds the D3 Example to `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (and records the round-035 note in `chat/dsl.md`), so the whole-stream residue row is exercised on the gated tool-tail path.

### D5 — The unit pin: the composite's yield ordering

A unit test in `internal/cli` over `compositeObserver` with recording doubles of the `call` and `spinner` halves:
- `OnCallEnd(_, _, _, final=false)` → the spinner's `BeforeToolLog` (the clear) is invoked **before** `call.OnCallEnd`, and `AfterToolLog` is **not** invoked.
- `OnCallEnd(_, _, _, final=true)` → neither the clear nor a resume is invoked; `call.OnCallEnd` is still invoked (the deferral is stored by the renderer).

Rationale: a flat unit capture cannot reproduce a terminal grid (the round-025 note), so the unit layer pins the **ordering contract** — the piece most directly encoded by the fix and cheap to guard on every run; the E2E Example (D3) witnesses the observable residue. **Alternative**: a flat-buffer assertion on `internal/ui.Spinner` — rejected (the fix is the composite's ordering, not the spinner's internals).

### D6 — Scope guard: nothing else changes

`Spinner.OnCallEnd` stays a no-op; `callRenderer` internals, the deferred final tail, the per-call frame cadence, the line formats, `Stop()` semantics, the `[Tool Output]`/tool-log yields, and the gated-off path are unchanged (FR-005). `stdout` stays byte-exact; `go.mod`/`go.sum` unchanged (stdlib-only).

## Recorded residual risk / forward items

- **Boundary fragility (residual, low — writer inventory folded, PR #73 review D6).** "Clear-only, no resume" depends on the invariant: **no diagnostic write occurs between a non-final tail and the next waiting-phase activation.** Verified (not merely asserted): in that window the **only** writer is the **next call's `OnCallBegin` frame** — safe, because the indicator is stopped and the line cleared, and `FormatTurnOpening` opens with a bare `\n` on an already-clean row. Call-end-adjacent paths that correctly need **no** yield, because `runTurn`'s idempotent `Stop()` already covers them: (a) the `no tools are registered` / `tool %q is not available` returns (`agentloop.go:161/165`) fire **no** call-end at all; (b) a `Complete` error returns before any call-end. A future phase that interleaves a write in that window must yield too; recorded so a future round re-diffs this inventory.
- Nothing else new; the round-034 siblings (G2 `Ready` overstatement, numbering skew) remain out of scope (Q3).
- **Literal line-count coverage (recorded candidate, PR #73 review F3, mechanism corrected F2-of-fold-review-#2).** SC-002 is carried by the whole-stream residue row only, and that narrowing is **permanent, not deferred**. The residue row catches the reported collision but — as the reviewer's simulation shows — **not** a `\n`/`\r` **inside** a reason (it looks for a braille frame + phase status, which neither produces; all three shapes PASS the residue row). The **mechanism, corrected**: a `\r` **truncates one reason's own row** (its visible text after the last `\r` survives, dropping the `[HH:MM:SS] [Tool Reason] ` prefix); it does **not** merge two reasons onto one row — each reason is newline-terminated by `Fprintln`. So the candidate carrier must name its formulation: a **marker-bearing count** catches the `\r` case (2 → 1) but **not** the `\n` case (marker count stays 2, plus a stray fragment line); catching `\n` needs a **total tail-block line count** or the underlying reason-sanitization rule (issue [#74](https://github.com/gosharplite/tellme/issues/74)). A feature change ⇒ an audit re-run.
- **`FormatToolReason` is unsanitized/uncapped — a pre-existing round-034 defect (PR #73 fold-review #2, G1).** `internal/ui/toolcall.go:42`'s `FormatToolReason` is a bare `fmt.Sprintf` — no `oneLine` fold, no `TrimSpace`, no `capRunes` — unlike its siblings `FormatToolAction`/`FormatToolResult`, and it **silently drops round-022's B1** one-line-fold guarantee (`specs/plans/022-tool-loop-log-line/research.md:116`). This is **out of round 035's Q2 scope** (no code change here) and is **pre-existing on `dev`/`main`** (introduced by round 034). Per the plan-package-frozen discipline, the finding is homed on issue [#74](https://github.com/gosharplite/tellme/issues/74) so it outlives this package's freeze; this bullet is the pointer, not a second copy.

## Truth impact

| Truth Spec | Action | Summary |
| --- | --- | --- |
| `specs/truth/techstack.md` — **Turn progress spinner (operator)** row | MODIFY | Round 035 adds: the per-call **tail** is written with the indicator **yielded** — the presenter synchronously clears the spinner **before** the tail's first line (a phase-boundary yield; **no** immediate resume — the next waiting phase re-activates it), so the tail's grouped `[Tool Reason]`/measured-payload/metrics/`Ready` lines begin on their own rows and no frame residue survives into the finished turn. The row's existing "yields the line to the answer / leaves no residue" intent otherwise holds. |
| `specs/truth/contracts/**` | NOOP (checked) | Single CLI end; no HTTP/OpenAPI surface; only a `stderr` diagnostic write changes. |
| `specs/truth/data/data-model.dbml` | NOOP (checked) | No persisted-state change; the fix alters only when/how the tail is written to the terminal. |
| `specs/truth/features/cli/chat/**` | MODIFY (delegated to `/axb-dsl-refine`) | Add the gated tool-tail residue Example (D3) + a `chat/dsl.md` round-035 note. |
