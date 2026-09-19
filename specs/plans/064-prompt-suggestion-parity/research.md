# Research — Prompt suggestion parity: the reference's newest-50 candidate pool (round 064)

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Anchor issue**: none (operator request)
**Truth owner**: `/axb-technical-research` — updates `specs/truth/techstack.md`

## Problem (grounded)

tellme's `-i` suggestion engine searches only the **newest 10 distinct** recorded prompts, so an older matching prompt can never be offered. Concretely (`internal/app/suggestions/service.go`): `const maxSuggestions = 10`, and `addPrompts` calls `s.prompts.RecentPrompts(ctx, maxSuggestions)` — the **same** constant governs both the *candidate pool depth* **and** the *surfaced cap*. Measured against the operator's real `~/.tellme/global_prompts.jsonl` (471 records / 259 distinct prompts; 43 contain "commit"), typing `commit` yields **3** suggestions — the subsequence matches among only the 10 newest distinct prompts.

The reference (`tell-me-go`, `internal/app/suggestions/service.go`) is different in exactly one load-bearing way: `NewMultiSourceSuggestionService` **pre-loads** `tracker.LoadTopN(ctx, 50)` (the newest **50** distinct prompts, newest-first, deduped, in memory), then `GetSuggestions` searches that slice with the **same** case-insensitive ordered-**subsequence** match and the **same** cap of **10**. Simulating the reference algorithm on the same file yields **10** (the cap). The matching rule, the dedup, the cap, the >3-line drop, the no-pre-selection rendering and the workspace-only-for-path-like rule are **already** aligned; the **pool depth** is the sole gap.

The operator's request ("behaviour only … 10 instead of 3") and clarify **Q1 → A** scope the round to that one behaviour.

## Decisions taken upstream (`/axb-clarify`, one at a time → `spec.md` S-1…S-6)

| # | Decision |
| --- | --- |
| **Q1 → A** | **Pool-only parity**: deepen the recent-prompt candidate pool to the newest **50** distinct prompts; the surfaced cap stays **10**. The reference's empty-query-first-5 and its session-last-prompt source are **not** adopted (recorded forward items); the CLI keeps its deliberate *no-startup-disk-I/O* property. |

## Decisions (this phase)

### D1 — One constant owns the depth; the surfaced cap is a distinct constant

