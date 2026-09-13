# System Analysis Plan — round 009 (`009-payload-status-line`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/009-payload-status-line/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── seeing-the-payload-status.feature
│       └── choosing-the-payload-budget.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (estimator, status line, budget, usage)
├── data/**                        # /axb-data-plan — NOOP this round (no persisted-state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **`NOOP`**: per-turn token counts are display-only and recomputed, so the session-history model is unchanged.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── agent/
│   └── agentloop.go               #   CHANGED — `Run` widened to surface the provider's reported usage
│                                  #   (returns an `AgentResult{Answer, Steps, Usage}` instead of dropping
│                                  #   `resp.Usage`), so the CLI can render the post-turn line (BLOCKER-2);
│                                  #   `BuildMessages` exported so the pre-flight estimate reuses the exact
│                                  #   conversation projection including tool steps (TD-1)
├── cli/
│   └── cli.go                     #   CHANGED — resolve the payload budget; emit the pre-flight status
│                                  #   line before the turn and the post-turn line after it (both to
│                                  #   stderr), reading `Usage` from the widened `AgentLoop.Run` (BLOCKER-2)
│                                  #   and rendering `<model>` from the provider's configured `MODEL`
│                                  #   (reference parity, TD-2); estimate the assembled conversation via
│                                  #   `agent.BuildMessages` (TD-1). Stays presentation/dispatch glue
├── config/
│   └── config.go                  #   CHANGED — `MAX_HISTORY_TOKENS` field + `EffectiveMaxHistoryTokens`
│                                  #   (env/config, default 1000000, `>= 0`; negative → configuration error),
│                                  #   mirroring `EffectiveMaxToolLoop`
├── domain/
│   └── llm/
│       ├── gateway.go             #   CHANGED — `Response` gains the provider's reported `Usage`
│       │                          #   (prompt/completion/total tokens)
│       └── token.go               #   NEW — `EstimateTokens(messages) int`: a deterministic, stdlib-only
│                                  #   heuristic over the assembled conversation (no dependency)
├── infrastructure/
│   └── llm/openai/client.go       #   CHANGED — parse the response's `usage` block into `Response.Usage`
└── ui/
    └── status.go                  #   NEW — the payload status-line formatter (pre-flight `~est/max`
                                   #   and post-turn `actual/max`; ingest mode/model + an injected clock
                                   #   seam; produces the reference-parity line; no `tellme:` prefix)

tests/e2e/                          # godog suite driving the built binary
├── harness/                        # CHANGED — the fake provider reports (or withholds) a `usage` block;
│                                   #   stderr is captured for the status-line assertions
└── steps/                          # NEW step files for the payload-status / budget scenarios
go.mod / go.sum                     # unchanged — stdlib-only (time, strings); no new dependency
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: The change stays **inside the existing CLI surface** and follows the round-004/007/008 seam discipline: it **widens two existing seams** — the provider port (`llm.Response` gains the reported `usage`) **and** the agent-loop seam (`AgentLoop.Run` surfaces that `usage` via an `AgentResult{Answer, Steps, Usage}` — BLOCKER-2 — and `BuildMessages` is exported so the CLI's pre-flight estimate reuses one conversation projection including tool steps — TD-1) — and adds **two pure pieces**: a deterministic estimator (`internal/domain/llm/token.go`) and a status-line formatter (`internal/ui/status.go`). `MAX_HISTORY_TOKENS` resolution joins `internal/config` beside the existing effective-value helpers. The status line is **emitted by the CLI** (`internal/cli`) to the diagnostic stream; it introduces **no** new domain port, **no** adapter, **no** persisted-state change, and **no** new dependency. `/axb-dsl-refine` owns the CLI truth. The behaviour is already settled by `research.md` Decisions 1–9, so this round adds no new architecture dimension.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface. Round 009 does not introduce a new system end; it extends the **CLI end**'s behaviour with a per-turn payload status line.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: **standard error** (the diagnostic stream) carries the per-turn payload status for the operator: a **pre-flight** line `[HH:MM:SS] Payload: ~<est>/<budget> tokens - <mode> - <model>` before the provider request, and a **post-turn** line `[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model>` after it (the actual read from the provider's reported usage; omitted when the provider reports none). The line is **always-on** — never gated by terminal-ness and never suppressed by `-r/--raw` — and carries **no** `tellme:` prefix. **Standard output** (the answer stream) is unchanged, and **standard input** piping is unchanged. The payload budget is `MAX_HISTORY_TOKENS` (env/config, default 1000000). Non-prompt paths (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`) emit no status line. The frozen `tellme: {phrase}` class-phrase vocabulary and the exit-code table are **unchanged** (10 phrases; `0/2/3/4/5/6/7`).
   - Requirement evidence: `FR-001`–`FR-014`, `NFR-001`–`NFR-005`; acceptance features `seeing-the-payload-status.feature`, `choosing-the-payload-budget.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own). The **outbound** provider request is unchanged; round 009 only **parses** the provider response's existing `usage` block — the provider remains an external dependency reached outbound, not a contract tellme authors.
> - `/axb-data-plan` = **`NOOP`** — the per-turn token counts are **display-only** and recomputed from history text on resume; the persisted session model (`history_entry`) is **not** widened (research Decision 7).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **token estimator** and the **status-line formatter** are **not** separate system interfaces: they have no schema/contract artifact and no persisted state to model — their operator-facing behaviour (the status line, the budget, the estimate-vs-actual distinction) belongs to the **CLI end** and is carried to `/axb-dsl-refine` with it.
> - The **provider** remains an **external dependency reached outbound by the CLI**; round 009 adds no new outbound request shape (only response parsing), so it introduces no new tellme-authored contract.
> - **Truth amendments carried to `/axb-dsl-refine`**: **ADD** the payload-status behaviour to `specs/truth/features/cli/**` (a new `status` module — or an extension of `chat/`); the root `cli/dsl.md` class-phrase vocabulary stays **10** (the status line carries no `tellme:` prefix); and **MODIFY** any existing `chat/*` feature that must now also account for the status line appearing on `stderr` during a prompt turn (the `stdout` answer contract is unchanged). A `-l`/`-d`/`--version`/boot run stays byte-identical to rounds 001–008.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound provider request is unchanged (only response `usage` parsing is added).
  - **Data** → **`NOOP`**: token counts are display-only and recomputed; the persisted session model is unchanged.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: add the payload-status behaviour to `specs/truth/features/cli/**` — the pre-flight estimate line, the post-turn actual line, the always-on/stderr posture, the no-`tellme:`-prefix rule, the non-prompt-path exclusion, and the `MAX_HISTORY_TOKENS` budget (default 1000000; overridable; negative rejected) — and **MODIFY** any existing `chat/*` feature that must now account for the status line on `stderr` for a prompt turn.
- Scheduling rationale: there is **one** interface and the behaviour is **already fixed** by `research.md` Decisions 1–9, so there is nothing to sequence; a single wave suffices. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a sole interface forms one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the payload-status behaviour (`seeing-the-payload-status`, `choosing-the-payload-budget`) plus any `chat/*` **MODIFY** for the status line on `stderr`, in `specs/truth/features/cli/**`.

*Handoff payload (for the next phase)*: plan package `specs/plans/009-payload-status-line`; truth root `specs/truth`; truth-delta `specs/plans/009-payload-status-line/truth-delta.md`; interfaces `CLI end`; analysis focus as above; acceptance features `features/acceptance/seeing-the-payload-status.feature` + `features/acceptance/choosing-the-payload-budget.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the stream, the estimate-vs-actual scope, and the always-on posture; `research.md` Decisions 1–9 settled the estimator, the provider-`usage` widening, the line format + clock seam, the budget resolution, what the estimate counts, no persistence, and the testing strategy. **Open (non-blocking):** the exact estimator constant, the `MAX_HISTORY_TOKENS: 0` semantics (mirrors the tool-loop resolver), and the timestamp rendering — all held as research-level defaults, none gating this round.)*
