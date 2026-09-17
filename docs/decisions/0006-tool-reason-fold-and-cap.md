# ADR 0006 — Tool reason fold, trim, and cap: the model-authored reason joins its sibling sanitize+cap family

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** issue [#74](https://github.com/gosharplite/tellme/issues/74);
  ADR 0005 (D5 — the rune-safe caps this ADR extends to the reason axis, and D6 — the `reason`-excluded argument list);
  round 034 (`specs/plans/034-tool-call-log-parity` — where the guarantee was dropped);
  round 022 (`specs/plans/022-tool-loop-log-line` — fold **B1**, the guarantee this ADR restores);
  round 035 (the spinner-yield fix this ADR's row-integrity fix is deliberately **independent** of);
  round 036 (`specs/plans/036-tool-reason-sanitize` — this ADR's round)

## Context

`internal/ui/toolcall.go`'s `FormatToolReason` (introduced by round 034, PR #71 `806cede`) renders `[HH:MM:SS] [Tool Reason] <reason>` as a bare `fmt.Sprintf` — **no fold, no trim, no cap** — while its siblings sanitize **and** cap (`FormatToolAction` → `argValueCap = 189`; `FormatToolResult` → `resultValueCap = 200`; both via `capRunes(oneLine(…), …)`).

The `reason` is the **only model-authored free-text field** in the tool log: `toolReason` (`internal/agent/agentloop.go`) unmarshals the top-level `reason` and returns it **raw** (untrimmed), and `logAction` guards on the raw value. So a provider can make it multi-line: a `\n` **splits** the reason row across lines and a `\r` **truncates** it (dropping the `[HH:MM:SS] [Tool Reason] ` prefix). A whitespace-only value renders a dangling prefix + trailing whitespace.

Round 022's fold **B1** fixed exactly this class — `FormatToolLog` folded and trimmed the reason (`strings.TrimSpace(oneLine(reason))`), pinned by `TestFormatToolLogFoldsNewlines`. Round 034 **removed** `FormatToolLog` and replaced it with `FormatToolReason`, which does **not** fold — so B1's guarantee was dropped **without being recorded as a divergence**. Under this project's discipline that is an **unrecorded** divergence, not a decided one (issue #74).

## Problem

Two project-level items need a citable home that other artifacts and future rounds can depend on:

1. **The reason axis of the cap family.** ADR 0005 **D5** enumerates exactly two rune-safe caps (`189` argument value, `200` result snippet). The reason's cap had no home, and the truth was asymmetric (the `chat/dsl.md` reason row promised "the reason's own line" while the result row said "on a single line").
2. **The historical drop.** B1's fold guarantee was silently retired by round 034. A future reader must be able to see that this was an *unrecorded* divergence, now recorded — not a behaviour that was never specified.

## Decision

**D1 — The reason joins its siblings: fold + trim + cap, single-site in the pure formatter.** `FormatToolReason` renders `capRunes(oneLine(strings.TrimSpace(reason)), reasonValueCap)` with `const reasonValueCap = 200`. The fold (`\n`/`\r` → space) and trim are applied **inside** the pure `internal/ui` formatter (the round-022 B1 discipline), so the single formatter repairs **both** production callers (`agentloop.logAction` and `cli/call_renderer.go OnCallEnd`) with no caller-side re-implementation.

**D2 — The cap is `200` runes, one U+2026 inside the cap, rune-boundary cut, evaluated on the folded value.** This extends ADR 0005 **D5**'s cap family to the reason axis: the cap set grows from `{189, 200}` to `{189, 200, 200}`. The value reuses the sibling free-text scale (`resultValueCap = 200`); it is a *single numeric threshold* and is not escalated to clarification (per the specify rule) — it is recorded here and ratified in round 036's `research.md`.

**D3 — A blank reason emits no reason line.** A `reason` that is empty or whitespace-only **after folding + trimming** emits **no** `[Tool Reason]` line on either surface. The suppression is evaluated at the **two callers** (a trim-check replacing the raw `reason != ""` guard); `FormatToolReason` stays a **pure** formatter and is **not** asked to return an empty-string sentinel — `callRenderer` emits the return via `Fprintln`, so a `""` return would still print a bare newline. This is round-022 B1's documented intent ("a whitespace-only reason takes the no-tail branch").

**D4 — The historical drop is recorded here; ADR 0005 is not edited.** Issue #74 suggested recording the divergence in ADR 0005, but ADR 0005 is `Accepted` and **immutable** (`docs/decisions/README.md`; ADR 0005's own Consequences: "a future change to the cadence, the caps, or the estimator seam supersedes this ADR rather than editing it"), and adding a reason cap is exactly "a future change to the caps". This successor ADR therefore **records** the round-034 drop and extends D5's cap family, rather than rewriting history.

**D5 — Independence from the round-035 spinner-yield invariant.** Sanitizing the reason keeps a **row intact**; round 035's phase-boundary yield keeps **the row the spinner shares** clear. They are independent invariants and MUST NOT be conflated — fixing one does not make the other redundant.

## Alternatives considered

1. **Edit ADR 0005's D5/D7 in place** (the issue's literal suggestion) — rejected: violates the repo's ADR immutability rule and ADR 0005's own supersession clause.
2. **Sanitize at each caller** — rejected: duplicates the rule in two places (the drift a single pure formatter prevents) and a third caller would silently miss it.
3. **Let `FormatToolReason` return `""` for a blank reason** — rejected: the tail caller's `Fprintln` still writes a bare `\n` unless every caller guards anyway, so the formatter would be coupled to an emitter detail for no gain.
4. **Cap the reason at `189`** (the `argValueCap` value) — rejected: that cap bounds rendered *argument* values; the reason is free-text prose, so the `resultValueCap` scale (`200`) is the closer sibling.
5. **Leave the reason uncapped** (fold + trim only) — rejected: leaves the log's one free-text field unbounded — the sibling asymmetry issue #74 flags.

## Consequences

- The tool log's three rendered value fields are now uniformly sanitized **and** capped: argument values `189`, reason `200`, result `200` — all via `capRunes(oneLine(…), …)` with one U+2026 inside the cap, rune-safe. The truth asymmetry issue #74 named is closed.
- The round-022 B1 fold guarantee is **restored and re-recorded**; the round-034 drop is no longer an unrecorded divergence.
- A blank reason produces no reason row on either surface (no dangling prefix, no blank line).
- **Recorded residual risk:** the E2E surface stays structurally blind to this defect class (round-035 residue rows look for a braille frame + phase status; either count formulation catches only one of `\n`/`\r`). The guarantee is carried by a **hostile-fixture unit pin** — a **permanent narrowing**, not a deferral.
- Immutable once `Accepted`; a future change to the reason cap, the fold, or the blank-suppression rule supersedes this ADR rather than editing it.
