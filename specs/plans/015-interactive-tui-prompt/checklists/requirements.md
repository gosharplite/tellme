# Spec Quality Checklist: tellme Interactive TUI Prompt (round 015)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/015-interactive-tui-prompt`

**Spec Path**: `specs/plans/015-interactive-tui-prompt/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (the `-i` TUI prompt: suggestions + dashboard + multi-line editor + keybindings; the shared prompt log; the coexist/non-TTY guarantees)
- [x] No implementation technology, framework, or code detail is written as a requirement (the TUI library, the store mechanism, and the compaction policy are named as context/assumptions, not as FR text; behaviour is stated observably)
- [x] Edge cases cover the main high-risk scenarios (positional prompt, non-TTY with `-i`, empty/abort, missing log, concurrent append / no `flock`, noisy dirs, large log)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (core prompt P1 · shared-log interop P1 → dashboard P2 · coexist/no-regression P2)
- [x] Each user story is independently verifiable (injected terminal seam + scripted keystrokes + injected suggestion source; no real pty)
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries (no-regression / frozen vocabulary; determinism)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify` (global-log record semantics; round-012 relationship; platform scope)
- [x] This round's clarify questions were held to the budget (Round 1 = 3 questions, within the 1–3 cap)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A7) rather than asked
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain — all three resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths (suggestions seed/refresh/insert/submit; shared-log read+write only under `-i`; dashboard; coexist/fallback)
- [x] Success criteria are measurable, verifiable, and technology-neutral (one turn per submission; identical-shape append; round-trip; no TUI on non-TTY; prior rounds green)
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-14)

- **Q1 → Option 1 (record semantics)** — the shared `output/global_prompts.jsonl` is **read + written**, recorded **only under `-i`** (`FR-008`–`FR-011`, `A1`).
- **Q2 → Option 1 (relationship)** — the TUI **coexists** with the round-012 plain reader; `-i`/`USE_TUI_PROMPT` opts in (`FR-014`, `A2`).
- **Q3 → Option 1 (platform)** — **POSIX-only** this round; no Windows variant (`A3`).

### Folded assumptions (accepted, not re-asked)

- **A4** — tool suggestions come from tellme's registered tool registry (projection is a research/DSL detail).
- **A5** — the shared-log read is newest-first + deduped; compaction policy is a `/axb-data-plan` determination; the no-`flock` carried item remains out of scope.
- **A6** — the round introduces tellme's first TUI-grade dependency (a research + Setup decision); the fallback keeps the no-TUI path.
- **A7** — rounds 001–014 semantics unchanged except the new prompt surface.

### Still open (non-blocking — later-phase determinations)

- **Exact record/DBML shape + compaction policy** for the shared log — a `/axb-data-plan` determination (the `output/global_prompts.jsonl` schema, ordering, dedupe, compaction).
- **TUI dependency + engine seam** — which TUI library and how suggestion/metrics sources are injected — a `/axb-technical-research` + Setup determination.
- **DSL step vocabulary** — Given/When/Then rows for "launch the interactive prompt", "suggestions appear", "submit/abort", "the shared log gains a line" — a `/axb-dsl-refine` determination.
- **`ui/**` terminal-mode UX artifact** — the plan-side `ui/ui-plan.md` + `ui/screens/*.txt` — a `/axb-ui-plan` (terminal mode) output; reviewed (not re-planned) by `/axb-system-analysis`.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Four stories stand (rich TUI prompt P1 · shared prompt log P1 · dashboard P2 · coexist/no-regression P2). All three clarify items are resolved; only data-plan / research / DSL-layer / ui-plan determinations remain. `/axb-spec-by-example`, `/axb-ui-plan` (terminal mode — un-gated), and `/axb-technical-research` can proceed in parallel.
