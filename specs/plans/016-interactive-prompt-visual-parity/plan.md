# System Analysis Plan — round 016 (`016-interactive-prompt-visual-parity`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/016-interactive-prompt-visual-parity/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 3 journeys — /axb-spec-by-example
│       ├── seeing-a-prompt-that-matches-the-reference.feature
│       ├── composing-with-settled-suggestions.feature
│       └── fitting-the-terminal-window.feature
├── ui/                            # PM plan-side TUI UX artifact — /axb-ui-plan (terminal mode)
│   ├── ui-plan.md
│   └── screens/{entry,10-suggestion,20-composed,30-narrow}.txt
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY this round (strict-parity chrome + debounce)
├── data/data-model.dbml           # /axb-data-plan — NOOP (no state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # ADD/MODIFY/DELETE — the strict-parity prompt contract + DSL rows
```

*(No `contracts/**` and no `data/**` change in this round — see the `NOOP` notes below.)*

### Source-code structure (repository root)

```text
internal/
├── ui/tui/prompt/
│   ├── model.go                   # CHANGED — reproduce the reference chrome (root padding; no dashboard header);
│   │                              #   handle WindowSizeMsg (width = msg.Width - 4); debounce + cancel the
│   │                              #   suggestion refresh; Tab/Shift+Tab insert the selection (last-token heuristic)
│   ├── textarea.go                # CHANGED — bordered editor (lipgloss.NormalBorder, fg 240), fixed height 10,
│   │                              #   the reference placeholder, ShowLineNumbers=false
│   ├── suggester.go               # CHANGED — the `Suggestions:` header + styled rows (selected bold fg 205 / bg 235,
│   │                              #   unselected fg 245) + the `>` cursor
│   └── run.go                     # unchanged — the stream-injected Bubble Tea run (stdout stays byte-exact)
├── cli/
│   └── cli.go                     # CHANGED (minor) — drop the dashboard wiring from the prompt launch; the `-i`
│                                  #   gating (terminal + flag/config) and the non-TTY/plain-reader paths are unchanged
└── app/suggestions/               # unchanged — the suggestion engine sources (log + session + FS + tools)
```

```text
go.mod / go.sum                     # unchanged — the Bubble Tea family is already present (no new module)
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 016 is a **bounded presentation/interaction change** to the existing `-i` prompt package. It reproduces the reference chrome on the **already-adopted** Bubble Tea family (no new module), adds the debounced/cancelable suggestion refresh and the `Tab`-inserts-selection behaviour, and **removes** the round-015 dashboard header from the prompt surface. There is **no** new endpoint, **no** new state, and **no** new dependency — consistent with `research.md` Decisions 1–8 and the operator's strict-parity decision (#39). The **executable contract** (the framed editor, the styled suggestion list, the absence of a metrics header, the debounced refresh, and `Tab`-inserts) is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`; the **TUI UX surface** was produced by the PM in `ui/**` (terminal mode) and is **reviewed** here, not re-planned; `specs/truth/data/**` and `specs/truth/contracts/**` are `NOOP`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** interfaces. Round 016 introduces **no** new system end: it changes the presentation/interaction of the existing **CLI end**'s `-i` prompt surface, and it exposes a plan-side **terminal UX surface**.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **`-i` interactive TUI prompt** — reproduce the reference chrome (a **bordered** multi-line editor above a **styled suggestion list**; the keybinding hints in the placeholder; **no** dashboard header, status line, or `?` overlay), debounce/cancel the suggestion refresh, drop over-long suggestions, let `Tab`/`Shift+Tab` **insert** the selection, and reflow on resize. A submit runs exactly one turn; an abort sends nothing; the non-`-i` / non-TTY / plain-reader paths are unchanged; `stdout` stays byte-exact and the frozen class-phrase vocabulary is unchanged.
   - Requirement evidence: `FR-001`–`FR-011`, `NFR-001`–`NFR-004`; acceptance features `seeing-a-prompt-that-matches-the-reference.feature`, `composing-with-settled-suggestions.feature`, `fitting-the-terminal-window.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates/creates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Interactive TUI UX surface (terminal)`
   - Endpoint type: `Terminal UI endpoint (frontend-like)`
   - Primary interface: the plan-side `ui/ui-plan.md` + `ui/screens/*.txt` frames — `entry` (framed editor + suggestion list), `10-suggestion` (path completion), `20-composed` (multi-line), `30-narrow` (degradation); the keybinding table and state-transition list make it traceable key → state → outcome.
   - Requirement evidence: `FR-001`–`FR-003`, `FR-009`; `ui/ui-plan.md`; `NFR-001`.
   - Handoff: **PM-produced** via `/axb-ui-plan` (**terminal mode**). `axb-system-analysis` **reviews** the artifact for feasibility under the current technical boundary and **does not redo/re-plan/re-delegate** it (mirroring the frontend handling — and the round-015 precedent).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; round 016 adds no tellme-owned request/response shape).
> - `/axb-data-plan` = **`NOOP`** — round 016 persists no new state; the shared global prompt log (`prompt_log_entry`, round 015) is unchanged.
> - `/axb-ui-plan` = **produced this round (terminal mode)** — the PM artifact `ui/ui-plan.md` + `ui/screens/*.txt`; **reviewed** here, **not** delegated/re-planned.
> - **Truth amendment carried to `/axb-dsl-refine`**: **DELETE** the rule "the interactive prompt reports the session dashboard" from `chat/using-the-interactive-prompt.feature` (dashboard retired); **ADD** a `chat` feature pinning the parity chrome + interaction (framed editor; suggestions beneath; no metrics header; over-long drop; `Tab` inserts the selection); **MODIFY** `chat/dsl.md` with the strict-parity rows (and remove any dashboard-header row). No new class phrase; the root `cli/dsl.md` vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Interactive TUI UX surface (terminal)` — review-only (PM artifact)
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own.
  - **Data** → **`NOOP`**: no state change.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the strict-parity prompt contract as mechanically assertable interface Rules — the framed (bordered) editor, the styled suggestion list beneath it, the **absence** of a metrics header, the debounced refresh, the over-long drop, and the `Tab`-inserts-selection behaviour — while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
  - **TUI UX surface** → **review** the PM `ui/**` artifact for feasibility under the current technical boundary; report gaps only.
- Scheduling rationale: the two interfaces do not depend on one another's *conclusions* — the CLI behaviour is fixed by `research.md` Decisions 1–3/5 and the UX artifact is already produced (review-only). Per `Wave依賴排序與平行分組判準.md` Rules 1–2, the review-only UI work and the low-coupling CLI-contract work run in the same wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the strict-parity prompt behaviour in `specs/truth/features/cli/**` (retire the dashboard rule; add the chrome + insertion rules; update `chat/dsl.md`), and confirm `acceptance-coverage` for the three round-016 acceptance features.

Not delegated:
- `/axb-ui-plan` — the PM already produced `ui/**` (terminal mode); **reviewed only**.
- `/axb-api-plan` / `/axb-data-plan` are invoked only to record their `NOOP`.

*Handoff payload (for the next phase)*: plan package `specs/plans/016-interactive-prompt-visual-parity`; truth root `specs/truth`; truth-delta `specs/plans/016-interactive-prompt-visual-parity/truth-delta.md`; interfaces `CLI end` + `Interactive TUI UX surface`; analysis focus as above; acceptance features `features/acceptance/**` (3); UI artifact `ui/**`.

---

### Gating blockers

*(none — the operator locked **strict parity** in issue [#39](https://github.com/gosharplite/tellme/issues/39) (the dashboard header is retired from the `-i` surface), and `research.md` Decisions 1–8 settled the chrome tokens, the debounced/cancelable refresh, the `Tab`-inserts behaviour, the resize handling, the hermetic no-pty verification, and that no new dependency is required. **Open (non-blocking):** the exact DSL step vocabulary (`/axb-dsl-refine`) and the exact prompt package layout (implementation). None gate this round.)*
