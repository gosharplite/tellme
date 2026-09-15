# System Analysis Plan — round 026 (`026-tool-usage-accounting`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/026-tool-usage-accounting/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 2 journeys — /axb-spec-by-example
│       ├── keeping-a-record-of-tool-use.feature
│       └── reviewing-how-the-tools-have-been-used.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (tool-usage accounting row + flag) ✓ done
├── data/data-model.dbml           # /axb-data-plan — ADD (the global tool-usage record)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/…                      # ADD a `chat` (or dedicated) module for the report + the outcome classification
```

*(No `contracts/**` change (`/axb-api-plan` = `NOOP`) and no `ui/**` artifact this round — a plain-CLI
measurement slice: `/axb-ui-plan` is skipped.)*

### Source-code structure (repository root)

```text
internal/
├── domain/
│   └── history/
│       └── usage.go               # CHANGED — add the ToolUsageSink port + the ToolUsageRecord / outcome type (nil = no-op)
├── agent/
│   └── agentloop.go               # CHANGED — classify each executed call ok/error/timeout (err + per-call deadline)
│                                  #   and Record it through the injected sink (best-effort); loop stays fs-free
├── infrastructure/
│   └── history/
│       └── tool_usage_store.go    # NEW — the file-backed sink adapter over ~/.tellme/tools-count.jsonl
│                                  #   (os.UserHomeDir; O_APPEND|O_CREATE; append-only JSONL; Load + aggregate)
├── ui/
│   └── toolusage.go               # NEW — the pure report formatter (per-tool roll-up; registry order)
└── cli/
    └── cli.go                     # CHANGED — wire the sink into AgentLoop; add the --tool-usage offline path
                                   #   (dispatch precedence before prompt/stdin); inject the user-home seam
internal/domain/history/usage_test.go / (new) tool_usage tests   # NEW/CHANGED — record round-trip + aggregation + classification
go.mod / go.sum                    # unchanged — no new module (stdlib only)
Makefile                           # unchanged (no new gate)
```

**Structure Decision**: Round 026 is a **measurement** slice on the **CLI end**. It (a) classifies each
executed agent-tool invocation `ok`/`error`/`timeout` in the **agent loop** (the only place holding the
tool's error and the per-call deadline) and records it through an **injected `ToolUsageSink`** port, and
(b) adds an **offline reporting flag** that reads the global log and prints a per-tool roll-up to
`stdout`. The log is a **global, append-only JSONL** file at `~/.tellme/tools-count.jsonl` (a new
`internal/infrastructure/history` adapter), modelled by `/axb-data-plan`. There is **no** new endpoint,
**no** new dependency, and **no** change to the tool semantics/ set or the round-018 token store. The
**executable contract** is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`;
`specs/truth/contracts/**` and `ui/**` are untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** interfaces.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the prompt turn's **tool-usage accounting** (each executed call classified and appended to the global log) and the **offline `--tool-usage` report** that prints, per registered tool, its invocations and `ok`/`error`/`timeout` breakdown to `stdout`. The log write is silent (no `stdout`/`stderr` emission); the prompt path's `stdout` stays byte-exact; the report is strictly offline (no provider, no stdin) and deterministic (registry order).
   - Requirement evidence: `FR-001`–`FR-011`; acceptance features `keeping-a-record-of-tool-use.feature`, `reviewing-how-the-tools-have-been-used.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which adds the executable Gherkin feature(s) and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Global tool-usage log (local-state interface)`
   - Endpoint type: `Local-state / file endpoint`
   - Primary interface: the **append-only JSONL record** `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"ok|error|timeout"}` at `~/.tellme/tools-count.jsonl` — one record per executed invocation; global (shared across repos/envs/modes); never reset by `--new`; read whole and aggregated by the report.
   - Requirement evidence: `FR-003`, `FR-010`; the key entity *Tool-usage record* / *Tool-usage log* in `spec.md`.
   - Planner: **`/axb-data-plan`** — ADD the record type + its location/lifecycle to `specs/truth/data/data-model.dbml`.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; the report authors no request/response shape).
> - `/axb-data-plan` = **ADD** — the global tool-usage record is new persisted local state (Decision 2), so the data truth is extended.
> - `/axb-ui-plan` = **skipped** — no user-facing UX screen; the report is line-oriented terminal output.
> - **Truth amendment carried to `/axb-dsl-refine`**: **ADD** the CLI interface truth for the accounting + the `--tool-usage` report (a new `chat` sub-area or a dedicated `usage`-style module; the exact home is the contract owner's call, honouring `dsl-single-authority`), and confirm `acceptance-coverage` for both round-026 acceptance features. No new class phrase; the root vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Global tool-usage log (local-state interface)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the executable contract for (i) the outcome classification recorded per executed tool call (`ok`/`error`/`timeout`; error = non-nil tool error, timeout = nil-error result at the per-call deadline, ok otherwise), (ii) the offline `--tool-usage` report (per registered tool: invocations + `ok`/`error`/`timeout`; deterministic registry order; strictly offline), keeping `stdout` byte-exact on the prompt path and the class-phrase vocabulary unchanged.
  - **Global tool-usage log** → handoff to **`/axb-data-plan`**: ADD the record shape (`{timestamp, tool, outcome}`), the global append-only location (`~/.tellme/tools-count.jsonl`), the "never reset by `--new`" lifecycle, and the recorded divergence from the `TellMeHome` namespace.
  - **API** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the round has two tightly-coupled interfaces on one end (the write and the state it writes); there is no cross-wave dependency, so the whole analysis is one wave, with the CLI end carried forward to its contract owner and the state delegated to the data planner.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Global tool-usage log`) → **ADD** the record to `specs/truth/data/data-model.dbml`.
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → add the CLI interface truth for the accounting + the `--tool-usage` report, and confirm `acceptance-coverage` for the two round-026 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/026-tool-usage-accounting`; truth root `specs/truth`; truth-delta `specs/plans/026-tool-usage-accounting/truth-delta.md`; interfaces `CLI end`, `Global tool-usage log`; analysis focus as above; acceptance features `keeping-a-record-of-tool-use.feature`, `reviewing-how-the-tools-have-been-used.feature`.

---

### Gating blockers

*(none — the operator locked the outcome taxonomy (Q1 → 1), the storage shape/location (Q2 → Others: `~/.tellme/tools-count.jsonl`), and the surfacing channel (Q3 → 1: a dedicated offline flag); `research.md` Decisions 1–6 settled the classification signals, the counting seam, the report shape, and the hermetic verification split. **Open (non-blocking):** the exact flag name (`--tool-usage` provisional), the report's line wording, and the CLI module home (`/axb-dsl-refine`), and the record's DBML prose (`/axb-data-plan`). None gate this round.)*
