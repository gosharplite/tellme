# ADR 0034 — Prompt suggestion parity: a deeper recent-prompt candidate pool (newest 50), the surfaced cap unchanged at 10

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** `tell-me-go` (the parity precedent: `NewMultiSourceSuggestionService` pre-loads `LoadTopN(ctx, 50)` and searches it by the same subsequence match, cap 10), [ADR 0004](0004-user-global-prompt-log.md) (the user-global prompt log tellme reads — a kept divergence), [ADR 0021](0021-ride-alongs-and-records.md) (the suggester's single-owned selection policy, untouched), round 015 (the suggestion engine), round 016 (the `-i` parity surface; the *no-startup-disk-I/O* retirement, D3), round 037/052 (the no-pre-selection policy), round 064 (`specs/plans/064-prompt-suggestion-parity` — this ADR's round)

## Context

The operator typed `commit` at the `-i` interactive prompt and saw only **3** suggestions, while the reference shows **10**. Grounding found the cause is **candidate-pool depth**, not the matching rule:

- tellme's engine (`internal/app/suggestions/service.go`) uses one constant, `maxSuggestions = 10`, for **both** the number of recent prompts it asks the history source for **and** the surfaced cap. So the matching pool is the **newest 10 distinct** prompts — anything older can never be offered.
- The reference pre-loads `LoadTopN(ctx, 50)` (the newest **50** distinct prompts) and searches it with the **same** case-insensitive ordered-**subsequence** match and the **same** cap of **10**.

Simulating both algorithms against the operator's real log (471 records / 259 distinct prompts; 43 contain "commit") reproduces the report exactly: pool **10** → **3** matches; pool **50** → **10** (the cap).

Everything else about the suggestion surface is **already** aligned or deliberately different: the subsequence match, dedup, cap 10, the >3-line drop, the debounce/cancellation, the no-pre-selection rendering, and the workspace-only-for-path-like rule are aligned; the tool-name source and the `~/.tellme/` log location are tellme's kept divergences; the reference's empty-query-first-5, its session-last-prompt source, its `WorkspacePolicy`, and its log compaction are not adopted.

A one-at-a-time clarify settled the scope: **Q1 → A** — pool-only parity (deepen the pool; everything else unchanged).

## Decision

**D1 — The recent-prompt candidate pool deepens to the newest 50 distinct prompts; the surfaced cap stays 10.** The engine asks its history source (`PromptTracker.Recent`) for the newest **50** distinct prompts (newest-first, deduped) and still surfaces **at most 10** suggestions. The two are now **distinct named constants** — `promptPoolDepth = 50` (the depth) and `maxSuggestions = 10` (the cap) — so the depth has a single owner and the surfaced limit cannot drift silently.

**D2 — The matching discipline is unchanged.** Case-insensitive ordered-**subsequence**, deduped, prompts-first (then workspace for path-like queries, then tool names), capped at 10. The round changes the *candidate depth* only.

**D3 — Read shape: tellme keeps its lazy, bound-per-query read.** The history source continues to be `RecentPrompts(ctx, promptPoolDepth)` — a fresh newest-first, deduped, **bounded** read per query — rather than the reference's construction-time `LoadTopN(50)` + in-memory slice. `Recent(ctx, n)` already returns the bounded slice the engine needs, so widening the requested count is the minimal, observable-only change; the reference's pre-load is a structure port with no behaviour gain (Q1 → A scopes the round to behaviour).

**D4 — The source set is unchanged.** No source is added or removed; **no** session-history read is introduced, so the CLI keeps its deliberate *no-startup-disk-I/O* property (round-016 D3). The empty-query behaviour stays today's (prompts only, ≤10).

**D5 — The rendered `-i` surface is untouched.** The debounce, cancellation, >3-line drop, `Suggestions:` header, no-pre-selection sentinel and keybindings are unchanged (a suggestion-*content* change only).

**D6 — Verification.** A **unit** pin (the engine asks the source for `promptPoolDepth`; the accumulator caps at `maxSuggestions`) and an **E2E** carrier (a scripted `-i` run whose pool holds a matching prompt **beyond the newest 10** offers it, through the round-016 `TELL_ME_FORCE_STDIN_TTY` + `TELL_ME_TUI_DEBOUNCE=0` seams). A falsifiability witness restores the depth to 10 and the carrier goes red. No new `make verify` member, no new dependency.

**D7 — Governance / kept divergences.** This ADR records the depth decision **and** the divergence boundary: tellme adopts **only** the reference's pool depth, keeping its tool-name source, its `~/.tellme/` log (ADR 0004), its hardcoded workspace ignore list, its plain entry names, and its no-compaction `wg` hook; the reference's empty-query-first-5 and session-last-prompt source remain unadopted forward items. No prior ADR is superseded.

**D8 — No new dependency; hermetic; stdlib-only.** The change is a constant split plus one argument; every gate stays offline.

## Consequences

- **Positive**: the operator's reported behaviour is fixed — a query like `commit` surfaces the reference's **10** (the cap) instead of **3**; the surfaced limit is provably unchanged (a separate constant); the edit is one argument wide.
- **Neutral / trade-off**: the per-query read now materialises up to 50 distinct prompts (bounded; still cheaper than the whole file). tellme's search remains strictly-local (no network) and the `-i` interaction surface is byte-unchanged.
- **Divergence kept**: the reference's empty-query-first-5, session-last-prompt source, `WorkspacePolicy`, trailing-separator dir form, and ≥150 KiB compaction are **not** adopted (recorded).

## §Forward (deferred, non-blocking)

- **RF-064-1** — **Full source parity**: adopting the reference's **empty-query-first-5** and its **active-session-last-prompt** source (the Q1 → B option). The session source would reintroduce a startup history read (superseding round-016 D3) and is deliberately not adopted here.
- **RF-064-2** — **Workspace-policy parity**: adopting the reference's injected `WorkspacePolicy` (hidden-dir + extension ignores) and its trailing-`os.PathSeparator` directory form; tellme keeps its hardcoded ignore list and plain names.
- **RF-064-3** — **The reference's ≈150 KiB log self-compaction** (tellme's `wg` drain hook stays a reserved no-op; a carried item, cf. ADR 0004).
- **RF-064-4** — **Pre-load vs per-query read**: adopting the reference's construction-time `LoadTopN(50)` + in-memory slice (a structure port, no behaviour gain; D3).
- **RF-064-5** — **Depth constant review**: if the reference (or the operator) moves `LoadTopN`'s 50, tellme's single `promptPoolDepth` follows; a periodic parity check is a candidate.
