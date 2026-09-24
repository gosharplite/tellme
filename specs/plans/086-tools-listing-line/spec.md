# Round 086 — `086-tools-listing-line`

**Theme**: the `-l`/`--list` listing gains a per-turn **tool-activity line** —
`[TOOLS] - M (N calls)`, **yellow**, printed **between** the turn's `[USER]` block and its `[MODEL]`
header — so a reader can see how much tool work a turn did without opening the turn log.
**Presentation-only** on the offline listing path: no provider request, no persisted-state change, no
new dependency.

**Anchor**: **operator request** (rounds 073/074/075/078/085 precedent — no anchor issue). The
operator locked the notation, placement, and the three sub-decisions in-session (§7 assumptions).

---

## 1. Why this round

The round-073 listing (`ADR 0045`) presents each turn as a role header line (`[USER]` / `[MODEL]`,
round 082 adds the backward turn index ` - M`) plus a body. It deliberately **omits** the tool
activity: the `history/dsl.md` row *"tellme lists only the operator's messages"* asserts stdout
carries **no** tool-step text (round-073 clarify **Q2 → A**), and the acceptance
`features/acceptance/*` journey is titled *"Listing after a tool-using turn shows only the operator's
messages"*.

That omission is a **readability loss** for a tool-heavy session: a reader cannot tell, from the
listing alone, which turns were tool rounds — the information is persisted (`history.Entry.Steps`)
but never surfaced. Round 086 surfaces a **single derived figure** — the number of tool steps the
turn recorded — as one compact line, without leaking any tool **content** (the round-073 no-content
rule is preserved; see FR-007).

This is a **tellme-specific presentation divergence**, not reference parity: `tell-me-go` emits
per-call `[Tool Call]` / `[Tool Response]` lines
(`tell-me-go/internal/ui/history.go`), which tellme does **not** adopt. It neighbours disclosure
**RF-073-4** (the `[Tool Call]`/`[Tool Response]` parity item) but takes a **count-shaped** form the
operator chose; it does **not** close RF-073-4.

## 2. The change

1. **`internal/ui/listing.go`** — for each **model** message, emit the turn's tool-activity line
   before the `[MODEL]` header:
   - `[TOOLS] - M (N calls)` when the message carries a positive turn index `M` (round-082 shape); a
     non-positive index falls back to the bare `[TOOLS] (N calls)` (EC-001).
   - the whole label is accented **yellow** (`\033[0;33m … \033[0m`) under the round-073 gate
     (`Colour && !Raw`); plain otherwise (FR-004/FR-005).
   - one blank line **before** and **after** the line — its own block (the round-073 separator
     discipline; FR-003).
2. **`internal/domain/render/ports.go`** — `ListingMessage` gains `ToolCount int` (the turn's
   tool-step count; the `TurnIndex` precedent — the CLI computes it, the adapter stays
   presentation-only). The `Listing` doc comment names the new line.
3. **`internal/cli/cli.go`** — `listingMessages()` sets `ToolCount: len(e.Steps)` on the turn's
   **model** message (the count is the persisted turn's executed tool steps — FR-009).
4. **Truth** — `specs/truth/features/cli/history/dsl.md` gains the round-086 rows and its prologue
   note; `inspecting-the-session-history.feature` gains a Rule/Examples and its
   *"shows only the operator's messages"* Rule is **modified** to the new contract (the count is
   surfaced; the tool **content** is not).
5. **Records** — **ADR 0057** amends **ADR 0045**; `specs/truth/techstack.md` *Output rendering* row
   notes the listing line; `docs/domain-model/**` colours the listing's `[TOOLS]` accent
   (ADR 0041 same-PR rule); `docs/decisions/README.md` gains the ADR 0057 row.

## 3. Requirements

### 使用者故事 1 — 我在 `-l` 看到每個回合的工具活動 (Priority: P1)

The operator lists their session and can see, at a glance, how many tool calls each turn made.

**驗收情境**

1. **Given** a session history holding a tool-using exchange, **When** the operator lists the last
   messages, **Then** the turn's block shows `[TOOLS] - M (1 calls)` between `[USER]` and `[MODEL]`.
2. **Given** a session history holding a plain (step-free) exchange, **When** the operator lists the
   last messages, **Then** the turn's block shows `[TOOLS] - M (0 calls)`.

**功能需求（FR）**

- **FR-001**: the `-l` listing MUST print, for every listed turn, a tool-activity line
  `[TOOLS] - M (N calls)` positioned **between** the turn's `[USER]` block and its `[MODEL]` header,
  where `M` is the turn's backward turn index (identical to the `[USER] - M` / `[MODEL] - M` index)
  and `N` is the number of tool steps the turn recorded. [Verification Intent: observable → 驗收情境 1]
- **FR-002**: the line MUST be printed for **every** listed turn, including a turn with **zero** tool
  steps (`[TOOLS] - M (0 calls)`). [Verification Intent: observable → 驗收情境 2]
- **FR-003**: the line MUST be its own block — exactly one blank line separates it from the `[USER]`
  block above and from the `[MODEL]` header below (the round-073 separator discipline).
  [Verification Intent: observable → 驗收情境 1]
- **FR-004**: on a terminal `stdout` with `-r` off, the **whole** `[TOOLS] - M (N calls)` label MUST
  be accented yellow (`\033[0;33m … \033[0m`); the accent MUST never reach `stderr`.
  [Verification Intent: observable → 驗收情境 1]
- **FR-005**: under `-r` (or when `stdout` is not a terminal) the line MUST be printed **plain** — no
  escape byte. [Verification Intent: observable → 驗收情境 1]
- **FR-006**: the line MUST NOT change `-l N` selection — `-l N` still selects the last `N`
  **messages** (two per turn); the `[TOOLS]` line is a **rider** and is never counted as a message.
  [Verification Intent: observable → 驗收情境 1]
- **FR-007**: the line MUST surface **only** the count — never any tool **step content** (the tool
  name, its arguments, or its result). [Verification Intent: observable → 驗收情境 1]
- **FR-008**: the line MUST NOT be emitted by any other surface — not `-t`/`turns.log`, not the turn
  chrome. [Verification Intent: observable → 驗收情境 1 (negative)]

### 全域需求

#### 功能需求

- **FR-009**: the reported count MUST be the **persisted turn's executed tool steps**
  (`len(history.Entry.Steps)`) — NOT the entry's `calls` field (the AI-endpoint inference-round
  count, a different figure). [Verification Intent: observable → 驗收情境 1]

