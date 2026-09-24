# ADR 0057 — A yellow `[TOOLS] - M (N calls)` line in the `-l`/`--list` listing

- **Status:** Accepted
- **Date:** 2026-09-24
- **Deciders:** tellme owner (operator-stated UI shape)
- **Related:** [ADR 0045](0045-list-role-headers-and-rendered-body.md) (the `-l` listing role headers + the stdout-terminal colour gate this **amends**), [ADR 0054](0054-list-backward-turn-indices.md) (the backward turn index this reuses and the whole-label colour rule), [ADR 0053](0053-back-rollback-turns.md) (the turn granularity the count describes), [ADR 0020](0020-cli-ui-decoupling.md) (the `render.*` domain ports the CLI uses — the new field keeps the CLI free of `internal/ui` types); the tool-step record is round 008 (`specs/truth/data/data-model.dbml`, `history_step`); round 086 (`specs/plans/086-tools-listing-line`)

## Context

`tellme -l [N]` (round 073; ADR 0045) prints each listed message as a role header line (`[USER]` / `[MODEL]`, round 082/ADR 0054 adds the backward turn index ` - M`) plus a body, and **deliberately omits** the tool activity — the round-073 clarify **Q2 → A** decision, asserted by the `history/dsl.md` row *"tellme lists only the operator's messages"*. The report that survived test review (round-082 follow-up) noted the cost: a reader cannot tell, from the listing alone, which turns were tool rounds — the figure is persisted (`history.Entry.Steps`) but never surfaced, and the tool **content** must stay omitted (Q2 → A on the content side).

The persisted record already carries what is needed: `history.Entry.Steps []Step` is the ordered list of tool executions of a completed turn (`specs/truth/data/data-model.dbml`, `history_step`), so `len(Entry.Steps)` is the number of tool calls the turn made. `Entry.Calls` is a **different** figure — the AI-endpoint inference-round count (round 027).

## Decision

**D1 — a `[TOOLS] - M (N calls)` line per listed turn.** Each listed turn shows exactly one tool-activity line between its `[USER]` block and its `[MODEL]` header. `M` is the turn's backward turn index — the **same** `M` the turn's `[USER] - M` / `[MODEL] - M` carry (round 082) — and `N` is the count of tool calls the turn made. The label is literally `[TOOLS] - M (N calls)` for **all** `N` (no singularization: `(0 calls)`, `(1 calls)`), matching the operator's stated shape.

**D2 — the count is `len(Entry.Steps)`.** The reported figure is the persisted turn's executed tool steps — **not** `Entry.Calls` (the inference-round count, a different concept). Both are visible in the same entry, so the two must not be conflated; the round pins the distinction.

**D3 — the count rides a new `render.ListingMessage.ToolCount int` field (the `TurnIndex` precedent).** The CLI's `listingMessages` sets `ToolCount: len(e.Steps)` on the turn's message; the `internal/ui` adapter formats the line from the field and never touches the store. The change is additive (existing callers compile unchanged). The line is emitted immediately **before** the `[MODEL]` header, so a **partial** `-l N` window whose first listed message is the turn's answer still shows the line (rather than dropping it with the turn's `[USER]`).

**D4 — the fallback.** When the message's turn index is **non-positive** (a non-history caller), the adapter prints the bare `[TOOLS] (N calls)` label — the **exact** round-082 bare-label fallback shape (never a ` - 0` suffix).

**D5 — colour: yellow, whole label, the existing gate.** The line reuses the reference's yellow SGR (`colorYellow = "\033[0;33m"`) and the empty-safe `yellow()`, so it is accented on a terminal `stdout` with `-r` off and plain otherwise — the **same** gate as the role headers (round 073). The **whole** label (role word + index + count) is the colour unit (the round-082 whole-label rule). The accent never reaches `stderr`, and under `-r` the listing still carries no escape byte.

**D6 — spacing: its own block.** The line is separated from the `[USER]` block above and the `[MODEL]` header below by exactly one blank line — the round-073 "one blank line after every block" discipline, so it cannot read as part of the model header. In a **partial** listing that begins at a turn's `[MODEL]` message (an odd `-l N` window whose `[USER]` partner is outside it) the line is the listing's **first** line — there is no blank above it (the "above" separator exists only when the `[USER]` block is listed); the blank before the `[MODEL]` header is always present (fold N-086-2).

**D7 — `-l N` is unchanged: the line is a rider.** `-l N` still selects the last `N` **messages** (two per turn); the `[TOOLS]` line is never counted as a message. The selection semantics, the odd-`N` partial-turn behaviour, the body rendering, and the blank separator are unchanged.

