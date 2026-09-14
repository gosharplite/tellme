# System Analysis Plan — round 015 (`015-interactive-tui-prompt`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/015-interactive-tui-prompt/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 4 journeys — /axb-spec-by-example
│       ├── composing-a-prompt-with-live-suggestions.feature
│       ├── sharing-the-prompt-log.feature
│       ├── seeing-the-session-dashboard.feature
│       └── choosing-between-the-interactive-prompt-and-plain-input.feature
├── ui/                            # PM plan-side TUI UX artifact — /axb-ui-plan (terminal mode)
│   ├── ui-plan.md
│   └── screens/{entry,10-suggestion,20-dashboard,30-error,40-help}.txt
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (TUI + suggestions + shared log)
├── data/data-model.dbml           # /axb-data-plan — ADD this round (the shared global prompt log)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/{chat,history}/…       # ADD/MODIFY — the interactive-prompt contract + DSL rows
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **in scope**: the round adds the shared global prompt log, which the `data-model-covers-all-state` invariant reaches.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── cli/
│   └── cli.go                     # CHANGED — parse `-i`/`--interactive`; engage the interactive TUI
│                                  #   prompt ONLY when enabled AND stdin is a terminal; otherwise the
│                                  #   round-012 plain reader / piped / non-interactive paths are unchanged
├── domain/
│   ├── ports/                     # NEW — the suggestion-source and prompt-tracker ports
│   │   └── suggestions.go
│   └── config/
│       └── config.go              # CHANGED — add the `USE_TUI_PROMPT` field (+ resolver)
├── ui/tui/prompt/                 # NEW — the Bubble Tea interactive prompt
│   ├── model.go                   #   the model: editor + suggestion list + dashboard + key handling
│   ├── suggester.go               #   the suggestion list (selection cursor)
│   └── textarea.go                #   the multi-line editor wrapper
├── app/suggestions/
│   └── service.go                 # NEW — the multi-source suggestion engine (log + session + FS + tools)
└── infrastructure/history/
    └── global_prompt_tracker.go   # NEW — the shared append-only prompt log store
                                   #   ($TELL_ME_HOME/output/global_prompts.jsonl)

tests/e2e/
├── harness/                       # EXTENDED — drive the built binary through the FORCE_STDIN_TTY seam
│                                  #   + scripted stdin for the interactive scenarios
└── steps/                          # NEW step files — the interactive-prompt journeys

