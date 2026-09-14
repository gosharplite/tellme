# System Analysis Plan — round 019 (`019-turn-spinner`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/019-turn-spinner/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 3 journeys — /axb-spec-by-example
│       ├── showing-a-progress-indicator.feature
│       ├── labelling-the-progress-indicator.feature
│       └── keeping-the-progress-indicator-bounded.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (spinner row + stderr gate + telemetry row) ✓ done
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/…                      # ADD/MODIFY — the turn-spinner contract + DSL rows
```

*(No `contracts/**` change and no `ui/**` artifact this round — a plain-CLI round: `/axb-api-plan` is
`NOOP`, `/axb-data-plan` is `NOOP`, and `/axb-ui-plan` is skipped (the round changes no UX surface — the
spinner is line-oriented terminal output, not a TUI; the `-i` TUI is explicitly out of scope).)*

### Source-code structure (repository root)

```text
internal/
├── ui/
│   ├── spinner.go                 # NEW — the round-019 spinner: the braille frame set, the phase-label
│   │                              #   + elapsed formatter (`{frame}{status} ({n}s)`), and a dependency-free
│   │                              #   ticker-driven presenter over the injected clock + stderr seams
│   ├── turn.go                    # unchanged — the round-017 turn chrome
│   ├── status.go                  # unchanged — the round-009 payload status formatter
│   ├── metrics.go                 # unchanged — the round-018 post-turn metrics/summary formatters
│   └── tui/prompt/…               # unchanged — the `-i` TUI (out of scope)
├── domain/metrics/…               # NEW — the SystemMetricsProvider port (machine-wide CPU/memory percentages)
├── infrastructure/telemetry/      # NEW — the POSIX adapters: system_metrics_linux.go,
│                                  #   system_metrics_darwin_cgo.go, system_metrics_darwin_nocgo.go
└── cli/cli.go                     # CHANGED — own the spinner lifecycle around each waiting phase; gate on the
                                   #   diagnostic stream (`isatty(stderr) && !-r`) + the `TELL_ME_FORCE_STDERR_TTY`
                                   #   seam (no stdout probe — round-006 / PR #16 Obs 1 stays OPEN)
```

```text
go.mod / go.sum                     # unchanged — no new module (stdlib + the existing internal packages)
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 019 is a **bounded presentation change** on the **CLI end**. It adds the
reference's **live progress spinner** to the non-TUI prompt surfaces so a run is never silently waiting —
brailled frames advancing on a ticker, a phase label (` Thinking [<model>]...` /
` Executing [<tool>]...` / ` Executing tools [<a>, <b>]...`), and, while tools run, the host CPU/memory
segment. It is written to `stderr` and gated on the **diagnostic stream** (`isatty(stderr) && !-r`) — mirroring the
reference's `IsTerminalContext()` (`ui.stderr`) — leaving `stdout` byte-exact; no standard-output probe is
wired, so round-006 / PR #16 **Obs 1** stays **OPEN**. There is **no** new endpoint, **no** new persisted state, and **no** new dependency,
consistent with `research.md` Decisions 1–9. The **executable contract** is pinned in
`specs/truth/features/cli/**` by `/axb-dsl-refine`; the spinner lives in a new `internal/ui/spinner.go`
beside the existing `internal/ui` formatters; `specs/truth/contracts/**`, `specs/truth/data/**`, and
`ui/**` are untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** interface. Round 019 introduces **no** new system end: it changes the
CLI end's prompt-turn output only, and persists nothing.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **prompt-bearing turn** (positional / piped / the round-012 reader) — while the run is waiting (awaiting the model, or executing tools) show a **live progress spinner** on the **diagnostic stream (`stderr`)**: a braille frame advancing on a ~200 ms ticker, a **phase label** (` Thinking [<model>]...` while awaiting the model; ` Executing [<tool>]...` / ` Executing tools [<a>, <b>]...` while tools run), and the whole-seconds elapsed counter `({n}s)`; the **tool-execution** state also shows the host CPU/memory segment (` [CPU: <c>% | MEM: <m>%]`). The spinner is drawn **only when the diagnostic stream (`stderr`) is a terminal and `-r` is off**, cleared before any interleaved write, absent on non-prompt paths and the `-i` TUI, and `stdout` stays byte-exact; the class-phrase vocabulary is unchanged.
   - Requirement evidence: `FR-001`–`FR-011`, `NFR-001`–`NFR-005`; acceptance features `showing-a-progress-indicator.feature`, `labelling-the-progress-indicator.feature`, `keeping-the-progress-indicator-bounded.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; the spinner authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** — the spinner is **ephemeral runtime presentation** (no persisted state); it reads only the active model name (config) and the current tool-call names. No entity / field / lifecycle / store.
> - `/axb-ui-plan` = **skipped** — the round changes no user-facing UX surface; the `-i` TUI (terminal-mode UI, round 015/016) is explicitly out of scope (`-i` precedes prompt capture). No plan-side `ui/**` artifact.
> - **Truth amendment carried to `/axb-dsl-refine`**: **ADD** `specs/truth/features/cli/chat/**` — the turn-spinner contract as executable Rules; **MODIFY** `chat/dsl.md` (spinner step rows); **MODIFY** the interface-root `cli/dsl.md` (**+1 cross-module Then row** for the cross-module negative, e.g. `the run shows no progress indicator`, used by `chat` + `diagnostics` + `history`); **MODIFY** the `diagnostics` / `history` / `presenting-the-turn` carrier features and **MODIFY** `chat/reporting-a-failed-provider-request.feature` (the failed-turn spinner-clear carrier). No new class phrase; the root vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the turn-spinner contract as mechanically assertable interface Rules — the spinner is present while the run waits (and animates), the phase label names the model / tools, the tool-execution state carries the resources segment, the spinner yields to interleaved output and clears before completion, it is drawn **only when the diagnostic stream (`stderr`) is a terminal and `-r` is off** (a `stderr` diagnostic, gated like the reference), and `stdout` stays byte-exact with the class-phrase vocabulary unchanged.
  - **API** → **`NOOP`**; **Data** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the round has a **single interface**; there is no second end to order against, so the whole analysis is one wave, carried forward to the CLI contract owner at delivery.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted state; the spinner is ephemeral).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the turn-spinner behaviour in `specs/truth/features/cli/**` (add the spinner Rules; extend `chat/dsl.md`; add the root cross-module negative row), and confirm `acceptance-coverage` for the three round-019 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/019-turn-spinner`; truth root `specs/truth`; truth-delta `specs/plans/019-turn-spinner/truth-delta.md`; interface `CLI end`; analysis focus as above; acceptance features `features/acceptance/**` (3).

---

### Gating blockers

*(none — the operator locked the spinner's visible form + labels (clarify Q1=2), the tool-execution metrics segment (Q2=2), the stop/resume behaviour (A), the `-i`-excluded / POSIX-only scope, and `research.md` Decisions 1–9 settled the frames/cadence, the label derivation, the elapsed semantics, the CPU/memory sampling, the gate + the forced stderr seam, the lifecycle placement, the verification, and the dependency footprint. **Open (non-blocking):** the exact DSL step vocabulary (`/axb-dsl-refine`) and the exact presenter/sampling layout (implementation). None gate this round.)*
