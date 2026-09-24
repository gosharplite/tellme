# Truth delta — round 086 `086-tools-listing-line`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | **ADD** a `## Then (round 086)` headed table with two rows — `the listing reports each turn's tool activity` (asserts `[TOOLS] - M (N calls)` between `[USER]` and `[MODEL]`, `N = len(Steps)` from the arranged turns + the `-l N` request) and `the listing accents the tool-activity line in yellow` (the whole-label yellow, the round-073 gate). **ADD** the round-086 prologue note. **MODIFY** the `Then` row `tellme lists only the operator's messages` → `tellme lists the tools' activity but not their contents` (the count is surfaced; the tool **content** is not). **MODIFY** the `the listing carries no accents` row to cover the yellow accent. **MODIFY** the `the session history already holds a tool-using exchange` Given to record the arranged turn's tool-step count. | The `-l` listing gains a per-turn tool-activity line (round 086; ADR 0057) — the round-073 "tool activity omitted" reading is **superseded in substance** (the count is now surfaced; the content still is not). |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | **ADD** 3 Rules / 6 Examples: *Listing a tool-using turn shows the tool count but none of the tool's contents*; *A listed turn reports how many tool calls it made* (a tool-using turn ⇒ `(1 calls)`; a plain turn ⇒ `(0 calls)`; a partial listing keeps the answer's line); *The tool-activity line is accented only when the listing goes to a terminal*. **MODIFY** the `Rule: Listing after a tool-using turn shows only the operator's messages` → the count-but-not-content Rule (the round-073 Q2 → A reading amended). **MODIFY** the feature header comment. | Same as above — the executable interface truth for the new line. |
| NOOP | `specs/truth/features/cli/{chat,configuration,diagnostics,usage,workspace}/**` | Checked — untouched (their rows/features are unaffected by a `history`-module-only presentation change). | Scope guard (I-5). |
| NOOP | `specs/truth/features/cli/dsl.md` (interface root) | Checked — no new cross-module row is introduced (the round's rows are `history`-module-only). | `dsl-single-authority` (no shared row). |

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*Output rendering*) | Note that the listing's yellow tool-activity line reuses the reference's `colorYellow` through the existing empty-safe `yellow()` helper — no new colour constant. | A presentation fact (no behaviour guarantee — upstream ADR 0006 Rule 6). |
| MODIFY | `specs/truth/techstack.md` (*Session lifecycle flags*) | Note the round-086 `[TOOLS] - M (N calls)` line: its shape, its placement, `N = len(history.Entry.Steps)`, the always-print rule, the yellow gate, and that it is a **rider** (the `-l N` last-`N`-messages selection is unchanged). | The `-l` listing presentation owner (round 073/082 precedent). |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no technology change; no new dependency. | stdlib-only (NFR-002). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a pure CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change; the line reads the existing `history.Entry.Steps` (round 008; schema unchanged). | `/axb-data-plan` NOOP. |
| NOOP | `ui/**` | Checked — no user-facing UX *plan* surface changes (a plain line-oriented CLI; `/axb-ui-plan` skipped). | CLI, plain line-oriented. |
