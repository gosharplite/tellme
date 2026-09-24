# Research — round 086 `086-tools-listing-line`

`/axb-technical-research` — the technical decisions behind the `-l` listing's new tool-activity line.
No technology change: the round reuses the existing renderer, colour helper, and port pattern.

## Context

The listing is presented by the domain-owned `render.Listing` port; `internal/ui/listing.go` owns the
**bytes** (`[USER]` / `[MODEL]` headers, the rendered model body, the blank separator, the
terminal-gated accents) and `internal/cli` computes the per-message **figures** (round 082 added
`render.ListingMessage.TurnIndex`, computed by `listingMessages()`). Round 086 adds one more derived
figure — the turn's tool-call count — and one more line.

The count source is already persisted: `history.Entry.Steps []Step` (the ordered tool executions of a
completed turn; `specs/truth/data/data-model.dbml`, `history_step`). `len(Entry.Steps)` is the number
of tool calls the turn made.

## Decisions

- **D1 — the count is `len(history.Entry.Steps)`, NOT `Entry.Calls`.** `Entry.Calls` is the number of
  **AI-endpoint** calls (provider inference rounds): 1 for a tool-less turn, `1 + tool rounds`
  otherwise (round 027), and the round-017 turn header sums it. The two figures are different; the
  operator's `N calls` means **tool** calls. Reusing `Steps` also means no new store read and no
  schema change.
- **D2 — the figure rides a new `render.ListingMessage.ToolCount int` field (the `TurnIndex`
  precedent).** The CLI computes it (`listingMessages()` sets `len(e.Steps)`); the adapter stays
  presentation-only and never touches the store. The field is additive — existing callers/tests
  compile unchanged.
- **D3 — the line rides the turn's MODEL message.** The adapter emits `[TOOLS] - M (N calls)`
  immediately **before** the `[MODEL]` header, so in a full turn it lands **between** the `[USER]`
  block and the `[MODEL]` header (the operator's placement), and in a **partial** `-l N` window whose
  first listed message is the turn's answer, the line still appears (EC-002) rather than being
  silently dropped with its `[USER]` sibling.
- **D4 — the label text.** Literally `[TOOLS] - M (N calls)` for **all** `N` (no singularization:
  `(0 calls)`, `(1 calls)`), matching the operator's stated format verbatim (A2). `M` is the turn's
  backward turn index — the **same** `M` the turn's `[USER] - M` / `[MODEL] - M` carry, so the three
  lines of a turn agree. A **non-positive** `M` (a non-history caller) falls back to the bare
  `[TOOLS] (N calls)` — the exact round-082 bare-label fallback shape (no ` - 0`).
- **D5 — colour.** Reuse the reference's yellow SGR (`colorYellow = "\033[0;33m"`) and the existing
  empty-safe `yellow()` wrapper — **no new constant**. The **whole** label (role word + index +
  count) is the colour unit, matching the round-082 whole-label rule. The gate is the **exact**
  round-073 one (`spec.Colour && !spec.Raw`: a terminal `stdout` AND `-r` off); under `-r` or a
  redirected stdout the line is plain, and the accent never reaches `stderr`.
- **D6 — spacing.** The line gets the round-073 treatment of every listed block: **one blank line
  each side** (its own block), so it cannot read as part of the `[MODEL]` header. Full separation was
  the operator's choice (3A).
- **D7 — `-l N` is unchanged: the line is a rider.** `-l N` still selects the last `N` **messages**
  (two per turn); the `[TOOLS]` line is never counted as a message, so the selection semantics, the
  E2E selection Thens, and the odd-`N` partial-turn behaviour are all untouched (the operator's
  choice, 2A).
- **D8 — records & scope.** **ADR 0057** (new; **amends ADR 0045**) records the line, the placement,
  the count source, the colour, and the divergence from the reference; `specs/truth/techstack.md`
  *Output rendering* / *Session lifecycle flags* rows are **MODIFY**-ed (a presentation fact — no
  behaviour guarantee, per upstream ADR 0006 Rule 6); the `history` feature + `history/dsl.md` rows
  are the **`/axb-dsl-refine`** owner's; `docs/domain-model/**` colours the listing's `[TOOLS]` accent
  (ADR 0041 same-PR rule); `specs/truth/contracts/**` + `specs/truth/data/**` are **NOOP** (no API
  surface, no persisted-state change).

## Divergence (recorded, not parity)

`tell-me-go` surfaces the listing's tool activity as **per-call** `[Tool Call] <name>` /
`[Tool Response] <name>` lines (`tell-me-go/internal/ui/history.go`). tellme adopts **neither** — it
emits **one count line per turn**, a form the operator chose. This is a **tellme-specific
presentation divergence**; it is adjacent to disclosure **RF-073-4** (recorded in ADR 0045 §Forward:
*"`[Tool Call]` / `[Tool Response]` listing parity (not adopted; supersedes only if requested)"*) but
does **not** close it, because the shape differs.

## Risks

- **R1 — the line could leak tool content.** Mitigated by construction: only an `int` count crosses
  the port (the tool name/arguments/result never leave the store); FR-007 pins the negative.
- **R2 — a partial listing's placement.** Mitigated by D3 (the line rides the MODEL message) and
  pinned by EC-002.
- **R3 — the count could be confused with `Calls`.** Mitigated by D1 and pinned by the CLI wiring pin
  (FR-009) using a `Steps`-only fixture whose `Calls` differs from `len(Steps)`.

## Claim → Witness preview (for `/axb-tasks`)

Every clause is **observable** through the CLI/listing, so it is carried by a `[BDD-GREEN]` (E2E)
or a unit pin — **no `[WITNESS]` task and no `accepted-unwitnessed` record** is required this round.