#### 非功能需求

- **NFR-001**: the `-l` listing MUST remain strictly **offline** — no provider request (the round-007
  contract). [Verification Intent: observable → the existing `tellme sends no request to any provider` row]
- **NFR-002**: the round MUST add **no dependency** (stdlib-only) — a presentation change.
  [Verification Intent: unobservable → `go.mod`/`go.sum` unchanged + `make lint`/`vet`]

### 邊界情況

- **EC-001**: for a non-history caller (a non-positive `M`), the line MUST print the bare
  `[TOOLS] (N calls)` label — never a ` - 0` suffix (the round-082 bare-label fallback).
  [Verification Intent: unobservable → unit pin]
- **EC-002**: a partially-listed turn (an odd `-l N` window whose first listed message is the turn's
  `[MODEL]`) MUST still show that turn's `[TOOLS]` line — the line rides the turn's MODEL message.
  [Verification Intent: observable → 驗收情境 1]
- **EC-003**: a legacy history line with **no** `steps` field MUST read as zero tool calls
  (`(0 calls)`), never an error. [Verification Intent: observable → 驗收情境 2]

### 關鍵實體

- **`history.Entry.Steps`** — the persisted turn's ordered tool steps; `len(Steps)` is the count.
  No new entity (the round surfaces a derived figure of an existing record).

## 4. Invariants

- **I-1 — `-l` selection semantics unchanged.** The last-`N`-**messages** selection, the two
  messages per turn, the true backward turn index, and the body rendering are unchanged; the
  `[TOOLS]` line is a rider (FR-006).
- **I-2 — No tool content is surfaced.** Only a count leaves the store (FR-007); the round-073
  no-content rule holds in substance (the *content* is still omitted; only a **derived figure** is
  added).
- **I-3 — Offline.** No provider request on the listing path (NFR-001).
- **I-4 — Colour is terminal-gated.** The yellow accent follows the exact round-073 gate
  (terminal `stdout` AND `-r` off); it never enters `stdout` when redirected and never reaches
  `stderr` (FR-004/FR-005).
- **I-5 — `-t`/turns.log and the turn chrome untouched** (FR-008).
- **I-6 — stdlib-only / POSIX-only / hermetic** (NFR-002).

## 5. Scope

**In**: the listing adapter's `[TOOLS]` line; the `ListingMessage.ToolCount` field; the CLI
projection wiring; the listing truth rows/feature; ADR 0057; the `techstack.md` and
`docs/domain-model/**` records; the E2E + unit carriers.

**Out**: any change to `-b`/`--back`, the store(s), `--new`, `-t`, the turn chrome, the answer path,
or the `-l N` count semantics; the reference's per-call `[Tool Call]`/`[Tool Response]` lines (that
is disclosure **RF-073-4**, not this round); `go.mod`/`go.sum`.

## 6. Success criteria

- **SC-001** A tool-using turn lists `[TOOLS] - M (1 calls)`; a plain turn lists `[TOOLS] - M (0 calls)`
  (the count line is present and correct for both).
- **SC-002** The listing's message shape is unchanged: `-l N` selects the last `N` messages; the E2E
  scenario count is unchanged (only Examples are added, none removed).
- **SC-003** `make check` (verify + test) is green; `go.mod`/`go.sum` unchanged.

## 7. Assumptions

- **A1** The count is `len(history.Entry.Steps)` (executed tool steps), **not** `Entry.Calls` (the
  AI-endpoint inference-round count) — the two are different figures and the operator's `N calls`
  means tool calls.
- **A2** The label is literally `[TOOLS] - M (N calls)` for **all** `N` — no singularization
  (`(1 calls)`, `(0 calls)`) — matching the operator's stated format verbatim.
- **A3** No `NEEDS CLARIFICATION`: the operator locked the notation (`[TOOLS] - M (N calls)`), the
  placement (between `[USER]` and `[MODEL]`), and — asked one at a time and answered — the zero case
  (**always print**, B), the `-l N` treatment (**rider**, A), and the spacing (**full separation**,
  A).
- **A4** The line rides the turn's **`[MODEL]`** message, so a partially-listed turn that begins at
  its answer still shows its tool line (EC-002).