- **Decision**: split the single `maxSuggestions = 10` into two named constants in `internal/app/suggestions`:
  - `promptPoolDepth = 50` — how many newest distinct prompts the **history source is asked for** (the reference's `LoadTopN(ctx, 50)`);
  - `maxSuggestions = 10` — the **surfaced cap** (unchanged).
  `addPrompts` asks the source for `promptPoolDepth`; the accumulator still caps at `maxSuggestions`.
- **Rationale**: FR-004 requires a single owner for the depth; keeping the cap as its own constant makes "deepen the pool, keep the cap" explicit and prevents a future edit from silently moving the surfaced limit. The two concepts were conflated; this round separates them.
- **Alternatives considered**: (a) hardcode `50` at the call site — rejected (FR-004: one named owner); (b) change `maxSuggestions` to 50 — **rejected**: it would raise the *surfaced* count to 50, violating I-1/FR-002 and the operator's "10, not 3".

### D2 — Read shape: keep tellme's lazy per-query read; do not adopt the reference's pre-load

- **Decision**: the history source stays tellme's `RecentPrompts(ctx, promptPoolDepth)` (a fresh newest-first, deduped, bounded read per query), **not** the reference's construction-time `LoadTopN(50)` + in-memory slice.
- **Rationale**: `Recent(ctx, n)` already returns the newest-first, deduped, **bounded** slice the engine needs, so widening the request from 10 to 50 is a **one-argument** change — the minimal-risk edit. The reference's pre-load exists to amortize `LoadTopN` across queries; tellme's per-query read is bounded by `n` and already yields to `ctx`, so adopting the pre-load is a structure port with no observable behaviour gain (and would widen the CLI's construction surface). Q1 → A scopes the round to behaviour; D2 honours that.
- **Alternatives considered**: (a) pre-load `LoadTopN(50)` in `runInteractiveTUI` and pass the slice — a faithful structure port, but not required by behaviour and it touches the CLI construction seam; rejected as out of scope (recorded).

### D3 — Keep the source set, the ordering, and the empty-query behaviour (Q1 → A)

- **Decision**: no source is added or removed — prompts (deepened) → workspace (path-like only) → tools, in that order; an **empty** query returns today's prompts-only list (≤10); **no** session-history read is added.
- **Rationale**: Q1 → A; keeps the *no-startup-disk-I/O* property (round-016 architect D3) intact and the change minimal.
- **Alternatives considered**: the reference's empty-query-first-5 and session source — **dropped** (Q1 → A; recorded).

### D4 — The matching rule and the rendered surface are untouched

- **Decision**: the case-insensitive **subsequence** match, the dedup, the >3-line drop, the debounce/cancellation, the `Suggestions:` header, the no-pre-selection sentinel and the keybindings are **unchanged**. Only the candidate pool depth moves.
- **Rationale**: they are already aligned with the reference (the grounding above proves it); changing them would be scope creep beyond the operator's ask.
- **Residual**: tellme's extra **tool-name** source (the reference has none) is a **kept** divergence (recorded, S-4).

### D5 — Verification: a depth-distinguishing carrier (unit + E2E), no new gate member

- **Decision**: pin the depth at the **unit** level (the engine asks the source for `promptPoolDepth`, and the accumulator caps at `maxSuggestions`) and at the **E2E** level (a scripted `-i` run whose pool contains a matching prompt **beyond the newest 10** offers it — the round-016 `TELL_ME_FORCE_STDIN_TTY` + `TELL_ME_TUI_DEBOUNCE=0` seams). A falsifiability witness restores the depth to 10 and the carrier goes red. No new `make verify` member; no new dependency.
- **Rationale**: SC-001/SC-002; the depth is only observable when a match sits beyond the shallow window, so the fixture must place one there.
- **Residual**: the E2E fixture must seed **>10** distinct prompts with the match **outside** the newest 10 — the natural carrier is the shared-log Givens (each `the shared prompt log already holds "…"` appends a record).

### D6 — Governance: a short ADR records the depth + the kept divergences

- **Decision**: **ADR 0034** records (a) the pool depth (10 → 50, cap 10) as a behaviour-parity decision with the reference, and (b) the **kept divergences** (no session source, no empty-query-first-5, no `WorkspacePolicy`, no compaction, tellme's tool source, the `~/.tellme/` log location). No prior ADR is superseded.
- **Rationale**: the repo records operator-request behaviour decisions as short ADRs (the 0023/0024/0027/0028/0032/0033 pattern); this round changes an observable behaviour and settles a divergence boundary worth a durable home.
- **Alternatives considered**: record it only in truth + forward items — rejected; the divergence boundary (pool-only) is exactly the kind of call a future reader must not have to re-derive.

## Truth impact

- **MODIFY** `specs/truth/techstack.md` — the *Prompt suggestion engine* row: the recent-prompt **candidate pool** is the newest **50** distinct prompts; the surfaced list remains **capped at 10** (the two are now distinct); everything else in the row (subsequence, dedup, the three sources, the empty-query behaviour, the >3-line drop, the no-pre-selection) is unchanged.
- **ADD** `docs/decisions/0034-prompt-suggestion-pool-depth.md` + the index row.
- **MODIFY** `specs/truth/features/cli/chat/prompting-with-suggestions.feature` + `chat/dsl.md` — a depth-distinguishing Example (a prompt beyond the newest 10 but within the newest 50 is offered) and its step row(s).
- **NOOP** `specs/truth/contracts/**` (no API) · **NOOP** `specs/truth/data/data-model.dbml` (the log record shape is unchanged).
- Domain model: expected **NOOP** (the product model's suggestion facts are source-shape, not depth); `make modelith-check` must stay green.
