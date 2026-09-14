# System Analysis Plan — round 017 (`017-turn-chrome-parity`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/017-turn-chrome-parity/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 3 journeys — /axb-spec-by-example
│       ├── announcing-the-captured-input.feature
│       ├── framing-the-turn.feature
│       └── keeping-the-new-chrome-off-the-other-surfaces.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (the turn-chrome row)
├── data/data-model.dbml           # /axb-data-plan — NOOP (no state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # MODIFY — the non-TUI turn-chrome contract + DSL rows
```

*(No `contracts/**` change and no `ui/**` artifact this round — a plain-CLI round: `/axb-api-plan` and
`/axb-data-plan` are `NOOP`, and `/axb-ui-plan` is skipped (the requirement changes no UX surface).)*

### Source-code structure (repository root)

```text
internal/
├── ui/
│   ├── status.go                  # unchanged — the round-009 payload status formatter
│   └── turn.go                    # NEW — the round-017 turn chrome formatter: the input-capture
│                                  #   acknowledgement + the 80-column rule + the `╭─⠿ Turn <N> - <mode>`
│                                  #   header + the leading/trailing blank-line spacing (plain text; the
│                                  #   injected clock supplies the timestamps)
├── cli/
│   └── cli.go                     # CHANGED — emit the acknowledgement + frame on the positional/piped
│                                  #   prompt turn (A) and the round-012 plain reader (B); carry a
│                                  #   *chrome* switch so the `-i` TUI submit path (C) is excluded
├── agent/…                        # unchanged — the turn runner keeps its pre-flight/post-turn behaviour
└── domain/llm/…                   # unchanged — no request/response change
```

```text
go.mod / go.sum                     # unchanged — no new module (stdlib + the existing internal/ui)
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 017 is a **bounded presentation change** to the **non-TUI** prompt turn. It adds the reference's operator chrome — the input-capture acknowledgement, the 80-column `─` rule, the `╭─⠿ Turn <N> - <mode>` header, the (unchanged) pre-flight payload line, and the blank-line spacing — emitted on the **positional/piped prompt turn** and the **round-012 plain reader** only. There is **no** new endpoint, **no** new state, and **no** new dependency, consistent with `research.md` Decisions 1–7. The `-i` TUI surface (round 016) and every non-prompt path are unchanged. The **executable contract** is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`; the chrome formatter lives beside the existing status formatter in `internal/ui`; `specs/truth/contracts/**`, `specs/truth/data/**`, and `ui/**` are untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** interface. Round 017 introduces **no** new system end: it changes the presentation of the existing **CLI end**'s non-TUI prompt turn (nothing else).

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **non-TUI prompt turn** on the **positional/piped** path and the **round-012 plain reader** — acknowledge the captured input (`[HH:MM:SS] Input captured. Processing...`), open the turn with an 80-column `─` rule and a `╭─⠿ Turn <N> - <mode>` header (N = the session's persisted turn count + 1) over the existing pre-flight payload line, and separate the frame from the answer with a blank gap. Emitted on the **diagnostic stream (`stderr`)** in plain text; `stdout` stays byte-exact; the frozen class-phrase vocabulary is unchanged; the `-i` TUI surface and every non-prompt path are excluded.
   - Requirement evidence: `FR-001`–`FR-009`, `NFR-001`–`NFR-003`; acceptance features `announcing-the-captured-input.feature`, `framing-the-turn.feature`, `keeping-the-new-chrome-off-the-other-surfaces.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; round 017 adds no tellme-owned request/response shape).
> - `/axb-data-plan` = **`NOOP`** — round 017 persists no new state; `history.jsonl`/`history.archive.jsonl` and the shared prompt log are unchanged.
> - `/axb-ui-plan` = **skipped** — the round changes no user-facing UX surface (the `-i` TUI surface is unchanged; the chrome is line-oriented terminal output, not a TUI). No plan-side `ui/**` artifact is produced.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** `specs/truth/features/cli/chat/**` to carry the turn-chrome contract as executable Rules + DSL rows — the capture acknowledgement, the rule/header/payload frame, the blank-line spacing, the `Turn <N>` counting rule, and the negative boundary ("the `-i` prompt and the non-prompt commands show no chrome"). No new class phrase; the root `cli/dsl.md` vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own.
  - **Data** → **`NOOP`**: no state change.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the non-TUI turn-chrome contract as mechanically assertable interface Rules — the capture acknowledgement, the 80-column rule, the `Turn <N> - <mode>` header (N = persisted turns + 1), the pre-flight payload line inside the frame, the blank-line spacing, and the **negative boundary** (no chrome on the `-i` TUI surface or on any non-prompt path) — while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: a single interface with no dependency on any other interface's conclusions; there is nothing to sequence, so the round runs in one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the non-TUI turn-chrome behaviour in `specs/truth/features/cli/**` (add the acknowledgement + frame + spacing + `Turn <N>` rules; extend `chat/dsl.md`), and confirm `acceptance-coverage` for the three round-017 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/017-turn-chrome-parity`; truth root `specs/truth`; truth-delta `specs/plans/017-turn-chrome-parity/truth-delta.md`; interface `CLI end`; analysis focus as above; acceptance features `features/acceptance/**` (3).

---

### Gating blockers

*(none — the operator locked the two high-impact decisions in `/axb-clarify` round 1 (surface scope **A+B**; **input-capture line only**), and `research.md` Decisions 1–7 settled the chrome tokens, the turn-number source, the plain-text/colour stance, the spacing, the single-seam scope, the hermetic verification, and that no new dependency is required. **Open (non-blocking):** the exact DSL step vocabulary (`/axb-dsl-refine`) and the exact formatter/call-site layout (implementation). None gate this round.)*
