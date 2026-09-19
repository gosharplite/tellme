# Feature Specification: Prompt suggestion parity — the reference's newest-50 candidate pool (round 064)

**Feature Branch**: `064-prompt-suggestion-parity`

**Created**: 2026-09-20

**Status**: Draft — produced by `/axb-specify` from an **operator request** (no anchor issue). **Clarify CLOSED** — **Q1 → A** (pool-only parity: widen the recent-prompt candidate pool to the newest 50; the empty-query count and the session-prompt source are **not** adopted) locked by the operator (2026-09-20). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-20, this session)**:

> *"In `-i` TUI, when I type 'commit', there are only 3 suggestions. Why?"* → (grounded answer: the candidate pool is the newest **10** distinct prompts; the reference reads the newest **50**) → *"I want behaviour only. I want to see 10 instead of 3 suggestions in tellme."*

**Anchor issue**: none (operator request). The round is a **behaviour-only** parity slice against [`tell-me-go`](https://github.com/gosharplite/tell-me-go) — it adopts the reference's **suggestion-source depth** (a deep, newest-first candidate pool) without porting the reference's code structure, its workspace-policy plumbing, or its log compaction.

**Behaviour intent**: **MODIFY (the `-i` suggestion engine's history candidate pool: 10 → the newest 50 distinct prompts).** Everything else about the suggestion surface is unchanged: the same three sources (recent prompts, workspace entries for path-like queries, registered tool names), the same case-insensitive **subsequence** match, the same **dedup-and-cap-at-10** output, the same debounce/drop-over-long/no-pre-selection rendering. The one observable change is **how deep the recent-prompt pool is** — which is exactly why typing `commit` surfaces **3** items today and **10** (the cap) under the reference.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `4092ae7`)*

| Site | Current shape |
| --- | --- |
| `internal/app/suggestions/service.go` | `const maxSuggestions = 10`; `Suggest` fills an accumulator cap 10 — `addPrompts` calls `s.prompts.RecentPrompts(ctx, maxSuggestions)` (**pool = 10**), then `addWorkspace` (path-like queries only), then `addTools`. An **empty** query returns prompts only (≤10). |
| `internal/app/suggestions/adapters.go` | `TrackerPrompts.RecentPrompts(ctx, n)` → `PromptTracker.Recent(ctx, n)` (newest-first, **deduped**, stop at `n`); `OSSWorkspace` (hardcoded ignore list `.git`/`node_modules`/…, plain entry names, 100-entry batches); `RegistryTools` (registered tool names). |
| `internal/domain/history/tracker.go` | Port `PromptTracker.Recent(ctx, n) ([]PromptLogEntry, error)` — the newest-first, deduped, bounded read of the shared user-global log. |
| `internal/domain/suggestions/suggestions.go` | Port `Service.Suggest(ctx, query) []Suggestion` — "up to a fixed cap of candidates"; "an empty query yields the newest recent prompts". |
| `internal/cli/cli.go` (`runInteractiveTUI`) | `engine := appsuggestions.New(TrackerPrompts{Tracker: tracker}, OSSWorkspace{}, RegistryTools{Registry: reg})`; the prompt path deliberately performs **no startup disk I/O** (the round-016 dashboard-header retirement, architect D3). |
| `internal/ui/tui/prompt/{model.go,suggester.go}` | Debounced, cancelable refresh; entries spanning >3 lines dropped; **no** pre-selection (`noChoice`); cap enforced by the engine. |
| `specs/truth/techstack.md` — *Prompt suggestion engine* | "aggregates recent prompts (shared log newest-first + active session), workspace paths (path-like queries only …), and registered tool names; **subsequence** match, **deduped**, capped at **10**." |
| `tell-me-go` (reference) | `NewMultiSourceSuggestionService` **pre-loads** `tracker.LoadTopN(ctx, 50)` (newest-first, deduped, in memory) and merges the active session's recent prompts; `GetSuggestions` searches that slice by the **same** case-insensitive **subsequence** match, cap **10**; an **empty** query returns the **first 5**; workspace entries are scanned only for a path-like query via an injected `WorkspacePolicy`; **no** tool-name source; the log self-compacts above **150 KiB**. |

**Reproduction of the gap (grounded):** with a real `~/.tellme/global_prompts.jsonl` (471 records / 259 distinct prompts, 43 distinct containing "commit"), typing `commit` in `-i` yields **3** suggestions — the subsequence matches among only the **10 newest** distinct prompts. A simulation of the reference's algorithm (newest **50**, same subsequence, cap 10) on the same file yields **10** (the cap). The gap is the **pool depth**, not the matching rule.

---

## Design (S-1 is the operator's locked intent; the rest is proposed — Q1 pending)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **The recent-prompt candidate pool deepens to the reference's newest 50 distinct prompts** (was 10); the surfaced list stays **capped at 10**. One named constant owns the depth (no scattered magic numbers). | **locked (operator intent)** |
| **S-2** | **The matching discipline is unchanged** — case-insensitive ordered-subsequence, dedup, cap 10, prompts-first ordering, workspace-only-for-path-like, tool names after. Only the *candidate depth* changes. | proposed |
| **S-3** | **The source set is unchanged** — prompts (**deepened** to 50) + workspace + tools. The reference's empty-query-first-5 and the session-last-prompt source are **not** adopted (**Q1 → A**); tellme's empty query keeps today's ≤10 and the CLI keeps its deliberate *no-startup-disk-I/O* property. | **locked (Q1 → A)** |
| **S-4** | **tellme's own features are kept** (recorded divergences): the extra **tool-name** source (the reference has none) and the round-028 user-global log location (`~/.tellme/`, ADR 0004). | proposed |
| **S-5** | **The workspace source and compaction are out of scope** — tellme keeps its `OSSWorkspace` ignore list and plain names; the reference's `WorkspacePolicy` plumbing and its ≈150 KiB self-compaction are not adopted this round (recorded as forward items). | proposed |
| **S-6** | **No new source, port, or dependency** — the change is confined to the existing engine + its history-prompt source; `internal/cli`'s "no startup disk I/O" note is preserved unless Q1 selects the session source. | proposed |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — **The rendered output cap is unchanged**: the prompt never shows more than **10** suggestions.
- **I-2** — **The matching rule is unchanged**: case-insensitive subsequence, deduped; a numbered acceptance criterion must not weaken into substring/word matching.
- **I-3** — **The prompt stays hermetic and offline**: the suggestion engine performs **no network** I/O; the shared log read is the only disk read on this path.
- **I-4** — **The `-i` interaction surface is unchanged**: debounce, cancellation, >3-line drop, `Suggestions:` header, no pre-selection, the `Tab`/`Esc`/`Ctrl+S` keybindings — all as today (a suggestion-content change only).
- **I-5** — **A missing/empty/unreadable log degrades to no suggestions**, never an error (unchanged).

---

## Clarify (CLOSED — Q1 → A)

> Per `/axb-clarify`, only **high-impact** gaps are asked; one question at a time.

| # | Question | Why it is high-impact | Answer |
| --- | --- | --- | --- |
| **Q1** ✅ **ANSWERED (A)** | What is the exact scope of **"behaviour only"** — **(A)** widen the recent-prompt pool to the newest **50** and nothing else, or **(B)** also adopt the reference's **other** source behaviours: an **empty** query shows the **first 5** **and** the **active session's last user prompt** is merged as an extra source? | It decides the requirement set (US2 appears or not), whether the CLI's deliberate "no startup disk I/O" property is preserved, and the acceptance scenarios. | **Operator chose (A) — pool-only** (2026-09-20). The recent-prompt pool deepens to the newest 50; the surfaced list stays capped at 10; the empty-query count stays today's; **no** session source is added, so the *no-startup-disk-I/O* property is preserved. Recorded as **S-1/S-3 (locked)** / **FR-001…FR-004**. **US2 and FR-005…FR-007 are dropped** (recorded as forward items). |

> **Q1 → A consequence**: US2 and **FR-005…FR-007** are **removed** from the requirement set; the empty-query count and the session source are carried as **forward items** (see *Out of scope*). `checklists/requirements.md` records the resolution.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - typing a query surfaces as many candidates as the reference (Priority: P1)

As the **operator** using the `-i` interactive prompt, I want the suggestion list to search a **deep** recent-prompt pool (the newest 50 distinct prompts, like the reference), so that typing a term like `commit` shows the **10** matches the reference shows instead of the **3** the shallow 10-prompt pool exposes.

**Why this priority**: it is the operator's request and the round's entire point.

**Independent verification**: seed the shared prompt log with more distinct recent prompts than the shallow pool holds, where only an **older-than-10** prompt matches the typed query; before the round the prompt does **not** offer it (the 10-pool never sees it), after the round it **does**. A falsifiability witness restores the 10-prompt depth and the check goes red.

**Acceptance Scenarios**:

1. **Given** a shared prompt log whose newest **50** distinct prompts contain at least 11 that match a query by subsequence, **When** the operator types that query at the interactive prompt, **Then** the prompt offers **10** suggestions (the cap) — including at least one drawn from **beyond** the newest 10.
2. **Given** the same log, **When** the typed query matches only a prompt recorded **deeper than the newest 10** but within the newest 50, **Then** the prompt offers that prompt (it was absent under the 10-prompt depth).
3. **Given** a log with fewer than 10 matching prompts, **When** the operator types, **Then** every match in the pool is offered and the list is still ≤10.

**Functional Requirements**:

- **FR-001**: The recent-prompt source MUST make available the newest **50** distinct prompts (newest-first, deduped) as the matching pool — not 10.
- **FR-002**: The surfaced suggestion list MUST remain capped at **10** (the cap is unchanged).
- **FR-003**: The match MUST remain a case-insensitive ordered **subsequence**; the round MUST NOT introduce substring/word matching or a scoring/ranking change.
- **FR-004**: The prompt pool depth MUST be a **single named constant** (one owner), so the value is changed in one place and asserted directly.
- **FR-005**: The public suggestion port (`domain/suggestions.Service`) MUST keep its contract (cap-only semantics); the deepened pool is an **engine detail**, not a port change.

### (Recorded, not adopted) the reference's other source behaviours — dropped by Q1 → A

Under **Q1 → A** the round adopts **only** the candidate-pool depth. The reference's other result-visible source behaviours are **deliberately not adopted this round** and are recorded as forward items:

- **An empty editor shows the first 5 recent prompts** (the reference's empty-query limit). tellme keeps today's empty-query behaviour (up to 10). *Forward item.*
- **The active session's last user prompt is merged as an extra recent-prompt source.** Not adopted, because it would reintroduce a startup history read (superseding the round-016 D3 "no startup disk I/O" property). *Forward item.*

*(These were the only other result-visible source behaviours; the workspace-policy plumbing, the tool-source, and log compaction are recorded in Out of scope.)*

## Edge Cases

- **Exactly 10 matching prompts** — the cap yields all 10; one fewer yields that many; the boundary is pinned both ways.
- **Fewer than 50 distinct prompts in the log** — the pool is the whole log (bounded by what exists); no error.
- **A prompt recorded deeper than 50** — still **not** offered (the pool is bounded at 50, matching the reference); only depth, not the whole file.
- **Empty/missing/unreadable log** — degrades to no prompts (unchanged, I-5); workspace/tool sources still apply.
- **Duplicates across the newest-50 window** — deduped newest-first before matching (unchanged).
- **A query matching nothing** — an empty suggestion list, no error.
- **An >3-line recent prompt** — dropped by the renderer (unchanged, I-4).
- **A very long file** — the read stays bounded by the pool depth, never the whole file materialised into the pool.

## Key Entities

- **The recent-prompt candidate pool** — the newest-first, deduped slice of the shared user-global prompt log that the engine matches against; its **depth** is the round's subject (10 → 50).
- **The suggestion engine (`internal/app/suggestions`)** — unchanged in structure: three sources, subsequence match, dedup, cap 10.
- **The suggestion port (`domain/suggestions.Service`)** — unchanged contract.
- **The shared global prompt log** — unchanged (round 028; `~/.tellme/global_prompts.jsonl`).

## Success Criteria

- **SC-001**: With a log whose newest-50 distinct prompts contain ≥10 subsequence matches for a query, the `-i` prompt offers **10** suggestions — including at least one drawn from **beyond** the newest 10 (the reported behaviour). *(Hermetic: E2E through the `TELL_ME_FORCE_STDIN_TTY` + `TELL_ME_TUI_DEBOUNCE=0` seams.)*
- **SC-002**: A red-capable carrier proves the depth: restoring the 10-prompt pool turns the SC-001-class check **red**.
- **SC-003**: The surfaced list is still **≤10** and the match is still subsequence-based (I-1/I-2) — pinned.
- **SC-004**: `make verify` + `go test -count=1 ./...` green (incl. the E2E contract); the topology/DSL audit adds **no** new findings; no new dependency.
- **SC-005**: Every changed behaviour has a reproduced **falsifiability witness**; long-standing `-i` behaviours (debounce, drop-over-long, no pre-selection, tool/workspace suggestions) are **unchanged**.

## Assumptions

- **A1**: The operator's "behaviour only" means **observable suggestion behaviour**, not a code-structure port; the reference's `WorkspacePolicy`, log compaction, and no-tool-source are **not** adopted (recorded divergences, S-4/S-5).
- **A2**: The **depth value 50** is taken from the reference (`LoadTopN(ctx, 50)`); the surfaced cap stays 10. If a different depth is wanted, it is a one-constant change (FR-004).
- **A3**: Plain line CLI round — `/axb-ui-plan` skipped; `/axb-api-plan` **NOOP**; `/axb-data-plan` expected **NOOP** (no persisted shape change; the log file's shape is unchanged).
- **A4**: Truth impact expected in `specs/truth/techstack.md` (*Prompt suggestion engine* row — the pool depth), possibly `specs/truth/features/cli/chat/prompting-with-suggestions.feature` + `chat/dsl.md` (a depth-distinguishing Example), and `docs/decisions/` if a durable decision is recorded (e.g. the pool depth / the no-startup-I/O supersession under Q1 → B). No ADR is expected under Q1 → A if no settled decision changes.
- **A5**: The round is testable **hermetically** (the existing fake suggestion source at unit level; the built binary behind the round-016 seams at E2E level); **no** real network/credential.
- **A6**: No new tool, no config key, no port change — the change is confined to the engine's history-prompt depth (plus the Q1-selected source behaviours).

## Out of scope (recorded forward items)

- **The reference's empty-query-first-5 behaviour** (Q1 → A) — tellme keeps today's empty-query limit.
- **The active session'`s last-user-prompt source** (Q1 → A) — not adopted; it would reintroduce a startup history read (round-016 D3).
- **The reference's `WorkspacePolicy` ignore set + trailing-separator dir form** (S-5) — tellme keeps its hardcoded ignore list and plain names.
- **The reference's ≈150 KiB log self-compaction** (S-5) — tellme's `wg` drain hook stays a reserved no-op.
- **Adopting the reference's log location** (`$TELL_ME_HOME/output/global_prompts.jsonl`) — tellme keeps `~/.tellme/` (ADR 0004, a recorded divergence).
- **Dropping tellme's tool-name suggestion source** (the reference has none) — kept (S-4).
- **The round-016 `-i` visual surface** (border/placeholder/styling) — unchanged.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion).