**D8 — presentation-only.** No change to `history.jsonl`, the stores, `-b`/`--back`, `--new`, `-t`/`turns.log`, the interactive prompt, or the live turn chrome (`╭─⠿ Turn N`). No new exit code (the set stays **ten**); no new class phrase; the listing stays offline (no provider request, no stdin read). No new dependency (stdlib `fmt` + the existing colour helper).

**D9 — records.** ADR 0057 (+ index) **amends ADR 0045**; `specs/truth/techstack.md` *Output rendering* + *Session lifecycle flags* **MODIFY** (the listing presentation owner — a presentation fact, no behaviour guarantee, per upstream ADR 0006 Rule 6); the `history` CLI feature + `history/dsl.md` rows (`/axb-dsl-refine`) — including the **modification** of the *"tellme lists only the operator's messages"* row (the count is now surfaced; the tool **content** still is not); `docs/domain-model/**` MODIFY (the listing's colour invariant names the `[TOOLS]` accent; ADR 0041 same-PR rule); `specs/truth/contracts/**` + `specs/truth/data/**` **NOOP** (no API surface, no persisted-state change).

## Consequences

- `tellme -l` shows, per turn, `[USER] - M` → `[TOOLS] - M (N calls)` → `[MODEL] - M`, so the operator can see which turns were tool-heavy without opening `turns.log`, and the three lines of a turn share their index.
- No tool **content** (name, arguments, result) is surfaced — only an `int` count crosses the port, so the round-073 no-content rule holds in substance.
- The change is byte-local to the listing: the `-l` last-N-**messages** selection, the body rendering, the separator, `-b`/the stores, and the turn chrome are unchanged.
- The `history/dsl.md` row *"tellme lists only the operator's messages"* and the acceptance *"shows only the operator's messages"* Rule are **superseded in substance** — they now assert *no tool content*, not *no tool activity* (the round records this as a MODIFY, not a silent contradiction).
- No new dependency; stdlib-only; POSIX-only; `go.mod`/`go.sum` unchanged.
- **Recorded divergence:** `tell-me-go` surfaces the listing's tool activity as **per-call** `[Tool Call]` / `[Tool Response]` lines (`tell-me-go/internal/ui/history.go`). tellme adopts **one count line per turn** instead — a tellme-specific presentation divergence that neighbours disclosure **RF-073-4** (recorded in ADR 0045 §Forward) but does **not** close it (the shape differs).

## Alternatives considered

- **Print the line only when `N ≥ 1`** (a tool-less turn would show two lines) — rejected by the operator (chose *always print*, so the three-line rhythm is uniform and a plain turn is explicit about `(0 calls)`).
- **Count the line as a listed message for `-l N`** — rejected (D7: it would change the last-N-**messages** selection, break the two-messages-per-turn pairing, and let an odd `N` split a turn).
- **Tight grouping** (the line hugging the `[MODEL]` header with no blank line) — rejected by the operator (full separation keeps it from reading as a model header).
- **Reuse the `[MODEL]` header's ` - M` slot for the count** (`[MODEL] - M (N calls)`) — rejected: the ` - M` slot already means the backward turn index (round 082); a second meaning in one slot is exactly the ambiguity round 082 removed.
- **The reference's per-call `[Tool Call]` / `[Tool Response]` lines** — not adopted (that is disclosure **RF-073-4**; a different shape from the operator's request).
- **`Entry.Calls` as the count** — rejected (D2: the inference-round count is a different figure).

## Forward (non-blocking)

> **⚠ Not open work.** A `§Forward` entry is a decision *deferred to a trigger* or a recorded
> divergence — **not** tasking. Do not re-raise absent its trigger.

- **RF-086-1** — the `(N calls)` wording is fixed for all `N` (incl. `(1 calls)`, `(0 calls)`); a
  singularization (`1 call`) is a one-line adapter edit if an operator ever wants it.
- **RF-086-2** — only the count crosses the port; there is no per-call rendering, no tool name, and
  no timing/outcome breakdown — the reference's per-call `[Tool Call]` / `[Tool Response]` lines stay
  **RF-073-4** (not closed; a different shape).
- **RF-086-3** — the line is `-l`-only; `-t`/`turns.log` and the turn chrome are untouched (a change
  there would be a separate decision).
- **RF-086-4** — the count is derived from the persisted `history.Entry.Steps`; a legacy line with no
  `steps` field reads as `(0 calls)` (there is no way to distinguish "no tools" from "unrecorded").
- **RF-086-5** — the yellow is the reference's `colorYellow`; a distinct listing-only colour is not
  adopted (the chrome and the listing share the palette).