specs/truth/data/data-model.dbml    # ADD — the shared global prompt log table + lifecycle
specs/truth/features/cli/**         # ADD/MODIFY — the interactive-prompt contract + DSL rows

go.mod / go.sum                     # CHANGED — bubbletea + bubbles (new modules); lipgloss promoted to direct
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 015 adds an **opt-in interactive surface** to the existing single CLI end and a **new shared data artifact**. The CLI dispatch (`internal/cli/cli.go`) gains the `-i`/`USE_TUI_PROMPT` branch that engages the TUI only when stdin is a terminal (reusing the round-012 real-isatty seam); the non-TTY / round-012-plain / piped paths are untouched. The interactive prompt lives in a new `internal/ui/tui/prompt` package (Bubble Tea model + suggester + editor, with injected I/O + suggestion source so it is testable without a pty); the suggestion engine is a new `internal/app/suggestions` service over a new `internal/domain/ports` seam; the shared prompt log is a new `internal/infrastructure/history/global_prompt_tracker.go` store. There is **no** new domain provider, transport, or HTTP surface — consistent with `research.md` Decisions 1–9 and Clarify Round 1 (`1,1,1`). The **executable contract** (the interactive prompt, its suggestions, the dashboard, the shared-log record, and the opt-in/non-TTY guarantees) is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`; the **shared-log model** is owned by `/axb-data-plan` in `specs/truth/data/data-model.dbml`; the **TUI UX surface** was produced by the PM in `ui/**` (terminal mode) and is **reviewed** here, not re-planned.

---

## Analysis Plan

### System interface inventory

This requirement inventories **3** system interfaces. Round 015 does not introduce a new system end: it adds an opt-in interactive surface to the **CLI end**, exposes a plan-side **terminal UX surface**, and adds a **shared local-persistence record** that reaches the `/axb-data-plan` trigger.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **`-i` interactive TUI prompt** — a live suggestion list (recent prompts + workspace paths + registered tools), a session dashboard (provider/model · tokens · turns), a multi-line editor, and terminal keybindings; a submit runs exactly one turn, an abort sends nothing; the shared prompt log is written **only under `-i`**. Non-`-i` runs, piped/non-terminal input, and a positional prompt are unchanged; `stdout` stays byte-exact and the frozen class-phrase vocabulary is unchanged.
   - Requirement evidence: `FR-001`–`FR-011`, `FR-014`, `FR-015`, `NFR-001`–`NFR-005`; acceptance features `composing-a-prompt-with-live-suggestions.feature`, `sharing-the-prompt-log.feature`, `seeing-the-session-dashboard.feature`, `choosing-between-the-interactive-prompt-and-plain-input.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates/creates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Interactive TUI UX surface (terminal)`
   - Endpoint type: `Terminal UI endpoint (frontend-like)`
   - Primary interface: the plan-side `ui/ui-plan.md` + `ui/screens/*.txt` frames — the launch/edit frame, the path-suggestion frame, the post-turn dashboard frame, the inline error frame, and the keybinding overlay; the keybinding table and state-transition list make it traceable key → state → outcome.
   - Requirement evidence: `FR-002`–`FR-006`, `FR-012`, `FR-013`; `ui/ui-plan.md`; `NFR-001`.
   - Handoff: **PM-produced** via `/axb-ui-plan` (**terminal mode**). `axb-system-analysis` **reviews** the artifact for feasibility under the current technical boundary and **does not redo/re-plan/re-delegate** it (mirroring the frontend handling).

3. `Shared global prompt log (local persisted state)`
   - Endpoint type: `Filesystem / local-persistence endpoint`
   - Primary interface: the shared, append-only prompt log at `$TELL_ME_HOME/output/global_prompts.jsonl` — one `{"timestamp":"<RFC3339>","prompt":"<text>"}` per line, shared across modes/personas; read **newest-first + deduplicated** to seed suggestions; appended **only under `-i`**; byte-identical shape to `tell-me-go` so lines round-trip; a bounded compaction policy keeps the file from growing without bound.
   - Requirement evidence: `FR-008`–`FR-011`, `NFR-003`, `NFR-004`; acceptance feature `sharing-the-prompt-log.feature`.
   - Planner: **`/axb-data-plan`** — the persisted record model, ordering/dedupe, lifecycle, and compaction are a data responsibility. The owner decides the table/columns in `specs/truth/data/data-model.dbml`.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; round 015 adds no tellme-owned API request/response shape — the TUI + shared log are a local terminal + local-file concern).
> - `/axb-data-plan` = **ADD** — `specs/truth/data/data-model.dbml`: a new shared **global prompt log** model (record shape `{timestamp, prompt}`, append-only lifecycle, newest-first + dedupe read, compaction), distinct from the per-session `history.jsonl`.
> - `/axb-ui-plan` = **produced this round (terminal mode)** — the PM artifact `ui/ui-plan.md` + `ui/screens/*.txt`; **reviewed** here, **not** delegated/re-planned.
> - **Truth amendment carried to `/axb-dsl-refine`**: **ADD** an executable `cli` feature for the interactive prompt (suggestions / dashboard / submit-abort / shared-log record / opt-in + non-TTY fallback) and **MODIFY** the `chat` and/or a new module's `dsl.md` with the interactive-prompt Given/When/Then rows; confirm **every round-015 acceptance rule is carried** (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10** (no new class phrase).

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Interactive TUI UX surface (terminal)` — review-only (PM artifact)
  - `Shared global prompt log (local persisted state)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own.
  - **Data** → handoff to **`/axb-data-plan`**: model the shared global prompt log — the record shape (`{timestamp, prompt}`, RFC3339), the append-only write (`O_APPEND|O_CREATE`, only under `-i`), the newest-first + deduped read, and the compaction lifecycle.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the interactive-prompt contract as mechanically assertable interface Rules — suggestions seed/refresh/accept, the dashboard, submit-vs-abort, the shared-log record under `-i` only, and the opt-in/non-TTY fallback — while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
  - **TUI UX surface** → **review** the PM `ui/**` artifact for feasibility under the current technical boundary; report gaps only.
- Scheduling rationale: the three interfaces do not depend on one another's *conclusions* — the record shape is fixed by `research.md` Decision 3, the CLI behaviour by Decisions 2/4/5, and the UX artifact is already produced (review-only). Per `Wave依賴排序與平行分組判準.md` Rules 1–2, the UI review and the low-coupling local-persistence work run in the same wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Shared global prompt log`) → **ADD** the shared prompt-log model to `specs/truth/data/data-model.dbml`.
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the interactive-prompt behaviour in `specs/truth/features/cli/**` (a new/updated feature + DSL rows), and confirm `acceptance-coverage` for all four round-015 acceptance features.

Not delegated:
- `/axb-ui-plan` — the PM already produced `ui/**` (terminal mode); **reviewed only**.
- `/axb-api-plan` is invoked only to record its `NOOP`.

*Handoff payload (for the next phase)*: plan package `specs/plans/015-interactive-tui-prompt`; truth root `specs/truth`; truth-delta `specs/plans/015-interactive-tui-prompt/truth-delta.md`; interfaces `CLI end` + `Interactive TUI UX surface` + `Shared global prompt log`; analysis focus as above; acceptance features `features/acceptance/**` (4); UI artifact `ui/**`.

---

### Gating blockers

*(none — Clarify Round 1 settled `1,1,1` (shared-log write only under `-i`; TUI coexists with the round-012 reader; POSIX-only), the upstream `/axb-ui-plan` terminal-mode gate is **RESOLVED** (aixbdd-tmg#13 → PR #14), and `research.md` Decisions 1–9 settled the TUI family, the suggestion engine, the shared-log store, the opt-in gating, the dashboard source, the hermetic no-pty verification, and the dependency footprint. **Open (non-blocking):** the exact DBML table/columns and compaction wording (`/axb-data-plan`), the DSL step vocabulary (`/axb-dsl-refine`), and the exact TUI package layout (implementation). None gate this round.)*
