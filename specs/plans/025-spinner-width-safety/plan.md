# System Analysis Plan — round 025 (`025-spinner-width-safety`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/025-spinner-width-safety/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 1 journey — /axb-spec-by-example
│       └── keeping-the-spinner-width-safe.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (spinner row + seam + unit-tests row) ✓ done
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # MODIFY — the round-019 spinner contract + DSL rows
```

*(No `contracts/**` change and no `ui/**` artifact this round — a plain-CLI defect fix: `/axb-api-plan` is
`NOOP`, `/axb-data-plan` is `NOOP`, and `/axb-ui-plan` is skipped (the spinner is line-oriented terminal
output, not a TUI screen).)*

### Source-code structure (repository root)

```text
internal/
├── ui/
│   └── spinner.go                 # CHANGED — `ExecutingToolsLabel` gains the BOUNDED several-tool label
│                                  #   (` Executing tools [<first> and <N-1> more]...`); the presenter gains a
│                                  #   `columns func() int` seam and tracks the last frame's rendered-row count,
│                                  #   erasing EVERY occupied row on redraw and on clear (the single-row
│                                  #   `clearControl` becomes a rows-aware clear)
├── cli/cli.go                     # CHANGED — inject the `columns` seam (default: probe the `stderr` fd width via
│                                  #   `golang.org/x/term.GetSize`, 0 when unknown) and honour the
│                                  #   `TELL_ME_FORCE_STDERR_COLS` diagnostic seam
└── …                              # unchanged — the agent loop, the `-i` TUI, the other `internal/ui` formatters
```

```text
go.mod / go.sum                     # unchanged — no new module (`golang.org/x/term` is already a direct dependency)
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 025 is a **bounded defect fix** on the **CLI end**'s round-019 spinner. It
bounds the several-tool phase label (so a multi-tool turn no longer overruns the terminal width and clips
the ` [CPU: … | MEM: …]` segment) and makes the spinner clear width-safe (the presenter tracks the last
frame's rendered-row count and erases every occupied row, so a soft-wrapped frame leaves no residue — the
round-019 teardown contract holds on a narrow terminal). Both changes live in `internal/ui/spinner.go`; a
`columns` seam (injected for tests, defaulting to a `stderr` fd width probe) supplies the terminal width.
There is **no** new endpoint, **no** new persisted state, and **no** new dependency, consistent with
`research.md` Decisions 1–4. The **executable contract** is pinned in `specs/truth/features/cli/chat/**` by
`/axb-dsl-refine`; `specs/truth/contracts/**`, `specs/truth/data/**`, and `ui/**` are untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** interface. Round 025 introduces **no** new system end: it changes the
CLI end's prompt-turn spinner output only, and persists nothing.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **prompt-bearing turn** on the non-TUI surfaces — the **tool-phase spinner label** and the spinner **clear** on the **diagnostic stream (`stderr`)**. The tool-phase label is now **bounded** for several tools (` Executing tools [<first> and <N-1> more]...`; the single-tool ` Executing [<name>]...` and no-names ` Executing tools...` forms are unchanged), so the line no longer overruns the terminal and the trailing host CPU/memory segment stays readable; the clear now erases **every** terminal row the last frame occupied, so an over-wide (soft-wrapped) frame leaves no residue (`the progress spinner no longer appears once the answer is written`). `stdout` stays byte-exact; the class-phrase vocabulary is unchanged.
   - Requirement evidence: `FR-001`–`FR-007`, `NFR-001`–`NFR-003`; acceptance feature `keeping-the-spinner-width-safe.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which modifies the executable Gherkin feature and DSL rows under `specs/truth/features/cli/chat/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; the spinner authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** — the spinner is **ephemeral runtime presentation** (no persisted state); it reads only the active model name (config), the current tool-call names, and the terminal width. No entity / field / lifecycle / store.
> - `/axb-ui-plan` = **skipped** — the round changes no user-facing UX screen; the spinner is line-oriented terminal output, and the `-i` TUI (terminal-mode UI, round 015/016) already consumes the same spinner but adds no new surface. No plan-side `ui/**` artifact.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (the several-tool Example's Then becomes the bounded assertion; add a Rule/Example for the residue-free clear); **MODIFY** `chat/dsl.md` (the several-tool spinner row → the bounded form; add the bounded-label and row-aware-clear rows/notes); reuse the existing `the progress spinner no longer appears once the answer is written` row for the teardown witness. No new class phrase; the root vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: amend the turn-spinner contract so the **several-tool** label is bounded (` Executing tools [<first> and <N-1> more]...`) while the single-tool / no-names forms stay unchanged, and pin the **residue-free clear** (the teardown Then must hold when a frame is wider than the terminal); keep `stdout` byte-exact and the class-phrase vocabulary unchanged.
  - **API** → **`NOOP`**; **Data** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the round has a **single interface**; there is no second end to order against, so the whole analysis is one wave, carried forward to the CLI contract owner at delivery.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted state; the spinner is ephemeral).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → amend the spinner behaviour in `specs/truth/features/cli/chat/**` (bound the several-tool label; add the residue-free clear), and confirm `acceptance-coverage` for the round-025 acceptance feature.

Not delegated:
- `/axb-ui-plan` — **skipped** (no new UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/025-spinner-width-safety`; truth root `specs/truth`; truth-delta `specs/plans/025-spinner-width-safety/truth-delta.md`; interface `CLI end`; analysis focus as above; acceptance feature `features/acceptance/keeping-the-spinner-width-safe.feature`.

---

### Gating blockers

*(none — the operator locked the scope (Q1 → 1: fix both defects), the width-safe mechanism (Q2 → 1: track the last frame's rows and erase all of them), and the reference finding (no upstream fix to port); `research.md` Decisions 1–4 settled the bounded-label shape, the row-aware clear, the `columns` seam, and the verification split. **Open (non-blocking):** the exact DSL step vocabulary (`/axb-dsl-refine`) and the exact presenter/clear byte sequence + width probe (implementation). None gate this round.)*
